package handler

import (
	"strconv"

	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/response"
	"ai_system_oncall/internal/service"
	"ai_system_oncall/pkg/jwt"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// ListUsers lists users with pagination
// @Summary 用户列表
// @Tags 用户
// @Produce json
// @Param page query int false "页码"
// @Param page_size query int false "每页数量"
// @Param keyword query string false "关键词"
// @Param role query string false "角色"
// @Param status query int false "状态"
// @Success 200 {object} response.Response{data=dto.UserListResponse}
// @Router /api/users [get]
func (h *UserHandler) ListUsers(c *gin.Context) {
	var req dto.UserListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam, "参数错误: "+err.Error())
		return
	}

	result, err := h.userService.ListUsers(&req)
	if err != nil {
		response.Fail(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, result)
}

// GetUser gets a user by ID
// @Summary 用户详情
// @Tags 用户
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response{data=dto.UserInfo}
// @Router /api/users/{id} [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的用户ID")
		return
	}

	userInfo, err := h.userService.GetUserByID(id)
	if err != nil {
		response.Fail(c, 10004, err.Error())
		return
	}

	response.Success(c, userInfo)
}

// UpdateUser updates a user
// @Summary 更新用户
// @Tags 用户
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body dto.UpdateUserRequest true "更新请求"
// @Success 200 {object} response.Response{data=dto.UserInfo}
// @Router /api/users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的用户ID")
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam, "参数错误: "+err.Error())
		return
	}

	// Check permission: only self or system admin can update
	claims, _ := c.Get("claims")
	userClaims := claims.(*jwt.Claims)
	if userClaims.UserID != id && userClaims.Role != "system_admin" {
		response.Fail(c, response.CodeForbidden, "无权限修改该用户")
		return
	}

	userInfo, err := h.userService.UpdateUser(id, &req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, userInfo)
}

// UpdateUserStatus updates a user's status
// @Summary 更新用户状态
// @Tags 用户
// @Accept json
// @Produce json
// @Param id path int true "用户ID"
// @Param request body dto.UpdateUserStatusRequest true "状态请求"
// @Success 200 {object} response.Response
// @Router /api/users/{id}/status [patch]
func (h *UserHandler) UpdateUserStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的用户ID")
		return
	}

	var req dto.UpdateUserStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam, "参数错误: "+err.Error())
		return
	}

	claims, _ := c.Get("claims")
	userClaims := claims.(*jwt.Claims)

	if err := h.userService.UpdateUserStatus(userClaims.Role, id, req.Status); err != nil {
		response.Fail(c, 10002, err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteUser deletes a user
// @Summary 删除用户
// @Tags 用户
// @Produce json
// @Param id path int true "用户ID"
// @Success 200 {object} response.Response
// @Router /api/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的用户ID")
		return
	}

	claims, _ := c.Get("claims")
	userClaims := claims.(*jwt.Claims)

	if err := h.userService.DeleteUser(userClaims.Role, id); err != nil {
		response.Fail(c, 10002, err.Error())
		return
	}

	response.Success(c, nil)
}
