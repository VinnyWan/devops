package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"devops-platform/config"
	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/service"
	cmdbterminal "devops-platform/internal/modules/cmdb/terminal"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var terminalUpgrader = websocket.Upgrader{
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
		return false
	},
}

func TerminalList(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	svc := getTerminalService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	req := service.TerminalListRequest{
		Page:     parsePositiveInt(c.DefaultQuery("page", "1"), 1),
		PageSize: parsePositiveInt(c.DefaultQuery("pageSize", "10"), 10),
		Keyword:  c.Query("keyword"),
		Username: c.Query("username"),
		Status:   c.Query("status"),
	}

	sessions, total, err := svc.ListInTenant(tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取终端会话列表失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":     200,
		"message":  "获取成功",
		"data":     sessions,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	})
}

func TerminalDetail(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	id, err := parseUintQuery(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的 ID"})
		return
	}

	svc := getTerminalService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	session, err := svc.DetailInTenant(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "终端会话不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取终端会话详情失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    session,
	})
}

func TerminalRecording(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	id, err := parseUintQuery(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的 ID"})
		return
	}

	svc := getTerminalService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	session, err := svc.DetailInTenant(tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "终端会话不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"message": "获取终端录屏失败", "error": err.Error()})
		return
	}

	if session.Status == "active" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "终端录屏尚未结束，暂不可回放"})
		return
	}

	payload, err := cmdbterminal.ParseCastFile(session.RecordingPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "解析终端录屏失败", "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    payload,
	})
}

func TerminalConnect(c *gin.Context) {
	tenantID, err := getCurrentTenantID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": err.Error()})
		return
	}

	hostID, err := parseUintQuery(c, "hostId")
	if err != nil || hostID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"message": "无效的 hostId"})
		return
	}

	userID := c.GetUint("userID")
	username := "unknown"
	if rawUsername, exists := c.Get("username"); exists {
		if value, ok := rawUsername.(string); ok && strings.TrimSpace(value) != "" {
			username = value
		}
	}

	svc := getTerminalService()
	if svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "服务未初始化"})
		return
	}

	host, credential, err := svc.GetConnectTarget(tenantID, hostID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"message": "主机或凭据不存在"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"message": "获取连接目标失败", "error": err.Error()})
		return
	}

	cols := parsePositiveInt(c.DefaultQuery("cols", "120"), 120)
	rows := parsePositiveInt(c.DefaultQuery("rows", "30"), 30)

	conn, err := terminalUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	sendStartupError := func(message string) {
		_ = conn.WriteJSON(cmdbterminal.WSMessage{Operation: "error", Data: message})
		_ = conn.WriteMessage(websocket.CloseMessage, cmdbterminal.EncodeClosedMessage(message))
		_ = conn.Close()
	}

	writeSocketMessage := func(message cmdbterminal.WSMessage) {
		if err := conn.WriteJSON(message); err != nil && !isNormalTerminalClose(err) {
			_ = err
		}
	}

	tempSession, err := svc.CreateSession(tenantID, userID, username, host, credential.ID, c.ClientIP(), "pending")
	if err != nil {
		sendStartupError(fmt.Sprintf("创建终端会话失败: %v", err))
		return
	}

	recordingDir := config.Cfg.GetString("terminal.recording_dir")
	if strings.TrimSpace(recordingDir) == "" {
		recordingDir = "./data/recordings/cmdb-terminal"
	}
	recordingPath := svc.BuildRecordingPath(recordingDir, tempSession.StartedAt, tempSession.ID)
	if !filepath.IsAbs(recordingPath) {
		if absPath, absErr := filepath.Abs(recordingPath); absErr == nil {
			recordingPath = absPath
		} else {
			recordingPath = filepath.Clean(recordingPath)
		}
	}

	if err := cmdbDB.Model(tempSession).Update("recording_path", recordingPath).Error; err != nil {
		sendStartupError(fmt.Sprintf("更新录屏路径失败: %v", err))
		return
	}
	tempSession.RecordingPath = recordingPath

	recorder, err := cmdbterminal.NewRecorder(recordingPath, cols, rows)
	if err != nil {
		markTerminalInterrupted(svc, tenantID, tempSession)
		sendStartupError(fmt.Sprintf("创建录屏文件失败: %v", err))
		return
	}

	sshTerm, err := cmdbterminal.NewSSHSession(host.Ip, host.Port, credential, cols, rows)
	if err != nil {
		_ = recorder.Close()
		markTerminalInterrupted(svc, tenantID, tempSession)
		sendStartupError(fmt.Sprintf("建立 SSH 会话失败: %v", err))
		return
	}

	bridge := cmdbterminal.NewBridge(conn, recorder, sshTerm)
	idleTimeout := time.Duration(config.Cfg.GetInt("terminal.idle_timeout")) * time.Second
	maxSessionDuration := time.Duration(config.Cfg.GetInt("terminal.max_session_duration")) * time.Second
	errCh := make(chan error, 4)
	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		errCh <- bridge.ForwardWSInput()
	}()
	go func() {
		defer wg.Done()
		errCh <- bridge.ForwardSSHOutput(sshTerm.Stdout, "stdout")
	}()
	go func() {
		defer wg.Done()
		errCh <- bridge.ForwardSSHOutput(sshTerm.Stderr, "stderr")
	}()
	go func() {
		defer wg.Done()
		errCh <- bridge.MonitorSession(idleTimeout, maxSessionDuration)
	}()

	firstErr := <-errCh
	status, closeReason, closeMessage := resolveTerminalCloseOutcome(firstErr)

	bridge.CloseWithReason(closeMessage)
	wg.Wait()

	fileSize := terminalRecordingFileSize(recordingPath)
	finishedAt := time.Now()
	if err := svc.CloseSession(tenantID, tempSession.ID, finishedAt, fileSize, status, closeReason); err != nil {
		writeSocketMessage(cmdbterminal.WSMessage{Operation: "error", Data: fmt.Sprintf("关闭终端会话失败: %v", err)})
	}
}

func parseUintQuery(c *gin.Context, key string) (uint, error) {
	value := c.Query(key)
	parsed, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(parsed), nil
}

func parsePositiveInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func terminalRecordingFileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func resolveTerminalCloseOutcome(err error) (status string, closeReason string, closeMessage string) {
	status = "closed"
	closeReason = "用户主动关闭或连接正常结束"
	closeMessage = "Connection closed."

	switch {
	case errors.Is(err, cmdbterminal.ErrTerminalIdleTimeout):
		return "idle_timeout", "空闲超时自动断开", "Terminal idle timeout."
	case errors.Is(err, cmdbterminal.ErrTerminalMaxDuration):
		return "max_duration", "会话时长超限自动断开", "Terminal max session duration exceeded."
	case err == nil || isNormalTerminalClose(err):
		return status, closeReason, closeMessage
	default:
		return "interrupted", "连接异常中断", err.Error()
	}
}

func markTerminalInterrupted(svc *service.TerminalService, tenantID uint, session *model.TerminalSession) {
	if svc == nil || session == nil {
		return
	}
	_ = svc.CloseSession(tenantID, session.ID, time.Now(), terminalRecordingFileSize(session.RecordingPath), "interrupted", "连接异常中断")
}

func isNormalTerminalClose(err error) bool {
	if err == nil || errors.Is(err, os.ErrClosed) || errors.Is(err, cmdbterminal.ErrTerminalIdleTimeout) || errors.Is(err, cmdbterminal.ErrTerminalMaxDuration) {
		return true
	}
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "use of closed network connection") || strings.Contains(message, "websocket: close") || strings.Contains(message, "session closed") || strings.Contains(message, "eof")
}
