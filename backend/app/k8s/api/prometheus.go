package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/gin-gonic/gin"
	"gkube/pkg/response"
)

type prometheus struct{}

var Prometheus = new(prometheus)

type PrometheusConfig struct {
	URL string `json:"url"`
}

func getPrometheusURL() string {
	// Try to read from config file
	data, err := os.ReadFile("config/prometheus.json")
	if err == nil {
		var config PrometheusConfig
		if json.Unmarshal(data, &config) == nil && config.URL != "" {
			return config.URL
		}
	}
	// Default
	return "http://prometheus:9090"
}

// QueryPrometheus executes an instant query against Prometheus
func (p *prometheus) QueryPrometheus(c *gin.Context) {
	query := c.Query("query")
	if query == "" {
		response.Fail(c, "query parameter is required")
		return
	}

	prometheusURL := getPrometheusURL()
	endpoint := fmt.Sprintf("%s/api/v1/query?query=%s", prometheusURL, url.QueryEscape(query))

	resp, err := http.Get(endpoint)
	if err != nil {
		response.Fail(c, fmt.Sprintf("Failed to query Prometheus: %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Fail(c, fmt.Sprintf("Failed to read response: %s", err.Error()))
		return
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		response.Fail(c, fmt.Sprintf("Failed to parse response: %s", err.Error()))
		return
	}

	response.Success(c, "执行成功", result)
}

// QueryPrometheusRange executes a range query against Prometheus
func (p *prometheus) QueryPrometheusRange(c *gin.Context) {
	query := c.Query("query")
	start := c.Query("start")
	end := c.Query("end")
	step := c.DefaultQuery("step", "60s")

	if query == "" {
		response.Fail(c, "query parameter is required")
		return
	}

	prometheusURL := getPrometheusURL()
	endpoint := fmt.Sprintf("%s/api/v1/query_range?query=%s&start=%s&end=%s&step=%s",
		prometheusURL, url.QueryEscape(query), start, end, step)

	resp, err := http.Get(endpoint)
	if err != nil {
		response.Fail(c, fmt.Sprintf("Failed to query Prometheus: %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Fail(c, fmt.Sprintf("Failed to read response: %s", err.Error()))
		return
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		response.Fail(c, fmt.Sprintf("Failed to parse response: %s", err.Error()))
		return
	}

	response.Success(c, "执行成功", result)
}

// GetPrometheusTargets returns the list of Prometheus targets
func (p *prometheus) GetPrometheusTargets(c *gin.Context) {
	prometheusURL := getPrometheusURL()
	endpoint := fmt.Sprintf("%s/api/v1/targets", prometheusURL)

	resp, err := http.Get(endpoint)
	if err != nil {
		response.Fail(c, fmt.Sprintf("Failed to get targets: %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Fail(c, fmt.Sprintf("Failed to read response: %s", err.Error()))
		return
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		response.Fail(c, fmt.Sprintf("Failed to parse response: %s", err.Error()))
		return
	}

	response.Success(c, "执行成功", result)
}

// GetPrometheusAlerts returns the list of Prometheus alerts
func (p *prometheus) GetPrometheusAlerts(c *gin.Context) {
	prometheusURL := getPrometheusURL()
	endpoint := fmt.Sprintf("%s/api/v1/alerts", prometheusURL)

	resp, err := http.Get(endpoint)
	if err != nil {
		response.Fail(c, fmt.Sprintf("Failed to get alerts: %s", err.Error()))
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		response.Fail(c, fmt.Sprintf("Failed to read response: %s", err.Error()))
		return
	}

	var result map[string]any
	if err := json.Unmarshal(body, &result); err != nil {
		response.Fail(c, fmt.Sprintf("Failed to parse response: %s", err.Error()))
		return
	}

	response.Success(c, "执行成功", result)
}

// UpdatePrometheusConfig updates the Prometheus configuration
func (p *prometheus) UpdatePrometheusConfig(c *gin.Context) {
	var config PrometheusConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		response.Fail(c, fmt.Sprintf("参数错误: %s", err.Error()))
		return
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		response.Fail(c, fmt.Sprintf("Failed to serialize config: %s", err.Error()))
		return
	}

	if err := os.WriteFile("config/prometheus.json", data, 0644); err != nil {
		response.Fail(c, fmt.Sprintf("Failed to save config: %s", err.Error()))
		return
	}

	response.Success(c, "配置保存成功", nil)
}

// GetPrometheusConfig returns the current Prometheus configuration
func (p *prometheus) GetPrometheusConfig(c *gin.Context) {
	config := PrometheusConfig{
		URL: getPrometheusURL(),
	}
	response.Success(c, "执行成功", config)
}
