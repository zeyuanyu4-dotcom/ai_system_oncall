package repository

import (
	"gorm.io/gorm"

	"ai_system_oncall/internal/model"
)

// KnowledgeDocAttachmentRepository 文档附件仓库
type KnowledgeDocAttachmentRepository struct {
	db *gorm.DB
}

// NewKnowledgeDocAttachmentRepository 创建文档附件仓库
func NewKnowledgeDocAttachmentRepository(db *gorm.DB) *KnowledgeDocAttachmentRepository {
	return &KnowledgeDocAttachmentRepository{db: db}
}

// Create 创建附件记录
func (r *KnowledgeDocAttachmentRepository) Create(attachment *model.KnowledgeDocAttachment) error {
	return r.db.Create(attachment).Error
}

// Delete 删除附件记录
func (r *KnowledgeDocAttachmentRepository) Delete(id uint64) error {
	return r.db.Delete(&model.KnowledgeDocAttachment{}, id).Error
}

// FindByID 根据ID查询
func (r *KnowledgeDocAttachmentRepository) FindByID(id uint64) (*model.KnowledgeDocAttachment, error) {
	var attachment model.KnowledgeDocAttachment
	err := r.db.Preload("Creator").First(&attachment, id).Error
	if err != nil {
		return nil, err
	}
	return &attachment, nil
}

// ListByDocumentID 查询文档的所有附件
func (r *KnowledgeDocAttachmentRepository) ListByDocumentID(documentID uint64) ([]model.KnowledgeDocAttachment, error) {
	var attachments []model.KnowledgeDocAttachment
	err := r.db.Preload("Creator").Where("document_id = ?", documentID).
		Order("created_at DESC").Find(&attachments).Error
	if err != nil {
		return nil, err
	}
	return attachments, nil
}