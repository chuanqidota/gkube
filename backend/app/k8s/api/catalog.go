package api

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gin-gonic/gin"
	"gkube/pkg/response"
)

type catalogHandler struct{}

var Catalog = new(catalogHandler)

type ChartInfo struct {
	Name        string `json:"name"`
	Version     string `json:"version"`
	AppVersion  string `json:"appVersion"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Repository  string `json:"repository"`
}

type InstallRequest struct {
	Chart     string `json:"chart"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   string `json:"version"`
	Values    string `json:"values"`
}

// ListCharts lists available Helm charts
func (h *catalogHandler) ListCharts(c *gin.Context) {
	charts, err := listHelmCharts()
	source := "helm"
	if err != nil {
		// Return sample charts if Helm is not available
		charts = getSampleCharts()
		source = "sample"
	}

	response.Success(c, "获取成功", gin.H{
		"charts":  charts,
		"source":  source,
		"message": getSourceMessage(source, "helm"),
	})
}

// GetChartDetails gets details of a specific chart
func (h *catalogHandler) GetChartDetails(c *gin.Context) {
	chartName := c.Query("chart")
	if chartName == "" {
		response.Fail(c, "图表名称不能为空")
		return
	}

	details, err := getChartDetails(chartName)
	if err != nil {
		details = &ChartInfo{
			Name:        chartName,
			Version:     "1.0.0",
			Description: "Sample chart",
			Category:    "Other",
		}
	}

	response.Success(c, "获取成功", details)
}

// InstallChart installs a Helm chart
func (h *catalogHandler) InstallChart(c *gin.Context) {
	var req InstallRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	if req.Chart == "" || req.Name == "" || req.Namespace == "" {
		response.Fail(c, "图表名称、发布名称和命名空间不能为空")
		return
	}

	err := installHelmChart(req)
	if err != nil {
		response.Fail(c, fmt.Sprintf("安装失败: %s", err.Error()))
		return
	}

	response.Success(c, "安装成功", nil)
}

// ListReleases lists installed Helm releases
func (h *catalogHandler) ListReleases(c *gin.Context) {
	namespace := c.Query("namespace")

	releases, err := listHelmReleases(namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取发布列表失败: %s", err.Error()))
		return
	}

	response.Success(c, "获取成功", releases)
}

// UninstallRelease uninstalls a Helm release
func (h *catalogHandler) UninstallRelease(c *gin.Context) {
	name := c.Query("name")
	namespace := c.Query("namespace")

	if name == "" {
		response.Fail(c, "发布名称不能为空")
		return
	}

	err := uninstallHelmRelease(name, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("卸载失败: %s", err.Error()))
		return
	}

	response.Success(c, "卸载成功", nil)
}

