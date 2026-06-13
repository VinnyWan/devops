package service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/pkg/logger"

	"go.uber.org/zap"
)

// aliyunDescribeInstancesResponse maps the JSON response from Alibaba Cloud ECS DescribeInstances.
type aliyunDescribeInstancesResponse struct {
	RequestID  string `json:"RequestId"`
	TotalCount int    `json:"TotalCount"`
	PageNumber int    `json:"PageNumber"`
	PageSize   int    `json:"PageSize"`
	Instances  struct {
		Instance []aliyunInstance `json:"Instance"`
	} `json:"Instances"`
}

type aliyunInstance struct {
	InstanceID   string `json:"InstanceId"`
	InstanceName string `json:"InstanceName"`
	Status       string `json:"Status"`
	OSType       string `json:"OSType"`
	OSName       string `json:"OSName"`
	CPU          int    `json:"Cpu"`
	Memory       int    `json:"Memory"`
	ZoneID       string `json:"ZoneId"`
	RegionID     string `json:"RegionId"`
	VpcAttributes struct {
		PrivateIPAddress struct {
			IPAddress []string `json:"IpAddress"`
		} `json:"PrivateIpAddress"`
	} `json:"VpcAttributes"`
	PublicIPAddress struct {
		IPAddress []string `json:"IpAddress"`
	} `json:"PublicIpAddress"`
}

func (s *CloudAccountService) syncAliyun(account *model.CloudAccount, secretID, secretKey string) error {
	regions := getDefaultRegions()
	var syncErrors []string

	for _, region := range regions {
		if err := s.syncAliyunECS(account, secretID, secretKey, region); err != nil {
			syncErrors = append(syncErrors, fmt.Sprintf("%s: %s", region, err.Error()))
		}
	}

	if len(syncErrors) > 0 {
		return fmt.Errorf("部分资源同步失败: %s", strings.Join(syncErrors, "; "))
	}
	return nil
}

func (s *CloudAccountService) syncAliyunECS(account *model.CloudAccount, accessKeyID, accessKeySecret, region string) error {
	endpoint := fmt.Sprintf("ecs.%s.aliyuncs.com", region)

	return s.paginateSyncAliyun(func(pageNumber int) (int, error) {
		params := map[string]string{
			"Action":      "DescribeInstances",
			"Version":     "2014-05-26",
			"RegionId":    region,
			"PageSize":    fmt.Sprintf("%d", cloudSyncPageSize),
			"PageNumber":  fmt.Sprintf("%d", pageNumber),
		}

		resp, err := callAliyunAPI(endpoint, accessKeyID, accessKeySecret, params)
		if err != nil {
			return 0, err
		}

		for _, instance := range resp.Instances.Instance {
			privateIP := ""
			if len(instance.VpcAttributes.PrivateIPAddress.IPAddress) > 0 {
				privateIP = instance.VpcAttributes.PrivateIPAddress.IPAddress[0]
			} else if len(instance.PublicIPAddress.IPAddress) > 0 {
				privateIP = instance.PublicIPAddress.IPAddress[0]
			}

			state := mapAliyunStatus(instance.Status)
			zone := instance.ZoneID

			specJSON, _ := json.Marshal(map[string]interface{}{
				"cpu": instance.CPU, "memory": instance.Memory, "zone": zone,
				"os_type": instance.OSType, "os_name": instance.OSName,
			})

			resource := &model.CloudResource{
				TenantID:       account.TenantID,
				CloudAccountID: account.ID,
				ResourceType:   "ecs",
				ResourceID:     instance.InstanceID,
				Region:         region,
				Zone:           zone,
				Name:           instance.InstanceName,
				State:          state,
				Spec:           string(specJSON),
				SyncedAt:       time.Now(),
			}
			if err := s.repo.UpsertResource(resource); err != nil {
				logger.Log.Warn("同步阿里云资源失败", zap.String("region", region), zap.String("resourceID", instance.InstanceID), zap.Error(err))
			}

			existing, hostErr := s.repo.GetHostByCloudInstanceID(account.TenantID, instance.InstanceID)
			if hostErr == nil && existing != nil {
				existing.Hostname = instance.InstanceName
				if privateIP != "" {
					existing.Ip = privateIP
				}
				existing.OsName = instance.OSName
				existing.CloudAccountID = &account.ID
				if err := s.repo.UpdateHost(existing); err != nil {
					logger.Log.Warn("更新阿里云主机关联 Host 失败", zap.String("instanceID", instance.InstanceID), zap.Error(err))
				}
			} else {
				tenantID := account.TenantID
				accountID := account.ID
				host := &model.Host{
					TenantID:        &tenantID,
					Hostname:        instance.InstanceName,
					Ip:              privateIP,
					Port:            22,
					OsName:          instance.OSName,
					Status:          state,
					CloudAccountID:  &accountID,
					CloudInstanceID: instance.InstanceID,
				}
				if err := s.repo.CreateHost(host); err != nil {
					logger.Log.Warn("创建阿里云主机关联 Host 失败", zap.String("instanceID", instance.InstanceID), zap.Error(err))
				}
			}
		}

		return len(resp.Instances.Instance), nil
	})
}

