package service

import (
	"devops-platform/internal/modules/user/model"
	"devops-platform/internal/modules/user/repository"
	"fmt"
	"strings"
	"time"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/mssola/useragent"
	"gorm.io/gorm"
)

type LoginLogService struct {
	repo     *repository.LoginLogRepo
	searcher *xdb.Searcher
}

type LoginLogListRequest struct {
	Username string
	Status   string
	StartAt  string
	EndAt    string
	Page     int
	PageSize int
}

var loginLogIPData []byte

func NewLoginLogService(repo *repository.LoginLogRepo, db *gorm.DB) *LoginLogService {
	s := &LoginLogService{repo: repo}

	// 加载 ip2region 数据到内存
	if loginLogIPData == nil {
		data, err := xdb.LoadContentFromFile("ip2region.xdb")
		if err != nil {
			data, err = xdb.LoadContentFromFile("backend/ip2region.xdb")
			if err != nil {
				data, err = xdb.LoadContentFromFile("third_party/ip2region/ip2region.xdb")
				if err != nil {
					// IP 解析不可用，不影响核心功能
					return s
				}
			}
		}
		loginLogIPData = data
	}

	searcher, err := xdb.NewWithBuffer(xdb.IPvx, loginLogIPData)
	if err == nil {
		s.searcher = searcher
	}

	return s
}

// CreateLoginLog 创建登录日志（内部处理 IP 地理位置和 UA 解析）
func (s *LoginLogService) CreateLoginLog(username, ip, userAgentStr, status, message string) error {
	log := &model.LoginLog{
		Username:  username,
		IP:        ip,
		Location:  s.parseIPLocation(ip),
		Browser:   parseBrowser(userAgentStr),
		OS:        parseOS(userAgentStr),
		Status:    status,
		Message:   message,
		UserAgent: userAgentStr,
		LoginAt:   time.Now(),
	}
	return s.repo.Create(log)
}

// List 分页查询登录日志
func (s *LoginLogService) List(req LoginLogListRequest) ([]map[string]interface{}, int64, error) {
	query := repository.LoginLogQuery{
		Username: req.Username,
		Status:   req.Status,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	if req.StartAt != "" {
		startAt, err := time.Parse(time.RFC3339, req.StartAt)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid startAt format: %w", err)
		}
		query.StartAt = &startAt
	}
	if req.EndAt != "" {
		endAt, err := time.Parse(time.RFC3339, req.EndAt)
		if err != nil {
			return nil, 0, fmt.Errorf("invalid endAt format: %w", err)
		}
		query.EndAt = &endAt
	}

	logs, total, err := s.repo.List(query)
	if err != nil {
		return nil, 0, err
	}

	result := make([]map[string]interface{}, 0, len(logs))
	for _, item := range logs {
		result = append(result, formatLoginLog(item))
	}

	return result, total, nil
}

func (s *LoginLogService) parseIPLocation(ip string) string {
	if s.searcher == nil {
		return ""
	}
	region, err := s.searcher.Search(ip)
	if err != nil {
		return ""
	}
	// ip2region 返回格式: 国家|区域|省份|城市|ISP
	// 提取省份+城市，去掉空段
	parts := strings.Split(region, "|")
	var location []string
	for i, p := range parts {
		if i > 3 {
			break
		}
		if p != "" && p != "0" {
			location = append(location, p)
		}
	}
	if len(location) == 0 {
		return ""
	}
	return strings.Join(location, " ")
}

func parseBrowser(uaStr string) string {
	ua := useragent.New(uaStr)
	browserName, browserVersion := ua.Browser()
	if browserName == "" {
		return "Unknown"
	}
	if browserVersion != "" {
		return browserName + " " + browserVersion
	}
	return browserName
}

func parseOS(uaStr string) string {
	ua := useragent.New(uaStr)
	osInfo := ua.OS()
	if osInfo == "" {
		return "Unknown"
	}
	return osInfo
}

func formatLoginLog(item model.LoginLog) map[string]interface{} {
	return map[string]interface{}{
		"id":        item.ID,
		"username":  item.Username,
		"ip":        item.IP,
		"location":  item.Location,
		"browser":   item.Browser,
		"os":        item.OS,
		"status":    item.Status,
		"message":   item.Message,
		"userAgent": item.UserAgent,
		"loginAt":   item.LoginAt,
		"createdAt": item.CreatedAt,
	}
}
