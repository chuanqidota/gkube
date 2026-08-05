package k8s

import (
	"fmt"
	"sync"
	"time"

	clustermodel "gkube/internal/cluster/model"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/logger"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// cachedClient wraps a generic client with an expiration time.
type cachedClient[T any] struct {
	client    T
	expiresAt time.Time
}

// Client cache TTL.
const clientCacheTTL = 5 * time.Minute

var (
	clientCache   = make(map[string]cachedClient[*kubernetes.Clientset])
	clientCacheMu sync.RWMutex

	aeClientCache   = make(map[string]cachedClient[*apiextensionsclientset.Clientset])
	aeClientCacheMu sync.RWMutex

	restConfigCache   = make(map[string]cachedClient[*rest.Config])
	restConfigCacheMu sync.RWMutex

	dynamicClientCache   = make(map[string]cachedClient[dynamic.Interface])
	dynamicClientCacheMu sync.RWMutex
)

// getOrCreateCached is a generic cache lookup with double-checked locking,
// eliminating duplicate client creation under concurrent cache misses.
func getOrCreateCached[T any](
	cache map[string]cachedClient[T],
	mu *sync.RWMutex,
	name string,
	builder func() (T, error),
) (T, error) {
	cacheKey := "name:" + name

	// Fast path: read lock only.
	mu.RLock()
	if cached, ok := cache[cacheKey]; ok && time.Now().Before(cached.expiresAt) {
		defer mu.RUnlock()
		return cached.client, nil
	}
	mu.RUnlock()

	// Slow path: acquire write lock and double-check.
	mu.Lock()
	if cached, ok := cache[cacheKey]; ok && time.Now().Before(cached.expiresAt) {
		defer mu.Unlock()
		return cached.client, nil
	}

	client, err := builder()
	if err != nil {
		mu.Unlock()
		var zero T
		return zero, err
	}

	cache[cacheKey] = cachedClient[T]{
		client:    client,
		expiresAt: time.Now().Add(clientCacheTTL),
	}
	mu.Unlock()

	return client, nil
}

// buildRestConfig 从 kubeconfig 字符串构造 rest.Config。
// 不再自动降级 Insecure:无 CA 证书时让 clientcmd 走默认行为(失败即报错),
// 避免静默跳过 TLS 校验带来的中间人风险。
func buildRestConfig(kubeConf string) (*rest.Config, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConf))
	if err != nil {
		return nil, fmt.Errorf("初始化客户端配置错误:%w", err)
	}
	// 限流与超时,避免单个慢集群拖垮服务
	config.QPS = 50
	config.Burst = 100
	config.Timeout = 30 * time.Second
	return config, nil
}

// getCachedKubeConfig retrieves the kubeconfig for a cluster, using DB lookup.
func getCachedKubeConfig(name string) (string, error) {
	var k8sCluster clustermodel.K8SCluster
	if err := database.DB.Model(&clustermodel.K8SCluster{}).
		Where(map[string]any{"cluster_name": name}).
		First(&k8sCluster).Error; err != nil {
		return "", err
	}
	kubeConfig, err := auth.DecryptAES(k8sCluster.KubeConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("解密集群 %s 凭证失败:%s", name, err.Error()))
		return "", fmt.Errorf("解密集群凭证失败")
	}
	return kubeConfig, nil
}

// GetK8sClient creates a k8s client from kubeconfig string.
func GetK8sClient(k8sConf string) (*kubernetes.Clientset, error) {
	config, err := buildRestConfig(k8sConf)
	if err != nil {
		return nil, err
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("初始化客户端错误:%w", err)
	}
	return clientSet, nil
}

// GetK8sClientByName retrieves a k8s client by cluster name with caching.
func GetK8sClientByName(name string) (*kubernetes.Clientset, error) {
	return getOrCreateCached(clientCache, &clientCacheMu, name, func() (*kubernetes.Clientset, error) {
		kubeConfig, err := getCachedKubeConfig(name)
		if err != nil {
			return nil, err
		}
		return GetK8sClient(kubeConfig)
	})
}

// GetK8sConf retrieves the kubeconfig string by cluster name.
func GetK8sConf(name string) (string, error) {
	return getCachedKubeConfig(name)
}

// GetApiExtensionsClientByName retrieves an apiextensions client by cluster name with caching.
func GetApiExtensionsClientByName(name string) (*apiextensionsclientset.Clientset, error) {
	return getOrCreateCached(aeClientCache, &aeClientCacheMu, name, func() (*apiextensionsclientset.Clientset, error) {
		kubeConfig, err := getCachedKubeConfig(name)
		if err != nil {
			return nil, err
		}
		config, err := buildRestConfig(kubeConfig)
		if err != nil {
			return nil, err
		}
		clientSet, err := apiextensionsclientset.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("初始化apiextensions客户端错误:%w", err)
		}
		return clientSet, nil
	})
}

// GetRestConfigByName retrieves the REST config by cluster name with caching.
func GetRestConfigByName(name string) (*rest.Config, error) {
	return getOrCreateCached(restConfigCache, &restConfigCacheMu, name, func() (*rest.Config, error) {
		kubeConfig, err := getCachedKubeConfig(name)
		if err != nil {
			return nil, err
		}
		return buildRestConfig(kubeConfig)
	})
}

// GetDynamicClientByName retrieves a dynamic client by cluster name with caching.
func GetDynamicClientByName(name string) (dynamic.Interface, error) {
	return getOrCreateCached(dynamicClientCache, &dynamicClientCacheMu, name, func() (dynamic.Interface, error) {
		config, err := GetRestConfigByName(name)
		if err != nil {
			return nil, err
		}
		client, err := dynamic.NewForConfig(config)
		if err != nil {
			return nil, fmt.Errorf("初始化dynamic客户端错误:%w", err)
		}
		return client, nil
	})
}

// InvalidateClient 删除指定集群的全部缓存项(client/ae/rest/dynamic)。
// 在集群更新或删除后调用,确保后续请求重建客户端。
func InvalidateClient(name string) {
	cacheKey := "name:" + name

	clientCacheMu.Lock()
	delete(clientCache, cacheKey)
	clientCacheMu.Unlock()

	aeClientCacheMu.Lock()
	delete(aeClientCache, cacheKey)
	aeClientCacheMu.Unlock()

	restConfigCacheMu.Lock()
	delete(restConfigCache, cacheKey)
	restConfigCacheMu.Unlock()

	dynamicClientCacheMu.Lock()
	delete(dynamicClientCache, cacheKey)
	dynamicClientCacheMu.Unlock()
}