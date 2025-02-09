package init

import (
	"fmt"

	"gkube/app/k8s/model"
	"gkube/pkg/database"

	"k8s.io/client-go/kubernetes"
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
	config.TLSClientConfig.Insecure = true
	if err != nil {
		return nil, fmt.Errorf("初始化客户端配置错误:%s", err.Error())
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
	clientSet, err := GetK8sClient(k8sCluster.KubeConfig)
	return clientSet, err
}
