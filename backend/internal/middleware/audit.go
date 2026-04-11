package middleware

import (
	"bytes"
	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/pkg/logger"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var db *gorm.DB

var sensitiveFields = []string{
	"password", "old_password", "new_password", "confirm_password",
	"token", "secret", "kubeconfig", "ca_data", "client_key",
	"client_certificate", "authorization", "cookie",
}

const (
	auditChannelSize = 1024
	auditWorkerCount = 3
)

var auditChannel chan *model.AuditLog
var auditDropCount int64
var auditInitMu sync.Mutex

// SetDB 设置审计中间件使用的数据库连接，并启动异步写入 worker
func SetDB(database *gorm.DB) {
	auditInitMu.Lock()
	defer auditInitMu.Unlock()

	db = database
	if auditChannel == nil {
		auditChannel = make(chan *model.AuditLog, auditChannelSize)
		for i := 0; i < auditWorkerCount; i++ {
			go auditWorker()
		}
	}
}

func auditWorker() {
	for record := range auditChannel {
		if db == nil {
			continue
		}
		if err := db.Create(record).Error; err != nil {
			logger.Log.Error("审计日志保存失败", zap.Error(err))
		}
	}
}

// SetAuditOperation 设置审计操作名称
func SetAuditOperation(operation string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("audit_operation", operation)
		c.Next()
	}
}

// SetAuditRetention 设置审计日志保留天数
func SetAuditRetention(days int) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("audit_retention", days)
		c.Next()
	}
}

// maskSensitiveFields 对 JSON 请求体中的敏感字段进行脱敏
func maskSensitiveFields(bodyBytes []byte) string {
	if len(bodyBytes) == 0 {
		return ""
	}

	var bodyAny interface{}
	if err := json.Unmarshal(bodyBytes, &bodyAny); err != nil {
		return string(bodyBytes)
	}
	maskedBody := maskSensitiveValue("", bodyAny)
	masked, err := json.Marshal(maskedBody)
	if err != nil {
		return string(bodyBytes)
	}
	return string(masked)
}

func isSensitiveField(fieldKey string) bool {
	lowerKey := strings.ToLower(fieldKey)
	for _, sensitive := range sensitiveFields {
		if lowerKey == sensitive || strings.Contains(lowerKey, sensitive) {
			return true
		}
	}
	return false
}

func maskSensitiveValue(fieldKey string, value interface{}) interface{} {
	if isSensitiveField(fieldKey) {
		return "***"
	}
	switch typed := value.(type) {
	case map[string]interface{}:
		masked := make(map[string]interface{}, len(typed))
		for key, nested := range typed {
			masked[key] = maskSensitiveValue(key, nested)
		}
		return masked
	case []interface{}:
		masked := make([]interface{}, len(typed))
		for i, nested := range typed {
			masked[i] = maskSensitiveValue(fieldKey, nested)
		}
		return masked
	default:
		return value
	}
}

func maskQueryValues(values url.Values) map[string]interface{} {
	masked := make(map[string]interface{}, len(values))
	for key, vals := range values {
		if len(vals) == 1 {
			masked[key] = maskSensitiveValue(key, vals[0])
			continue
		}
		list := make([]interface{}, len(vals))
		for i := range vals {
			list[i] = maskSensitiveValue(key, vals[i])
		}
		masked[key] = list
	}
	return masked
}

func marshalToJSONString(payload interface{}) string {
	data, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(data)
}

func parseAndMaskPayload(contentType string, bodyBytes []byte) interface{} {
	if len(bodyBytes) == 0 {
		return nil
	}
	if strings.Contains(contentType, "application/json") {
		var parsed interface{}
		if err := json.Unmarshal(bodyBytes, &parsed); err == nil {
			return maskSensitiveValue("", parsed)
		}
	}
	return maskSensitiveFields(bodyBytes)
}

type auditResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *auditResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *auditResponseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// Audit 审计日志中间件
func Audit() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		writer := &auditResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer(nil),
		}
		c.Writer = writer

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()
		operation, hasOperation := c.Get("audit_operation")

		latency := time.Since(start)

		retentionDays := 3
		if val, exists := c.Get("audit_retention"); exists {
			if days, ok := val.(int); ok {
				retentionDays = days
			}
		}

		userID, _ := c.Get("userID")
		username, _ := c.Get("username")
		requestAt, _ := c.Get("requestAt")
		timezone, _ := c.Get("timezone")
		locale, _ := c.Get("locale")
		timeFormat, _ := c.Get("time_format")
		requestAtTime := start.UTC()
		if ts, ok := requestAt.(time.Time); ok {
			requestAtTime = ts.UTC()
		}
		requestAtStr := requestAtTime.Format(time.RFC3339)
		queryPayload := maskQueryValues(c.Request.URL.Query())
		requestPayload := map[string]interface{}{
			"request_time": requestAtStr,
			"timezone":     "Asia/Shanghai",
			"locale":       "zh-CN",
			"time_format":  "2006-01-02 15:04:05",
			"username":     "anonymous",
			"method":       c.Request.Method,
			"path":         c.Request.URL.Path,
			"route":        c.FullPath(),
			"query":        queryPayload,
			"payload":      parseAndMaskPayload(c.ContentType(), bodyBytes),
		}

		auditLog := &model.AuditLog{
			UserID:        0,
			Username:      "anonymous",
			Operation:     fmt.Sprintf("%s %s", c.Request.Method, c.Request.URL.Path),
			Method:        c.Request.Method,
			Path:          c.Request.URL.Path,
			Params:        "{}",
			Result:        "{}",
			ErrorMessage:  "",
			IP:            c.ClientIP(),
			Status:        c.Writer.Status(),
			Latency:       latency.Milliseconds(),
			RetentionDays: retentionDays,
			RequestAt:     requestAtTime,
		}

		if uid, ok := userID.(uint); ok {
			auditLog.UserID = uid
		}
		if uname, ok := username.(string); ok {
			auditLog.Username = uname
		}
		requestPayload["username"] = auditLog.Username
		if hasOperation {
			if op, ok := operation.(string); ok {
				auditLog.Operation = op
			}
		} else if route := c.FullPath(); route != "" {
			auditLog.Operation = fmt.Sprintf("%s %s", c.Request.Method, route)
		}
		timezoneStr := "Asia/Shanghai"
		if tz, ok := timezone.(string); ok && tz != "" {
			timezoneStr = tz
		}
		localeStr := "zh-CN"
		if lc, ok := locale.(string); ok && lc != "" {
			localeStr = lc
		}
		timeFormatStr := "2006-01-02 15:04:05"
		if tf, ok := timeFormat.(string); ok && tf != "" {
			timeFormatStr = tf
		}
		requestPayload["timezone"] = timezoneStr
		requestPayload["locale"] = localeStr
		requestPayload["time_format"] = timeFormatStr
		responsePayload := map[string]interface{}{
			"status":     c.Writer.Status(),
			"latency_ms": auditLog.Latency,
			"success":    c.Writer.Status() < 400,
			"body":       parseAndMaskPayload(writer.Header().Get("Content-Type"), writer.body.Bytes()),
		}
		if len(c.Errors) > 0 {
			errMsg := c.Errors.String()
			auditLog.ErrorMessage = errMsg
			responsePayload["error"] = errMsg
		}
		auditLog.Params = marshalToJSONString(requestPayload)
		auditLog.Result = marshalToJSONString(responsePayload)

		// 通过 channel 发送给 worker，避免无限制创建 goroutine
		select {
		case auditChannel <- auditLog:
		default:
			dropped := atomic.AddInt64(&auditDropCount, 1)
			logger.Log.Warn("审计日志 channel 已满，丢弃日志",
				zap.String("path", auditLog.Path),
				zap.String("method", auditLog.Method),
				zap.Int64("dropped_total", dropped),
				zap.Int("queue_capacity", auditChannelSize),
				zap.Int("queue_current", len(auditChannel)))
		}
	}
}

func AuditStats() (queued int, dropped int64, capacity int) {
	if auditChannel == nil {
		return 0, atomic.LoadInt64(&auditDropCount), 0
	}
	return len(auditChannel), atomic.LoadInt64(&auditDropCount), cap(auditChannel)
}
