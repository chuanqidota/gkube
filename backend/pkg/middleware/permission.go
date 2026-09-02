package middleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gkube/internal/cluster/model"
	"gkube/pkg/auth"
	"gkube/pkg/database"
	"gkube/pkg/logger"
	"gkube/pkg/response"
)

// 集群名 → ID 缓存（集群很少改名，简单缓存即可）
var clusterIDCache sync.Map // map[string]uint

// InvalidateClusterIDCache 失效指定集群名的 ID 缓存（集群删除后调用）。
func InvalidateClusterIDCache(clusterName string) {
	if clusterName != "" {
		clusterIDCache.Delete(clusterName)
	}
}

// RequirePermission 基于 RBAC 的权限检查中间件。
// 从请求路径自动推断 resourceGroup 和 verb，检查用户是否有对应权限。
func RequirePermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 提取用户 ID
		userIDVal, exists := c.Get("userID")
		if !exists {
			response.FailWithStatus(c, http.StatusUnauthorized, "未认证")
			return
		}
		userID, ok := userIDVal.(uint)
		if !ok {
			response.FailWithStatus(c, http.StatusUnauthorized, "未认证")
			return
		}

		// 2. 提取作用域
		clusterName := c.Query("clusterName")
		namespace := extractNamespace(c)

		// 解析集群 ID（K8s 路由通过 clusterName 传参，中间件用 ID 匹配）
		var clusterID uint
		if clusterName != "" {
			if cached, ok := clusterIDCache.Load(clusterName); ok {
				clusterID = cached.(uint)
			} else {
				var cluster model.K8SCluster
				if err := database.DB.Where("cluster_name = ?", clusterName).First(&cluster).Error; err == nil {
					clusterID = cluster.ID
					clusterIDCache.Store(clusterName, clusterID)
				}
			}
		}

		// 3. 查询用户权限（缓存）
		username, _ := c.Get("username")
		name, _ := username.(string)

		// 超级管理员：config 白名单 OR DB is_super_admin
		if auth.IsAdmin(name) {
			c.Next()
			return
		}

		cached := auth.GetUserPermissions(userID)
		if cached.IsSuperAdmin {
			c.Next()
			return
		}

		// 4. 从路径推断 resourceGroup + verb
		path := c.Request.URL.Path
		resourceGroup := resolveResourceGroup(path)
		if resourceGroup == "" {
			response.FailWithStatus(c, http.StatusForbidden, "权限不足")
			return
		}
		verb := resolveVerb(path, c.Request.Method)

		// 5. 检查任一绑定的角色 permissions 是否包含 resourceGroup + verb
		for _, binding := range cached.Bindings {
			// 集群级绑定：匹配集群
			if binding.ClusterID == clusterID {
				if binding.Namespace == "" || binding.Namespace == namespace {
					perms := binding.Role.GetPermissionsMap()
					if verbs, ok := perms[resourceGroup]; ok {
						for _, v := range verbs {
							if v == verb {
								c.Next()
								return
							}
						}
					}
				}
			}
		}

		logger.Warn(fmt.Sprintf("RequirePermission: 用户 %s 无权访问 %s %s (集群=%s, NS=%s)",
			name, verb, resourceGroup, clusterName, namespace))
		response.FailWithStatus(c, http.StatusForbidden, "权限不足")
	}
}

// extractNamespace 从请求中提取命名空间。
// GET 从 query 参数读，POST/PUT/DELETE 从 body 读。
// ShouldBindBodyWith 会把 body 缓存进 context（BodyBytesKey），但 c.Request.Body
// 流本身已被消费——必须用缓存副本重置回去，否则下游 handler 的
// ShouldBindJSON 会读到 EOF（历史上曾因此导致全部 JSON 写接口 400）。
func extractNamespace(c *gin.Context) string {
	if ns := c.Query("namespace"); ns != "" {
		return ns
	}
	if c.Request.Method != "GET" {
		defer restoreRequestBody(c)
		var body struct {
			Namespace string `json:"namespace"`
		}
		if err := c.ShouldBindBodyWith(&body, binding.JSON); err == nil && body.Namespace != "" {
			return body.Namespace
		}
	}
	return ""
}

