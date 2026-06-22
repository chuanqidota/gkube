package container

// 使用 gorilla/websocket 处理 WebSocket
import (
	"encoding/json"
	"fmt"
	"time"

	"context"

	"gkube/pkg/asciinema"
	"gkube/pkg/k8s"

	"gkube/pkg/audit"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type AsciinemaEvent struct {
	Type   string          `json:"type" label:"type"`
	Key    string          `json:"key" label:"key"`
	Time   time.Time       `json:"time" label:"time"`
	Record *audit.EsRecord `json:"record" label:"记录"`
	Data   string          `json:"data" label:"数据"`
	Width  uint16          `json:"width" label:"宽"`
	Height uint16          `json:"height" label:"高"`
}

func ExecToPod(key, clusterName, namespace, podName, containerName string, conn *websocket.Conn, record *audit.EsRecord) error {
	// 创建Clientset
	clientset, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		return fmt.Errorf("创建Clientset失败: %v", err)
	}

	// 构造Exec请求
	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(namespace).
		Name(podName).
		SubResource("exec").
		VersionedParams(&corev1.PodExecOptions{
			Container: containerName,
			Command:   []string{"/bin/bash"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	// 获取配置
	conf, err := k8s.GetK8sConf(clusterName)
	if err != nil {
		return err
	}
	confByte, err := json.Marshal(conf)
	if err != nil {
		return err
	}
	var config rest.Config
	err = json.Unmarshal(confByte, &config)
	if err != nil {
		return err
	}

	// 创建SPDY Executor
	executor, err := remotecommand.NewSPDYExecutor(&config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("创建SPDY执行器失败: %v", err)
	}

	// 创建带缓冲的事件通道（缓冲区大小可根据需要调整）
	eventChan := make(chan AsciinemaEvent, 1024*1024*1024)
	defer close(eventChan)
	// 启动事件处理协程
	go handleAsciinemaEvents(eventChan)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 确保在操作完成后取消上下文

	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             &TerminalReader{Conn: conn},
		Stdout:            &TerminalWriter{Conn: conn, Record: record, Key: key, Event: eventChan},
		Stderr:            &TerminalWriter{Conn: conn, Record: record, Key: key, Event: eventChan},
		Tty:               true,
		TerminalSizeQueue: &TerminalSizeHandler{Conn: conn, Record: record, Key: key, Event: eventChan},
	})
	return err
}

// TerminalReader 从WebSocket读取输入
type TerminalReader struct {
	Conn *websocket.Conn
}

// 从websocket中读取数据
func (r *TerminalReader) Read(p []byte) (int, error) {
	_, msg, err := r.Conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	return copy(p, msg), nil
}

// TerminalWriter 将输出写入WebSocket
type TerminalWriter struct {
	Conn   *websocket.Conn
	Record *audit.EsRecord
	Key    string
	Event  chan AsciinemaEvent
}

// 将数据输出到websocket
func (w *TerminalWriter) Write(p []byte) (int, error) {
	err := w.Conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	// 把数据写到chan中
	w.Event <- AsciinemaEvent{
		Type:   "data",
		Data:   string(p),
		Record: w.Record,
		Time:   time.Now(),
	}
	asciinema.WriteData(w.Key, time.Now(), string(p), w.Record)
	return len(p), nil
}

// TerminalSizeHandler 处理终端尺寸调整
type TerminalSizeHandler struct {
	Conn   *websocket.Conn
	Record *audit.EsRecord
	Key    string
	Event  chan AsciinemaEvent
}

// 调整终端尺寸
func (t *TerminalSizeHandler) Next() *remotecommand.TerminalSize {
	var data map[string][]int
	if err := t.Conn.ReadJSON(&data); err != nil {
		fmt.Printf("读取终端尺寸失败: %v", err)
		return nil
	}

	resizeData, ok := data["resize"]
	if !ok || len(resizeData) < 2 {
		fmt.Println("无效的尺寸数据")
		return nil
	}

	width := uint16(resizeData[0])
	height := uint16(resizeData[1])

	// 把数据推送到chan中
	t.Event <- AsciinemaEvent{
		Type:   "resize",
		Width:  width,
		Height: height,
		Record: t.Record,
		Time:   time.Now(),
	}

	return &remotecommand.TerminalSize{
		Width:  width,
		Height: height,
	}
}

// 消费数据
func handleAsciinemaEvents(eventChan <-chan AsciinemaEvent) {
	for event := range eventChan {
		switch event.Type {
		case "data":
			asciinema.WriteData(event.Key, event.Time, event.Data, event.Record)
		case "resize":
			asciinema.WriteSize(event.Key, event.Time, event.Width, event.Height, event.Record)
		}
	}
}
