package k8s

import (
	"context"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	k8sclient "gkube/pkg/k8s"
	k8sHpa "gkube/pkg/k8s/hpa"
	"gkube/pkg/response"
)

type hpa struct{}

var Hpa = new(hpa)

type nsKindName struct {
	namespace string
	kind      string
	name      string
}

type nsKind struct {
	namespace string
	kind      string
}

// buildTargetSet builds a lookup map of existing workload targets per namespace.
// For each HPA target_kind + namespace, it batch-fetches all workloads once
// and stores their names in a set for O(1) existence checks.
// Returns (exists set, errored pairs) — errored pairs should not be marked as orphan.
func buildTargetSet(client *kubernetes.Clientset, hpaList []map[string]any) (map[nsKindName]bool, map[nsKind]bool) {
	// Collect unique (namespace, kind) pairs
	seen := make(map[nsKind]bool)
	for _, h := range hpaList {
		ns, _ := h["namespace"].(string)
		kind, _ := h["target_kind"].(string)
		if ns != "" && kind != "" {
			seen[nsKind{ns, kind}] = true
		}
	}

	// Batch-fetch workloads per (namespace, kind) and build name sets
	exists := make(map[nsKindName]bool)
	errored := make(map[nsKind]bool)
	for nk := range seen {
		switch nk.kind {
		case "Deployment":
			deps, err := client.AppsV1().Deployments(nk.namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Printf("orphan detection: failed to list Deployments in %s: %v", nk.namespace, err)
				errored[nk] = true
				continue
			}
			for _, d := range deps.Items {
				exists[nsKindName{nk.namespace, nk.kind, d.Name}] = true
			}
		case "StatefulSet":
			sts, err := client.AppsV1().StatefulSets(nk.namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Printf("orphan detection: failed to list StatefulSets in %s: %v", nk.namespace, err)
				errored[nk] = true
				continue
			}
			for _, s := range sts.Items {
				exists[nsKindName{nk.namespace, nk.kind, s.Name}] = true
			}
		case "ReplicaSet":
			rss, err := client.AppsV1().ReplicaSets(nk.namespace).List(context.TODO(), metav1.ListOptions{})
			if err != nil {
				log.Printf("orphan detection: failed to list ReplicaSets in %s: %v", nk.namespace, err)
				errored[nk] = true
				continue
			}
			for _, r := range rss.Items {
				exists[nsKindName{nk.namespace, nk.kind, r.Name}] = true
			}
		}
	}
	return exists, errored
}

func (h *hpa) GetHPAList(c *gin.Context) {
	namespace := c.Query("namespace")
	clusterName := c.Query("clusterName")
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	hpaList, err := k8sHpa.GetHPAList(client, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取HPA列表失败:%s", err.Error()))
		return
	}

	var result []map[string]any
	for _, hpa := range hpaList {
		var minReplicas int32
		if hpa.Spec.MinReplicas != nil {
			minReplicas = *hpa.Spec.MinReplicas
		}
		// Check paused status from annotations
		paused := false
		if hpa.Annotations != nil && hpa.Annotations["gkube.io/paused"] == "true" {
			paused = true
		}
		result = append(result, map[string]any{
			"name":             hpa.Name,
			"namespace":        hpa.Namespace,
			"min_replicas":     minReplicas,
			"max_replicas":     hpa.Spec.MaxReplicas,
			"current_replicas": hpa.Status.CurrentReplicas,
			"desired_replicas": hpa.Status.DesiredReplicas,
			"target":           hpa.Spec.ScaleTargetRef.Name,
			"target_kind":      hpa.Spec.ScaleTargetRef.Kind,
			"conditions":       hpa.Status.Conditions,
			"metrics":          hpa.Spec.Metrics,
			"current_metrics":  hpa.Status.CurrentMetrics,
			"age":              hpa.CreationTimestamp.Time.Format("2006-01-02 15:04:05"),
			"labels":           hpa.Labels,
			"paused":           paused,
		})
	}

	// Batch-check target existence (orphan detection)
	if len(result) > 0 {
		targetSet, errored := buildTargetSet(client, result)
		for _, r := range result {
			ns, _ := r["namespace"].(string)
			kind, _ := r["target_kind"].(string)
			name, _ := r["target"].(string)
			nk := nsKind{ns, kind}
			if errored[nk] {
				// API error for this (ns, kind) — don't mark as orphan
				r["target_exists"] = true
			} else {
				r["target_exists"] = targetSet[nsKindName{ns, kind, name}]
			}
		}
	}

	response.Success(c, "执行成功", result)
}

func (h *hpa) GetHPADetail(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	hpa, err := k8sHpa.GetHPADetail(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取HPA详情失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", hpa)
}

func (h *hpa) GetHPAYaml(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	yamlContent, err := k8sHpa.GetHPAYaml(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取HPA YAML失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", map[string]string{"yaml": yamlContent})
}

func (h *hpa) CreateHPA(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(req.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sHpa.CreateHPA(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("创建HPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "创建HPA成功", nil)
}

func (h *hpa) UpdateHPA(c *gin.Context) {
	var req struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(req.ClusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sHpa.UpdateHPA(client, req.Namespace, req.Yaml); err != nil {
		response.Fail(c, fmt.Sprintf("更新HPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新HPA成功", nil)
}

func (h *hpa) DeleteHPA(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sHpa.DeleteHPA(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("删除HPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除HPA成功", nil)
}

func (h *hpa) GetHPAEvents(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	events, err := k8sHpa.GetHPAEvents(client, namespace, name)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取HPA事件失败:%s", err.Error()))
		return
	}
	response.Success(c, "执行成功", events)
}

func (h *hpa) PauseHPA(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sHpa.PauseHPA(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("暂停HPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "暂停HPA成功", nil)
}

func (h *hpa) ResumeHPA(c *gin.Context) {
	namespace := c.Query("namespace")
	name := c.Query("name")
	clusterName := c.Query("clusterName")
	if name == "" {
		response.Fail(c, "name参数不能为空")
		return
	}
	client, err := k8sclient.GetK8sClientByName(clusterName)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sHpa.ResumeHPA(client, namespace, name); err != nil {
		response.Fail(c, fmt.Sprintf("恢复HPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "恢复HPA成功", nil)
}
