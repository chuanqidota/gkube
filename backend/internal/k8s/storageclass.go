package k8s

import (
	"context"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sStorageClass "gkube/pkg/k8s/storageclass"
	"gkube/pkg/response"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetStorageClassByName = ClusterGetHandler(
	func(ctx context.Context, client *kubernetes.Clientset, name string) (any, error) {
		return k8sStorageClass.GetStorageClassByName(ctx, client, name)
	},
	"执行成功",
)

var GetStorageClassYaml = ClusterGetHandler(
	func(ctx context.Context, client *kubernetes.Clientset, name string) (any, error) {
		yaml, err := k8sStorageClass.GetStorageClassYaml(ctx, client, name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"yaml": yaml}, nil
	},
	"执行成功",
)

var CreateStorageClass = ClusterCreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, yaml string) error {
		return k8sStorageClass.CreateStorageClass(ctx, client, yaml)
	},
	"执行成功",
)

var UpdateStorageClass = ClusterCreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, yaml string) error {
		return k8sStorageClass.UpdateStorageClass(ctx, client, yaml)
	},
	"执行成功",
)

var DeleteStorageClassByName = ClusterDeleteHandler(
	func(ctx context.Context, client *kubernetes.Clientset, name string) error {
		return k8sStorageClass.DeleteStorageClassByName(ctx, client, name)
	},
	"执行成功",
)

// ---------------------------------------------------------------------------
// 特殊 handler
// ---------------------------------------------------------------------------

// GetStorageClassList 列表 —— 非分页，不支持 ClusterListHandler
func GetStorageClassList(c *gin.Context) {
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
	storageClasses, err := k8sStorageClass.GetStorageClassList(c.Request.Context(), client, selector)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", storageClasses)
}

// GetStorageClassEvents 事件 —— 内联 K8s 调用（集群级资源，需跨命名空间搜索事件）
func GetStorageClassEvents(c *gin.Context) {
	var p ClusterScopedParams
	if err := c.ShouldBindQuery(&p); err != nil {
		response.FailWithError(c, apperr.Validation("参数校验失败", err))
		return
	}
	if p.Name == "" {
		response.FailWithError(c, apperr.BadRequest("name参数不能为空", nil))
		return
	}
	client, err := k8sclient.GetK8sClientByName(p.ClusterName)
	if err != nil {
		response.FailWithError(c, apperr.K8sClientFail("获取k8s客户端失败", err))
		return
	}
	events, err := client.CoreV1().Events(corev1.NamespaceAll).List(c.Request.Context(), metav1.ListOptions{
		FieldSelector: fields.AndSelectors(
			fields.OneTermEqualSelector("involvedObject.name", p.Name),
			fields.OneTermEqualSelector("involvedObject.kind", "StorageClass"),
		).String(),
	})
	if err != nil {
		response.FailWithError(c, apperr.K8sAPIFail("获取StorageClass事件失败", err))
		return
	}
	var result []map[string]any
	for _, event := range events.Items {
		lastSeen := ""
		if !event.LastTimestamp.IsZero() {
			lastSeen = event.LastTimestamp.Time.Format("2006-01-02 15:04:05")
		}
		result = append(result, map[string]any{
			"type":      event.Type,
			"reason":    event.Reason,
			"message":   event.Message,
			"last_seen": lastSeen,
		})
	}
	response.Success(c, "执行成功", result)
}
