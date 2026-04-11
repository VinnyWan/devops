package service

type MenuOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type AppMenuService struct{}

func NewAppMenuService() *AppMenuService {
	return &AppMenuService{}
}

func (s *AppMenuService) GetAppManagementMenuOptions() []MenuOption {
	return []MenuOption{
		{Label: "应用配置", Value: "app-config"},
		{Label: "构建配置", Value: "build-config"},
		{Label: "部署配置", Value: "deploy-config"},
		{Label: "容器配置", Value: "container-config"},
	}
}
