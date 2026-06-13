package service

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"devops-platform/internal/modules/task/model"
)

// LogWriter captures execution output for streaming.
type LogWriter struct {
	Buffer *bytes.Buffer
	Stream chan string
}

func NewLogWriter() *LogWriter {
	return &LogWriter{
		Buffer: bytes.NewBuffer(nil),
		Stream: make(chan string, 100),
	}
}

func (w *LogWriter) Write(p []byte) (int, error) {
	line := string(p)
	w.Buffer.WriteString(line)
	select {
	case w.Stream <- line:
	default:
	}
	return len(p), nil
}

// RunShell executes a shell script on the local machine or via SSH.
func RunShell(ctx context.Context, content string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", content)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("execution failed: %w\nstderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

// RunAnsible executes an Ansible playbook by piping content to ansible-playbook.
func RunAnsible(ctx context.Context, playbookContent string, inventory string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	args := []string{}
	if inventory != "" {
		args = append(args, "-i", inventory)
	}
	cmd := exec.CommandContext(ctx, "ansible-playbook", args...)
	cmd.Stdin = strings.NewReader(playbookContent)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("ansible execution failed: %w\nstderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func (s *TaskService) runExecution(ctx context.Context, exec *model.TaskExecution, task *model.Task) {
	s.execRepo.SetStarted(exec.ID)
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(ctx, time.Duration(task.Timeout)*time.Second)
	defer cancel()

	var output string
	var execErr error

	switch task.Type {
	case model.TaskTypeShell, model.TaskTypePython:
		output, execErr = RunShell(ctx, task.Content, time.Duration(task.Timeout)*time.Second)
	case model.TaskTypeAnsible:
		output, execErr = RunAnsible(ctx, task.Content, "", time.Duration(task.Timeout)*time.Second)
	default:
		execErr = fmt.Errorf("unsupported task type: %s", task.Type)
	}

	durationMs := time.Since(startTime).Milliseconds()
	status := model.TaskStatusSuccess
	if execErr != nil {
		status = model.TaskStatusFailed
		if ctx.Err() != nil {
			status = model.TaskStatusTimeout
		}
		output = execErr.Error()
	}
	s.execRepo.UpdateStatus(exec.ID, status, output, durationMs)
}
