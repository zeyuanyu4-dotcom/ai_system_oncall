package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ai_system_oncall/internal/middleware"
	"ai_system_oncall/internal/response"
	"ai_system_oncall/internal/service"
)

// KnowledgeDocAttachmentHandler 文档附件处理器
type KnowledgeDocAttachmentHandler struct {
	attachmentService *service.KnowledgeDocAttachmentService
}

// NewKnowledgeDocAttachmentHandler 创建文档附件处理器
func NewKnowledgeDocAttachmentHandler(attachmentService *service.KnowledgeDocAttachmentService) *KnowledgeDocAttachmentHandler {
	return &KnowledgeDocAttachmentHandler{attachmentService: attachmentService}
}

// UploadAttachment 上传附件
func (h *KnowledgeDocAttachmentHandler) UploadAttachment(c *gin.Context) {
	documentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文档ID")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	userID := middleware.GetUserID(c)
	info, err := h.attachmentService.UploadAttachment(documentID, userID, file)
	if err != nil {
		response.Fail(c, 50001, err.Error())
		return
	}

	response.Success(c, info)
}

// GetAttachments 获取附件列表
func (h *KnowledgeDocAttachmentHandler) GetAttachments(c *gin.Context) {
	documentID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的文档ID")
		return
	}

	attachments, err := h.attachmentService.GetAttachments(documentID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, attachments)
}

// DownloadAttachment 下载附件
func (h *KnowledgeDocAttachmentHandler) DownloadAttachment(c *gin.Context) {
	attachmentID, err := strconv.ParseUint(c.Param("aid"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的附件ID")
		return
	}

	filePath, fileName, err := h.attachmentService.GetAttachmentPath(attachmentID)
	if err != nil {
		response.Fail(c, 40401, err.Error())
		return
	}

	c.FileAttachment(filePath, fileName)
}

// GetAttachmentContent 获取附件内容（供 Agent 读取）
func (h *KnowledgeDocAttachmentHandler) GetAttachmentContent(c *gin.Context) {
	attachmentID, err := strconv.ParseUint(c.Param("aid"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的附件ID")
		return
	}

	content, fileType, err := h.attachmentService.ReadAttachmentContent(attachmentID)
	if err != nil {
		response.Fail(c, 40401, err.Error())
		return
	}

	response.Success(c, gin.H{
		"content":   content,
		"file_type": fileType,
	})
}

// DeleteAttachment 删除附件
func (h *KnowledgeDocAttachmentHandler) DeleteAttachment(c *gin.Context) {
	attachmentID, err := strconv.ParseUint(c.Param("aid"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的附件ID")
		return
	}

	if err := h.attachmentService.DeleteAttachment(attachmentID); err != nil {
		response.Fail(c, 50001, err.Error())
		return
	}

	response.Success(c, nil)
}

// ParseAttachmentToContent 解析附件生成文档内容
func (h *KnowledgeDocAttachmentHandler) ParseAttachmentToContent(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}

	content, err := h.attachmentService.ParseAttachmentToContent(file)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, gin.H{"content": content})
}