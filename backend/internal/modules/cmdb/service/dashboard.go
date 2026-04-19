package service

import (
	"sort"

	"devops-platform/internal/modules/cmdb/model"
	"devops-platform/internal/modules/cmdb/repository"
)

type DashboardService struct {
	repo *repository.DashboardRepo
}

func NewDashboardService(repo *repository.DashboardRepo) *DashboardService {
	return &DashboardService{repo: repo}
}

const dashboardActivityLimit = 10
const dashboardMyHostsLimit = 8

func (s *DashboardService) GetDashboard(tenantID, userID uint) (*model.DashboardResponse, error) {
	stats, err := s.getStats(tenantID)
	if err != nil {
		return nil, err
	}

	activity, err := s.getActivity(tenantID)
	if err != nil {
		return nil, err
	}

	myHosts, err := s.repo.GetMyHosts(tenantID, userID, dashboardMyHostsLimit)
	if err != nil {
		return nil, err
	}

	return &model.DashboardResponse{
		Stats:    *stats,
		Activity: activity,
		MyHosts:  myHosts,
	}, nil
}

func (s *DashboardService) getStats(tenantID uint) (*model.DashboardStats, error) {
	hostCounts, err := s.repo.GetHostStatusCounts(tenantID)
	if err != nil {
		return nil, err
	}

	hostByGroup, err := s.repo.GetHostCountByGroup(tenantID, 10)
	if err != nil {
		return nil, err
	}

	activeTerminals, err := s.repo.GetActiveTerminalCount(tenantID)
	if err != nil {
		return nil, err
	}

	todayTerminals, err := s.repo.GetTodayTerminalCount(tenantID)
	if err != nil {
		return nil, err
	}

	onlineUsers, err := s.repo.GetOnlineTerminalUsers(tenantID)
	if err != nil {
		return nil, err
	}

	cloudInstances, err := s.repo.GetCloudInstanceCount(tenantID)
	if err != nil {
		return nil, err
	}

	lastSyncAt, err := s.repo.GetLastCloudSyncAt(tenantID)
	if err != nil {
		return nil, err
	}

	todayFileOps, err := s.repo.GetTodayFileOps(tenantID)
	if err != nil {
		return nil, err
	}

	return &model.DashboardStats{
		Hosts: model.HostStats{
			Total:   hostCounts["online"] + hostCounts["offline"] + hostCounts["warning"] + hostCounts["unknown"],
			Online:  hostCounts["online"],
			Warning: hostCounts["warning"],
			Offline: hostCounts["offline"],
			Unknown: hostCounts["unknown"],
			ByGroup: hostByGroup,
		},
		Terminals: model.TerminalStats{
			ActiveCount: activeTerminals,
			TodayCount:  todayTerminals,
			OnlineUsers: onlineUsers,
		},
		Cloud: model.CloudStats{
			InstanceCount: cloudInstances,
			LastSyncAt:    lastSyncAt,
		},
		Files: model.FileStats{
			TodayOps: todayFileOps,
		},
	}, nil
}

func (s *DashboardService) getActivity(tenantID uint) ([]model.ActivityEvent, error) {
	terminalEvents, err := s.repo.GetRecentTerminalActivity(tenantID, dashboardActivityLimit)
	if err != nil {
		return nil, err
	}

	fileEvents, err := s.repo.GetRecentFileActivity(tenantID, dashboardActivityLimit)
	if err != nil {
		return nil, err
	}

	all := append(terminalEvents, fileEvents...)
	sort.Slice(all, func(i, j int) bool {
		return all[i].Timestamp > all[j].Timestamp
	})
	if len(all) > dashboardActivityLimit {
		all = all[:dashboardActivityLimit]
	}
	return all, nil
}
