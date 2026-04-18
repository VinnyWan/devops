package terminal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
)

type CastHeader struct {
	Version   int   `json:"version"`
	Width     int   `json:"width"`
	Height    int   `json:"height"`
	Timestamp int64 `json:"timestamp"`
}

type ReplayEvent struct {
	Time float64 `json:"time"`
	Type string  `json:"type"`
	Data string  `json:"data"`
}

type ReplayPayload struct {
	Width  int           `json:"width"`
	Height int           `json:"height"`
	Events []ReplayEvent `json:"events"`
}

func ParseCastFile(path string) (*ReplayPayload, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024)
	scanner.Buffer(buffer, 1024*1024)

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("cast file is empty")
	}

	var header CastHeader
	if err := json.Unmarshal(scanner.Bytes(), &header); err != nil {
		return nil, fmt.Errorf("decode cast header: %w", err)
	}

	payload := &ReplayPayload{
		Width:  header.Width,
		Height: header.Height,
		Events: make([]ReplayEvent, 0),
	}

	for lineNumber := 2; scanner.Scan(); lineNumber++ {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		event, err := decodeCastEvent(line, lineNumber)
		if err != nil {
			return nil, err
		}
		payload.Events = append(payload.Events, event)
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return payload, nil
}

func decodeCastEvent(line []byte, lineNumber int) (ReplayEvent, error) {
	var row []json.RawMessage
	if err := json.Unmarshal(line, &row); err != nil {
		return ReplayEvent{}, fmt.Errorf("decode cast event line %d: %w", lineNumber, err)
	}
	if len(row) < 3 {
		return ReplayEvent{}, fmt.Errorf("decode cast event line %d: expected at least 3 fields", lineNumber)
	}

	var event ReplayEvent
	if err := json.Unmarshal(row[0], &event.Time); err != nil {
		return ReplayEvent{}, fmt.Errorf("decode cast event line %d: %w", lineNumber, err)
	}
	if err := json.Unmarshal(row[1], &event.Type); err != nil {
		return ReplayEvent{}, fmt.Errorf("decode cast event line %d: %w", lineNumber, err)
	}
	if err := json.Unmarshal(row[2], &event.Data); err != nil {
		return ReplayEvent{}, fmt.Errorf("decode cast event line %d: %w", lineNumber, err)
	}

	return event, nil
}
