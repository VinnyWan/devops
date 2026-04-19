package repository

import (
	"time"

	"devops-platform/internal/modules/cmdb/model"

	"gorm.io/gorm"
)

type DashboardRepo struct {
	db *gorm.DB
}

func NewDashboardRepo(db *gorm.DB) *DashboardRepo {
	return &DashboardRepo{db: db}
}

// GetHostStatusCounts 按状态统计主机数量
func (r *DashboardRepo) GetHostStatusCounts(tenantID uint) (map[string]int64, error) {
	type result struct {
		Status string
		Count  int64
	}
	var results []result
	err := r.db.Model(&model.Host{}).
		Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Select("status, count(*) as count").
		Group("status").
		Find(&results).Error
	if err != nil {
		return nil, err
	}
	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}

// GetHostCountByGroup 按分组统计主机数量
func (r *DashboardRepo) GetHostCountByGroup(tenantID uint, limit int) ([]model.GroupCount, error) {
	var counts []model.GroupCount
	err := r.db.Model(&model.Host{}).
		Where("hosts.tenant_id = ? AND hosts.deleted_at IS NULL", tenantID).
		Joins("LEFT JOIN host_groups ON hosts.group_id = host_groups.id").
		Select("COALESCE(host_groups.id, 0) as group_id, COALESCE(host_groups.name, '未分组') as group_name, count(*) as count").
		Group("host_groups.id, host_groups.name").
		Order("count DESC").
		Limit(limit).
		Find(&counts).Error
	return counts, err
}

// GetActiveTerminalCount 获取当前活跃终端会话数
func (r *DashboardRepo) GetActiveTerminalCount(tenantID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.TerminalSession{}).
		Where("tenant_id = ? AND status = 'active' AND deleted_at IS NULL", tenantID).
		Count(&count).Error
	return count, err
}

// GetTodayTerminalCount 获取今日终端会话数
func (r *DashboardRepo) GetTodayTerminalCount(tenantID uint) (int64, error) {
	today := time.Now().Truncate(24 * time.Hour)
	var count int64
	err := r.db.Model(&model.TerminalSession{}).
		Where("tenant_id = ? AND started_at >= ? AND deleted_at IS NULL", tenantID, today).
		Count(&count).Error
	return count, err
}

// GetOnlineTerminalUsers 获取当前在线终端用户数
func (r *DashboardRepo) GetOnlineTerminalUsers(tenantID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.TerminalSession{}).
		Where("tenant_id = ? AND status = 'active' AND deleted_at IS NULL", tenantID).
		Distinct("user_id").
		Count(&count).Error
	return count, err
}

// GetCloudInstanceCount 获取云资源实例数
func (r *DashboardRepo) GetCloudInstanceCount(tenantID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.CloudResource{}).
		Where("tenant_id = ? AND resource_type = 'cvm' AND deleted_at IS NULL", tenantID).
		Count(&count).Error
	return count, err
}

// GetLastCloudSyncAt 获取最后一次云同步时间
func (r *DashboardRepo) GetLastCloudSyncAt(tenantID uint) (string, error) {
	var account model.CloudAccount
	err := r.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Select("last_sync_at").
		Order("last_sync_at DESC").
		First(&account).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	if account.LastSyncAt == nil {
		return "", nil
	}
	return account.LastSyncAt.Format("2006-01-02 15:04:05"), nil
}

// GetTodayFileOps 获取今日文件操作数
func (r *DashboardRepo) GetTodayFileOps(tenantID uint) (int64, error) {
	today := time.Now().Truncate(24 * time.Hour)
	var count int64
	err := r.db.Model(&model.FileOperationLog{}).
		Where("tenant_id = ? AND created_at >= ? AND deleted_at IS NULL", tenantID, today).
		Count(&count).Error
	return count, err
}

// GetRecentTerminalActivity 获取最近终端活动
func (r *DashboardRepo) GetRecentTerminalActivity(tenantID uint, limit int) ([]model.ActivityEvent, error) {
	var sessions []model.TerminalSession
	err := r.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("started_at DESC").
		Limit(limit).
		Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	events := make([]model.ActivityEvent, 0, len(sessions))
	for _, s := range sessions {
		msg := s.Username + " 连接到 " + s.HostName + " (" + s.HostIP + ")"
		if s.Status != "active" {
			msg = s.Username + " 断开 " + s.HostName + " (" + s.HostIP + ")"
		}
		events = append(events, model.ActivityEvent{
			ID:        s.ID,
			Type:      "terminal",
			Message:   msg,
			User:      s.Username,
			Timestamp: s.StartedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return events, nil
}

// GetRecentFileActivity 获取最近文件操作活动
func (r *DashboardRepo) GetRecentFileActivity(tenantID uint, limit int) ([]model.ActivityEvent, error) {
	var logs []model.FileOperationLog
	err := r.db.Where("tenant_id = ? AND deleted_at IS NULL", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	if err != nil {
		return nil, err
	}
	events := make([]model.ActivityEvent, 0, len(logs))
	for _, l := range logs {
		events = append(events, model.ActivityEvent{
			ID:        l.ID,
			Type:      "file",
			Message:   l.Username + " " + l.OpType + " " + l.FilePath + " on " + l.HostName,
			User:      l.Username,
			Timestamp: l.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	return events, nil
}

// GetMyHosts 获取用户最近访问的主机
func (r *DashboardRepo) GetMyHosts(tenantID, userID uint, limit int) ([]model.MyHostInfo, error) {
	type recentSession struct {
		HostID    uint
		StartedAt time.Time
	}
	var recent []recentSession
	err := r.db.Model(&model.TerminalSession{}).
		Select("host_id, MAX(started_at) as started_at").
		Where("tenant_id = ? AND user_id = ? AND deleted_at IS NULL", tenantID, userID).
		Group("host_id").
		Order("started_at DESC").
		Limit(limit).
		Find(&recent).Error
	if err != nil {
		return nil, err
	}
	if len(recent) == 0 {
		return []model.MyHostInfo{}, nil
	}

	hostIDs := make([]uint, len(recent))
	sessionMap := make(map[uint]time.Time)
	for i, rs := range recent {
		hostIDs[i] = rs.HostID
		sessionMap[rs.HostID] = rs.StartedAt
	}

	var hosts []model.Host
	err = r.db.Where("id IN ? AND deleted_at IS NULL", hostIDs).Find(&hosts).Error
	if err != nil {
		return nil, err
	}

	result := make([]model.MyHostInfo, 0, len(hosts))
	for _, h := range hosts {
		lastActive := ""
		if t, ok := sessionMap[h.ID]; ok {
			lastActive = t.Format("2006-01-02 15:04:05")
		}
		result = append(result, model.MyHostInfo{
			ID:           h.ID,
			Hostname:     h.Hostname,
			IP:           h.Ip,
			Status:       h.Status,
			LastActiveAt: lastActive,
		})
	}
	return result, nil
}
