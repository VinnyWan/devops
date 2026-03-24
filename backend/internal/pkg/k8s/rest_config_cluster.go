package k8s

import (
	"fmt"

	"devops-platform/internal/modules/k8s/model"
	"devops-platform/internal/pkg/utils"
	"k8s.io/client-go/rest"
)

func BuildRestConfigFromCluster(cluster *model.Cluster) (*rest.Config, error) {
	var server, caData, certData, keyData, token string
	server = cluster.Url

	if cluster.AuthType == "kubeconfig" {
		if cluster.Kubeconfig == "" {
			return nil, fmt.Errorf("集群 %d 的 kubeconfig 为空", cluster.ID)
		}
		decryptedKubeconfig, err := utils.Decrypt(cluster.Kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("解密 kubeconfig 失败: %w", err)
		}

		kubeconfigData, err := utils.ParseKubeconfig(decryptedKubeconfig)
		if err != nil {
			return nil, fmt.Errorf("解析 kubeconfig 失败: %w", err)
		}

		server = kubeconfigData.Server
		caData = kubeconfigData.CaData
		if kubeconfigData.AuthType == "cert" {
			certData = kubeconfigData.CertData
			keyData = kubeconfigData.KeyData
		} else {
			token = kubeconfigData.Token
		}
	} else if cluster.AuthType == "token" {
		if cluster.Token == "" {
			return nil, fmt.Errorf("集群 %d 的 token 为空", cluster.ID)
		}
		decryptedToken, err := utils.Decrypt(cluster.Token)
		if err != nil {
			return nil, fmt.Errorf("解密 token 失败: %w", err)
		}
		token = decryptedToken

		if cluster.CaData != "" {
			decryptedCa, err := utils.Decrypt(cluster.CaData)
			if err == nil {
				caData = decryptedCa
			}
		}
	} else {
		return nil, fmt.Errorf("不支持的认证类型: %s", cluster.AuthType)
	}

	clientCfg := &ClientConfig{
		Server:   server,
		CaData:   caData,
		CertData: certData,
		KeyData:  keyData,
		Token:    token,
	}

	return BuildRestConfig(clientCfg)
}
