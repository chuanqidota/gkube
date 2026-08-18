package middleware

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gkube/pkg/audit"
)

// auditSkipPaths 审计自身 + 终端/日志(有独立记录)
var auditSkipPaths = map[string]bool{
	"audit":          true,
	"container/exec": true,
	"log":            true,
	"log/stream":     true,
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

		// 跳过审计自身和终端/日志
		if auditSkipPaths[resource] || auditSkipPaths[resource+"/"+action] {
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
