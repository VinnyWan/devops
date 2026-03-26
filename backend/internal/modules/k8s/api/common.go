package api

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"devops-platform/internal/modules/k8s/service"
	"devops-platform/internal/pkg/k8s"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

var (
	k8sServiceInstance     *service.K8sService
	clusterServiceInstance *service.ClusterService
	k8sDB                  *gorm.DB
	clientFactory          *k8s.ClientFactory

	clusterOnce sync.Once
	k8sMu       sync.Mutex
)

// Response 标准响应结构
type Response struct {
	Code    int         `json:"code" example:"200"`
	Message string      `json:"message" example:"Success"`
	Data    interface{} `json:"data"`
}

// SetK8sDB 设置数据库实例和客户端工厂（在主程序初始化时调用）
func SetK8sDB(database *gorm.DB, factory *k8s.ClientFactory) {
	k8sMu.Lock()
	defer k8sMu.Unlock()
	k8sDB = database
	clientFactory = factory
	k8sServiceInstance = nil
	clusterServiceInstance = nil
	clusterOnce = sync.Once{}
}

// getK8sService 获取服务实例（延迟初始化）
func getK8sService() (*service.K8sService, error) {
	k8sMu.Lock()
	defer k8sMu.Unlock()

	if k8sServiceInstance != nil {
		return k8sServiceInstance, nil
	}
	if k8sDB == nil {
		return nil, errors.New("数据库未初始化，请先调用 SetK8sDB")
	}
	if clientFactory == nil {
		return nil, errors.New("客户端工厂未初始化，请先调用 SetK8sDB")
	}

	clusterSvc := getService()
	if clusterSvc == nil {
		return nil, errors.New("集群服务未初始化")
	}

	k8sServiceInstance = service.NewK8sService(clusterSvc, clientFactory)
	return k8sServiceInstance, nil
}

// K8sListRequest K8s 资源列表通用请求参数
type K8sListRequest struct {
	ClusterID uint   `form:"clusterId"`
	Namespace string `form:"namespace"`
	Page      int    `form:"page,default=1"`
	PageSize  int    `form:"pageSize,default=10"`
	Keyword   string `form:"keyword"`
}

// resolveListClusterID 从 K8sListRequest 解析集群ID，为0时回退到默认集群
func resolveListClusterID(c *gin.Context, id uint) (uint, error) {
	if id != 0 {
		return id, nil
	}
	return resolveClusterID(c)
}

func parseNamespaceAndAll(rawNamespace string, rawAll string) (string, bool, error) {
	namespace := strings.TrimSpace(rawNamespace)
	allStr := strings.TrimSpace(rawAll)

	if allStr != "" {
		all, err := strconv.ParseBool(allStr)
		if err != nil {
			return "", false, errors.New("无效的 all 参数")
		}
		if all {
			return "", true, nil
		}
		if namespace == "" {
			return "", false, errors.New("namespace 不能为空")
		}
		return namespace, false, nil
	}

	if strings.EqualFold(namespace, "all") {
		return "", true, nil
	}
	if namespace == "" {
		return "", true, nil
	}
	return namespace, false, nil
}

// handleK8sError 将 K8s API 错误映射为对应 HTTP 状态码
func handleK8sError(c *gin.Context, err error) {
	switch {
	case apierrors.IsNotFound(err):
		c.JSON(http.StatusNotFound, gin.H{"message": err.Error()})
	case apierrors.IsForbidden(err):
		c.JSON(http.StatusForbidden, gin.H{"message": err.Error()})
	case apierrors.IsConflict(err):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	case apierrors.IsUnauthorized(err):
		c.JSON(http.StatusUnauthorized, gin.H{"message": err.Error()})
	case apierrors.IsBadRequest(err):
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
	case apierrors.IsAlreadyExists(err):
		c.JSON(http.StatusConflict, gin.H{"message": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"message": err.Error()})
	}
}

func resolveClusterID(c *gin.Context) (uint, error) {
	raw := strings.TrimSpace(c.Query("clusterId"))
	if raw != "" {
		clusterId, err := strconv.ParseUint(raw, 10, 32)
		if err != nil {
			return 0, errors.New("无效的 clusterId")
		}
		return uint(clusterId), nil
	}

	clusterSvc := getService()
	if clusterSvc == nil {
		return 0, errors.New("集群服务未初始化")
	}

	cluster, err := clusterSvc.GetDefaultOrFirst()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("clusterId 不能为空且未配置默认集群")
		}
		return 0, err
	}
	return cluster.ID, nil
}
