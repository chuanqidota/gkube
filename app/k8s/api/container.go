package api

import (
	"net/http"

	"gkube/app/k8s/container"
	"gkube/app/k8s/params"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// WebSocket处理函数
func HandleWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin:     func(r *http.Request) bool { return true }, // 生产环境需严格限制
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		
		return
	}
	defer conn.Close()

	// 获取Pod参数
	var reqQueryParams params.ContainerQueryParams
	if err = c.ShouldBindQuery(&reqQueryParams); err != nil {
		conn.WriteMessage(websocket.CloseMessage, []byte("参数缺失:必须提供pod/namespace/container参数"))
		return
	}
	clusterName := reqQueryParams.ClusterName
	namespace := reqQueryParams.Namespace
	podName := reqQueryParams.PodName
	containerName := reqQueryParams.Container

	// 执行Exec到Pod
	if err := container.ExecToPod(clusterName, namespace, podName, containerName, conn); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
	}
}
