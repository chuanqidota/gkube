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
	"k8s.io/client-go/tools/clientcmd"
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

func ExecToPod(key, clusterName, namespace, podName, containerName string, conn *websocket.Conn, record *audit.EsRecord, initCols, initRows int) error {
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
	kubeConf, err := k8s.GetK8sConf(clusterName)
	if err != nil {
		return err
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConf))
	if err != nil {
		return fmt.Errorf("解析kubeconfig失败: %v", err)
	}

	// 创建SPDY Executor
	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
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

	// 使用channel分离stdin和resize消息，避免多个goroutine竞争读取同一个WebSocket
	resizeChan := make(chan *remotecommand.TerminalSize, 16)
	stdinChan := make(chan []byte, 1024)
	// 第一个尺寸来自HandleWebSocket已读取的初始消息
	resizeChan <- &remotecommand.TerminalSize{Width: uint16(initCols), Height: uint16(initRows)}

	// 单个goroutine从WebSocket读取消息，根据格式分发到不同channel
	go func() {
		defer close(resizeChan)
		defer close(stdinChan)
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			// 尝试解析为resize消息
			var data map[string][]int
			if json.Unmarshal(msg, &data) == nil {
				if resizeData, ok := data["resize"]; ok && len(resizeData) >= 2 {
					resizeChan <- &remotecommand.TerminalSize{
						Width:  uint16(resizeData[0]),
						Height: uint16(resizeData[1]),
					}
					continue
				}
			}
			// 不是resize消息，当作stdin输入
			stdinChan <- msg
		}
	}()

	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             &TerminalReader{StdinChan: stdinChan},
		Stdout:            &TerminalWriter{Conn: conn, Record: record, Key: key, Event: eventChan},
		Stderr:            &TerminalWriter{Conn: conn, Record: record, Key: key, Event: eventChan},
		Tty:               true,
		TerminalSizeQueue: &TerminalSizeHandler{Record: record, Key: key, Event: eventChan, ResizeChan: resizeChan},
	})
	return err
}

// TerminalReader 从stdinChan读取输入
type TerminalReader struct {
	StdinChan <-chan []byte
}

// 从stdinChan中读取数据
func (r *TerminalReader) Read(p []byte) (int, error) {
	msg, ok := <-r.StdinChan
	if !ok {
		return 0, fmt.Errorf("stdin channel closed")
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
	Record     *audit.EsRecord
	Key        string
	Event      chan AsciinemaEvent
	ResizeChan <-chan *remotecommand.TerminalSize
}

// 调整终端尺寸
func (t *TerminalSizeHandler) Next() *remotecommand.TerminalSize {
	size, ok := <-t.ResizeChan
	if !ok {
		return nil
	}

	// 把数据推送到chan中
	t.Event <- AsciinemaEvent{
		Type:   "resize",
		Width:  size.Width,
		Height: size.Height,
		Record: t.Record,
		Time:   time.Now(),
	}

	return size
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
