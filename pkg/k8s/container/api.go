package container

// 使用 gorilla/websocket 处理 WebSocket
import (
	"encoding/json"
	"fmt"

	"context"

	"gkube/pkg/k8s"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

func ExecToPod(clusterName, namespace, podName, containerName string, conn *websocket.Conn) error {
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // 确保在操作完成后取消上下文

	err = executor.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             &TerminalReader{conn: conn},
		Stdout:            &TerminalWriter{conn: conn},
		Stderr:            &TerminalWriter{conn: conn},
		Tty:               true,
		TerminalSizeQueue: &TerminalSizeHandler{conn: conn},
	})
	return err
}

// TerminalReader 从WebSocket读取输入
type TerminalReader struct {
	conn *websocket.Conn
}

func (r *TerminalReader) Read(p []byte) (int, error) {
	_, msg, err := r.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	return copy(p, msg), nil
}

// TerminalWriter 将输出写入WebSocket
type TerminalWriter struct {
	conn *websocket.Conn
}

func (w *TerminalWriter) Write(p []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.BinaryMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// TerminalSizeHandler 处理终端尺寸调整
type TerminalSizeHandler struct {
	conn *websocket.Conn
}

func (t *TerminalSizeHandler) Next() *remotecommand.TerminalSize {
	var size remotecommand.TerminalSize
	if err := t.conn.ReadJSON(&size); err != nil {
		fmt.Printf("读取终端尺寸失败: %v", err)
		return nil
	}
	return &size
}
