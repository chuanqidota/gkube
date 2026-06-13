package api

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gkube/pkg/response"
)

type approvalHandler struct{}

var Approval = new(approvalHandler)

const approvalFile = "config/approvals.json"

type ApprovalRequest struct {
	ID          string            `json:"id"`
	Type        string            `json:"type"` // deploy, scale, delete, config
	Resource    string            `json:"resource"`
	Namespace   string            `json:"namespace"`
	Cluster     string            `json:"cluster"`
	RequestedBy string            `json:"requestedBy"`
	RequestedAt time.Time         `json:"requestedAt"`
	Status      string            `json:"status"` // pending, approved, rejected
	ReviewedBy  string            `json:"reviewedBy,omitempty"`
	ReviewedAt  *time.Time        `json:"reviewedAt,omitempty"`
	Reason      string            `json:"reason,omitempty"`
	Details     map[string]string `json:"details,omitempty"`
}

type ApprovalStore struct {
	Requests []ApprovalRequest `json:"requests"`
}

func loadApprovals() *ApprovalStore {
	store := &ApprovalStore{Requests: []ApprovalRequest{}}
	data, err := os.ReadFile(approvalFile)
	if err == nil {
		json.Unmarshal(data, store)
	}
	return store
}

func saveApprovals(store *ApprovalStore) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(approvalFile, data, 0644)
}

// ListApprovals lists all approval requests
func (h *approvalHandler) ListApprovals(c *gin.Context) {
	status := c.Query("status")

	store := loadApprovals()
	var result []ApprovalRequest

	for _, req := range store.Requests {
		if status == "" || req.Status == status {
			result = append(result, req)
		}
	}

	response.Success(c, "获取成功", result)
}

// GetApproval gets a specific approval request
func (h *approvalHandler) GetApproval(c *gin.Context) {
	id := c.Query("id")

	store := loadApprovals()
	for _, req := range store.Requests {
		if req.ID == id {
			response.Success(c, "获取成功", req)
			return
		}
	}

	response.Fail(c, "审批请求不存在")
}

// CreateApproval creates a new approval request
func (h *approvalHandler) CreateApproval(c *gin.Context) {
	var req ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	if req.Type == "" || req.Resource == "" {
		response.Fail(c, "类型和资源不能为空")
		return
	}

	req.ID = fmt.Sprintf("apr-%d", time.Now().UnixNano())
	req.RequestedAt = time.Now()
	req.Status = "pending"

	store := loadApprovals()
	store.Requests = append(store.Requests, req)

	if err := saveApprovals(store); err != nil {
		response.Fail(c, "保存审批请求失败: "+err.Error())
		return
	}

	response.Success(c, "审批请求已创建", req)
}

// ApproveRequest approves an approval request
func (h *approvalHandler) ApproveRequest(c *gin.Context) {
	var body struct {
		ID         string `json:"id"`
		ReviewedBy string `json:"reviewedBy"`
		Reason     string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	store := loadApprovals()
	for i, req := range store.Requests {
		if req.ID == body.ID {
			if req.Status != "pending" {
				response.Fail(c, "该请求已被处理")
				return
			}
			now := time.Now()
			store.Requests[i].Status = "approved"
			store.Requests[i].ReviewedBy = body.ReviewedBy
			store.Requests[i].ReviewedAt = &now
			store.Requests[i].Reason = body.Reason

			if err := saveApprovals(store); err != nil {
				response.Fail(c, "保存失败: "+err.Error())
				return
			}

			response.Success(c, "已批准", store.Requests[i])
			return
		}
	}

	response.Fail(c, "审批请求不存在")
}

// RejectRequest rejects an approval request
func (h *approvalHandler) RejectRequest(c *gin.Context) {
	var body struct {
		ID         string `json:"id"`
		ReviewedBy string `json:"reviewedBy"`
		Reason     string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	store := loadApprovals()
	for i, req := range store.Requests {
		if req.ID == body.ID {
			if req.Status != "pending" {
				response.Fail(c, "该请求已被处理")
				return
			}
			now := time.Now()
			store.Requests[i].Status = "rejected"
			store.Requests[i].ReviewedBy = body.ReviewedBy
			store.Requests[i].ReviewedAt = &now
			store.Requests[i].Reason = body.Reason

			if err := saveApprovals(store); err != nil {
				response.Fail(c, "保存失败: "+err.Error())
				return
			}

			response.Success(c, "已拒绝", store.Requests[i])
			return
		}
	}

	response.Fail(c, "审批请求不存在")
}

// DeleteApproval deletes an approval request
func (h *approvalHandler) DeleteApproval(c *gin.Context) {
	id := c.Query("id")

	store := loadApprovals()
	for i, req := range store.Requests {
		if req.ID == id {
			store.Requests = append(store.Requests[:i], store.Requests[i+1:]...)
			if err := saveApprovals(store); err != nil {
				response.Fail(c, "删除失败: "+err.Error())
				return
			}
			response.Success(c, "删除成功", nil)
			return
		}
	}

	response.Fail(c, "审批请求不存在")
}

// GetApprovalStats gets approval statistics
func (h *approvalHandler) GetApprovalStats(c *gin.Context) {
	store := loadApprovals()

	stats := map[string]int{
		"total":    len(store.Requests),
		"pending":  0,
		"approved": 0,
		"rejected": 0,
	}

	for _, req := range store.Requests {
		stats[req.Status]++
	}

	response.Success(c, "获取成功", stats)
}
