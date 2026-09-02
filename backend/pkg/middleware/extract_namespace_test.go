package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	rbacmodel "gkube/internal/rbac/model"
	"gkube/pkg/auth"
)

// TestExtractNamespaceBodyRegression 回归：RequirePermission 读过 body 后，
// 下游 handler 的 ShouldBindJSON 必须仍能读到完整 body。
// 历史缺陷：extractNamespace 用 ShouldBindBodyWith 消费 c.Request.Body 后未恢复，
// 导致全部 JSON 写接口返回"参数校验失败"。
func TestExtractNamespaceBodyRegression(t *testing.T) {
	gin.SetMode(gin.TestMode)
	clusterIDCache.Store("c1", uint(1))

	editorRole := rbacmodel.Role{
		ID: 11, Name: "ns-editor", ScopeType: "namespace", IsSystem: true,
		Permissions: `{"workload":["read","create","update"]}`,
	}
	auth.SetPermissionsForTest(uint(3), &auth.CachedPermissions{
		IsSuperAdmin: false,
		Bindings: []rbacmodel.PermissionBinding{{
			ID: 2, UserID: 3, RoleID: 11, ClusterID: 1, Namespace: "default",
			Role: editorRole,
		}},
		ExpireAt: time.Now().Add(time.Hour),
	})

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", uint(3))
		c.Set("username", "editoruser")
		c.Next()
	})
	r.Use(RequirePermission())
	r.PUT("/v1/k8s/deployment/update-yaml", func(c *gin.Context) {
		var body struct {
			ClusterName string `json:"clusterName" binding:"required"`
			Namespace   string `json:"namespace"`
			Name        string `json:"name" binding:"required"`
			Yaml        string `json:"yaml" binding:"required"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			t.Errorf("handler ShouldBindJSON 失败（body 被中间件消费）: %v", err)
			return
		}
		if body.Namespace != "default" || body.Name != "nginx" {
			t.Errorf("body 字段值异常: %+v", body)
		}
	})

	req := httptest.NewRequest("PUT", "/v1/k8s/deployment/update-yaml?clusterName=c1",
		strings.NewReader(`{"clusterName":"c1","namespace":"default","name":"nginx","yaml":"a: b"}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(httptest.NewRecorder(), req)
}
