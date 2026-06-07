package handler

import (
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/response"
	"ai_system_oncall/internal/service"
	"ai_system_oncall/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register handles user registration
// @Summary 用户注册
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "注册请求"
// @Success 200 {object} response.Response
// @Router /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam, "参数错误: "+err.Error())
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, dto.ToUserInfo(user))
}

// Login handles user login
// @Summary 用户登录
// @Tags 认证
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "登录请求"
// @Success 200 {object} response.Response{data=dto.LoginResponse}
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam, "参数错误: "+err.Error())
		return
	}

	loginResp, err := h.authService.Login(&req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, loginResp)
}

// GetCurrentUser gets current user info
// @Summary 获取当前用户信息
// @Tags 认证
// @Produce json
// @Success 200 {object} response.Response{data=dto.UserInfo}
// @Router /api/auth/me [get]
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		response.Fail(c, response.CodeUnauthorized, "未登录")
		return
	}

	userClaims := claims.(*jwt.Claims)
	userInfo, err := h.authService.GetCurrentUser(userClaims.UserID)
	if err != nil {
		response.Fail(c, 10004, "用户不存在")
		return
	}

	response.Success(c, userInfo)
}

// Logout handles user logout
// @Summary 用户退出
// @Tags 认证
// @Produce json
// @Success 200 {object} response.Response
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT is stateless, just return success
	response.SuccessWithMessage(c, "退出成功", nil)
}
