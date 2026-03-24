package utils

import (
	"encoding/base64"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// KubeconfigData kubeconfig 解析后的数据
type KubeconfigData struct {
	Server   string
	CaData   string
	CertData string
	KeyData  string
	Token    string // 如果使用token认证
	AuthType string // "cert" 或 "token"
}

// Kubeconfig kubeconfig 文件结构
type Kubeconfig struct {
	Clusters []struct {
		Name    string `yaml:"name"`
		Cluster struct {
			Server                   string `yaml:"server"`
			CertificateAuthorityData string `yaml:"certificate-authority-data"`
		} `yaml:"cluster"`
	} `yaml:"clusters"`
	Users []struct {
		Name string `yaml:"name"`
		User struct {
			ClientCertificateData string `yaml:"client-certificate-data"`
			ClientKeyData         string `yaml:"client-key-data"`
			Token                 string `yaml:"token"`
		} `yaml:"user"`
	} `yaml:"users"`
	Contexts []struct {
		Name    string `yaml:"name"`
		Context struct {
			Cluster string `yaml:"cluster"`
			User    string `yaml:"user"`
		} `yaml:"context"`
	} `yaml:"contexts"`
	CurrentContext string `yaml:"current-context"`
}

// ParseKubeconfig 解析 kubeconfig YAML
func ParseKubeconfig(kubeconfigYAML string) (*KubeconfigData, error) {
	// 进一步清洗数据函数定义
	cleanString := func(s string) string {
		// 先处理转义的换行
		s = strings.ReplaceAll(s, "\\n", " ")
		s = strings.ReplaceAll(s, "\\r", " ")

		// 将所有空白字符（换行、制表符等）替换为空格
		s = strings.Map(func(r rune) rune {
			if r == '\n' || r == '\r' || r == '\t' {
				return ' '
			}
			return r
		}, s)

		// 使用 Fields 切分，只取第一部分（处理同一行拼接了其他 key: value 的情况）
		fields := strings.Fields(s)
		if len(fields) == 0 {
			return ""
		}

		// 取第一个非空部分，并去除首尾的引号、反引号和空格
		res := fields[0]
		res = strings.Trim(res, "\"`'\" ")
		return res
	}

	// 0. 预清洗：处理常见的粘贴错误
	// 将所有反引号替换为双引号
	kubeconfigYAML = strings.ReplaceAll(kubeconfigYAML, "`", "\"")
	// 处理一些可能被意外转义的换行符
	kubeconfigYAML = strings.ReplaceAll(kubeconfigYAML, "\\n", "\n")
	kubeconfigYAML = strings.TrimSpace(kubeconfigYAML)

	var config Kubeconfig
	err := yaml.Unmarshal([]byte(kubeconfigYAML), &config)
	if err != nil {
		// 如果 YAML 解析失败，尝试正则提取
		data := &KubeconfigData{}
		extract := func(key string) string {
			re := regexp.MustCompile(key + `\s*[:=]\s*("[^"]*"|'[^']*'|` + "`" + `[^` + "`" + `]*` + "`" + `|[^\s\n,]+)`)
			match := re.FindStringSubmatch(kubeconfigYAML)
			if len(match) > 1 {
				return match[1]
			}
			return ""
		}

		data.Server = cleanString(extract("server"))
		data.CaData = cleanString(extract("certificate-authority-data"))
		token := cleanString(extract("token"))
		certData := cleanString(extract("client-certificate-data"))
		keyData := cleanString(extract("client-key-data"))

		if data.Server != "" && (token != "" || (certData != "" && keyData != "")) {
			if token != "" {
				data.AuthType = "token"
				data.Token = token
			} else {
				data.AuthType = "cert"
				data.CertData = certData
				data.KeyData = keyData
			}
			return data, nil
		}

		return nil, fmt.Errorf("解析 kubeconfig 失败 (请检查 YAML 格式): %w", err)
	}

	// 校验必要字段
	if len(config.Clusters) == 0 {
		return nil, errors.New("kubeconfig 中未找到集群配置 (clusters 字段为空)")
	}
	if len(config.Users) == 0 {
		return nil, errors.New("kubeconfig 中未找到用户配置 (users 字段为空)")
	}

	// 获取当前上下文
	currentContext := config.CurrentContext
	if currentContext == "" && len(config.Contexts) > 0 {
		currentContext = config.Contexts[0].Name
	}

	var clusterName, userName string
	for _, ctx := range config.Contexts {
		if ctx.Name == currentContext {
			clusterName = ctx.Context.Cluster
			userName = ctx.Context.User
			break
		}
	}

	if clusterName == "" && len(config.Clusters) > 0 {
		clusterName = config.Clusters[0].Name
	}
	if userName == "" && len(config.Users) > 0 {
		userName = config.Users[0].Name
	}

	var clusterInfo *struct {
		Server                   string `yaml:"server"`
		CertificateAuthorityData string `yaml:"certificate-authority-data"`
	}
	for _, c := range config.Clusters {
		if c.Name == clusterName {
			clusterInfo = &c.Cluster
			break
		}
	}
	if clusterInfo == nil {
		return nil, errors.New("未找到对应的集群配置 (cluster)")
	}

	var userInfo *struct {
		ClientCertificateData string `yaml:"client-certificate-data"`
		ClientKeyData         string `yaml:"client-key-data"`
		Token                 string `yaml:"token"`
	}
	for _, u := range config.Users {
		if u.Name == userName {
			userInfo = &u.User
			break
		}
	}
	if userInfo == nil {
		return nil, errors.New("未找到对应的用户配置 (user)")
	}

	data := &KubeconfigData{
		Server: cleanString(clusterInfo.Server),
		CaData: cleanString(clusterInfo.CertificateAuthorityData),
	}

	if userInfo.Token != "" {
		data.AuthType = "token"
		data.Token = cleanString(userInfo.Token)
	} else if userInfo.ClientCertificateData != "" && userInfo.ClientKeyData != "" {
		data.AuthType = "cert"
		data.CertData = cleanString(userInfo.ClientCertificateData)
		data.KeyData = cleanString(userInfo.ClientKeyData)
	} else {
		return nil, errors.New("未找到有效的认证信息 (需要 token 或 client-certificate-data/client-key-data)")
	}

	if data.Server == "" {
		return nil, errors.New("未找到有效的 API Server 地址")
	}

	return data, nil
}

// ValidateBase64 验证 base64 编码的数据
func ValidateBase64(data string) error {
	if data == "" {
		return nil
	}
	_, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return fmt.Errorf("无效的 base64 编码: %w", err)
	}
	return nil
}

// BuildKubeconfig 构建 kubeconfig YAML（用于存储）
func BuildKubeconfig(server, caData, certData, keyData, token string) string {
	if token != "" {
		return fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: cluster
  cluster:
    server: %s
    certificate-authority-data: %s
users:
- name: user
  user:
    token: %s
contexts:
- name: context
  context:
    cluster: cluster
    user: user
current-context: context
`, server, caData, token)
	} else {
		return fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- name: cluster
  cluster:
    server: %s
    certificate-authority-data: %s
users:
- name: user
  user:
    client-certificate-data: %s
    client-key-data: %s
contexts:
- name: context
  context:
    cluster: cluster
    user: user
current-context: context
`, server, caData, certData, keyData)
	}
}
