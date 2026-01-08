package user

import (
	"devops/common"
	"devops/common/config"
	"devops/middleware"
	usermodels "devops/models/user"
	userservice "devops/service/user"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *userservice.UserService
}

func NewUserController() *UserController {
	return &UserController{
		userService: &userservice.UserService{},
	}
}

// Login 用户登录
// @Summary 用户登录
// @Description 用户登录接口，登录成功后返回JWT Token。在Swagger中登录后，点击右上角Authorize按钮，输入 "Bearer <token>" 即可测试需要认证的接口
// @Tags 认证管理
// @Accept json
// @Produce json
// @Param body body LoginRequest true "登录信息"
// @Success 200 {object} common.Response{data=LoginResponse} "登录成功，返回token和用户信息"
// @Failure 500 {object} common.Response "登录失败"
// @Router /api/auth/login [post]
func (ctrl *UserController) Login(c *gin.Context) {
	var req LoginRequest
	// 支持JSON和URL参数两种方式
	if err := c.ShouldBind(&req); err != nil {
		common.Fail(c, "参数错误: "+err.Error())
		return
	}

	// 验证码校验（如果启用）
	captchaConfig := config.GetCaptchaConfig()
	if captchaConfig.Enabled {
		captchaService := &userservice.CaptchaService{}
		if !captchaService.Verify(req.CaptchaID, req.CaptchaCode) {
			common.Fail(c, "验证码错误")
			return
		}
	}

	user, token, err := ctrl.userService.Login(req.Username, req.Password)
	if err != nil {
		common.Fail(c, err.Error())
		return
	}

	common.Success(c, gin.H{
		"token": token,
		"user":  user,
	})
}

// GetInfo 获取当前用户信息
// @Summary 获取当前用户信息
// @Description 获取当前登录用户的详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} common.Response
// @Router /api/user/info [get]
func (ctrl *UserController) GetInfo(c *gin.Context) {
	userID := middleware.GetUserID(c)
	user, err := ctrl.userService.GetByID(userID)
	if err != nil {
		common.Fail(c, "获取用户信息失败")
		return
	}

	common.Success(c, user)
}

// Create 创建用户
// @Summary 创建用户
// @Description 创建新用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param body body usermodels.User true "用户信息"
// @Success 200 {object} common.Response
// @Router /api/users [post]
func (ctrl *UserController) Create(c *gin.Context) {
	var user usermodels.User
	if err := c.ShouldBindJSON(&user); err != nil {
		common.Fail(c, "参数错误")
		return
	}

	if err := ctrl.userService.Create(&user); err != nil {
		common.Fail(c, "创建用户失败:"+err.Error())
		return
	}

	common.Success(c, user)
}

// Update 更新用户
// @Summary 更新用户
// @Description 根据ID更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Param body body usermodels.User true "用户信息"
// @Success 200 {object} common.Response
// @Router /api/users/{id} [put]
func (ctrl *UserController) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var user usermodels.User
	if err := c.ShouldBindJSON(&user); err != nil {
		common.Fail(c, "参数错误")
		return
	}

	if err := ctrl.userService.Update(uint(id), &user); err != nil {
		common.Fail(c, "更新用户失败:"+err.Error())
		return
	}

	common.Success(c, nil)
}

// UpdateByUsername 根据用户名更新用户
// @Summary 根据用户名更新用户
// @Description 根据用户名更新用户信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param username path string true "用户名"
// @Param body body usermodels.User true "用户信息"
// @Success 200 {object} common.Response
// @Router /api/users/username/{username} [put]
func (ctrl *UserController) UpdateByUsername(c *gin.Context) {
	username := c.Param("username")
	var user usermodels.User
	if err := c.ShouldBindJSON(&user); err != nil {
		common.Fail(c, "参数错误")
		return
	}

	if err := ctrl.userService.UpdateByUsername(username, &user); err != nil {
		common.Fail(c, "更新用户失败:"+err.Error())
		return
	}

	common.Success(c, nil)
}

// Delete 删除用户
// @Summary 删除用户
// @Description 根据ID删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Success 200 {object} common.Response
// @Router /api/users/{id} [delete]
func (ctrl *UserController) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	if err := ctrl.userService.Delete(uint(id)); err != nil {
		common.Fail(c, "删除用户失败")
		return
	}

	common.Success(c, nil)
}

// DeleteByUsername 根据用户名删除用户
// @Summary 根据用户名删除用户
// @Description 根据用户名删除用户
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param username path string true "用户名"
// @Success 200 {object} common.Response
// @Router /api/users/username/{username} [delete]
func (ctrl *UserController) DeleteByUsername(c *gin.Context) {
	username := c.Param("username")
	if err := ctrl.userService.DeleteByUsername(username); err != nil {
		common.Fail(c, "删除用户失败")
		return
	}

	common.Success(c, nil)
}

// GetByID 根据ID获取用户
// @Summary 根据ID获取用户
// @Description 根据用户ID获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Success 200 {object} common.Response
// @Router /api/users/{id} [get]
func (ctrl *UserController) GetByID(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	user, err := ctrl.userService.GetByID(uint(id))
	if err != nil {
		common.Fail(c, "用户不存在")
		return
	}

	common.Success(c, user)
}

// GetByUsername 根据用户名获取用户
// @Summary 根据用户名获取用户
// @Description 根据用户名获取用户详细信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param username path string true "用户名"
// @Success 200 {object} common.Response
// @Router /api/users/username/{username} [get]
func (ctrl *UserController) GetByUsername(c *gin.Context) {
	username := c.Param("username")
	user, err := ctrl.userService.GetByUsername(username)
	if err != nil {
		common.Fail(c, "用户不存在")
		return
	}

	common.Success(c, user)
}

// GetList 获取用户列表
// @Summary 获取用户列表
// @Description 分页查询用户列表
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param page query int false "页码" default(1)
// @Param pageSize query int false "每页数量" default(10)
// @Param username query string false "用户名"
// @Param phone query string false "手机号"
// @Param status query string false "状态"
// @Success 200 {object} common.Response
// @Router /api/users [get]
func (ctrl *UserController) GetList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	username := c.Query("username")
	phone := c.Query("phone")
	status := c.Query("status")

	users, total, err := ctrl.userService.GetList(page, pageSize, username, phone, status)
	if err != nil {
		common.Fail(c, "获取用户列表失败")
		return
	}

	common.PageSuccess(c, users, total, page, pageSize)
}

// AssignRoles 分配角色
// @Summary 分配角色
// @Description 为用户分配角色
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "用户ID"
// @Param body body AssignRolesRequest true "角色ID列表"
// @Success 200 {object} common.Response
// @Router /api/users/{id}/roles [post]
func (ctrl *UserController) AssignRoles(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 32)
	var req AssignRolesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.Fail(c, "参数错误")
		return
	}

	if err := ctrl.userService.AssignRoles(uint(id), req.RoleIDs); err != nil {
		common.Fail(c, "分配角色失败")
		return
	}

	common.Success(c, nil)
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username    string `json:"username" form:"username" binding:"required" example:"admin"`
	Password    string `json:"password" form:"password" binding:"required" example:"admin123"`
	CaptchaID   string `json:"captchaId" form:"captchaId" binding:"required" example:"xxxx"`
	CaptchaCode string `json:"captchaCode" form:"captchaCode" binding:"required" example:"1234"`
}

// LoginResponse 登录响应
type LoginResponse struct {
	Token string      `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  interface{} `json:"user"`
}

// AssignRolesRequest 分配角色请求
type AssignRolesRequest struct {
	RoleIDs []uint `json:"roleIds" binding:"required"`
}
