package cluster

import (
	"context"
	"sync"
	"time"

	"gkube/internal/cluster/model"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/k8s"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	healthCheckTimeout  = 10 * time.Second
	healthCheckParallel = 5
)

// HealthChecker performs periodic health checks on all registered clusters.
type HealthChecker struct {
	interval time.Duration
	stopCh   chan struct{}
	stopOnce sync.Once
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewHealthChecker creates a HealthChecker that runs checks at the given interval.
func NewHealthChecker(interval time.Duration) *HealthChecker {
	ctx, cancel := context.WithCancel(context.Background())
	return &HealthChecker{
		interval: interval,
		stopCh:   make(chan struct{}),
		ctx:      ctx,
		cancel:   cancel,
	}
}

// Start begins the background health-check loop.
func (hc *HealthChecker) Start() {
	go func() {
		ticker := time.NewTicker(hc.interval)
		defer ticker.Stop()

		// Run once immediately on start.
		hc.checkAll()

		for {
			select {
			case <-ticker.C:
				hc.checkAll()
			case <-hc.stopCh:
				logrus.Info("HealthChecker stopped")
				return
			}
		}
	}()
	logrus.Infof("HealthChecker started with interval %s", hc.interval)
}

// Stop signals the background goroutine to exit. Safe to call multiple times.
func (hc *HealthChecker) Stop() {
	hc.stopOnce.Do(func() {
		close(hc.stopCh)
		hc.cancel()
	})
}

// checkAll queries all clusters from the database and checks each one concurrently,
// bounded by a semaphore. Stop 后及时退出。
func (hc *HealthChecker) checkAll() {
	// 已停止则不再执行
	select {
	case <-hc.ctx.Done():
		return
	default:
	}

	var clusters []model.K8SCluster
	if err := database.DB.Find(&clusters).Error; err != nil {
		logrus.Errorf("HealthChecker: failed to query clusters: %v", err)
		return
	}

	sem := make(chan struct{}, healthCheckParallel)
	var wg sync.WaitGroup
loop:
	for _, cluster := range clusters {
		// 已停止则不再调度新检查
		select {
		case <-hc.ctx.Done():
			break loop
		default:
		}

		cluster := cluster
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case sem <- struct{}{}:
			case <-hc.ctx.Done():
				return
			}
			defer func() { <-sem }()
			hc.checkOne(cluster)
		}()
	}
	wg.Wait()
}

// checkOne decrypts the kubeconfig, tests connectivity, and updates the cluster status.
func (hc *HealthChecker) checkOne(cluster model.K8SCluster) {
	now := time.Now()

	// 单集群检查带超时,并继承 Stop ctx
	ctx, cancel := context.WithTimeout(hc.ctx, healthCheckTimeout)
	defer cancel()

	// Decrypt kubeconfig.
	kubeconfig, err := auth.DecryptAES(cluster.KubeConfig)
	if err != nil {
		logrus.Errorf("HealthChecker: cluster %s decrypt kubeconfig failed: %v", cluster.ClusterName, err)
		hc.updateStatus(cluster.ID, "offline", "", 0, now)
		return
	}

	// Create k8s client.
	clientset, err := k8s.GetK8sClient(kubeconfig)
	if err != nil {
		logrus.Errorf("HealthChecker: cluster %s get k8s client failed: %v", cluster.ClusterName, err)
		hc.updateStatus(cluster.ID, "offline", "", 0, now)
		return
	}

	// Get server version.
	version, err := clientset.Discovery().ServerVersion()
	if err != nil {
		logrus.Errorf("HealthChecker: cluster %s get server version failed: %v", cluster.ClusterName, err)
		hc.updateStatus(cluster.ID, "offline", "", 0, now)
		return
	}

	// Get node count.
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		logrus.Errorf("HealthChecker: cluster %s list nodes failed: %v", cluster.ClusterName, err)
		hc.updateStatus(cluster.ID, "offline", "", 0, now)
		return
	}

	logrus.Infof("HealthChecker: cluster %s is online (version=%s, nodes=%d)", cluster.ClusterName, version.GitVersion, len(nodes.Items))
	hc.updateStatus(cluster.ID, "online", version.GitVersion, len(nodes.Items), now)
}

// updateStatus writes the health-check result to the database.
func (hc *HealthChecker) updateStatus(id uint, status, clusterVersion string, nodeCount int, lastCheck time.Time) {
	updates := map[string]interface{}{
		"status":            status,
		"cluster_version":   clusterVersion,
		"node_count":        nodeCount,
		"last_health_check": lastCheck,
	}
	if err := database.DB.Model(&model.K8SCluster{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		logrus.Errorf("HealthChecker: failed to update cluster %d status: %v", id, err)
	}
}