// restoreRequestBody 用 gin 缓存的 body 副本重置 c.Request.Body，
// 保证下游 handler 仍可通过 ShouldBindJSON 读到完整请求体。
func restoreRequestBody(c *gin.Context) {
	if raw, ok := c.Get(gin.BodyBytesKey); ok {
		if bodyBytes, ok := raw.([]byte); ok {
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}
}

// resolveResourceGroup 从 URL 路径推断资源组。长前缀优先匹配。
func resolveResourceGroup(path string) string {
	switch {
	case strings.Contains(path, "/volumesnapshotclass/"):
		return "storage"
	case strings.Contains(path, "/volumesnapshot/"):
		return "storage"
	case strings.Contains(path, "/networkpolicy/"):
		return "network"
	case strings.Contains(path, "/resourcequota/"):
		return "config"
	case strings.Contains(path, "/limitrange/"):
		return "config"
	case strings.Contains(path, "/storageclass/"):
		return "storage"
	case strings.Contains(path, "/configmap/"):
		return "config"
	case strings.Contains(path, "/replicaset/"):
		return "workload"
	case strings.Contains(path, "/statefulset/"):
		return "workload"
	case strings.Contains(path, "/daemonset/"):
		return "workload"
	case strings.Contains(path, "/cronjob/"):
		return "workload"
	case strings.Contains(path, "/deployment/"):
		return "workload"
	case strings.Contains(path, "/namespace/"):
		return "namespace"
	case strings.Contains(path, "/ingress/"):
		return "network"
	case strings.Contains(path, "/service/"):
		return "network"
	case strings.Contains(path, "/secret/"):
		return "config"
	case strings.Contains(path, "/container/exec"):
		return "terminal"
	case strings.Contains(path, "/pod/"):
		return "workload"
	case strings.Contains(path, "/event/"):
		return "event"
	case strings.Contains(path, "/node/"):
		return "node"
	case strings.Contains(path, "/job/"):
		return "workload"
	case strings.Contains(path, "/pvc/"):
		return "storage"
	case strings.Contains(path, "/pv/"):
		return "storage"
	case strings.Contains(path, "/hpa/"):
		return "workload"
	case strings.Contains(path, "/audit/"):
		return "audit"
	case strings.Contains(path, "/crd/"):
		return "crd"
	// 注意：/cluster/nodes 必须在 /cluster/ 之前，否则节点路由误归 cluster_mgmt
	case strings.Contains(path, "/cluster/nodes"):
		return "node"
	case strings.Contains(path, "/cluster/"):
		return "cluster_mgmt"
	case strings.Contains(path, "/log"):
		return "terminal"
	default:
		logger.Warn(fmt.Sprintf("RequirePermission: 未识别的资源路径 %s，拒绝访问", path))
		return ""
	}
}

// specialVerbOverride 特殊路由 verb 覆写
type specialVerbOverride struct {
	Suffix string
	Method string
	Verb   string
}

var specialVerbOverrides = []specialVerbOverride{
	// 节点管理
	{"/node/cordon", "PUT", "cordon"},
	{"/node/taints", "PUT", "taint"},
	{"/node/drain", "PUT", "drain"},
	{"/node/labels", "PUT", "taint"},
	{"/node/update-yaml", "PUT", "taint"},
	// 工作负载操作
	{"/deployment/scale", "PUT", "update"},
	{"/deployment/restart", "POST", "update"},
	{"/deployment/rollback", "POST", "update"},
	{"/statefulset/scale", "PUT", "update"},
	{"/statefulset/restart", "POST", "update"},
	{"/statefulset/rollback", "POST", "update"},
	{"/daemonset/restart", "POST", "update"},
	{"/daemonset/rollback", "POST", "update"},
	{"/cronjob/suspend", "PUT", "update"},
	{"/cronjob/resume", "PUT", "update"},
	{"/cronjob/trigger", "POST", "update"},
	{"/hpa/pause", "POST", "update"},
	{"/hpa/resume", "POST", "update"},
	// 命名空间特殊
	{"/namespace/create", "POST", "create"},
	{"/namespace/list", "GET", "read"},
	{"/namespace/detail", "GET", "read"},
	// 终端/日志
	{"/container/exec", "GET", "terminal"},
	{"/log", "GET", "terminal"},
	{"/log/stream", "GET", "terminal"},
}

// resolveVerb 根据特殊路由表和 HTTP Method 确定操作 verb。
func resolveVerb(path, method string) string {
	// 检查特殊覆写表
	for _, ov := range specialVerbOverrides {
		if strings.HasSuffix(path, ov.Suffix) && method == ov.Method {
			return ov.Verb
		}
	}
	// 默认映射
	switch method {
	case "GET":
		return "read"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "read"
	}
}
