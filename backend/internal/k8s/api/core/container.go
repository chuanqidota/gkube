package core

import (
	"bufio"
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"

	"encoding/json"
	"gkube/internal/k8s/model"
	"gkube/internal/k8s/params"
	"gkube/pkg/asciinema"
	"gkube/pkg/audit"
	"gkube/pkg/database"
	"gkube/pkg/k8s/container"
	"gkube/pkg/logger"
	"gkube/pkg/middleware"
	"gkube/pkg/response"
	"gkube/pkg/s3"

	"gkube/config"

	"gkube/pkg/k8s"
)

// WebSocket处理函数
func HandleWebSocket(c *gin.Context) {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// 校验 Origin 在配置的白名单内,空配置仅允许同源(无 Origin)
		CheckOrigin: func(r *http.Request) bool {
			return middleware.IsOriginAllowed(r.Header.Get("Origin"))
		},
	}
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 设置pong handler，保持连接活跃
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	})

	// 获取Pod参数
	var reqQueryParams params.ContainerQueryParams
	if err = c.ShouldBindQuery(&reqQueryParams); err != nil {
		_ = conn.WriteMessage(websocket.CloseMessage, []byte("参数缺失:必须提供pod/namespace/container参数"))
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
		logger.Error(err.Error())
		_ = conn.WriteMessage(websocket.CloseMessage, []byte("数据库错误"))
		return
	}

	// 接受第一次消息（窗口大小）
	_, firstMessage, err := conn.ReadMessage()
	if err != nil {
		return
	}
	var firstData map[string][]int
	err = json.Unmarshal(firstMessage, &firstData)
	if err != nil {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("接收窗口大小失败"))
		return
	}
	resizeData, ok := firstData["resize"]
	if !ok || len(resizeData) < 2 {
		_ = conn.WriteMessage(websocket.TextMessage, []byte("无效的尺寸数据"))
		return
	}
	cols := resizeData[0]
	rows := resizeData[1]

	// 记录操作到es中
	startTime := time.Now()
	record := audit.NewEsRecord()
	asciinema.WriteHeader(key, cols, rows, startTime, record)

	// 执行Exec到Pod，传入初始终端尺寸(心跳与并发写由 ExecToPod 内部管理)
	if err := container.ExecToPod(key, clusterName, namespace, podName, containerName, conn, record, cols, rows); err != nil {
		logger.Error(err.Error())
		_ = conn.WriteMessage(websocket.TextMessage, []byte("终端执行失败"))
	}
}

// 所有的操作记录
func RecordList(c *gin.Context) {
	limit := c.DefaultQuery("limit", "10")
	offset := c.DefaultQuery("offset", "0")
	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)
	if limitInt <= 0 {
		limitInt = 10
	}
	if limitInt > 500 {
		limitInt = 500
	}
	if offsetInt < 0 {
		offsetInt = 0
	}

	db := database.DB.Model(&model.TerminalRecord{})
	var count int64
	if err := db.Session(&gorm.Session{}).Count(&count).Error; err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "获取记录总数失败")
		return
	}

	result := make([]model.TerminalRecord, 0)
	if err := db.Limit(limitInt).Offset(offsetInt).Find(&result).Error; err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusInternalServerError, "获取失败")
		return
	}
	response.Success(c, "获取成功", map[string]any{"count": count, "result": result})
}

// 获取记录的url
func RecordUrl(c *gin.Context) {
	key := c.Query("key")
	if key == "" {
		response.Fail(c, "key参数不能为空")
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
	// 上传到s3中-会覆盖更新
	s3.UploadFile(key, buffer.Bytes())

	url := fmt.Sprintf("http://%s/%s/%s", endpoint, bucket, key)
	response.Success(c, "执行成功", url)
}

// 获取日志且包含日志行数
func PodContainerLog(c *gin.Context) {
	var query params.ContainerLogQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}

	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}
	log, err := container.GetPodContainerLog(client, query.Namespace, query.PodName, query.Container, query.TailLines)
	if err != nil {
		logger.Error(err.Error())
		response.FailWithStatus(c, http.StatusBadGateway, "获取日志失败")
		return
	}
	response.Success(c, "获取成功", log)
}

// 通过SSE获取日志信息
func StreamPodContainerLogs(c *gin.Context) {
	var query params.ContainerLogQueryParams
	if err := c.ShouldBindQuery(&query); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	client, err := k8s.GetK8sClientByName(query.ClusterName)
	if err != nil {
		logger.Error(err.Error())
		response.Fail(c, "获取k8s客户端失败")
		return
	}

	// 使用请求的context，客户端断开时自动取消K8s流
	ctx := c.Request.Context()
	stream, err := container.GetPodContainerLogStream(ctx, client, query.Namespace, query.PodName, query.Container, query.TailLines)
	if err != nil {
		logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "日志流创建失败"})
		return
	}
	defer stream.Close()

	// 设置 SSE 响应头
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// 创建带缓冲的读取器
	reader := bufio.NewReader(stream)

	// 用channel传递读取结果，避免阻塞select
	type readResult struct {
		line []byte
		err  error
	}
	ch := make(chan readResult, 1)

	// 在goroutine中读取日志流
	go func() {
		defer close(ch)
		for {
			line, err := reader.ReadBytes('\n')
			ch <- readResult{line: line, err: err}
			if err != nil {
				return
			}
		}
	}()

	// 主循环：同时检查客户端断开和日志数据
	for {
		select {
		case <-ctx.Done():
			return
		case result, ok := <-ch:
			if !ok {
				return
			}
			if result.err != nil {
				return
			}

			c.SSEvent("message", string(result.line))
			c.Writer.Flush()
		}
	}
}
