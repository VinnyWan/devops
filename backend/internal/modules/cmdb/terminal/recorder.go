package terminal

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Recorder struct {
	mu        sync.Mutex
	file      *os.File
	startedAt time.Time
}

func NewRecorder(path string, width, height int) (*Recorder, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create recording dir: %w", err)
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open recording file: %w", err)
	}

	startedAt := time.Now()
	header := CastHeader{
		Version:   2,
		Width:     width,
		Height:    height,
		Timestamp: startedAt.Unix(),
	}
	encodedHeader, err := json.Marshal(header)
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("encode cast header: %w", err)
	}
	if _, err := file.Write(append(encodedHeader, '\n')); err != nil {
		file.Close()
		return nil, fmt.Errorf("write cast header: %w", err)
	}

	return &Recorder{
		file:      file,
		startedAt: startedAt,
	}, nil
}

func (r *Recorder) Record(eventType, data string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file == nil {
		return os.ErrClosed
	}

	elapsed := time.Since(r.startedAt).Seconds()
	if elapsed < 0 {
		elapsed = 0
	}

	encodedEvent, err := json.Marshal([]any{elapsed, eventType, data})
	if err != nil {
		return fmt.Errorf("encode cast event: %w", err)
	}
	if _, err := r.file.Write(append(encodedEvent, '\n')); err != nil {
		return fmt.Errorf("write cast event: %w", err)
	}

	return nil
}

func (r *Recorder) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.file == nil {
		return nil
	}

	err := r.file.Close()
	r.file = nil
	return err
}
