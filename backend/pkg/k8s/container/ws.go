package container

// 使用 gorilla/websocket 处理 WebSocket
import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

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

// safeConn 包装 websocket.Conn,用互斥锁序列化所有写操作(含心跳 ping),
// 规避 gorilla/websocket 禁止并发写的约束。
type safeConn struct {
	conn    *websocket.Conn
	writeMu *sync.Mutex
}

func (s *safeConn) writeMessage(messageType int, data []byte) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.conn.WriteMessage(messageType, data)
}

func ExecToPod(key, clusterName, namespace, podName, containerName string, conn *websocket.Conn, record *audit.EsRecord, initCols, initRows int) error {
	// 创建Clientset
	clientset, err := k8s.GetK8sClientByName(clusterName)
	if err != nil {
		return fmt.Errorf("创建Clientset失败")
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
		return fmt.Errorf("解析kubeconfig失败")
	}
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeConf))
	if err != nil {
		return fmt.Errorf("解析kubeconfig失败")
	}

	// 创建SPDY Executor
	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("创建SPDY执行器失败")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 确保在操作完成后取消上下文,联动心跳与读循环退出

	// 创建带缓冲的事件通道
	eventChan := make(chan AsciinemaEvent, 4096)
	// 启动事件处理协程(消费完成后关闭通道)
	go func() {
		handleAsciinemaEvents(ctx, eventChan)
		close(eventChan)
	}()

	// 共享写锁,序列化 TerminalWriter 与心跳的写操作
	writeMu := &sync.Mutex{}
	safe := &safeConn{conn: conn, writeMu: writeMu}

	// 心跳 goroutine:与 ctx 串联,StreamWithContext 返回后自动停止
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := safe.writeMessage(websocket.PingMessage, nil); err != nil {
					return
				}
			}
		}
	}()

	// 使用channel分离stdin和resize消息,避免多个goroutine竞争读取同一个WebSocket
	resizeChan := make(chan *remotecommand.TerminalSize, 16)
	stdinChan := make(chan []byte, 1024)
	// 第一个尺寸来自HandleWebSocket已读取的初始消息(加 select 防阻塞)
	select {
	case resizeChan <- &remotecommand.TerminalSize{Width: uint16(initCols), Height: uint16(initRows)}:
	case <-ctx.Done():
	}

	// 单个goroutine从WebSocket读取消息,根据格式分发到不同channel
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
					select {
					case resizeChan <- &remotecommand.TerminalSize{
						Width:  uint16(resizeData[0]),
						Height: uint16(resizeData[1]),
					}:
					case <-ctx.Done():
						return
					}
					continue
				}
			}
			// 不是resize消息,当作stdin输入(加 select 防阻塞)
			select {
			case stdinChan <- msg:
			case <-ctx.Done():
				return
			}
		}
	}()

	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  &TerminalReader{StdinChan: stdinChan},
		Stdout: &TerminalWriter{Safe: safe, Record: record, Key: key, Event: eventChan, Done: ctx.Done()},
		Stderr: &TerminalWriter{Safe: safe, Record: record, Key: key, Event: eventChan, Done: ctx.Done()},
		Tty:    true,
		TerminalSizeQueue: &TerminalSizeHandler{Record: record, Key: key, Event: eventChan, ResizeChan: resizeChan, Done: ctx.Done()},
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
	Safe   *safeConn
	Record *audit.EsRecord
	Key    string
	Event  chan<- AsciinemaEvent
	Done   <-chan struct{}
}

// 将数据输出到websocket。asciinema 数据由 handleAsciinemaEvents 统一写入,此处不再重复写。
func (w *TerminalWriter) Write(p []byte) (int, error) {
	if err := w.Safe.writeMessage(websocket.BinaryMessage, p); err != nil {
		return 0, err
	}
	// 把数据写到 chan 中(加 select 防阻塞)
	select {
	case w.Event <- AsciinemaEvent{
		Type:   "data",
		Data:   string(p),
		Record: w.Record,
		Time:   time.Now(),
	}:
	case <-w.Done:
	}
	return len(p), nil
}

// TerminalSizeHandler 处理终端尺寸调整
type TerminalSizeHandler struct {
	Record     *audit.EsRecord
	Key        string
	Event      chan<- AsciinemaEvent
	ResizeChan <-chan *remotecommand.TerminalSize
	Done       <-chan struct{}
}

// 调整终端尺寸
func (t *TerminalSizeHandler) Next() *remotecommand.TerminalSize {
	size, ok := <-t.ResizeChan
	if !ok {
		return nil
	}

	// 把数据推送到chan中(加 select 防阻塞)
	select {
	case t.Event <- AsciinemaEvent{
		Type:   "resize",
		Width:  size.Width,
		Height: size.Height,
		Record: t.Record,
		Time:   time.Now(),
	}:
	case <-t.Done:
	}

	return size
}

// 消费数据
func handleAsciinemaEvents(ctx context.Context, eventChan <-chan AsciinemaEvent) {
	for {
		select {
		case <-ctx.Done():
			return
		case event, ok := <-eventChan:
			if !ok {
				return
			}
			switch event.Type {
			case "data":
				asciinema.WriteData(event.Key, event.Time, event.Data, event.Record)
			case "resize":
				asciinema.WriteSize(event.Key, event.Time, event.Width, event.Height, event.Record)
			}
		}
	}
}
