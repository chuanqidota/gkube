package k8s

import (
	"gkube/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetPaginationParams extracts limit and continue token from query params (GET).
// For POST list endpoints, handlers should read limit/continue from the bound params struct directly.
func GetPaginationParams(c *gin.Context) (int64, string) {
	limitStr := c.DefaultQuery("limit", "0")
	continueToken := c.DefaultQuery("continue", "")
	limit, _ := strconv.ParseInt(limitStr, 10, 64)
	return limit, continueToken
}

// BuildPaginatedData builds a paginated response from K8s list metadata
func BuildPaginatedData(items interface{}, continueToken string, remainingItemCount int64, limit int64) response.PaginatedResponse {
	return response.PaginatedResponse{
		Items:     items,
		Continue:  continueToken,
		Remaining: remainingItemCount,
		PageSize:  int(limit),
		HasMore:   continueToken != "",
	}
}
