package api

import (
	"net/http"

	"devops-platform/internal/modules/task/model"
	"devops-platform/internal/modules/task/service"

	"github.com/gin-gonic/gin"
)

var taskService *service.TaskService

func InitTaskService(svc *service.TaskService) {
	taskService = svc
}

type createTaskReq struct {
	Name        string `json:"name" binding:"required,max=255"`
	Description string `json:"description"`
	Type        string `json:"type" binding:"required,oneof=shell python ansible"`
	Content     string `json:"content" binding:"required"`
	Timeout     int    `json:"timeout"`
}

type updateTaskReq struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Content     string `json:"content"`
	Timeout     int    `json:"timeout"`
}

type executeTaskReq struct {
	Targets []string `json:"targets"`
}

type setScheduleReq struct {
	CronExpr string `json:"cron_expr" binding:"required"`
}

func getTenantAndUser(c *gin.Context) (uint, uint) {
	return c.GetUint("tenantID"), c.GetUint("userID")
}

func CreateTask(c *gin.Context) {
	tenantID, userID := getTenantAndUser(c)
	var req createTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	t, err := taskService.Create(tenantID, userID, req.Name, req.Description, model.TaskType(req.Type), req.Content, req.Timeout)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": t})
}

func UpdateTask(c *gin.Context) {
	tenantID, _ := getTenantAndUser(c)
	id := parseID(c.Param("id"))
	var req updateTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}
	if err := taskService.Update(id, tenantID, req.Name, req.Description, req.Content, req.Timeout); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}

func DeleteTask(c *gin.Context) {
	tenantID, _ := getTenantAndUser(c)
	id := parseID(c.Param("id"))
	if err := taskService.Delete(id, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已删除"})
}

func GetTask(c *gin.Context) {
	tenantID, _ := getTenantAndUser(c)
	id := parseID(c.Param("id"))
	t, err := taskService.Get(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "任务不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": t})
}

func ListTasks(c *gin.Context) {
	tenantID, _ := getTenantAndUser(c)
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 20)
	keyword := c.Query("keyword")
	tasks, total, err := taskService.List(tenantID, keyword, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"total": total, "items": tasks}})
}

func ExecuteTask(c *gin.Context) {
	tenantID, _ := getTenantAndUser(c)
	id := parseID(c.Param("id"))
	var req executeTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Targets = []string{}
	}
	exec, err := taskService.Execute(c.Request.Context(), id, tenantID, req.Targets)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": exec})
}

func GetExecution(c *gin.Context) {
	tenantID, _ := getTenantAndUser(c)
	id := parseID(c.Param("execId"))
	exec, err := taskService.GetExecution(id, tenantID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "执行记录不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": exec})
}

func ListExecutions(c *gin.Context) {
	tenantID, _ := getTenantAndUser(c)
	taskID := parseID(c.Param("id"))
	page := queryInt(c, "page", 1)
	pageSize := queryInt(c, "pageSize", 20)
	execs, total, err := taskService.ListExecutions(taskID, tenantID, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"total": total, "items": execs}})
}

func SetSchedule(c *gin.Context) {
	tenantID, _ := getTenantAndUser(c)
	id := parseID(c.Param("id"))
	var req setScheduleReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if err := taskService.SetSchedule(id, tenantID, req.CronExpr); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "调度已设置"})
}

func DeleteSchedule(c *gin.Context) {
	tenantID, _ := getTenantAndUser(c)
	id := parseID(c.Param("id"))
	if err := taskService.DeleteSchedule(id, tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "调度已删除"})
}

func InstallTaskRoutes(r *gin.RouterGroup) {
	r.POST("/tasks", CreateTask)
	r.GET("/tasks", ListTasks)
	r.GET("/tasks/:id", GetTask)
	r.PUT("/tasks/:id", UpdateTask)
	r.DELETE("/tasks/:id", DeleteTask)
	r.POST("/tasks/:id/execute", ExecuteTask)
	r.GET("/tasks/:id/executions", ListExecutions)
	r.GET("/tasks/:id/executions/:execId", GetExecution)
	r.POST("/tasks/:id/schedule", SetSchedule)
	r.DELETE("/tasks/:id/schedule", DeleteSchedule)
}
