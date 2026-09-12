package k8s

import (
	"context"

	"github.com/gin-gonic/gin"
	apperr "gkube/pkg/errors"
	k8sclient "gkube/pkg/k8s"
	k8sPv "gkube/pkg/k8s/pv"
	"gkube/pkg/response"
	"k8s.io/client-go/kubernetes"
)

// ---------------------------------------------------------------------------
// 标准 handler（使用 wrapper）
// ---------------------------------------------------------------------------

var GetPVByName = ClusterGetHandler(
	func(ctx context.Context, client *kubernetes.Clientset, name string) (any, error) {
		return k8sPv.GetPVByName(ctx, client, name)
	},
	"执行成功",
)

var GetPVYaml = ClusterGetHandler(
	func(ctx context.Context, client *kubernetes.Clientset, name string) (any, error) {
		yaml, err := k8sPv.GetPVYaml(ctx, client, name)
		if err != nil {
			return nil, err
		}
		return map[string]string{"yaml": yaml}, nil
	},
	"执行成功",
)

var CreatePV = ClusterCreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, yaml string) error {
		return k8sPv.CreatePV(ctx, client, yaml)
	},
	"执行成功",
)

var UpdatePV = ClusterCreateHandler(
	func(ctx context.Context, client *kubernetes.Clientset, yaml string) error {
		return k8sPv.UpdatePV(ctx, client, yaml)
	},
	"执行成功",
)

var DeletePVByName = ClusterDeleteHandler(
	func(ctx context.Context, client *kubernetes.Clientset, name string) error {
		return k8sPv.DeletePVByName(ctx, client, name)
	},
	"执行成功",
)

// ---------------------------------------------------------------------------
// 特殊 handler —— GetPVList 不支持分页，无法使用 ClusterListHandler
// ---------------------------------------------------------------------------

// GetPVList 列表 —— 非分页
func GetPVList(c *gin.Context) {
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
	pvList, err := k8sPv.GetPVList(c.Request.Context(), client, selector)
	if err != nil {
		response.FailWithError(c, ensureAppError(err))
		return
	}
	response.Success(c, "执行成功", pvList)
}
