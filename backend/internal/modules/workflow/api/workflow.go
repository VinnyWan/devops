package api

import (
	"net/http"
	"strconv"

	workflow "devops-platform/internal/modules/workflow/service"

	"github.com/gin-gonic/gin"
)

var workflowService *workflow.WorkflowService

func InitWorkflowService(svc *workflow.WorkflowService) {
	workflowService = svc
}

type createOrderReq struct {
	Title          string `json:"title" binding:"required"`
	Description    string `json:"description"`
	Type           string `json:"type" binding:"required"`
	ApprovalLevels int    `json:"approval_levels"`
}

type approvalReq struct {
	Comment string `json:"comment"`
}

func CreateOrder(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	userID := c.GetUint("userID")
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	order, err := workflowService.CreateOrder(tenantID, userID, req.Title, req.Description, req.Type, req.ApprovalLevels)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": order})
}

func SubmitOrder(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	id := parseID(c.Param("id"))
	if err := workflowService.SubmitForReview(id, tenantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已提交审批"})
}

func ApproveOrder(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	userID := c.GetUint("userID")
	id := parseID(c.Param("id"))
	var req approvalReq
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Comment = "同意"
	}
	if err := workflowService.Approve(id, tenantID, userID, req.Comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已批准"})
}

func RejectOrder(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	userID := c.GetUint("userID")
	id := parseID(c.Param("id"))
	var req approvalReq
	if err := c.ShouldBindJSON(&req); err != nil || req.Comment == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请填写驳回原因"})
		return
	}
	if err := workflowService.Reject(id, tenantID, userID, req.Comment); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已驳回"})
}

func ListOrders(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))
	status := c.Query("status")
	orders, total, err := workflowService.ListOrders(tenantID, page, pageSize, status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"total": total, "items": orders}})
}

func GetOrder(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	id := parseID(c.Param("id"))
	order, err := workflowService.GetOrder(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "工单不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": order})
}

func ExecuteOrder(c *gin.Context) {
	tenantID := c.GetUint("tenantID")
	id := parseID(c.Param("id"))
	if err := workflowService.ExecuteOrder(id, tenantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "执行已触发"})
}

func InstallWorkflowRoutes(r *gin.RouterGroup) {
	r.POST("/orders", CreateOrder)
	r.GET("/orders", ListOrders)
	r.GET("/orders/:id", GetOrder)
	r.POST("/orders/:id/submit", SubmitOrder)
	r.POST("/orders/:id/approve", ApproveOrder)
	r.POST("/orders/:id/reject", RejectOrder)
	r.POST("/orders/:id/execute", ExecuteOrder)
}
