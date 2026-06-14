package k8s

import (
	"fmt"

	"gkube/app/k8s/model"
	"gkube/pkg/auth"
	"gkube/pkg/database"

	apiextensionsclientset "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// GetK8sClient
//
//	@Description: 初始化k8s客户端
//	@param k8sConf
//	@return *kubernetes.Clientset
//	@return error
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

// GetK8sClientClusterID
//
//	@Description: 通过id获取k8s客户端
//	@param id
//	@return *kubernetes.Clientset
//	@return error
func GetK8sClientClusterID(id uint) (*kubernetes.Clientset, error) {
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
	return clientSet, err
}

// GetK8sClientByName
//
//	@Description: 根据名称获取k8s客户端
//	@param name
//	@return *kubernetes.Clientset
//	@return error
func GetK8sClientByName(name string) (*kubernetes.Clientset, error) {
	var k8sCluster model.K8SCluster
	if err := database.DB.Model(&model.K8SCluster{}).
		Where(map[string]any{"cluster_name": name}).
		Scan(&k8sCluster).Error; err != nil {
		return nil, err
	}
	kubeConfig, err := auth.DecryptAES(k8sCluster.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("解密集群凭证失败:%s", err.Error())
	}
	clientSet, err := GetK8sClient(kubeConfig)
	return clientSet, err
}

// GetK8sConf
//
//	@Description: 根据名称获取k8s配置信息
//	@param name
//	@return string
//	@return error
func GetK8sConf(name string) (string, error) {
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

// CreateDynamicClient
//
//	@Description: 从 kubeconfig 字符串创建动态客户端
//	@param kubeConf
//	@return dynamic.Interface
//	@return error
func CreateDynamicClient(kubeConf string) (dynamic.Interface, error) {
	// 将字符串转换为 clientcmdapi.Config 对象
	config, err := clientcmd.Load([]byte(kubeConf))
	if err != nil {
		return nil, fmt.Errorf("加载 kubeconfig 失败: %v", err)
	}

	// 创建客户端配置
	clientConfig := clientcmd.NewDefaultClientConfig(
		*config,
		&clientcmd.ConfigOverrides{}, // 可在此处覆盖配置参数
	)

	// 转换为 rest.Config
	restConfig, err := clientConfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("创建 REST 配置失败: %v", err)
	}

	// 初始化动态客户端
	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("创建动态客户端失败: %v", err)
	}

	return dynamicClient, nil
}

// GetApiExtensionsClientByName
//
//	@Description: 根据名称获取apiextensions客户端（用于CRD操作）
//	@param name
//	@return *apiextensionsclientset.Clientset
//	@return error
func GetApiExtensionsClientByName(name string) (*apiextensionsclientset.Clientset, error) {
	var k8sCluster model.K8SCluster
	if err := database.DB.Model(&model.K8SCluster{}).
		Where(map[string]any{"cluster_name": name}).
		Scan(&k8sCluster).Error; err != nil {
		return nil, err
	}
	kubeConfig, err := auth.DecryptAES(k8sCluster.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("解密集群凭证失败:%s", err.Error())
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
	return clientSet, nil
}

// GetRestConfigByName
//
//	@Description: 根据名称获取REST配置（用于动态客户端操作）
//	@param name
//	@return *rest.Config
//	@return error
func GetRestConfigByName(name string) (*rest.Config, error) {
	var k8sCluster model.K8SCluster
	if err := database.DB.Model(&model.K8SCluster{}).
		Where(map[string]any{"cluster_name": name}).
		Scan(&k8sCluster).Error; err != nil {
		return nil, err
	}
	kubeConfig, err := auth.DecryptAES(k8sCluster.KubeConfig)
	if err != nil {
		return nil, fmt.Errorf("解密集群凭证失败:%s", err.Error())
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConfig))
	if err != nil {
		return nil, fmt.Errorf("初始化客户端配置错误:%s", err.Error())
	}
	if config.TLSClientConfig.CAFile == "" && len(config.TLSClientConfig.CAData) == 0 {
		config.TLSClientConfig.Insecure = true
	}
	return config, nil
}
