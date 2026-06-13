package api

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gkube/pkg/response"
)

type auditHandler struct{}

var Audit = new(auditHandler)

const auditFile = "config/audit-logs.json"

type AuditLog struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	User      string            `json:"user"`
	Action    string            `json:"action"`
	Resource  string            `json:"resource"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Cluster   string            `json:"cluster"`
	Details   map[string]string `json:"details"`
	IP        string            `json:"ip"`
	UserAgent string            `json:"userAgent"`
	Status    string            `json:"status"` // success, failure
	Error     string            `json:"error,omitempty"`
}

type AuditStore struct {
	Logs []AuditLog `json:"logs"`
}

func loadAuditLogs() *AuditStore {
	store := &AuditStore{Logs: []AuditLog{}}
	data, err := os.ReadFile(auditFile)
	if err == nil {
		json.Unmarshal(data, store)
	}
	return store
}

func saveAuditLogs(store *AuditStore) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(auditFile, data, 0644)
}

// ListAuditLogs lists audit logs with filtering
func (h *auditHandler) ListAuditLogs(c *gin.Context) {
	user := c.Query("user")
	action := c.Query("action")
	resource := c.Query("resource")
	status := c.Query("status")

	store := loadAuditLogs()
	var result []AuditLog

	for _, log := range store.Logs {
		if user != "" && log.User != user {
			continue
		}
		if action != "" && log.Action != action {
			continue
		}
		if resource != "" && log.Resource != resource {
			continue
		}
		if status != "" && log.Status != status {
			continue
		}
		result = append(result, log)
	}

	response.Success(c, "获取成功", result)
}

// GetAuditLog gets a specific audit log
func (h *auditHandler) GetAuditLog(c *gin.Context) {
	id := c.Query("id")

	store := loadAuditLogs()
	for _, log := range store.Logs {
		if log.ID == id {
			response.Success(c, "获取成功", log)
			return
		}
	}

	response.Fail(c, "审计日志不存在")
}

// CreateAuditLog creates a new audit log entry
func (h *auditHandler) CreateAuditLog(c *gin.Context) {
	var log AuditLog
	if err := c.ShouldBindJSON(&log); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	log.ID = fmt.Sprintf("audit-%d", time.Now().UnixNano())
	log.Timestamp = time.Now()
	log.IP = c.ClientIP()
	log.UserAgent = c.GetHeader("User-Agent")

	store := loadAuditLogs()
	store.Logs = append(store.Logs, log)

	if err := saveAuditLogs(store); err != nil {
		response.Fail(c, "保存审计日志失败: "+err.Error())
		return
	}

	response.Success(c, "审计日志已创建", log)
}

// GetAuditStats gets audit log statistics
func (h *auditHandler) GetAuditStats(c *gin.Context) {
	store := loadAuditLogs()

	stats := map[string]interface{}{
		"total":   len(store.Logs),
		"byUser":  make(map[string]int),
		"byAction": make(map[string]int),
		"byResource": make(map[string]int),
		"byStatus": make(map[string]int),
	}

	byUser := stats["byUser"].(map[string]int)
	byAction := stats["byAction"].(map[string]int)
	byResource := stats["byResource"].(map[string]int)
	byStatus := stats["byStatus"].(map[string]int)

	for _, log := range store.Logs {
		byUser[log.User]++
		byAction[log.Action]++
		byResource[log.Resource]++
		byStatus[log.Status]++
	}

	response.Success(c, "获取成功", stats)
}

// ClearAuditLogs clears all audit logs
func (h *auditHandler) ClearAuditLogs(c *gin.Context) {
	store := &AuditStore{Logs: []AuditLog{}}

	if err := saveAuditLogs(store); err != nil {
		response.Fail(c, "清除审计日志失败: "+err.Error())
		return
	}

	response.Success(c, "审计日志已清除", nil)
}