func listHelmCharts() ([]ChartInfo, error) {
	// Check if helm is available
	_, err := exec.LookPath("helm")
	if err != nil {
		return nil, fmt.Errorf("helm not found")
	}

	// List repos
	cmd := exec.Command("helm", "repo", "list", "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var repos []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	}
	json.Unmarshal(output, &repos)

	// Search charts
	var charts []ChartInfo
	for _, repo := range repos {
		searchCmd := exec.Command("helm", "search", "repo", repo.Name+"/", "-o", "json")
		searchOutput, err := searchCmd.Output()
		if err != nil {
			continue
		}

		var results []struct {
			Name        string `json:"name"`
			Version     string `json:"version"`
			AppVersion  string `json:"app_version"`
			Description string `json:"description"`
		}
		json.Unmarshal(searchOutput, &results)

		for _, r := range results {
			name := strings.TrimPrefix(r.Name, repo.Name+"/")
			charts = append(charts, ChartInfo{
				Name:        name,
				Version:     r.Version,
				AppVersion:  r.AppVersion,
				Description: r.Description,
				Category:    categorizeChart(name),
				Repository:  repo.Name,
			})
		}
	}

	return charts, nil
}

func getChartDetails(chartName string) (*ChartInfo, error) {
	cmd := exec.Command("helm", "show", "chart", chartName, "-o", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var details struct {
		Name        string `json:"name"`
		Version     string `json:"version"`
		AppVersion  string `json:"appVersion"`
		Description string `json:"description"`
	}
	json.Unmarshal(output, &details)

	return &ChartInfo{
		Name:        details.Name,
		Version:     details.Version,
		AppVersion:  details.AppVersion,
		Description: details.Description,
		Category:    categorizeChart(details.Name),
	}, nil
}

func installHelmChart(req InstallRequest) error {
	args := []string{"install", req.Name, req.Chart, "--namespace", req.Namespace, "--create-namespace"}

	if req.Version != "" {
		args = append(args, "--version", req.Version)
	}

	if req.Values != "" {
		args = append(args, "--set", req.Values)
	}

	cmd := exec.Command("helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), string(output))
	}

	return nil
}

func listHelmReleases(namespace string) ([]map[string]interface{}, error) {
	args := []string{"list", "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	} else {
		args = append(args, "-A")
	}

	cmd := exec.Command("helm", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var releases []map[string]interface{}
	json.Unmarshal(output, &releases)

	return releases, nil
}

func uninstallHelmRelease(name, namespace string) error {
	args := []string{"uninstall", name}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	cmd := exec.Command("helm", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), string(output))
	}

	return nil
}

func getSourceMessage(source, tool string) string {
	if source == "sample" {
		return fmt.Sprintf("Using sample data. Install %s for live data.", tool)
	}
	return "Connected to " + tool
}

func categorizeChart(name string) string {
	name = strings.ToLower(name)
	categories := map[string][]string{
		"Web":       {"nginx", "apache", "httpd", "caddy", "traefik", "haproxy"},
		"Database":  {"mysql", "postgresql", "postgres", "redis", "mongodb", "mariadb", "cassandra", "couchdb", "influxdb"},
		"Search":    {"elasticsearch", "solr", "opensearch"},
		"Monitoring": {"prometheus", "grafana", "alertmanager", "loki", "jaeger", "zipkin"},
		"CI/CD":     {"jenkins", "gitlab", "argocd", "flux", "tekton", "drone"},
		"Registry":  {"harbor", "registry", "nexus"},
		"Storage":   {"minio", "ceph", "rook", "longhorn", "nfs"},
		"Security":  {"vault", "keycloak", "oauth2-proxy"},
		"Messaging": {"kafka", "rabbitmq", "nats", "redis"},
	}

	for category, keywords := range categories {
		for _, keyword := range keywords {
			if strings.Contains(name, keyword) {
				return category
			}
		}
	}

	return "Other"
}

func getSampleCharts() []ChartInfo {
	return []ChartInfo{
		{Name: "nginx", Version: "15.0.0", AppVersion: "1.25.0", Description: "High performance web server", Category: "Web", Repository: "bitnami"},
		{Name: "redis", Version: "18.0.0", AppVersion: "7.2.0", Description: "In-memory data store", Category: "Database", Repository: "bitnami"},
		{Name: "mysql", Version: "9.0.0", AppVersion: "8.0.0", Description: "Relational database", Category: "Database", Repository: "bitnami"},
		{Name: "postgresql", Version: "13.0.0", AppVersion: "16.0.0", Description: "Advanced relational database", Category: "Database", Repository: "bitnami"},
		{Name: "mongodb", Version: "14.0.0", AppVersion: "7.0.0", Description: "NoSQL document database", Category: "Database", Repository: "bitnami"},
		{Name: "elasticsearch", Version: "19.0.0", AppVersion: "8.10.0", Description: "Search and analytics engine", Category: "Search", Repository: "elastic"},
		{Name: "grafana", Version: "7.0.0", AppVersion: "10.0.0", Description: "Observability platform", Category: "Monitoring", Repository: "grafana"},
		{Name: "prometheus", Version: "25.0.0", AppVersion: "2.47.0", Description: "Monitoring system", Category: "Monitoring", Repository: "prometheus"},
		{Name: "jenkins", Version: "4.0.0", AppVersion: "2.420.0", Description: "CI/CD automation server", Category: "CI/CD", Repository: "jenkins"},
		{Name: "harbor", Version: "1.0.0", AppVersion: "2.9.0", Description: "Container registry", Category: "Registry", Repository: "harbor"},
		{Name: "minio", Version: "5.0.0", AppVersion: "2023.09.07", Description: "Object storage", Category: "Storage", Repository: "bitnami"},
		{Name: "keycloak", Version: "16.0.0", AppVersion: "22.0.0", Description: "Identity and access management", Category: "Security", Repository: "bitnami"},
	}
}
