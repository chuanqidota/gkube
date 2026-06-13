package api

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	"gkube/pkg/es"
	"gkube/pkg/response"
)

type auditHandler struct{}

var Audit = new(auditHandler)

const auditFile = "config/audit-logs.json"
const auditIndex = "gkube-audit-logs"

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

// isElasticsearchAvailable checks if ES client is initialized
func isElasticsearchAvailable() bool {
	return es.ElasticSearch != nil
}

// saveToElasticsearch saves audit log to ES
func saveToElasticsearch(log AuditLog) error {
	if !isElasticsearchAvailable() {
		return fmt.Errorf("elasticsearch not available")
	}
	_, err := es.ElasticSearch.Index().
		Index(auditIndex).
		Id(log.ID).
		BodyJson(log).
		Do(context.Background())
	return err
}

// searchFromElasticsearch searches audit logs from ES
func searchFromElasticsearch(user, action, resource, status string, limit int) ([]AuditLog, error) {
	if !isElasticsearchAvailable() {
		return nil, fmt.Errorf("elasticsearch not available")
	}

	boolQuery := elastic.NewBoolQuery()
	if user != "" {
		boolQuery.Must(elastic.NewTermQuery("user", user))
	}
	if action != "" {
		boolQuery.Must(elastic.NewTermQuery("action", action))
	}
	if resource != "" {
		boolQuery.Must(elastic.NewTermQuery("resource", resource))
	}
	if status != "" {
		boolQuery.Must(elastic.NewTermQuery("status", status))
	}

	searchResult, err := es.ElasticSearch.Search().
		Index(auditIndex).
		Query(boolQuery).
		Sort("timestamp", false).
		Size(limit).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	var logs []AuditLog
	for _, hit := range searchResult.Hits.Hits {
		var log AuditLog
		if err := json.Unmarshal(hit.Source, &log); err == nil {
			logs = append(logs, log)
		}
	}
	return logs, nil
}

// getStatsFromElasticsearch gets audit stats from ES
func getStatsFromElasticsearch() (map[string]interface{}, error) {
	if !isElasticsearchAvailable() {
		return nil, fmt.Errorf("elasticsearch not available")
	}

	// Total count
	totalResult, err := es.ElasticSearch.Count(auditIndex).Do(context.Background())
	if err != nil {
		return nil, err
	}

	stats := map[string]interface{}{
		"total":      totalResult,
		"byUser":     make(map[string]int),
		"byAction":   make(map[string]int),
		"byResource": make(map[string]int),
		"byStatus":   make(map[string]int),
	}

	// Aggregations for breakdowns
	aggResult, err := es.ElasticSearch.Search().
		Index(auditIndex).
		Size(0).
		Aggregation("by_user", elastic.NewTermsAggregation().Field("user.keyword")).
		Aggregation("by_action", elastic.NewTermsAggregation().Field("action.keyword")).
		Aggregation("by_resource", elastic.NewTermsAggregation().Field("resource.keyword")).
		Aggregation("by_status", elastic.NewTermsAggregation().Field("status.keyword")).
		Do(context.Background())
	if err != nil {
		return stats, nil // Return total even if aggs fail
	}

	// Parse aggregations
	if agg, found := aggResult.Aggregations.Terms("by_user"); found {
		byUser := stats["byUser"].(map[string]int)
		for _, bucket := range agg.Buckets {
			byUser[bucket.Key.(string)] = int(bucket.DocCount)
		}
	}
	if agg, found := aggResult.Aggregations.Terms("by_action"); found {
		byAction := stats["byAction"].(map[string]int)
		for _, bucket := range agg.Buckets {
			byAction[bucket.Key.(string)] = int(bucket.DocCount)
		}
	}
	if agg, found := aggResult.Aggregations.Terms("by_resource"); found {
		byResource := stats["byResource"].(map[string]int)
		for _, bucket := range agg.Buckets {
			byResource[bucket.Key.(string)] = int(bucket.DocCount)
		}
	}
	if agg, found := aggResult.Aggregations.Terms("by_status"); found {
		byStatus := stats["byStatus"].(map[string]int)
		for _, bucket := range agg.Buckets {
			byStatus[bucket.Key.(string)] = int(bucket.DocCount)
		}
	}

	return stats, nil
}

// clearElasticsearchAuditLogs clears all audit logs from ES
func clearElasticsearchAuditLogs() error {
	if !isElasticsearchAvailable() {
		return fmt.Errorf("elasticsearch not available")
	}
	_, err := es.ElasticSearch.DeleteByQuery(auditIndex).
		Query(elastic.NewMatchAllQuery()).
		Do(context.Background())
	return err
}

// File-based fallback functions
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

	// Try Elasticsearch first
	if isElasticsearchAvailable() {
		logs, err := searchFromElasticsearch(user, action, resource, status, 500)
		if err == nil {
			response.Success(c, "获取成功", logs)
			return
		}
	}

	// Fallback to file storage
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

	// Try Elasticsearch first
	if isElasticsearchAvailable() {
		getResult, err := es.ElasticSearch.Get().
			Index(auditIndex).
			Id(id).
			Do(context.Background())
		if err == nil && getResult.Found {
			var log AuditLog
			if err := json.Unmarshal(getResult.Source, &log); err == nil {
				response.Success(c, "获取成功", log)
				return
			}
		}
	}

	// Fallback to file storage
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

	// Try Elasticsearch first
	if isElasticsearchAvailable() {
		if err := saveToElasticsearch(log); err == nil {
			response.Success(c, "审计日志已创建", log)
			return
		}
	}

	// Fallback to file storage
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
	// Try Elasticsearch first
	if isElasticsearchAvailable() {
		stats, err := getStatsFromElasticsearch()
		if err == nil {
			response.Success(c, "获取成功", stats)
			return
		}
	}

	// Fallback to file storage
	store := loadAuditLogs()

	stats := map[string]interface{}{
		"total":      len(store.Logs),
		"byUser":     make(map[string]int),
		"byAction":   make(map[string]int),
		"byResource": make(map[string]int),
		"byStatus":   make(map[string]int),
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
	// Try Elasticsearch first
	if isElasticsearchAvailable() {
		if err := clearElasticsearchAuditLogs(); err == nil {
			response.Success(c, "审计日志已清除", nil)
			return
		}
	}

	// Fallback to file storage
	store := &AuditStore{Logs: []AuditLog{}}

	if err := saveAuditLogs(store); err != nil {
		response.Fail(c, "清除审计日志失败: "+err.Error())
		return
	}

	response.Success(c, "审计日志已清除", nil)
}
