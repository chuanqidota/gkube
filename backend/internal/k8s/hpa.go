package k8s

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	k8sclient "gkube/pkg/k8s"
	k8sHpa "gkube/pkg/k8s/hpa"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

// ---------------------------------------------------------------------------
// 标准 handler
// ---------------------------------------------------------------------------

var GetHPADetail = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sHpa.GetHPADetail(client, namespace, name)
	},
	"执行成功", "获取HPA详情失败",
)

var GetHPAYaml = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		yaml, err := k8sHpa.GetHPAYaml(client, namespace, name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"yaml": yaml}, nil
	},
	"执行成功", "获取HPA YAML失败",
)

var GetHPAEvents = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sHpa.GetHPAEvents(client, namespace, name)
	},
	"执行成功", "获取HPA事件失败",
)

var PauseHPA = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return nil, k8sHpa.PauseHPA(client, namespace, name)
	},
	"暂停HPA成功", "暂停HPA失败",
)

var ResumeHPA = NamespacedHandler(
	func(client *kubernetes.Clientset, namespace, name string) (any, error) {
		return nil, k8sHpa.ResumeHPA(client, namespace, name)
	},
	"恢复HPA成功", "恢复HPA失败",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// CreateHPA 创建
func CreateHPA(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sHpa.CreateHPA(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("创建HPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "创建HPA成功", nil)
}

// UpdateHPA 更新
func UpdateHPA(c *gin.Context) {
	var body struct {
		ClusterName string `json:"clusterName" binding:"required"`
		Namespace   string `json:"namespace" binding:"required"`
		Yaml        string `json:"yaml" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取k8s客户端失败:%s", err.Error()))
		return
	}
	if err := k8sHpa.UpdateHPA(client, body.Namespace, body.Yaml); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("更新HPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "更新HPA成功", nil)
}

// DeleteHPA 删除 —— 原代码用 ShouldBindQuery
func DeleteHPA(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	if err := k8sHpa.DeleteHPA(client, p.Namespace, p.Name); err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("删除HPA失败:%s", err.Error()))
		return
	}
	response.Success(c, "删除HPA成功", nil)
}

// ---------------------------------------------------------------------------
// 内部类型
// ---------------------------------------------------------------------------

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
func buildTargetSet(client *kubernetes.Clientset, hpaList []map[string]any) (map[nsKindName]bool, map[nsKind]bool) {
	seen := make(map[nsKind]bool)
	for _, h := range hpaList {
		ns, _ := h["namespace"].(string)
		kind, _ := h["target_kind"].(string)
		if ns != "" && kind != "" {
			seen[nsKind{ns, kind}] = true
		}
	}
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

// GetHPAList 列表 —— 非分页 + orphan detection
func GetHPAList(c *gin.Context) {
	var p ListParams
	if err := c.ShouldBind(&p); err != nil {
		response.Fail(c, "参数校验失败")
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取k8s客户端失败")
		return
	}
	selector, err := buildLabelSelector(p.LabelFilters)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadGateway, err.Error())
		return
	}
	hpaList, err := k8sHpa.GetHPAList(client, p.Namespace, selector)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, fmt.Sprintf("获取HPA列表失败:%s", err.Error()))
		return
	}
	var result []map[string]any
	for _, hpa := range hpaList {
		var minReplicas int32
		if hpa.Spec.MinReplicas != nil {
			minReplicas = *hpa.Spec.MinReplicas
		}
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
	if len(result) > 0 {
		targetSet, errored := buildTargetSet(client, result)
		for _, r := range result {
			ns, _ := r["namespace"].(string)
			kind, _ := r["target_kind"].(string)
			name, _ := r["target"].(string)
			nk := nsKind{ns, kind}
			if errored[nk] {
				r["target_exists"] = true
			} else {
				r["target_exists"] = targetSet[nsKindName{ns, kind, name}]
			}
		}
	}
	response.Success(c, "执行成功", result)
}
