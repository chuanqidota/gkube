package k8s

import (
	"fmt"
	"sync"
	"time"

	"gkube/app/k8s/model"
	"gkube/pkg/auth"
	"gkube/pkg/database"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// cachedClient wraps a kubernetes.Clientset with an expiration time
type cachedClient[T any] struct {
	client    T
	expiresAt time.Time
}

// Client cache with TTL
const clientCacheTTL = 5 * time.Minute

var (
	clientCache   = make(map[string]cachedClient[*kubernetes.Clientset])
	clientCacheMu sync.RWMutex

	aeClientCache   = make(map[string]cachedClient[*apiextensionsclientset.Clientset])
	aeClientCacheMu sync.RWMutex

	dynamicClientCache   = make(map[string]cachedClient[dynamic.Interface])
	dynamicClientCacheMu sync.RWMutex

	restConfigCache   = make(map[string]cachedClient[*rest.Config])
	restConfigCacheMu sync.RWMutex
)

// getCachedKubeConfig retrieves the kubeconfig for a cluster, using DB lookup
func getCachedKubeConfig(name string) (string, error) {
	var k8sCluster model.K8SCluster
	if err := database.DB.Model(&model.K8SCluster{}).
		Where(map[string]any{"cluster_name": name}).
		Scan(&k8sCluster).Error; err != nil {
		return "", err
	}
	kubeConfig, err := auth.DecryptAES(k8sCluster.KubeConfig)
	if err != nil {
		return "", fmt.Errorf("解密集群凭证失败:%s", err.Error())
	}
	return kubeConfig, nil
}

// GetK8sClient creates a k8s client from kubeconfig string
func GetK8sClient(k8sConf string) (*kubernetes.Clientset, error) {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(k8sConf))
	if err != nil {
		return nil, fmt.Errorf("初始化客户端配置错误:%s", err.Error())
	}
	// 仅在 kubeconfig 未配置 CA 证书时启用 Insecure 跳过 TLS 验证
	if config.TLSClientConfig.CAFile == "" && len(config.TLSClientConfig.CAData) == 0 {
		config.TLSClientConfig.Insecure = true
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("初始化客户端错误:%s", err.Error())
	}
	return clientSet, nil
}

// GetK8sClientClusterID retrieves a k8s client by cluster ID with caching
func GetK8sClientClusterID(id uint) (*kubernetes.Clientset, error) {
	cacheKey := fmt.Sprintf("id:%d", id)

	clientCacheMu.RLock()
	if cached, ok := clientCache[cacheKey]; ok && time.Now().Before(cached.expiresAt) {
		clientCacheMu.RUnlock()
		return cached.client, nil
	}
	clientCacheMu.RUnlock()

	var k8sCluster model.K8SCluster
	if err := database.DB.Model(&model.K8SCluster{}).
		Where(map[string]any{"id": id}).
		Scan(&k8sCluster).Error; err != nil {
		return nil, err
	}
	kubeConfig, err := auth.DecryptAES(k8sCluster.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("解密集群凭证失败:%s", err.Error())
	}
	clientSet, err := GetK8sClient(kubeConfig)
	if err != nil {
		return nil, err
	}

	clientCacheMu.Lock()
	clientCache[cacheKey] = cachedClient[*kubernetes.Clientset]{
		client:    clientSet,
		expiresAt: time.Now().Add(clientCacheTTL),
	}
	clientCacheMu.Unlock()

	return clientSet, nil
}

// GetK8sClientByName retrieves a k8s client by cluster name with caching
func GetK8sClientByName(name string) (*kubernetes.Clientset, error) {
	cacheKey := "name:" + name

	clientCacheMu.RLock()
	if cached, ok := clientCache[cacheKey]; ok && time.Now().Before(cached.expiresAt) {
		clientCacheMu.RUnlock()
		return cached.client, nil
	}
	clientCacheMu.RUnlock()

	kubeConfig, err := getCachedKubeConfig(name)
	if err != nil {
		return nil, err
	}
	clientSet, err := GetK8sClient(kubeConfig)
	if err != nil {
		return nil, err
	}

	clientCacheMu.Lock()
	clientCache[cacheKey] = cachedClient[*kubernetes.Clientset]{
		client:    clientSet,
		expiresAt: time.Now().Add(clientCacheTTL),
	}
	clientCacheMu.Unlock()

	return clientSet, nil
}

// GetK8sConf retrieves the kubeconfig string by cluster name
func GetK8sConf(name string) (string, error) {
	return getCachedKubeConfig(name)
}

// CreateDynamicClient creates a dynamic client from kubeconfig string
func CreateDynamicClient(kubeConf string) (dynamic.Interface, error) {
	config, err := clientcmd.Load([]byte(kubeConf))
	if err != nil {
		return nil, fmt.Errorf("加载 kubeconfig 失败: %v", err)
	}

	clientConfig := clientcmd.NewDefaultClientConfig(
		*config,
		&clientcmd.ConfigOverrides{},
	)

	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("创建 REST 配置失败: %v", err)
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("创建动态客户端失败: %v", err)
	}

	return dynamicClient, nil
}

// GetApiExtensionsClientByName retrieves an apiextensions client by cluster name with caching
func GetApiExtensionsClientByName(name string) (*apiextensionsclientset.Clientset, error) {
	cacheKey := "name:" + name

	aeClientCacheMu.RLock()
	if cached, ok := aeClientCache[cacheKey]; ok && time.Now().Before(cached.expiresAt) {
		aeClientCacheMu.RUnlock()
		return cached.client, nil
	}
	aeClientCacheMu.RUnlock()

	kubeConfig, err := getCachedKubeConfig(name)
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return nil, fmt.Errorf("初始化客户端配置错误:%s", err.Error())
	}
	if config.TLSClientConfig.CAFile == "" && len(config.TLSClientConfig.CAData) == 0 {
		config.TLSClientConfig.Insecure = true
	}
	clientSet, err := apiextensionsclientset.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("初始化apiextensions客户端错误:%s", err.Error())
	}

	aeClientCacheMu.Lock()
	aeClientCache[cacheKey] = cachedClient[*apiextensionsclientset.Clientset]{
		client:    clientSet,
		expiresAt: time.Now().Add(clientCacheTTL),
	}
	aeClientCacheMu.Unlock()

	return clientSet, nil
}

// GetRestConfigByName retrieves the REST config by cluster name with caching
func GetRestConfigByName(name string) (*rest.Config, error) {
	cacheKey := "name:" + name

	restConfigCacheMu.RLock()
	if cached, ok := restConfigCache[cacheKey]; ok && time.Now().Before(cached.expiresAt) {
		restConfigCacheMu.RUnlock()
		return cached.client, nil
	}
	restConfigCacheMu.RUnlock()

	kubeConfig, err := getCachedKubeConfig(name)
	if err != nil {
		return nil, err
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return nil, fmt.Errorf("初始化客户端配置错误:%s", err.Error())
	}
	if config.TLSClientConfig.CAFile == "" && len(config.TLSClientConfig.CAData) == 0 {
		config.TLSClientConfig.Insecure = true
	}

	restConfigCacheMu.Lock()
	restConfigCache[cacheKey] = cachedClient[*rest.Config]{
		client:    config,
		expiresAt: time.Now().Add(clientCacheTTL),
	}
	restConfigCacheMu.Unlock()

	return config, nil
}
