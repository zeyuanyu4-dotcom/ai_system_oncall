package handler

import (
	"strconv"

	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/middleware"
	"ai_system_oncall/internal/response"
	"ai_system_oncall/internal/service"

	"github.com/gin-gonic/gin"
)

type IssueHandler struct {
	issueService *service.IssueService
}

func NewIssueHandler(issueService *service.IssueService) *IssueHandler {
	return &IssueHandler{issueService: issueService}
}

// CreateIssue creates a new issue
func (h *IssueHandler) CreateIssue(c *gin.Context) {
	var req dto.CreateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	issueInfo, err := h.issueService.CreateIssue(userID, &req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, issueInfo)
}

// GetIssue gets an issue by ID
func (h *IssueHandler) GetIssue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	issueInfo, err := h.issueService.GetIssue(id)
	if err != nil {
		response.Fail(c, 10004, err.Error())
		return
	}

	response.Success(c, issueInfo)
}

// ListIssues lists issues
func (h *IssueHandler) ListIssues(c *gin.Context) {
	var req dto.IssueListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.issueService.ListIssues(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// UpdateIssue updates an issue
func (h *IssueHandler) UpdateIssue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	var req dto.UpdateIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	issueInfo, err := h.issueService.UpdateIssue(id, userID, &req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, issueInfo)
}

// DeleteIssue deletes an issue
func (h *IssueHandler) DeleteIssue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	if err := h.issueService.DeleteIssue(id); err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetOperationLogs gets operation logs of an issue
func (h *IssueHandler) GetOperationLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	logs, err := h.issueService.GetOperationLogs(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, logs)
}

// SearchHistoryIssues 搜索历史问题
func (h *IssueHandler) SearchHistoryIssues(c *gin.Context) {
	var req dto.HistoryIssueQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	result, err := h.issueService.SearchHistoryIssues(&req)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// GetSimilarIssues 获取相似问题推荐
func (h *IssueHandler) GetSimilarIssues(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))
	result, err := h.issueService.GetSimilarIssues(id, limit)
	if err != nil {
		response.Fail(c, 10004, err.Error())
		return
	}

	response.Success(c, result)
}

type StatusHandler struct {
	statusService *service.StatusService
}

func NewStatusHandler(statusService *service.StatusService) *StatusHandler {
	return &StatusHandler{statusService: statusService}
}

// AssignIssue assigns an issue to a user
func (h *StatusHandler) AssignIssue(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	var req dto.AssignIssueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	globalRole := middleware.GetUserRole(c)

	if err := h.statusService.AssignIssue(id, userID, req.AssigneeID, globalRole); err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, nil)
}

// ChangeStatus changes issue status
func (h *StatusHandler) ChangeStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	var req dto.ChangeStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	globalRole := middleware.GetUserRole(c)

	if err := h.statusService.ChangeStatus(id, userID, globalRole, req.Status, req.Reason); err != nil {
		response.Fail(c, 10005, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetStatusLogs gets status logs of an issue
func (h *StatusHandler) GetStatusLogs(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	logs, err := h.statusService.GetStatusLogs(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, logs)
}

type CommentHandler struct {
	commentService *service.CommentService
}

func NewCommentHandler(commentService *service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

// CreateComment creates a new comment
func (h *CommentHandler) CreateComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	var req dto.CreateCommentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	commentInfo, err := h.commentService.CreateComment(id, userID, &req)
	if err != nil {
		response.Fail(c, 10003, err.Error())
		return
	}

	response.Success(c, commentInfo)
}

// ListComments lists comments of an issue
func (h *CommentHandler) ListComments(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	commentType := c.Query("comment_type")
	result, err := h.commentService.ListComments(id, commentType)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, result)
}

// DeleteComment deletes a comment
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论ID")
		return
	}

	userID := middleware.GetUserID(c)
	globalRole := middleware.GetUserRole(c)

	if err := h.commentService.DeleteComment(id, userID, globalRole); err != nil {
		response.Fail(c, 10002, err.Error())
		return
	}

	response.Success(c, nil)
}
