package dto

import (
	"time"

	"ai_system_oncall/internal/constant"
	"ai_system_oncall/internal/model"
)

// ============ 请求 DTO ============

// CreateKnowledgeDocRequest 创建文档请求
type CreateKnowledgeDocRequest struct {
	Title         string  `json:"title" binding:"required,min=2,max=255"`
	Content       string  `json:"content" binding:"required"`
	DocType       string  `json:"doc_type" binding:"required"`
	ProjectID     *uint64 `json:"project_id"`
	ServiceID     *uint64 `json:"service_id"`
	ChangeSummary string  `json:"change_summary"` // 版本变更摘要
}

// UpdateKnowledgeDocRequest 更新文档请求
type UpdateKnowledgeDocRequest struct {
	Title         string `json:"title" binding:"required,min=2,max=255"`
	Content       string `json:"content" binding:"required"`
	DocType       string `json:"doc_type" binding:"required"`
	ChangeSummary string `json:"change_summary"`
}

// KnowledgeDocListRequest 查询文档列表请求
type KnowledgeDocListRequest struct {
	ProjectID *uint64 `form:"project_id"`
	ServiceID *uint64 `form:"service_id"`
	DocType   string  `form:"doc_type"`
	Keyword   string  `form:"keyword"`
	Page      int     `form:"page"`
	PageSize  int     `form:"page_size"`
}

// UpdateVectorStatusRequest 更新向量化状态请求
type UpdateVectorStatusRequest struct {
	VectorStatus string `json:"vector_status" binding:"required"`
}

// ============ 响应 DTO ============

// KnowledgeDocInfo 文档信息
type KnowledgeDocInfo struct {
	ID            uint64  `json:"id"`
	ProjectID     *uint64 `json:"project_id,omitempty"`
	ProjectName   string  `json:"project_name,omitempty"`
	ServiceID     *uint64 `json:"service_id,omitempty"`
	ServiceName   string  `json:"service_name,omitempty"`
	Title         string  `json:"title"`
	Content       string  `json:"content"`
	DocType       string  `json:"doc_type"`
	DocTypeName   string  `json:"doc_type_name"`
	CreatorID     uint64  `json:"creator_id"`
	CreatorName   string  `json:"creator_name"`
	UpdaterID     uint64  `json:"updater_id"`
	UpdaterName   string  `json:"updater_name"`
	VectorStatus  string  `json:"vector_status"`
	VectorStatusName string `json:"vector_status_name"`
	Version       int     `json:"version"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ToKnowledgeDocInfo 将模型转换为 DTO
func ToKnowledgeDocInfo(doc *model.KnowledgeDocument) *KnowledgeDocInfo {
	if doc == nil {
		return nil
	}
	info := &KnowledgeDocInfo{
		ID:           doc.ID,
		Title:        doc.Title,
		Content:      doc.Content,
		DocType:      doc.DocType,
		DocTypeName:  constant.GetDocTypeName(doc.DocType),
		CreatorID:    doc.CreatorID,
		UpdaterID:    doc.UpdaterID,
		VectorStatus: doc.VectorStatus,
		VectorStatusName: constant.GetVectorStatusName(doc.VectorStatus),
		Version:      doc.CurrentVersion,
		CreatedAt:    doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}
	
	if doc.ProjectID != nil {
		info.ProjectID = doc.ProjectID
	}
	if doc.Project != nil {
		info.ProjectName = doc.Project.Name
	}
	if doc.ServiceID != nil {
		info.ServiceID = doc.ServiceID
	}
	if doc.Service != nil {
		info.ServiceName = doc.Service.Name
	}
	if doc.Creator != nil {
		info.CreatorName = doc.Creator.Username
	}
	if doc.Updater != nil {
		info.UpdaterName = doc.Updater.Username
	}
	
	return info
}

// ============ 分页响应 ============

// KnowledgeDocListResponse 文档列表响应
type KnowledgeDocListResponse struct {
	List       []*KnowledgeDocInfo `json:"list"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}