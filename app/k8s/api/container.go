package api

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"encoding/json"
	"gkube/app/k8s/model"
	"gkube/app/k8s/params"
	"gkube/pkg/asciinema"
	"gkube/pkg/audit"
	"gkube/pkg/database"
	"gkube/pkg/k8s/container"
	"gkube/pkg/s3"
	"time"

	"gkube/config"

	"gkube/pkg/response"

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

	// 写入操作记录
	key := strings.ReplaceAll(uuid.New().String(), "-", "")
	if err := database.DB.Model(&model.TerminalRecord{}).Create(&model.TerminalRecord{
		Key:         key,
		ClusterName: clusterName,
		Namespace:   namespace,
		PodName:     podName,
	}).Error; err != nil {
		conn.WriteMessage(websocket.CloseMessage, []byte("数据库错误"))
		return
	}

	// 接受第一次消息
	_, firstMessage, _ := conn.ReadMessage()
	var firstData map[string][]int
	err = json.Unmarshal(firstMessage, &firstData)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("接收窗口大小失败"))
		return
	}
	cols := firstData["resize"][0]
	rows := firstData["resize"][1]

	// 记录操作到es中
	startTime := time.Now()
	record := audit.NewEsRecord()
	asciinema.WriteHeader(key, cols, rows, startTime, record)

	// 执行Exec到Pod
	if err := container.ExecToPod(key, clusterName, namespace, podName, containerName, conn, record); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
	}
}

// 所有的操作记录
func RecordList(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")
	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)

	db := database.DB.Model(&model.TerminalRecord{})
	var count int64
	db.Count(&count)

	result := make([]model.TerminalRecord, 0)
	if err := db.Limit(limitInt).Offset(offsetInt).Find(&result); err != nil {
		response.Fail(c, "获取失败")
		return
	}
	response.Success(c, "获取成功", map[string]any{"count": count, "result": result})
}

// 获取记录的url
func RecordUrl(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		return
	}
	endpoint := config.Conf.S3.EndPoint
	bucket := config.Conf.S3.Bucket
	// 从es中读取数据
	record := audit.NewEsRecord()
	result := record.ReadData(key)

	var buffer bytes.Buffer
	for _, value := range result {
		history, _ := value["history"].(string)
		buffer.Write([]byte(history))
		buffer.WriteByte('\n')
	}
	// 上传到as3中-会覆盖更新
	s3.UploadFile(key, buffer.Bytes())

	url := fmt.Sprintf("http://%s/%s/%s", endpoint, bucket, key)
	response.Success(c, "执行成功", url)

}
