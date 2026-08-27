package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
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
		cluster := c.Query("cluster")
		if cluster == "" {
			cluster = c.Query("clusterId")
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
// 不匹配则返回 nil。
func parseK8sPath(path string) []string {
	const prefix = "/v1/k8s/"
	if !strings.HasPrefix(path, prefix) {
		return nil
	}
	rest := strings.TrimPrefix(path, prefix)
	parts := strings.SplitN(rest, "/", 2)
	if len(parts) < 2 {
		return nil
	}
	return parts
}
