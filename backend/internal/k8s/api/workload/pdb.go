package workload

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gkube/pkg/k8s"
	k8sPdb "gkube/pkg/k8s/pdb"
	"gkube/pkg/response"
)

type pdb struct{}

var Pdb = new(pdb)

func (p *pdb) GetPDBList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	pdbList, err := k8sPdb.GetPDBList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取PDB列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, pdb := range pdbList {
		var minAvailable, maxUnavailable string
		if pdb.Spec.MinAvailable != nil {
			minAvailable = pdb.Spec.MinAvailable.String()
		}
		if pdb.Spec.MaxUnavailable != nil {
			maxUnavailable = pdb.Spec.MaxUnavailable.String()
		}
		var selector string
		if pdb.Spec.Selector != nil && pdb.Spec.Selector.MatchLabels != nil {
			for k, v := range pdb.Spec.Selector.MatchLabels {
				if selector != "" {
					selector += ", "
				}
				selector += k + "=" + v
			}
		}
		if selector == "" {
			selector = "All pods"
		}
		result = append(result, map[string]any{
			"name":            pdb.Name,
			"namespace":       pdb.Namespace,
			"min_available":   minAvailable,
			"max_unavailable": maxUnavailable,
			"selector":        selector,
			"current":         pdb.Status.CurrentHealthy,
			"desired":         pdb.Status.DesiredHealthy,
			"allowed":         pdb.Status.DisruptionsAllowed,
			"age":             pdb.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
			"labels":          pdb.Labels,
		})
	}
	response.Success(c, "执行成功", result)
}

func (p *pdb) GetPDBDetail(c *gin.Context) {
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
	pdbList, err := k8sPdb.GetPDBList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取PDB详情失败:%s", err.Error()))
		return
	}
	for _, p := range pdbList {
		if p.Name == name {
			response.Success(c, "执行成功", p)
			return
		}
	}
	response.Fail(c, "PDB不存在")
}

func (p *pdb) GetPDBYaml(c *gin.Context) {
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
	yamlContent, err := k8sPdb.GetPDBYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取PDB YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (p *pdb) CreatePDB(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace"`
		Yaml string `json:"yaml"`
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
	if err := k8sPdb.CreatePDB(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建PDB失败:%s", err.Error()))
		return
	}
	response.Success(c, "创建PDB成功", nil)
}

func (p *pdb) UpdatePDB(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName"`
		Namespace   string `json:"namespace"`
		Yaml string `json:"yaml"`
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
	if err := k8sPdb.UpdatePDB(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新PDB失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新PDB成功", nil)
}

func (p *pdb) DeletePDB(c *gin.Context) {
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
	if err := k8sPdb.DeletePDB(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除PDB失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除PDB成功", nil)
}
