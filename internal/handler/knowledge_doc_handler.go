package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/middleware"
	"ai_system_oncall/internal/response"
	"ai_system_oncall/internal/service"
)

// KnowledgeDocHandler 知识库文档处理器
type KnowledgeDocHandler struct {
	docService *service.KnowledgeDocService
}

// NewKnowledgeDocHandler 创建知识库文档处理器
func NewKnowledgeDocHandler(docService *service.KnowledgeDocService) *KnowledgeDocHandler {
	return &KnowledgeDocHandler{docService: docService}
}

// CreateDocument 创建文档
func (h *KnowledgeDocHandler) CreateDocument(c *gin.Context) {
	var req dto.CreateKnowledgeDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	info, err := h.docService.CreateDocument(userID, &req)
	if err != nil {
		response.Fail(c, 40302, err.Error())
		return
	}

	response.Success(c, info)
}

// ListDocuments 查询文档列表
func (h *KnowledgeDocHandler) ListDocuments(c *gin.Context) {
	var req dto.KnowledgeDocListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam, "参数错误: "+err.Error())
		return
	}

	result, err := h.docService.ListDocuments(&req)
	if err != nil {
		response.Fail(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, result)
}

// GetDocument 查询文档详情
func (h *KnowledgeDocHandler) GetDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的文档ID")
		return
	}

	info, err := h.docService.GetDocument(id)
	if err != nil {
		response.Fail(c, 40401, err.Error())
		return
	}

	response.Success(c, info)
}

// UpdateDocument 修改文档
func (h *KnowledgeDocHandler) UpdateDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的文档ID")
		return
	}

	var req dto.UpdateKnowledgeDocRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam, "参数错误: "+err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	info, err := h.docService.UpdateDocument(id, userID, &req)
	if err != nil {
		response.Fail(c, 40303, err.Error())
		return
	}

	response.Success(c, info)
}

// DeleteDocument 删除文档
func (h *KnowledgeDocHandler) DeleteDocument(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的文档ID")
		return
	}

	if err := h.docService.DeleteDocument(id); err != nil {
		response.Fail(c, 40304, err.Error())
		return
	}

	response.Success(c, nil)
}

// SearchDocuments 搜索文档
func (h *KnowledgeDocHandler) SearchDocuments(c *gin.Context) {
	keyword := c.Query("q")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if keyword == "" {
		response.Fail(c, response.CodeInvalidParam, "请提供搜索关键词")
		return
	}

	result, err := h.docService.SearchDocuments(keyword, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, result)
}

// GetDocumentsByProject 按项目查询
func (h *KnowledgeDocHandler) GetDocumentsByProject(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的项目ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.docService.GetDocumentsByProject(projectID, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, result)
}

// GetDocumentsByService 按服务查询
func (h *KnowledgeDocHandler) GetDocumentsByService(c *gin.Context) {
	serviceID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的服务ID")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.docService.GetDocumentsByService(serviceID, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, result)
}

// GetDocumentsByType 按类型查询
func (h *KnowledgeDocHandler) GetDocumentsByType(c *gin.Context) {
	docType := c.Param("type")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	result, err := h.docService.GetDocumentsByType(docType, page, pageSize)
	if err != nil {
		response.Fail(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, result)
}

// UpdateVectorStatus 更新向量化状态
func (h *KnowledgeDocHandler) UpdateVectorStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的文档ID")
		return
	}

	var req dto.UpdateVectorStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, response.CodeInvalidParam, "参数错误: "+err.Error())
		return
	}

	if err := h.docService.UpdateVectorStatus(id, req.VectorStatus); err != nil {
		response.Fail(c, 40301, err.Error())
		return
	}

	response.Success(c, nil)
}

// GetVersions 获取版本历史
func (h *KnowledgeDocHandler) GetVersions(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的文档ID")
		return
	}

	versions, err := h.docService.GetVersions(id)
	if err != nil {
		response.Fail(c, response.CodeInternalError, err.Error())
		return
	}

	response.Success(c, versions)
}

// TriggerVectorization 触发文档向量化
func (h *KnowledgeDocHandler) TriggerVectorization(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Fail(c, response.CodeInvalidParam, "无效的文档ID")
		return
	}

	// 异步执行向量化
	go func() {
		h.docService.TriggerVectorization(id)
	}()

	response.Success(c, gin.H{"message": "向量化任务已提交"})
}