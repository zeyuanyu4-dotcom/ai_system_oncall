package repository

import (
	"gorm.io/gorm"

	"ai_system_oncall/internal/model"
)

// KnowledgeDocVersionRepository 文档版本仓库
type KnowledgeDocVersionRepository struct {
	db *gorm.DB
}

// NewKnowledgeDocVersionRepository 创建文档版本仓库
func NewKnowledgeDocVersionRepository(db *gorm.DB) *KnowledgeDocVersionRepository {
	return &KnowledgeDocVersionRepository{db: db}
}

// Create 创建版本记录
func (r *KnowledgeDocVersionRepository) Create(version *model.KnowledgeDocVersion) error {
	return r.db.Create(version).Error
}

// FindByDocumentIDAndVersion 根据文档ID和版本号查询
func (r *KnowledgeDocVersionRepository) FindByDocumentIDAndVersion(documentID uint64, version int) (*model.KnowledgeDocVersion, error) {
	var ver model.KnowledgeDocVersion
	err := r.db.Preload("Editor").Where("document_id = ? AND version = ?", documentID, version).First(&ver).Error
	if err != nil {
		return nil, err
	}
	return &ver, nil
}

// ListByDocumentID 查询文档的所有版本
func (r *KnowledgeDocVersionRepository) ListByDocumentID(documentID uint64) ([]model.KnowledgeDocVersion, error) {
	var versions []model.KnowledgeDocVersion
	err := r.db.Preload("Editor").Where("document_id = ?", documentID).
		Order("version DESC").Find(&versions).Error
	if err != nil {
		return nil, err
	}
	return versions, nil
}

// GetMaxVersion 获取文档的最大版本号
func (r *KnowledgeDocVersionRepository) GetMaxVersion(documentID uint64) (int, error) {
	var maxVersion int
	err := r.db.Model(&model.KnowledgeDocVersion{}).
		Where("document_id = ?", documentID).
		Select("COALESCE(MAX(version), 0)").
		Scan(&maxVersion).Error
	if err != nil {
		return 0, err
	}
	return maxVersion, nil
}