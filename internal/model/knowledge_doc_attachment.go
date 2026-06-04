package model

import (
	"time"
)

// KnowledgeDocAttachment 文档附件
type KnowledgeDocAttachment struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	DocumentID uint64    `gorm:"index;not null" json:"document_id"`
	FileName   string    `gorm:"type:varchar(255);not null" json:"file_name"`
	FilePath   string    `gorm:"type:varchar(500);not null" json:"file_path"`
	FileSize   int       `gorm:"not null" json:"file_size"`
	FileType   string    `gorm:"type:varchar(100)" json:"file_type"`
	CreatorID  uint64    `gorm:"not null" json:"creator_id"`
	CreatedAt  time.Time `json:"created_at"`

	// 关联关系
	Document *KnowledgeDocument `gorm:"foreignKey:DocumentID" json:"document,omitempty"`
	Creator  *User              `gorm:"foreignKey:CreatorID" json:"creator,omitempty"`
}

func (KnowledgeDocAttachment) TableName() string {
	return "knowledge_doc_attachments"
}