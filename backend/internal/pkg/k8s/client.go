package k8s

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

// ClientConfig 客户端配置
type ClientConfig struct {
	Server   string
	CaData   string // base64 编码的 CA 证书
	CertData string // base64 编码的客户端证书（可选）
	KeyData  string // base64 编码的客户端密钥（可选）
	Token    string // Bearer Token（可选）
}

// BuildRestConfig 构建可复用的 K8s rest.Config
func BuildRestConfig(cfg *ClientConfig) (*rest.Config, error) {
	if cfg.Server == "" {
		return nil, errors.New("server 地址不能为空")
	}

	config := &rest.Config{
		Host:    cfg.Server,
		Timeout: 10 * time.Second,
	}

	// 设置 TLS 配置
	tlsConfig := &tls.Config{}

	// 设置 CA 证书
	if cfg.CaData != "" {
		caBytes, err := base64.StdEncoding.DecodeString(cfg.CaData)
		if err != nil {
			return nil, fmt.Errorf("解码 CA 证书失败: %w", err)
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caBytes) {
			return nil, errors.New("解析 CA 证书失败")
		}
		tlsConfig.RootCAs = caCertPool
	} else {
		// 如果没有提供 CA，则跳过证书验证（不推荐用于生产环境）
		tlsConfig.InsecureSkipVerify = true
	}

	config.TLSClientConfig = rest.TLSClientConfig{
		Insecure: tlsConfig.InsecureSkipVerify,
		CAData:   []byte{},
	}

	if !tlsConfig.InsecureSkipVerify {
		caBytes, _ := base64.StdEncoding.DecodeString(cfg.CaData)
		config.TLSClientConfig.CAData = caBytes
	}

	// 设置认证方式
	if cfg.Token != "" {
		// Token 认证
		config.BearerToken = cfg.Token
	} else if cfg.CertData != "" && cfg.KeyData != "" {
		// 证书认证
		certBytes, err := base64.StdEncoding.DecodeString(cfg.CertData)
		if err != nil {
			return nil, fmt.Errorf("解码客户端证书失败: %w", err)
		}
		keyBytes, err := base64.StdEncoding.DecodeString(cfg.KeyData)
		if err != nil {
			return nil, fmt.Errorf("解码客户端密钥失败: %w", err)
		}
		config.TLSClientConfig.CertData = certBytes
		config.TLSClientConfig.KeyData = keyBytes
	} else {
		return nil, errors.New("必须提供 token 或证书认证信息")
	}

	return config, nil
}

// NewClient 创建 K8s 客户端（通用方法）
func NewClient(cfg *ClientConfig) (*Client, error) {
	config, err := BuildRestConfig(cfg)
	if err != nil {
		return nil, err
	}

	// 创建 clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("创建 kubernetes 客户端失败: %w", err)
	}

	return &Client{
		clientset: clientset,
		config:    config,
	}, nil
}

// HealthCheck 健康检查
func (c *Client) HealthCheck(ctx context.Context) error {
	// 尝试获取版本信息
	_, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return fmt.Errorf("健康检查失败: %w", err)
	}
	return nil
}

// GetServerVersion 获取服务器版本
func (c *Client) GetServerVersion(ctx context.Context) (string, error) {
	version, err := c.clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	return version.GitVersion, nil
}

// GetNodeCount 获取节点数量
func (c *Client) GetNodeCount(ctx context.Context) (int, error) {
	nodes, err := c.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}
	return len(nodes.Items), nil
}

// GetNamespaceCount 获取命名空间数量
func (c *Client) GetNamespaceCount(ctx context.Context) (int, error) {
	namespaces, err := c.clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return 0, err
	}
	return len(namespaces.Items), nil
}

// Ping 简单的连接测试
func (c *Client) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 使用 HTTP 请求测试连接
	req, err := http.NewRequestWithContext(ctx, "GET", c.config.Host+"/healthz", nil)
	if err != nil {
		return err
	}

	if c.config.BearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.config.BearerToken)
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: c.config.TLSClientConfig.Insecure,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("连接失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("健康检查返回非200状态码: %d", resp.StatusCode)
	}

	return nil
}
