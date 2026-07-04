package service

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

// HealthChecker performs periodic health checks on all registered clusters.
type HealthChecker struct {
	interval time.Duration
	stopCh   chan struct{}
}

// NewHealthChecker creates a HealthChecker that runs checks at the given interval.
func NewHealthChecker(interval time.Duration) *HealthChecker {
	return &HealthChecker{
		interval: interval,
		stopCh:   make(chan struct{}),
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

// Stop signals the background goroutine to exit.
func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}

// checkAll queries all clusters from the database and checks each one concurrently.
func (hc *HealthChecker) checkAll() {
	var clusters []model.K8SCluster
	if err := database.DB.Find(&clusters).Error; err != nil {
		logrus.Errorf("HealthChecker: failed to query clusters: %v", err)
		return
	}

	var wg sync.WaitGroup
	for _, cluster := range clusters {
		wg.Add(1)
		go func(c model.K8SCluster) {
			defer wg.Done()
			hc.checkOne(c)
		}(cluster)
	}
	wg.Wait()
}

// checkOne decrypts the kubeconfig, tests connectivity, and updates the cluster status.
func (hc *HealthChecker) checkOne(cluster model.K8SCluster) {
	now := time.Now()

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
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
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
