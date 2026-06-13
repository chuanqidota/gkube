package api

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	k8sNp "gkube/pkg/k8s/networkpolicy"
	"gkube/pkg/response"
)

type networkPolicy struct{}

var NetworkPolicy = new(networkPolicy)

func (np *networkPolicy) GetNetworkPolicyList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	npList, err := k8sNp.GetNetworkPolicyList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取NetworkPolicy列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, np := range npList {
		var podSelector string
		if np.Spec.PodSelector.MatchLabels != nil {
			for k, v := range np.Spec.PodSelector.MatchLabels {
				if podSelector != "" {
					podSelector += ", "
				}
				podSelector += k + "=" + v
			}
		}
		if podSelector == "" {
			podSelector = "All pods"
		}
		var policyTypes []string
		for _, pt := range np.Spec.PolicyTypes {
			policyTypes = append(policyTypes, string(pt))
		}
		result = append(result, map[string]any{
			"name":          np.Name,
			"namespace":     np.Namespace,
			"pod_selector":  podSelector,
			"policy_types":  policyTypes,
			"ingress_rules": len(np.Spec.Ingress),
			"egress_rules":  len(np.Spec.Egress),
			"age":           np.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
			"labels":        np.Labels,
		})
	}
	response.Success(c, "执行成功", result)
}

func (np *networkPolicy) GetNetworkPolicyDetail(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	npList, err := k8sNp.GetNetworkPolicyList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取NetworkPolicy详情失败:%s", err.Error()))
		return
	}
	for _, n := range npList {
		if n.Name == name {
			response.Success(c, "执行成功", n)
			return
		}
	}
	response.Fail(c, "NetworkPolicy不存在")
}

func (np *networkPolicy) GetNetworkPolicyYaml(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	yamlContent, err := k8sNp.GetNetworkPolicyYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取NetworkPolicy YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (np *networkPolicy) CreateNetworkPolicy(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace"`
		YamlContent string `json:"yamlContent"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(req.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sNp.CreateNetworkPolicy(client, req.Namespace, req.YamlContent); err != nil {
		response.Fail(c, fmt.Sprintf("创建NetworkPolicy失败:%s", err.Error()))
		return
	}
	response.Success(c, "创建NetworkPolicy成功", nil)
}

func (np *networkPolicy) UpdateNetworkPolicy(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace"`
		YamlContent string `json:"yamlContent"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}
	client, err := k8s.GetK8sClientByName(req.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sNp.UpdateNetworkPolicy(client, req.Namespace, req.YamlContent); err != nil {
		response.Fail(c, fmt.Sprintf("更新NetworkPolicy失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新NetworkPolicy成功", nil)
}

func (np *networkPolicy) DeleteNetworkPolicy(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sNp.DeleteNetworkPolicy(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除NetworkPolicy失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除NetworkPolicy成功", nil)
}
