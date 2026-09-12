package k8s

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sHpa "gkube/pkg/k8s/hpa"
	"gkube/pkg/logger"
	"gkube/pkg/response"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetHPADetail = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sHpa.GetHPADetail(ctx, client, namespace, name)
	},
	"执行成功",
)

var GetHPAYaml = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		yaml, err := k8sHpa.GetHPAYaml(ctx, client, namespace, name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"yaml": yaml}, nil
	},
	"执行成功",
)

var GetHPAEvents = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return k8sHpa.GetHPAEvents(ctx, client, namespace, name)
	},
	"执行成功",
)

var PauseHPA = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return nil, k8sHpa.PauseHPA(ctx, client, namespace, name)
	},
	"暂停HPA成功",
)

var ResumeHPA = NamespacedHandler(
	func(ctx context.Context, client *kubernetes.Clientset, namespace, name string) (any, error) {
		return nil, k8sHpa.ResumeHPA(ctx, client, namespace, name)
	},
	"恢复HPA成功",
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
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sHpa.CreateHPA(c.Request.Context(), client, body.Namespace, body.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(body.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sHpa.UpdateHPA(c.Request.Context(), client, body.Namespace, body.Yaml); err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "更新HPA成功", nil)
}

// DeleteHPA 删除
func DeleteHPA(c *gin.Context) {
	var p NamespacedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	if err := k8sHpa.DeleteHPA(c.Request.Context(), client, p.Namespace, p.Name); err != nil {
		response.FailWithError(c, ensureAppError(err))
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
func buildTargetSet(ctx context.Context, client *kubernetes.Clientset, hpaList []map[string]any) (map[nsKindName]bool, map[nsKind]bool) {
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
			deps, err := client.AppsV1().Deployments(nk.namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Error(fmt.Sprintf("orphan detection: failed to list Deployments in %s: %v", nk.namespace, err))
				errored[nk] = true
				continue
			}
			for _, d := range deps.Items {
				exists[nsKindName{nk.namespace, nk.kind, d.Name}] = true
			}
		case "StatefulSet":
			sts, err := client.AppsV1().StatefulSets(nk.namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Error(fmt.Sprintf("orphan detection: failed to list StatefulSets in %s: %v", nk.namespace, err))
				errored[nk] = true
				continue
			}
			for _, s := range sts.Items {
				exists[nsKindName{nk.namespace, nk.kind, s.Name}] = true
			}
		case "ReplicaSet":
			rss, err := client.AppsV1().ReplicaSets(nk.namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Error(fmt.Sprintf("orphan detection: failed to list ReplicaSets in %s: %v", nk.namespace, err))
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
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	selector, err := buildLabelSelector(p.LabelFilters)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	hpaList, err := k8sHpa.GetHPAList(c.Request.Context(), client, p.Namespace, selector)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
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
		targetSet, errored := buildTargetSet(c.Request.Context(), client, result)
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
