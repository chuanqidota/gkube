package k8s

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/olivere/elastic/v7"
	auditlog "gkube/pkg/audit"
	"gkube/pkg/es"
	"gkube/pkg/response"
)

type auditHandler struct{}

var Audit = new(auditHandler)

const auditFile = "config/audit-logs.json"

var auditFileMu sync.Mutex

type auditStore struct {
	Logs []auditlog.AuditLog `json:"logs"`
}

func isElasticsearchAvailable() bool {
	return es.ElasticSearch != nil
}

// searchFromElasticsearch searches audit logs from ES
func searchFromElasticsearch(user, action, resource, status string, limit int) ([]auditlog.AuditLog, error) {
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
		Index(auditlog.AuditIndex).
		Query(boolQuery).
		Sort("timestamp", false).
		Size(limit).
		Do(context.Background())
	if err != nil {
		return nil, err
	}

	var logs []auditlog.AuditLog
	for _, hit := range searchResult.Hits.Hits {
		var log auditlog.AuditLog
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

	totalResult, err := es.ElasticSearch.Count(auditlog.AuditIndex).Do(context.Background())
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

	aggResult, err := es.ElasticSearch.Search().
		Index(auditlog.AuditIndex).
		Size(0).
		Aggregation("by_user", elastic.NewTermsAggregation().Field("user.keyword")).
		Aggregation("by_action", elastic.NewTermsAggregation().Field("action.keyword")).
		Aggregation("by_resource", elastic.NewTermsAggregation().Field("resource.keyword")).
		Aggregation("by_status", elastic.NewTermsAggregation().Field("status.keyword")).
		Do(context.Background())
	if err != nil {
		return stats, nil
	}

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
	_, err := es.ElasticSearch.DeleteByQuery(auditlog.AuditIndex).
		Query(elastic.NewMatchAllQuery()).
		Do(context.Background())
	return err
}

// File-based fallback functions
func loadAuditLogs() *auditStore {
	auditFileMu.Lock()
	defer auditFileMu.Unlock()
	store := &auditStore{Logs: []auditlog.AuditLog{}}
	data, err := os.ReadFile(auditFile)
	if err == nil {
		json.Unmarshal(data, store)
	}
	return store
}

func saveAuditLogs(store *auditStore) error {
	auditFileMu.Lock()
	defer auditFileMu.Unlock()
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

	if isElasticsearchAvailable() {
		logs, err := searchFromElasticsearch(user, action, resource, status, 500)
		if err == nil {
			response.Success(c, "获取成功", logs)
			return
		}
	}

	store := loadAuditLogs()
	var result []auditlog.AuditLog

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

	if isElasticsearchAvailable() {
		getResult, err := es.ElasticSearch.Get().
			Index(auditlog.AuditIndex).
			Id(id).
			Do(context.Background())
		if err == nil && getResult.Found {
			var log auditlog.AuditLog
			if err := json.Unmarshal(getResult.Source, &log); err == nil {
				response.Success(c, "获取成功", log)
				return
			}
		}
	}

	store := loadAuditLogs()
	for _, log := range store.Logs {
		if log.ID == id {
			response.Success(c, "获取成功", log)
			return
		}
	}

	response.Fail(c, "审计日志不存在")
}

// CreateAuditLog creates a new audit log entry (manual, via API)
func (h *auditHandler) CreateAuditLog(c *gin.Context) {
	var log auditlog.AuditLog
	if err := c.ShouldBindJSON(&log); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误:%s", err.Error()))
		return
	}

	log.IP = c.ClientIP()
	log.UserAgent = c.GetHeader("User-Agent")

	auditlog.RecordAuditLog(log)
	response.Success(c, "审计日志已创建", log)
}

// GetAuditStats gets audit log statistics
func (h *auditHandler) GetAuditStats(c *gin.Context) {
	if isElasticsearchAvailable() {
		stats, err := getStatsFromElasticsearch()
		if err == nil {
			response.Success(c, "获取成功", stats)
			return
		}
	}

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
	if isElasticsearchAvailable() {
		if err := clearElasticsearchAuditLogs(); err == nil {
			response.Success(c, "审计日志已清除", nil)
			return
		}
	}

	store := &auditStore{Logs: []auditlog.AuditLog{}}

	if err := saveAuditLogs(store); err != nil {
		response.Fail(c, fmt.Sprintf("清除审计日志失败:%s", err.Error()))
		return
	}

	response.Success(c, "审计日志已清除", nil)
}
