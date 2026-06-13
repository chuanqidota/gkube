package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"gkube/app/k8s/model"
	"gkube/pkg/database"
	"gkube/pkg/response"
)

type approvalHandler struct{}

var Approval = new(approvalHandler)

type CreateApprovalRequest struct {
	Type        string            `json:"type"`
	Resource    string            `json:"resource"`
	Namespace   string            `json:"namespace"`
	Cluster     string            `json:"cluster"`
	RequestedBy uint              `json:"requestedBy"`
	Details     map[string]string `json:"details"`
}

// ListApprovals lists all approval requests
func (h *approvalHandler) ListApprovals(c *gin.Context) {
	status := c.Query("status")

	var approvals []model.ApprovalRequest
	query := database.DB.Order("created_at DESC")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Find(&approvals).Error; err != nil {
		response.Fail(c, "获取审批列表失败: "+err.Error())
		return
	}

	response.Success(c, "获取成功", approvals)
}

// GetApproval gets a specific approval request
func (h *approvalHandler) GetApproval(c *gin.Context) {
	id := c.Query("id")

	var approval model.ApprovalRequest
	if err := database.DB.Where("request_id = ?", id).First(&approval).Error; err != nil {
		response.Fail(c, "审批请求不存在")
		return
	}

	response.Success(c, "获取成功", approval)
}

// CreateApproval creates a new approval request
func (h *approvalHandler) CreateApproval(c *gin.Context) {
	var req CreateApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	if req.Type == "" || req.Resource == "" {
		response.Fail(c, "类型和资源不能为空")
		return
	}

	approval := model.ApprovalRequest{
		RequestID:   fmt.Sprintf("apr-%d", time.Now().UnixNano()),
		Type:        req.Type,
		Resource:    req.Resource,
		Namespace:   req.Namespace,
		Cluster:     req.Cluster,
		RequestedBy: req.RequestedBy,
		RequestedAt: time.Now(),
		Status:      "pending",
	}

	if err := database.DB.Create(&approval).Error; err != nil {
		response.Fail(c, "创建审批请求失败: "+err.Error())
		return
	}

	response.Success(c, "审批请求已创建", approval)
}

// ApproveRequest approves an approval request
func (h *approvalHandler) ApproveRequest(c *gin.Context) {
	var body struct {
		ID         string `json:"id"`
		ReviewedBy uint   `json:"reviewedBy"`
		Reason     string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	var approval model.ApprovalRequest
	if err := database.DB.Where("request_id = ?", body.ID).First(&approval).Error; err != nil {
		response.Fail(c, "审批请求不存在")
		return
	}

	if approval.Status != "pending" {
		response.Fail(c, "该请求已被处理")
		return
	}

	now := time.Now()
	approval.Status = "approved"
	approval.ReviewedBy = &body.ReviewedBy
	approval.ReviewedAt = &now
	approval.Reason = body.Reason

	if err := database.DB.Save(&approval).Error; err != nil {
		response.Fail(c, "审批失败: "+err.Error())
		return
	}

	response.Success(c, "已批准", approval)
}

// RejectRequest rejects an approval request
func (h *approvalHandler) RejectRequest(c *gin.Context) {
	var body struct {
		ID         string `json:"id"`
		ReviewedBy uint   `json:"reviewedBy"`
		Reason     string `json:"reason"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "参数错误: "+err.Error())
		return
	}

	var approval model.ApprovalRequest
	if err := database.DB.Where("request_id = ?", body.ID).First(&approval).Error; err != nil {
		response.Fail(c, "审批请求不存在")
		return
	}

	if approval.Status != "pending" {
		response.Fail(c, "该请求已被处理")
		return
	}

	now := time.Now()
	approval.Status = "rejected"
	approval.ReviewedBy = &body.ReviewedBy
	approval.ReviewedAt = &now
	approval.Reason = body.Reason

	if err := database.DB.Save(&approval).Error; err != nil {
		response.Fail(c, "拒绝失败: "+err.Error())
		return
	}

	response.Success(c, "已拒绝", approval)
}

// DeleteApproval deletes an approval request
func (h *approvalHandler) DeleteApproval(c *gin.Context) {
	id := c.Query("id")

	if err := database.DB.Where("request_id = ?", id).Delete(&model.ApprovalRequest{}).Error; err != nil {
		response.Fail(c, "删除失败: "+err.Error())
		return
	}

	response.Success(c, "删除成功", nil)
}

// GetApprovalStats gets approval statistics
func (h *approvalHandler) GetApprovalStats(c *gin.Context) {
	var total, pending, approved, rejected int64

	database.DB.Model(&model.ApprovalRequest{}).Count(&total)
	database.DB.Model(&model.ApprovalRequest{}).Where("status = ?", "pending").Count(&pending)
	database.DB.Model(&model.ApprovalRequest{}).Where("status = ?", "approved").Count(&approved)
	database.DB.Model(&model.ApprovalRequest{}).Where("status = ?", "rejected").Count(&rejected)

	stats := map[string]int64{
		"total":    total,
		"pending":  pending,
		"approved": approved,
		"rejected": rejected,
	}

	response.Success(c, "获取成功", stats)
}
