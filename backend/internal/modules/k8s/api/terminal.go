package api

import (
	"context"
	"encoding/json"
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

type wsBinaryWriter struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsBinaryWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if err := w.conn.WriteMessage(websocket.BinaryMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
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
	pod := c.Query("pod")
	container := c.Query("container")
	shell := c.Query("shell")
	cols, _ := strconv.Atoi(c.DefaultQuery("cols", "120"))
	rows, _ := strconv.Atoi(c.DefaultQuery("rows", "30"))

	if namespace == "" || pod == "" {
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
		podDetail, err := svc.GetPodObject(clusterID, namespace, pod)
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

	conn, err := podTerminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	stdinR, stdinW := io.Pipe()
	defer stdinR.Close()

	resizeCh := make(chan remotecommand.TerminalSize, 8)
	if cols > 0 && rows > 0 {
		resizeCh <- remotecommand.TerminalSize{Width: uint16(cols), Height: uint16(rows)}
	}

	go func() {
		defer func() {
			stdinW.Close()
			close(resizeCh)
			cancel()
		}()
		for {
			mt, data, err := conn.ReadMessage()
			if err != nil {
				return
			}

			if mt == websocket.TextMessage {
				var msg struct {
					Type string `json:"type"`
					Cols int    `json:"cols"`
					Rows int    `json:"rows"`
					Data string `json:"data"`
				}
				if json.Unmarshal(data, &msg) == nil && msg.Type != "" {
					if msg.Type == "resize" && msg.Cols > 0 && msg.Rows > 0 {
						resizeCh <- remotecommand.TerminalSize{Width: uint16(msg.Cols), Height: uint16(msg.Rows)}
						continue
					}
					if msg.Type == "stdin" && msg.Data != "" {
						_, _ = stdinW.Write([]byte(msg.Data))
						continue
					}
				}
				_, _ = stdinW.Write(data)
				continue
			}

			if mt == websocket.BinaryMessage {
				_, _ = stdinW.Write(data)
				continue
			}
		}
	}()

	out := &wsBinaryWriter{conn: conn}
	queue := &terminalSizeQueue{ch: resizeCh}

	// Shell 自动探测：未指定时先尝试 bash，失败回退 sh
	shells := []string{shell}
	if shell == "" || shell == "bash" {
		shells = []string{"bash", "sh"}
	}

	var execErr error
	for _, sh := range shells {
		execErr = svc.ExecPodTerminal(ctx, clusterID, namespace, pod, container, []string{sh}, stdinR, out, out, true, queue)
		if execErr == nil {
			return
		}
		// 如果只有一个 shell 或者 context 已取消，不再重试
		if ctx.Err() != nil {
			break
		}
	}
	if execErr != nil {
		_ = conn.WriteJSON(gin.H{"type": "error", "message": execErr.Error()})
	}
}
