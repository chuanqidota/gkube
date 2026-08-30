package middleware

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"gkube/pkg/audit"
)

// auditSkipPaths 不需要审计记录的精确路径(终端/日志有独立记录机制,audit/create 会自引用)
var auditSkipPaths = map[string]bool{
	"container/exec": true,
	"log/stream":     true,
	"audit/create":   true, // 避免自引用审计条目
}

// auditSkipResources 跳过整个子树的资源前缀
var auditSkipResources = map[string]bool{
	"log": true, // log 和 log/stream 都跳过,log 有独立记录
}

// AuditLog 自动记录 K8s 写操作的审计日志。
// 挂在 JWTAuth 之后,确保 c.GetString("username") 可用。
func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		// 只记录写操作
		method := c.Request.Method
		if method == "GET" {
			return
		}

		// 解析路径: /v1/k8s/{resource}/{action}
		path := c.Request.URL.Path
		segments := parseK8sPath(path)
		if segments == nil {
			return
		}

		resource := segments[0]
		action := segments[1]

		// RBAC bindings 路由: /v1/rbac/bindings 或 /v1/rbac/bindings/:id
		if resource == "bindings" {
			if action == "" {
				// POST /v1/rbac/bindings → create-binding
				if method == "POST" {
					action = "create-binding"
				}
			} else if _, err := strconv.Atoi(action); err == nil {
				// PUT/DELETE /v1/rbac/bindings/:id → semantic action
				switch method {
				case "PUT":
					action = "update-binding"
				case "DELETE":
					action = "delete-binding"
				}
			}
		}

		// 跳过终端/日志等有独立记录的路径
		if auditSkipResources[resource] || auditSkipPaths[resource+"/"+action] {
			return
		}

		// 判断成功/失败
		status := "success"
		if c.Writer.Status() >= 400 {
			status = "failure"
		}

		// 提取资源名称和命名空间
		name := c.Query("name")
		namespace := c.Query("namespace")
		if namespace == "" {
			namespace = c.Query("ns")
		}
		cluster := c.Query("clusterName")
		if cluster == "" {
			cluster = c.Query("cluster")
		}
		if cluster == "" {
			cluster = c.Query("clusterId")
		}
		// RBAC 路由的 clusterName 在 body（POST/PUT）中
		if cluster == "" && (method == "POST" || method == "PUT") {
			var body struct{ ClusterName string }
			if err := c.ShouldBindBodyWith(&body, binding.JSON); err == nil {
				cluster = body.ClusterName
			}
		}

		log := audit.AuditLog{
			User:      c.GetString("username"),
			Action:    action,
			Resource:  resource,
			Name:      name,
			Namespace: namespace,
			Cluster:   cluster,
			IP:        c.ClientIP(),
			UserAgent: c.GetHeader("User-Agent"),
			Status:    status,
			Details: map[string]string{
				"method":   method,
				"path":     path,
				"duration": time.Since(start).String(),
			},
		}

		audit.RecordAuditLog(log)
	}
}

// parseK8sPath 从 URL 路径中提取 resource 和 action。
// 输入: /v1/k8s/deployment/create  输出: ["deployment", "create"]
// 输入: /v1/rbac/bindings          输出: ["bindings", ""]
// 不匹配则返回 nil。
func parseK8sPath(path string) []string {
	prefixes := []string{"/v1/k8s/", "/v1/rbac/"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			rest := strings.TrimPrefix(path, prefix)
			parts := strings.SplitN(rest, "/", 2)
			if len(parts) == 1 {
				parts = append(parts, "")
			}
			return parts
		}
	}
	return nil
}
