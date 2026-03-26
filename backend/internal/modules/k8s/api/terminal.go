package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"k8s.io/client-go/tools/remotecommand"
)

var podTerminalUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return true
		}
		u, err := url.Parse(origin)
		if err != nil {
			return false
		}
		if strings.EqualFold(u.Host, r.Host) {
			return true
		}
		if strings.HasPrefix(u.Host, "localhost:") || strings.HasPrefix(u.Host, "127.0.0.1:") {
			return true
		}
		return false
	},
}

// K8sMessage WebSocket消息结构（与前端约定）
type K8sMessage struct {
	Operation string      `json:"operation"`
	Data      interface{} `json:"data"`
	Cols      int         `json:"cols,omitempty"`
	Rows      int         `json:"rows,omitempty"`
}

// K8sWebSocketStream K8s WebSocket流处理
type K8sWebSocketStream struct {
	sync.RWMutex
	Conn     *websocket.Conn
	executor remotecommand.Executor
	Ctx      context.Context
	cancel   context.CancelFunc
	closed   bool
	reader   *io.PipeReader
	writer   *io.PipeWriter
}

// terminalConn 实现io.ReadWriter接口用于K8s executor
type terminalConn struct {
	stream *K8sWebSocketStream
}

func (tc *terminalConn) Read(p []byte) (n int, err error) {
	return tc.stream.reader.Read(p)
}

func (tc *terminalConn) Write(p []byte) (n int, err error) {
	return tc.stream.WriteToWebSocket(p)
}

// Close 关闭WebSocket流
func (kws *K8sWebSocketStream) Close() error {
	kws.Lock()
	defer kws.Unlock()

	if kws.closed {
		return nil
	}

	kws.closed = true
	if kws.cancel != nil {
		kws.cancel()
	}
	return nil
}

// IsClosed 检查是否已关闭
func (kws *K8sWebSocketStream) IsClosed() bool {
	kws.RLock()
	defer kws.RUnlock()
	return kws.closed
}

// WriteToWebSocket 写入数据到WebSocket
func (kws *K8sWebSocketStream) WriteToWebSocket(p []byte) (n int, err error) {
	if kws.IsClosed() {
		return 0, io.EOF
	}

	message := K8sMessage{
		Operation: "stdout",
		Data:      string(p),
	}

	if err = kws.Conn.WriteJSON(message); err != nil {
		go kws.Close()
		return 0, err
	}

	return len(p), nil
}

// ReadFromWebSocket 从WebSocket读取数据
func (kws *K8sWebSocketStream) ReadFromWebSocket() {
	defer func() {
		kws.Close()
		if kws.writer != nil {
			kws.writer.Close()
		}
	}()

	for {
		if kws.IsClosed() {
			return
		}

		var message K8sMessage
		err := kws.Conn.ReadJSON(&message)
		if err != nil {
			return
		}

		switch message.Operation {
		case "stdin":
			if kws.writer != nil && !kws.IsClosed() {
				var data string
				if str, ok := message.Data.(string); ok {
					data = str
				} else {
					continue
				}

				_, err := kws.writer.Write([]byte(data))
				if err != nil {
					return
				}
			}
		case "resize":
			// 终端大小调整已在初始连接时处理，动态resize暂时忽略
			continue
		default:
			continue
		}
	}
}

type terminalSizeQueue struct {
	ch <-chan remotecommand.TerminalSize
}

func (q *terminalSizeQueue) Next() *remotecommand.TerminalSize {
	s, ok := <-q.ch
	if !ok {
		return nil
	}
	return &s
}

