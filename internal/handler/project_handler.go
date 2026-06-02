package handler

import (
	"strconv"

	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/middleware"
	"ai_system_oncall/internal/response"
	"ai_system_oncall/internal/service"

	"github.com/gin-gonic/gin"
)

type ProjectHandler struct {
	projectService *service.ProjectService
}

func NewProjectHandler(projectService *service.ProjectService) *ProjectHandler {
	return &ProjectHandler{projectService: projectService}
}

// CreateProject creates a new project
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	var req dto.CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	projectInfo, err := h.projectService.CreateProject(userID, &req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, projectInfo)
}

// GetProject gets a project by ID
func (h *ProjectHandler) GetProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的项目ID")
		return
	}

	projectInfo, err := h.projectService.GetProject(id)
	if err != nil {
		response.Fail(c, 10004, err.Error())
		return
	}

	response.Success(c, projectInfo)
}

// ListProjects lists projects
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	var req dto.ProjectListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.projectService.ListProjects(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// UpdateProject updates a project
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的项目ID")
		return
	}

	var req dto.UpdateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	projectInfo, err := h.projectService.UpdateProject(id, &req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, projectInfo)
}

// DeleteProject deletes a project
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的项目ID")
		return
	}

	if err := h.projectService.DeleteProject(id); err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, nil)
}

type ProjectMemberHandler struct {
	memberService *service.ProjectMemberService
}

func NewProjectMemberHandler(memberService *service.ProjectMemberService) *ProjectMemberHandler {
	return &ProjectMemberHandler{memberService: memberService}
}

// AddMember adds a member to a project
func (h *ProjectMemberHandler) AddMember(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的项目ID")
		return
	}

	var req dto.AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.memberService.AddMember(projectID, userID, &req); err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, nil)
}

// ListMembers lists members of a project
func (h *ProjectMemberHandler) ListMembers(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的项目ID")
		return
	}

	result, err := h.memberService.ListMembers(projectID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// UpdateMemberRole updates a member's role
func (h *ProjectMemberHandler) UpdateMemberRole(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的项目ID")
		return
	}

	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	var req dto.UpdateMemberRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.memberService.UpdateMemberRole(projectID, userID, &req); err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, nil)
}

// RemoveMember removes a member from a project
func (h *ProjectMemberHandler) RemoveMember(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的项目ID")
		return
	}

	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的用户ID")
		return
	}

	if err := h.memberService.RemoveMember(projectID, userID); err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, nil)
}
