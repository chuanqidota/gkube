package api

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/gin-gonic/gin"
	"gkube/pkg/response"
)

type gitopsHandler struct{}

var GitOps = new(gitopsHandler)

type Application struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Project    string            `json:"project"`
	RepoURL    string            `json:"repoURL"`
	Path       string            `json:"path"`
	TargetRev  string            `json:"targetRevision"`
	Status     string            `json:"status"`
	Health     string            `json:"health"`
	SyncStatus string            `json:"syncStatus"`
	Labels     map[string]string `json:"labels"`
}

type SyncRequest struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Revision  string `json:"revision"`
	Force     bool   `json:"force"`
	DryRun    bool   `json:"dryRun"`
}

// ListApplications lists ArgoCD applications
func (h *gitopsHandler) ListApplications(c *gin.Context) {
	namespace := c.Query("namespace")

	apps, err := listArgoApplications(namespace)
	source := "argocd"
	if err != nil {
		// Return sample apps if ArgoCD is not available
		apps = getSampleApplications()
		source = "sample"
	}

	response.Success(c, "获取成功", gin.H{
		"applications": apps,
		"source":       source,
		"message":      getSourceMessage(source, "argocd"),
	})
}

// GetApplication gets details of a specific application
func (h *gitopsHandler) GetApplication(c *gin.Context) {
	name := c.Query("name")
	namespace := c.Query("namespace")

	if name == "" {
		response.Fail(c, "应用名称不能为空")
		return
	}

	app, err := getArgoApplication(name, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取应用失败: %s", err.Error()))
		return
	}

	response.Success(c, "获取成功", app)
}

// SyncApplication syncs an ArgoCD application
func (h *gitopsHandler) SyncApplication(c *gin.Context) {
	var req SyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	if req.Name == "" {
		response.Fail(c, "应用名称不能为空")
		return
	}

	err := syncArgoApplication(req)
	if err != nil {
		response.Fail(c, fmt.Sprintf("同步失败: %s", err.Error()))
		return
	}

	response.Success(c, "同步成功", nil)
}

// GetApplicationHistory gets the sync history of an application
func (h *gitopsHandler) GetApplicationHistory(c *gin.Context) {
	name := c.Query("name")
	namespace := c.Query("namespace")

	if name == "" {
		response.Fail(c, "应用名称不能为空")
		return
	}

	history, err := getArgoHistory(name, namespace)
	if err != nil {
		response.Fail(c, fmt.Sprintf("获取历史失败: %s", err.Error()))
		return
	}

	response.Success(c, "获取成功", history)
}

// RollbackApplication rolls back to a previous revision
func (h *gitopsHandler) RollbackApplication(c *gin.Context) {
	name := c.Query("name")
	namespace := c.Query("namespace")
	revision := c.Query("revision")

	if name == "" || revision == "" {
		response.Fail(c, "应用名称和版本不能为空")
		return
	}

	err := rollbackArgoApplication(name, namespace, revision)
	if err != nil {
		response.Fail(c, fmt.Sprintf("回滚失败: %s", err.Error()))
		return
	}

	response.Success(c, "回滚成功", nil)
}

// CreateApplication creates a new ArgoCD application
func (h *gitopsHandler) CreateApplication(c *gin.Context) {
	var app Application
	if err := c.ShouldBindJSON(&app); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	if app.Name == "" || app.RepoURL == "" {
		response.Fail(c, "应用名称和仓库地址不能为空")
		return
	}

	err := createArgoApplication(app)
	if err != nil {
		response.Fail(c, fmt.Sprintf("创建失败: %s", err.Error()))
		return
	}

	response.Success(c, "创建成功", nil)
}

// DeleteApplication deletes an ArgoCD application
func (h *gitopsHandler) DeleteApplication(c *gin.Context) {
	name := c.Query("name")
	namespace := c.Query("namespace")
	cascade := c.Query("cascade") == "true"

	if name == "" {
		response.Fail(c, "应用名称不能为空")
		return
	}

	err := deleteArgoApplication(name, namespace, cascade)
	if err != nil {
		response.Fail(c, fmt.Sprintf("删除失败: %s", err.Error()))
		return
	}

	response.Success(c, "删除成功", nil)
}