// DetectPodShell godoc
// @Summary 检测Pod容器可用Shell
// @Description 检测指定Pod容器中可用的shell类型
// @Tags K8s资源管理
// @Accept json
// @Produce json
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param pod query string true "Pod名称"
// @Param container query string false "容器名称（可选，默认第一个容器）"
// @Success 200 {object} Response "成功"
// @Security BearerAuth
// @Router /k8s/pod/detect-shell [get]
func DetectPodShell(c *gin.Context) {
	namespace := c.Query("namespace")
	podName := c.Query("pod")
	container := c.Query("container")

	if namespace == "" || podName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数不完整"})
		return
	}

	clusterID, err := resolveClusterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	result, err := svc.DetectContainerShell(clusterID, namespace, podName, container)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("检测shell失败: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": result})
}

// PodTerminal godoc
// @Summary Pod 终端（WebSocket）
// @Description WebSocket 进入 Pod 命令行（shell=bash/sh，可选 container，支持 resize）
// @Tags K8s资源管理
// @Param clusterId query int false "集群ID（可选，未传则使用默认集群）"
// @Param namespace query string true "命名空间"
// @Param pod query string true "Pod 名称"
// @Param container query string false "容器名称（可选）"
// @Param shell query string false "shell（bash/sh，默认 sh）"
// @Param cols query int false "终端列数（可选）"
// @Param rows query int false "终端行数（可选）"
// @Security BearerAuth
// @Router /k8s/pod/terminal [get]
func PodTerminal(c *gin.Context) {
	namespace := c.Query("namespace")
	podName := c.Query("pod")
	container := c.Query("container")
	shell := c.Query("shell")
	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "120"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "30"))

	if namespace == "" || podName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "参数不完整"})
		return
	}

	clusterID, err := resolveClusterID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	svc, err := getK8sService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
		return
	}

	// 若未指定容器，自动选择 Pod 的第一个容器
	if container == "" {
		podDetail, err := svc.GetPodObject(clusterID, namespace, podName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": fmt.Sprintf("获取Pod信息失败: %v", err)})
			return
		}
		if len(podDetail.Spec.Containers) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"message": "Pod 中没有容器"})
			return
		}
		container = podDetail.Spec.Containers[0].Name
	}

	// 升级HTTP连接为WebSocket
	conn, err := podTerminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// 获取executor，支持bash到sh的自动降级
	executor, actualShell, err := svc.CreatePodExecutor(clusterID, namespace, podName, container, shell)
	if err != nil {
		_ = conn.WriteJSON(K8sMessage{Operation: "error", Data: fmt.Sprintf("创建executor失败: %v", err)})
		return
	}

	// 如果降级了shell，通知前端
	if shell != "" && shell != actualShell {
		_ = conn.WriteJSON(K8sMessage{Operation: "stdout", Data: fmt.Sprintf("注意：容器不支持%s，已自动切换到%s\r\n\r\n", shell, actualShell)})
	}

	// 创建长期运行的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建管道
	reader, writer := io.Pipe()

	// 创建流对象
	stream := &K8sWebSocketStream{
		Conn:     conn,
		executor: executor,
		Ctx:      ctx,
		cancel:   cancel,
		reader:   reader,
		writer:   writer,
	}

	// 创建终端连接
	termConn := &terminalConn{stream: stream}

	// 创建resize queue
	resizeCh := make(chan remotecommand.TerminalSize, 8)
	if cols > 0 && rows > 0 {
		resizeCh <- remotecommand.TerminalSize{Width: uint16(cols), Height: uint16(rows)}
	}
	close(resizeCh) // 初始大小设置后关闭
	queue := &terminalSizeQueue{ch: resizeCh}

	// 启动WebSocket读取goroutine
	go stream.ReadFromWebSocket()

	// 启动K8s executor goroutine
	go func() {
		defer func() {
			cancel()
			if reader != nil {
				reader.Close()
			}
		}()

		executor.StreamWithContext(ctx, remotecommand.StreamOptions{
			Stdin:             termConn,
			Stdout:            termConn,
			Stderr:            termConn,
			Tty:               true,
			TerminalSizeQueue: queue,
		})
	}()

	// 等待连接关闭
	<-ctx.Done()
}
