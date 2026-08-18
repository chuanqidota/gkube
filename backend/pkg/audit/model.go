package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"gkube/pkg/es"
)

const auditFile = "config/audit-logs.json"

// AuditIndex ES 审计日志索引名,同时被 internal/k8s/audit.go 的查询函数引用。
const AuditIndex = "gkube-audit-logs"

var fileMu sync.Mutex

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
	Status    string            `json:"status"`
	Error     string            `json:"error,omitempty"`
}

type auditStore struct {
	Logs []AuditLog `json:"logs"`
}

// RecordAuditLog 审计日志统一写入入口,异步写入,不阻塞调用方。
func RecordAuditLog(log AuditLog) {
	go func() {
		if log.ID == "" {
			log.ID = fmt.Sprintf("audit-%d", time.Now().UnixNano())
		}
		if log.Timestamp.IsZero() {
			log.Timestamp = time.Now()
		}
		if saveToES(log) {
			return
		}
		saveToFile(log)
	}()
}

func saveToES(log AuditLog) bool {
	if es.ElasticSearch == nil {
		return false
	}
	_, err := es.ElasticSearch.Index().
		Index(AuditIndex).
		Id(log.ID).
		BodyJson(log).
		Do(context.Background())
	return err == nil
}

func saveToFile(log AuditLog) {
	fileMu.Lock()
	defer fileMu.Unlock()

	store := &auditStore{Logs: []AuditLog{}}
	data, err := os.ReadFile(auditFile)
	if err == nil {
		json.Unmarshal(data, store)
	}
	store.Logs = append(store.Logs, log)
	out, _ := json.MarshalIndent(store, "", "  ")
	os.WriteFile(auditFile, out, 0644)
}
