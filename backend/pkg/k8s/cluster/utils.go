package cluster

import (
	"k8s.io/apimachinery/pkg/api/resource"
	"fmt"
)

// formatMemory 格式化内存信息（将字节转换为 GiB）
func formatMemory(res resource.Quantity) string {
	memoryGiB := res.Value() / (1024 * 1024 * 1024)
	return fmt.Sprintf("%v GiB (%s)", memoryGiB, res.String())
}