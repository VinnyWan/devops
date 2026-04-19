package terminal

import (
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"devops-platform/internal/modules/cmdb/model"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// SFTPClient wraps an SSH connection with SFTP subsystem.
type SFTPClient struct {
	sshClient  *ssh.Client
	sftpClient *sftp.Client
}

// FileEntry represents a file or directory entry.
type FileEntry struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
	ModTime string `json:"modTime"`
	Owner   string `json:"owner,omitempty"`
}

// NewSFTPClient establishes an SSH connection then opens the SFTP subsystem.
func NewSFTPClient(host string, port int, credential *model.Credential) (*SFTPClient, error) {
	if credential == nil {
		return nil, fmt.Errorf("凭据不能为空")
	}
	if host == "" {
		return nil, fmt.Errorf("主机地址不能为空")
	}
	if port <= 0 {
		port = 22
	}

	authMethods, err := buildAuthMethods(credential)
	if err != nil {
		return nil, err
	}
	if len(authMethods) == 0 {
		return nil, fmt.Errorf("不支持的凭据类型: %s", credential.Type)
	}

	address := net.JoinHostPort(host, strconv.Itoa(port))
	hostKeyCallback, err := BuildHostKeyCallback()
	if err != nil {
		return nil, err
	}
	config := &ssh.ClientConfig{
		User:            credential.Username,
		Auth:            authMethods,
		HostKeyCallback: hostKeyCallback,
		Timeout:         10 * time.Second,
	}

	sshClient, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("建立 SSH 连接失败: %w", err)
	}

	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, fmt.Errorf("创建 SFTP 会话失败: %w", err)
	}

	return &SFTPClient{
		sshClient:  sshClient,
		sftpClient: sftpClient,
	}, nil
}

// Close releases SFTP and SSH resources.
func (c *SFTPClient) Close() {
	if c == nil {
		return
	}
	if c.sftpClient != nil {
		_ = c.sftpClient.Close()
	}
	if c.sshClient != nil {
		_ = c.sshClient.Close()
	}
}

// Browse lists files and directories at the given path.
func (c *SFTPClient) Browse(path string) ([]FileEntry, error) {
	path = normalizePath(path)
	entries, err := c.sftpClient.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	result := make([]FileEntry, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		if name == "." || name == ".." {
			continue
		}
		result = append(result, FileEntry{
			Name:    name,
			Path:    filepath.Join(path, name),
			IsDir:   entry.IsDir(),
			Size:    entry.Size(),
			Mode:    entry.Mode().String(),
			ModTime: entry.ModTime().Format(time.DateTime),
		})
	}
	return result, nil
}

// Stat returns file info for the given path.
func (c *SFTPClient) Stat(path string) (*FileEntry, error) {
	path = normalizePath(path)
	info, err := c.sftpClient.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}
	return &FileEntry{
		Name:    info.Name(),
		Path:    path,
		IsDir:   info.IsDir(),
		Size:    info.Size(),
		Mode:    info.Mode().String(),
		ModTime: info.ModTime().Format(time.DateTime),
	}, nil
}

// Download reads a remote file and writes it to the given writer.
func (c *SFTPClient) Download(path string, writer io.Writer) (int64, error) {
	path = normalizePath(path)
	f, err := c.sftpClient.Open(path)
	if err != nil {
		return 0, fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()
	return io.Copy(writer, f)
}

// Upload reads from the given reader and writes to the remote path.
func (c *SFTPClient) Upload(path string, reader io.Reader, size int64) error {
	path = normalizePath(path)
	dir := filepath.Dir(path)
	if err := c.sftpClient.MkdirAll(dir); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}
	f, err := c.sftpClient.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	return err
}

// Delete removes a file or directory at the given path.
func (c *SFTPClient) Delete(path string) error {
	path = normalizePath(path)
	info, err := c.sftpClient.Stat(path)
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}
	if info.IsDir() {
		return c.removeDir(path)
	}
	return c.sftpClient.Remove(path)
}

func (c *SFTPClient) removeDir(path string) error {
	walker := c.sftpClient.Walk(path)
	var paths []string
	for walker.Step() {
		if err := walker.Err(); err != nil {
			return err
		}
		if walker.Path() != path {
			paths = append(paths, walker.Path())
		}
	}
	for i := len(paths) - 1; i >= 0; i-- {
		info, _ := c.sftpClient.Stat(paths[i])
		if info != nil && info.IsDir() {
			if err := c.sftpClient.RemoveDirectory(paths[i]); err != nil {
				return err
			}
		} else {
			if err := c.sftpClient.Remove(paths[i]); err != nil {
				return err
			}
		}
	}
	return c.sftpClient.RemoveDirectory(path)
}

// Rename renames a file or directory.
func (c *SFTPClient) Rename(oldPath, newPath string) error {
	return c.sftpClient.Rename(normalizePath(oldPath), normalizePath(newPath))
}

// Mkdir creates a directory (and any necessary parents).
func (c *SFTPClient) Mkdir(path string) error {
	return c.sftpClient.MkdirAll(normalizePath(path))
}

// Chmod changes file permissions.
func (c *SFTPClient) Chmod(path string, mode os.FileMode) error {
	return c.sftpClient.Chmod(normalizePath(path), mode)
}

// ReadFile reads the content of a remote file (for preview/edit).
func (c *SFTPClient) ReadFile(path string, maxBytes int64) (string, error) {
	path = normalizePath(path)
	info, err := c.sftpClient.Stat(path)
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %w", err)
	}
	if info.IsDir() {
		return "", fmt.Errorf("不能读取目录")
	}
	if info.Size() > maxBytes {
		return "", fmt.Errorf("文件大小超过限制(%d字节)", maxBytes)
	}
	f, err := c.sftpClient.Open(path)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	return string(data), nil
}

// WriteFile writes content to a remote file (for edit).
func (c *SFTPClient) WriteFile(path, content string) error {
	path = normalizePath(path)
	f, err := c.sftpClient.Create(path)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()
	_, err = f.Write([]byte(content))
	return err
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return filepath.Clean(path)
}
