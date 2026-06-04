package dto

import (
	"time"

	"ai_system_oncall/internal/constant"
	"ai_system_oncall/internal/model"
)

// KnowledgeDocVersionInfo 文档版本信息
type KnowledgeDocVersionInfo struct {
	ID            uint64    `json:"id"`
	DocumentID    uint64    `json:"document_id"`
	Version       int       `json:"version"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	DocType       string    `json:"doc_type"`
	DocTypeName   string    `json:"doc_type_name"`
	EditorID      uint64    `json:"editor_id"`
	EditorName    string    `json:"editor_name"`
	ChangeSummary string    `json:"change_summary"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToKnowledgeDocVersionInfo 将模型转换为 DTO
func ToKnowledgeDocVersionInfo(version *model.KnowledgeDocVersion) *KnowledgeDocVersionInfo {
	if version == nil {
		return nil
	}
	info := &KnowledgeDocVersionInfo{
		ID:            version.ID,
		DocumentID:    version.DocumentID,
		Version:       version.Version,
		Title:         version.Title,
		Content:       version.Content,
		DocType:       version.DocType,
		DocTypeName:   constant.GetDocTypeName(version.DocType),
		EditorID:      version.EditorID,
		ChangeSummary: version.ChangeSummary,
		CreatedAt:     version.CreatedAt,
	}
	if version.Editor != nil {
		info.EditorName = version.Editor.Username
	}
	return info
}