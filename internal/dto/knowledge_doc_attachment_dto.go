package dto

import (
	"time"

	"ai_system_oncall/internal/model"
)

// KnowledgeDocAttachmentInfo 文档附件信息
type KnowledgeDocAttachmentInfo struct {
	ID         uint64    `json:"id"`
	DocumentID uint64    `json:"document_id"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileSize   int       `json:"file_size"`
	FileType   string    `json:"file_type"`
	CreatorID  uint64    `json:"creator_id"`
	CreatorName string   `json:"creator_name"`
	CreatedAt  time.Time `json:"created_at"`
}

// ToKnowledgeDocAttachmentInfo 将模型转换为 DTO
func ToKnowledgeDocAttachmentInfo(attachment *model.KnowledgeDocAttachment) *KnowledgeDocAttachmentInfo {
	if attachment == nil {
		return nil
	}
	info := &KnowledgeDocAttachmentInfo{
		ID:         attachment.ID,
		DocumentID: attachment.DocumentID,
		FileName:   attachment.FileName,
		FilePath:   attachment.FilePath,
		FileSize:   attachment.FileSize,
		FileType:   attachment.FileType,
		CreatorID:  attachment.CreatorID,
		CreatedAt:  attachment.CreatedAt,
	}
	if attachment.Creator != nil {
		info.CreatorName = attachment.Creator.Username
	}
	return info
}