// paginateSyncAliyun paginates through Alibaba Cloud API using page numbers (1-based).
func (s *CloudAccountService) paginateSyncAliyun(fetchPage func(pageNumber int) (int, error)) error {
	for i := 1; i <= maxPaginationPages; i++ {
		count, err := fetchPage(i)
		if err != nil {
			if logger.Log != nil {
				logger.Log.Error("阿里云同步分页拉取失败，跳过该页", zap.Int("page", i), zap.Error(err))
			}
			continue
		}
		if count < cloudSyncPageSize {
			return nil
		}
	}
	return nil
}

// callAliyunAPI makes a signed GET request to the Alibaba Cloud API.
func callAliyunAPI(endpoint, accessKeyID, accessKeySecret string, actionParams map[string]string) (*aliyunDescribeInstancesResponse, error) {
	nonce := make([]byte, 16)
	rand.Read(nonce)

	params := url.Values{}
	params.Set("Format", "JSON")
	params.Set("Version", "2014-05-26")
	params.Set("AccessKeyId", accessKeyID)
	params.Set("SignatureMethod", "HMAC-SHA1")
	params.Set("Timestamp", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	params.Set("SignatureVersion", "1.0")
	params.Set("SignatureNonce", hex.EncodeToString(nonce))

	for k, v := range actionParams {
		params.Set(k, v)
	}

	signature := signAliyun(params, "GET", accessKeySecret)
	params.Set("Signature", signature)

	reqURL := fmt.Sprintf("https://%s/?%s", endpoint, params.Encode())
	resp, err := http.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("阿里云 API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取阿里云 API 响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("阿里云 API 返回状态码 %d: %s", resp.StatusCode, string(body))
	}

	var result aliyunDescribeInstancesResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("解析阿里云 API 响应失败: %w, body=%s", err, string(body))
	}

	return &result, nil
}

// signAliyun computes the HMAC-SHA1 signature for Alibaba Cloud API (Signature Version 1.0).
func signAliyun(params url.Values, httpMethod, accessKeySecret string) string {
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var encodedParts []string
	for _, k := range keys {
		encodedParts = append(encodedParts, percentEncode(k)+"="+percentEncode(params.Get(k)))
	}
	canonicalQuery := strings.Join(encodedParts, "&")

	stringToSign := httpMethod + "&" + percentEncode("/") + "&" + percentEncode(canonicalQuery)

	mac := hmac.New(sha1.New, []byte(accessKeySecret+"&"))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// percentEncode performs the Alibaba Cloud specific percent-encoding.
func percentEncode(s string) string {
	encoded := url.QueryEscape(s)
	encoded = strings.ReplaceAll(encoded, "+", "%20")
	encoded = strings.ReplaceAll(encoded, "*", "%2A")
	encoded = strings.ReplaceAll(encoded, "%7E", "~")
	// Alibaba Cloud expects uppercase hex in percent encoding
	return specialPercentUpper(encoded)
}

func specialPercentUpper(s string) string {
	runes := []rune(s)
	for i := 0; i < len(runes)-2; i++ {
		if runes[i] == '%' {
			runes[i+1] = []rune(strings.ToUpper(string(runes[i+1])))[0]
			runes[i+2] = []rune(strings.ToUpper(string(runes[i+2])))[0]
		}
	}
	return string(runes)
}

func mapAliyunStatus(status string) string {
	switch strings.ToLower(status) {
	case "running":
		return "RUNNING"
	case "stopped":
		return "STOPPED"
	case "starting":
		return "STARTING"
	case "stopping":
		return "STOPPING"
	default:
		return strings.ToUpper(status)
	}
}
