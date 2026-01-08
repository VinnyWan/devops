package user

import (
	"devops/common"
	userservice "devops/service/user"

	"github.com/dchest/captcha"
	"github.com/gin-gonic/gin"
)

type CaptchaController struct {
	captchaService *userservice.CaptchaService
}

func NewCaptchaController() *CaptchaController {
	return &CaptchaController{
		captchaService: &userservice.CaptchaService{},
	}
}

// Generate 生成验证码
// @Summary 生成验证码
// @Description 生成验证码ID和图片URL，验证码有效期5分钟，存储在Redis中
// @Tags 认证管理
// @Accept json
// @Produce json
// @Success 200 {object} common.Response{data=CaptchaResponse} "成功生成验证码"
// @Router /api/captcha [get]
func (ctrl *CaptchaController) Generate(c *gin.Context) {
	// 使用Service生成验证码（会自动存储到Redis）
	id, err := ctrl.captchaService.Generate()
	if err != nil {
		common.Fail(c, "生成验证码失败")
		return
	}

	common.Success(c, gin.H{
		"captchaId": id,
		"imageUrl":  "/api/captcha/" + id + ".png",
	})
}

// Serve 提供验证码图片
func (ctrl *CaptchaController) Serve(c *gin.Context) {
	captcha.Server(captcha.StdWidth, captcha.StdHeight).ServeHTTP(c.Writer, c.Request)
}

// CaptchaResponse 验证码响应
type CaptchaResponse struct {
	CaptchaID string `json:"captchaId" example:"xxxx"`
	ImageURL  string `json:"imageUrl" example:"/api/captcha/xxxx.png"`
}
