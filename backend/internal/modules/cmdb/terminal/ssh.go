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

	"devops-platform/config"
	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/pkg/utils"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

type SSHTerminal struct {
	Client  *ssh.Client
	Session *ssh.Session
	Stdin   io.WriteCloser
	Stdout  io.Reader
	Stderr  io.Reader
}

func NewSSHSession(host string, port int, credential *model.Credential, cols, rows int) (*SSHTerminal, error) {
	if credential == nil {
		return nil, fmt.Errorf("凭据不能为空")
	}
	if host == "" {
		return nil, fmt.Errorf("主机地址不能为空")
	}
	if port <= 0 {
		port = 22
	}
	if cols <= 0 {
		cols = 120
	}
	if rows <= 0 {
		rows = 30
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

	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("建立 SSH 连接失败: %w", err)
	}

	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("创建 SSH 会话失败: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("创建 SSH stdin 管道失败: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		stdin.Close()
		session.Close()
		client.Close()
		return nil, fmt.Errorf("创建 SSH stdout 管道失败: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		stdin.Close()
		session.Close()
		client.Close()
		return nil, fmt.Errorf("创建 SSH stderr 管道失败: %w", err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	if err := session.RequestPty("xterm-256color", rows, cols, modes); err != nil {
		stdin.Close()
		session.Close()
		client.Close()
		return nil, fmt.Errorf("请求终端失败: %w", err)
	}

	if err := session.Shell(); err != nil {
		stdin.Close()
		session.Close()
		client.Close()
		return nil, fmt.Errorf("启动 shell 失败: %w", err)
	}

	return &SSHTerminal{
		Client:  client,
		Session: session,
		Stdin:   stdin,
		Stdout:  stdout,
		Stderr:  stderr,
	}, nil
}

func (t *SSHTerminal) Close() {
	if t == nil {
		return
	}
	if t.Stdin != nil {
		_ = t.Stdin.Close()
	}
	if t.Session != nil {
		_ = t.Session.Close()
	}
	if t.Client != nil {
		_ = t.Client.Close()
	}
}

func BuildHostKeyCallback() (ssh.HostKeyCallback, error) {
	paths := candidateKnownHostsPaths()
	if len(paths) == 0 {
		return nil, fmt.Errorf("未找到 known_hosts 文件，请在 terminal.known_hosts_path 中配置有效路径")
	}
	callback, err := knownhosts.New(paths...)
	if err != nil {
		return nil, fmt.Errorf("加载 known_hosts 失败: %w", err)
	}
	return callback, nil
}

func candidateKnownHostsPaths() []string {
	paths := make([]string, 0, 3)
	configuredPath := strings.TrimSpace(config.Cfg.GetString("terminal.known_hosts_path"))
	if configuredPath != "" {
		if resolvedPath := resolveKnownHostsPath(configuredPath); resolvedPath != "" {
			paths = append(paths, resolvedPath)
		}
	}
	for _, path := range []string{"~/.ssh/known_hosts", "/etc/ssh/ssh_known_hosts"} {
		if resolvedPath := resolveKnownHostsPath(path); resolvedPath != "" {
			paths = append(paths, resolvedPath)
		}
	}
	return uniqueStrings(paths)
}

func resolveKnownHostsPath(path string) string {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return ""
	}
	if strings.HasPrefix(trimmed, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		trimmed = filepath.Join(homeDir, trimmed[2:])
	}
	if !filepath.IsAbs(trimmed) {
		if absPath, err := filepath.Abs(trimmed); err == nil {
			trimmed = absPath
		}
	}
	if info, err := os.Stat(trimmed); err == nil && !info.IsDir() {
		return trimmed
	}
	return ""
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func buildAuthMethods(credential *model.Credential) ([]ssh.AuthMethod, error) {
	switch credential.Type {
	case "password":
		password, err := utils.Decrypt(credential.Password)
		if err != nil {
			return nil, fmt.Errorf("解密密码失败: %w", err)
		}
		return []ssh.AuthMethod{ssh.Password(password)}, nil
	case "key":
		privateKey, err := utils.Decrypt(credential.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("解密私钥失败: %w", err)
		}

		var signer ssh.Signer
		if credential.Passphrase != "" {
			passphrase, err := utils.Decrypt(credential.Passphrase)
			if err != nil {
				return nil, fmt.Errorf("解密私钥口令失败: %w", err)
			}
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(privateKey), []byte(passphrase))
			if err != nil {
				return nil, fmt.Errorf("解析带口令私钥失败: %w", err)
			}
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(privateKey))
			if err != nil {
				return nil, fmt.Errorf("解析私钥失败: %w", err)
			}
		}

		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil
	default:
		return nil, nil
	}
}
