package model

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	Hosts     HostStats     `json:"hosts"`
	Terminals TerminalStats `json:"terminals"`
	Cloud     CloudStats    `json:"cloud"`
	Files     FileStats     `json:"files"`
}

// HostStats 主机统计
type HostStats struct {
	Total   int64        `json:"total"`
	Online  int64        `json:"online"`
	Warning int64        `json:"warning"`
	Offline int64        `json:"offline"`
	Unknown int64        `json:"unknown"`
	ByGroup []GroupCount `json:"byGroup"`
}

// GroupCount 分组主机计数
type GroupCount struct {
	GroupID   uint   `json:"groupId"`
	GroupName string `json:"groupName"`
	Count     int64  `json:"count"`
}

// TerminalStats 终端统计
type TerminalStats struct {
	ActiveCount int64 `json:"activeCount"`
	TodayCount  int64 `json:"todayCount"`
	OnlineUsers int64 `json:"onlineUsers"`
}

// CloudStats 云资源统计
type CloudStats struct {
	InstanceCount int64  `json:"instanceCount"`
	LastSyncAt    string `json:"lastSyncAt"`
}

// FileStats 文件操作统计
type FileStats struct {
	TodayOps int64 `json:"todayOps"`
}

// ActivityEvent 活动事件
type ActivityEvent struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"`
	Message   string `json:"message"`
	User      string `json:"user"`
	Timestamp string `json:"timestamp"`
}

// MyHostInfo 我的常用主机
type MyHostInfo struct {
	ID           uint   `json:"id"`
	Hostname     string `json:"hostname"`
	IP           string `json:"ip"`
	Status       string `json:"status"`
	LastActiveAt string `json:"lastActiveAt"`
}

// DashboardResponse 仪表盘完整响应
type DashboardResponse struct {
	Stats    DashboardStats  `json:"stats"`
	Activity []ActivityEvent `json:"activity"`
	MyHosts  []MyHostInfo    `json:"myHosts"`
}
