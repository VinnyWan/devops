package terminal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	ErrTerminalIdleTimeout = errors.New("terminal idle timeout")
	ErrTerminalMaxDuration = errors.New("terminal max session duration exceeded")
)

type WSMessage struct {
	Operation string      `json:"operation"`
	Data      interface{} `json:"data,omitempty"`
	Cols      int         `json:"cols,omitempty"`
	Rows      int         `json:"rows,omitempty"`
}

type Bridge struct {
	conn         *websocket.Conn
	recorder     *Recorder
	sshTerm      *SSHTerminal
	once         sync.Once
	writeMu      sync.Mutex
	activityMu   sync.Mutex
	closedMu     sync.Mutex
	closedClean  bool
	createdAt    time.Time
	lastActivity time.Time
	done         chan struct{}
}

func NewBridge(conn *websocket.Conn, recorder *Recorder, sshTerm *SSHTerminal) *Bridge {
	now := time.Now()
	return &Bridge{
		conn:         conn,
		recorder:     recorder,
		sshTerm:      sshTerm,
		createdAt:    now,
		lastActivity: now,
		done:         make(chan struct{}),
	}
}

func (b *Bridge) touchActivity() {
	b.activityMu.Lock()
	b.lastActivity = time.Now()
	b.activityMu.Unlock()
}

func (b *Bridge) Close() {
	b.once.Do(func() {
		close(b.done)
		if b.recorder != nil {
			_ = b.recorder.Close()
		}
		if b.sshTerm != nil {
			b.sshTerm.Close()
		}
		if b.conn != nil {
			_ = b.conn.Close()
		}
	})
}

func (b *Bridge) CloseWithReason(reason string) {
	b.closedMu.Lock()
	b.closedClean = true
	b.closedMu.Unlock()
	b.once.Do(func() {
		close(b.done)
		if b.conn != nil {
			b.writeMu.Lock()
			_ = b.conn.WriteJSON(WSMessage{Operation: "closed", Data: reason})
			_ = b.conn.WriteControl(websocket.CloseMessage, EncodeClosedMessage(reason), time.Now().Add(2*time.Second))
			b.writeMu.Unlock()
		}
		if b.recorder != nil {
			_ = b.recorder.Close()
		}
		if b.sshTerm != nil {
			b.sshTerm.Close()
		}
		if b.conn != nil {
			_ = b.conn.Close()
		}
	})
}

func (b *Bridge) ClosedCleanly() bool {
	b.closedMu.Lock()
	defer b.closedMu.Unlock()
	return b.closedClean
}

func (b *Bridge) MonitorSession(idleTimeout, maxSessionDuration time.Duration) error {
	if idleTimeout <= 0 && maxSessionDuration <= 0 {
		return nil
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-b.done:
			return nil
		case <-ticker.C:
			b.activityMu.Lock()
			createdAt := b.createdAt
			lastActivity := b.lastActivity
			b.activityMu.Unlock()

			now := time.Now()
			if maxSessionDuration > 0 && now.Sub(createdAt) >= maxSessionDuration {
				return ErrTerminalMaxDuration
			}
			if idleTimeout > 0 && now.Sub(lastActivity) >= idleTimeout {
				return ErrTerminalIdleTimeout
			}
		}
	}
}

func (b *Bridge) ForwardWSInput() error {
	for {
		var message WSMessage
		if err := b.conn.ReadJSON(&message); err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseNoStatusReceived) {
				return nil
			}
			messageText := strings.ToLower(err.Error())
			if strings.Contains(messageText, "use of closed network connection") || strings.Contains(messageText, "websocket: close") || strings.Contains(messageText, "eof") {
				return nil
			}
			return err
		}

		switch message.Operation {
		case "stdin":
			data, ok := message.Data.(string)
			if !ok {
				if dataBytes, marshalErr := json.Marshal(message.Data); marshalErr == nil {
					data = strings.Trim(string(dataBytes), "\"")
				} else {
					return fmt.Errorf("stdin 数据格式无效")
				}
			}
			if b.recorder != nil {
				if err := b.recorder.Record("i", data); err != nil {
					return err
				}
			}
			if _, err := io.WriteString(b.sshTerm.Stdin, data); err != nil {
				return err
			}
			b.touchActivity()
		case "resize":
			if message.Cols <= 0 || message.Rows <= 0 {
				continue
			}
			if err := b.sshTerm.Session.WindowChange(message.Rows, message.Cols); err != nil {
				return err
			}
			b.touchActivity()
		default:
			continue
		}
	}
}

func (b *Bridge) ForwardSSHOutput(reader io.Reader, operation string) error {
	buffer := make([]byte, 4096)
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			payload := string(buffer[:n])
			if b.recorder != nil {
				if recordErr := b.recorder.Record("o", payload); recordErr != nil {
					return recordErr
				}
			}
			if b.conn != nil {
				b.writeMu.Lock()
				writeErr := b.conn.WriteJSON(WSMessage{Operation: operation, Data: payload})
				b.writeMu.Unlock()
				if writeErr != nil {
					return writeErr
				}
			}
			b.touchActivity()
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

func EncodeClosedMessage(reason string) []byte {
	if len(reason) > 120 {
		reason = reason[:120]
	}
	return websocket.FormatCloseMessage(websocket.CloseNormalClosure, reason)
}