func listArgoApplications(namespace string) ([]Application, error) {
	// Check if argocd CLI is available
	_, err := exec.LookPath("argocd")
	if err != nil {
		return nil, fmt.Errorf("argocd not found")
	}

	args := []string{"app", "list", "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	cmd := exec.Command("argocd", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var apps []struct {
		Metadata struct {
			Name      string            `json:"name"`
			Namespace string            `json:"namespace"`
			Labels    map[string]string `json:"labels"`
		} `json:"metadata"`
		Spec struct {
			Project string `json:"project"`
			Source  struct {
				RepoURL        string `json:"repoURL"`
				Path           string `json:"path"`
				TargetRevision string `json:"targetRevision"`
			} `json:"source"`
		} `json:"spec"`
		Status struct {
			Health struct {
				Status string `json:"status"`
			} `json:"health"`
			Sync struct {
				Status string `json:"status"`
			} `json:"sync"`
			OperationState struct {
				Phase string `json:"phase"`
			} `json:"operationState"`
		} `json:"status"`
	}
	json.Unmarshal(output, &apps)

	var result []Application
	for _, app := range apps {
		result = append(result, Application{
			Name:       app.Metadata.Name,
			Namespace:  app.Metadata.Namespace,
			Project:    app.Spec.Project,
			RepoURL:    app.Spec.Source.RepoURL,
			Path:       app.Spec.Source.Path,
			TargetRev:  app.Spec.Source.TargetRevision,
			Status:     app.Status.OperationState.Phase,
			Health:     app.Status.Health.Status,
			SyncStatus: app.Status.Sync.Status,
			Labels:     app.Metadata.Labels,
		})
	}

	return result, nil
}

func getArgoApplication(name, namespace string) (*Application, error) {
	args := []string{"app", "get", name, "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	cmd := exec.Command("argocd", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var app struct {
		Metadata struct {
			Name      string            `json:"name"`
			Namespace string            `json:"namespace"`
			Labels    map[string]string `json:"labels"`
		} `json:"metadata"`
		Spec struct {
			Project string `json:"project"`
			Source  struct {
				RepoURL        string `json:"repoURL"`
				Path           string `json:"path"`
				TargetRevision string `json:"targetRevision"`
			} `json:"source"`
		} `json:"spec"`
		Status struct {
			Health struct {
				Status string `json:"status"`
			} `json:"health"`
			Sync struct {
				Status string `json:"status"`
			} `json:"sync"`
			OperationState struct {
				Phase string `json:"phase"`
			} `json:"operationState"`
		} `json:"status"`
	}
	json.Unmarshal(output, &app)

	return &Application{
		Name:       app.Metadata.Name,
		Namespace:  app.Metadata.Namespace,
		Project:    app.Spec.Project,
		RepoURL:    app.Spec.Source.RepoURL,
		Path:       app.Spec.Source.Path,
		TargetRev:  app.Spec.Source.TargetRevision,
		Status:     app.Status.OperationState.Phase,
		Health:     app.Status.Health.Status,
		SyncStatus: app.Status.Sync.Status,
		Labels:     app.Metadata.Labels,
	}, nil
}

func syncArgoApplication(req SyncRequest) error {
	args := []string{"app", "sync", req.Name}
	if req.Namespace != "" {
		args = append(args, "-n", req.Namespace)
	}
	if req.Revision != "" {
		args = append(args, "--revision", req.Revision)
	}
	if req.Force {
		args = append(args, "--force")
	}
	if req.DryRun {
		args = append(args, "--dry-run")
	}

	cmd := exec.Command("argocd", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), string(output))
	}

	return nil
}

func getArgoHistory(name, namespace string) ([]map[string]interface{}, error) {
	args := []string{"app", "history", name, "-o", "json"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	cmd := exec.Command("argocd", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var history []map[string]interface{}
	json.Unmarshal(output, &history)

	return history, nil
}

func rollbackArgoApplication(name, namespace, revision string) error {
	args := []string{"app", "rollback", name, revision}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}

	cmd := exec.Command("argocd", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), string(output))
	}

	return nil
}

func createArgoApplication(app Application) error {
	args := []string{"app", "create", app.Name,
		"--repo", app.RepoURL,
		"--path", app.Path,
		"--dest-server", "https://kubernetes.default.svc",
		"--dest-namespace", app.Namespace,
	}

	if app.TargetRev != "" {
		args = append(args, "--revision", app.TargetRev)
	}
	if app.Project != "" {
		args = append(args, "--project", app.Project)
	}

	cmd := exec.Command("argocd", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), string(output))
	}

	return nil
}

func deleteArgoApplication(name, namespace string, cascade bool) error {
	args := []string{"app", "delete", name, "-y"}
	if namespace != "" {
		args = append(args, "-n", namespace)
	}
	if cascade {
		args = append(args, "--cascade")
	}

	cmd := exec.Command("argocd", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %s", err.Error(), string(output))
	}

	return nil
}

func getSampleApplications() []Application {
	return []Application{
		{
			Name:       "guestbook",
			Namespace:  "argocd",
			Project:    "default",
			RepoURL:    "https://github.com/argoproj/argocd-example-apps.git",
			Path:       "guestbook",
			TargetRev:  "HEAD",
			Status:     "Succeeded",
			Health:     "Healthy",
			SyncStatus: "Synced",
		},
		{
			Name:       "helm-guestbook",
			Namespace:  "argocd",
			Project:    "default",
			RepoURL:    "https://github.com/argoproj/argocd-example-apps.git",
			Path:       "helm-guestbook",
			TargetRev:  "HEAD",
			Status:     "Running",
			Health:     "Progressing",
			SyncStatus: "OutOfSync",
		},
	}
}
