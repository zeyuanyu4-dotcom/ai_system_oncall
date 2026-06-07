package repository

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"ai_system_oncall/internal/cache"
	"ai_system_oncall/internal/model"
)

// KnowledgeDocRepository 知识库文档仓库
type KnowledgeDocRepository struct {
	db *gorm.DB
	sf *cache.SingleflightCache
}

// NewKnowledgeDocRepository 创建知识库文档仓库
func NewKnowledgeDocRepository(db *gorm.DB) *KnowledgeDocRepository {
	return &KnowledgeDocRepository{
		db: db,
		sf: cache.GetSingleflightCache(),
	}
}

// Create 创建文档
func (r *KnowledgeDocRepository) Create(doc *model.KnowledgeDocument) error {
	if err := r.db.Create(doc).Error; err != nil {
		return err
	}
	// 失效文档列表缓存
	ctx := context.Background()
	r.sf.InvalidateByPattern(ctx, "knowledge:list:*")
	return nil
}

// Update 更新文档
func (r *KnowledgeDocRepository) Update(doc *model.KnowledgeDocument) error {
	if err := r.db.Save(doc).Error; err != nil {
		return err
	}
	// 失效文档详情和列表缓存
	ctx := context.Background()
	r.sf.Invalidate(ctx, cache.KeyKnowledgeDocDetail, doc.ID)
	r.sf.InvalidateByPattern(ctx, "knowledge:list:*")
	r.sf.InvalidateByPattern(ctx, "knowledge:search:*")
	return nil
}

// Delete 软删除文档
func (r *KnowledgeDocRepository) Delete(id uint64) error {
	if err := r.db.Delete(&model.KnowledgeDocument{}, id).Error; err != nil {
		return err
	}
	// 失效所有相关缓存
	ctx := context.Background()
	r.sf.Invalidate(ctx, cache.KeyKnowledgeDocDetail, id)
	r.sf.InvalidateByPattern(ctx, "knowledge:list:*")
	r.sf.InvalidateByPattern(ctx, "knowledge:search:*")
	return nil
}

// FindByID 根据ID查询
func (r *KnowledgeDocRepository) FindByID(id uint64) (*model.KnowledgeDocument, error) {
	ctx := context.Background()
	var doc model.KnowledgeDocument

	err := r.sf.GetWithLoad(ctx, cache.KeyKnowledgeDocDetail, &doc, []interface{}{id}, func() (interface{}, error) {
		var d model.KnowledgeDocument
		if err := r.db.Preload("Project").Preload("Service").Preload("Creator").Preload("Updater").
			First(&d, id).Error; err != nil {
			return nil, err
		}
		return &d, nil
	})

	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// List 查询文档列表
func (r *KnowledgeDocRepository) List(page, pageSize int, filters map[string]interface{}) ([]model.KnowledgeDocument, int64, error) {
	var docs []model.KnowledgeDocument
	var total int64

	query := r.db.Model(&model.KnowledgeDocument{})

	// 应用筛选条件
	if projectID, ok := filters["project_id"]; ok && projectID != nil {
		if pid, ok := projectID.(*uint64); ok && pid != nil {
			query = query.Where("project_id = ?", *pid)
		}
	}
	if serviceID, ok := filters["service_id"]; ok && serviceID != nil {
		if sid, ok := serviceID.(*uint64); ok && sid != nil {
			query = query.Where("service_id = ?", *sid)
		}
	}
	if docType, ok := filters["doc_type"]; ok && docType != nil && docType != "" {
		query = query.Where("doc_type = ?", docType)
	}
	if keyword, ok := filters["keyword"]; ok && keyword != nil && keyword != "" {
		query = query.Where("(title LIKE ? OR content LIKE ?)", "%"+keyword.(string)+"%", "%"+keyword.(string)+"%")
	}

	// 统计总数
	query.Count(&total)

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Preload("Project").Preload("Service").Preload("Creator").Preload("Updater").
		Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&docs).Error
	if err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

// Search 全文搜索
func (r *KnowledgeDocRepository) Search(keyword string, page, pageSize int) ([]model.KnowledgeDocument, int64, error) {
	var docs []model.KnowledgeDocument
	var total int64

	// 使用 LIKE 进行搜索（生产环境可考虑全文索引）
	searchPattern := "%" + strings.Join(strings.Fields(keyword), "%") + "%"
	
	query := r.db.Model(&model.KnowledgeDocument{}).
		Where("title LIKE ? OR content LIKE ?", searchPattern, searchPattern)

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Project").Preload("Service").Preload("Creator").Preload("Updater").
		Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&docs).Error
	if err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

// ListByProjectID 按项目查询
func (r *KnowledgeDocRepository) ListByProjectID(projectID uint64, page, pageSize int) ([]model.KnowledgeDocument, int64, error) {
	var docs []model.KnowledgeDocument
	var total int64

	query := r.db.Model(&model.KnowledgeDocument{}).Where("project_id = ?", projectID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Project").Preload("Service").Preload("Creator").Preload("Updater").
		Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&docs).Error
	if err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

// ListByServiceID 按服务查询
func (r *KnowledgeDocRepository) ListByServiceID(serviceID uint64, page, pageSize int) ([]model.KnowledgeDocument, int64, error) {
	var docs []model.KnowledgeDocument
	var total int64

	query := r.db.Model(&model.KnowledgeDocument{}).Where("service_id = ?", serviceID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Project").Preload("Service").Preload("Creator").Preload("Updater").
		Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&docs).Error
	if err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

// ListByDocType 按类型查询
func (r *KnowledgeDocRepository) ListByDocType(docType string, page, pageSize int) ([]model.KnowledgeDocument, int64, error) {
	var docs []model.KnowledgeDocument
	var total int64

	query := r.db.Model(&model.KnowledgeDocument{}).Where("doc_type = ?", docType)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Preload("Project").Preload("Service").Preload("Creator").Preload("Updater").
		Offset(offset).Limit(pageSize).Order("updated_at DESC").Find(&docs).Error
	if err != nil {
		return nil, 0, err
	}

	return docs, total, nil
}

// UpdateVectorStatus 更新向量化状态
func (r *KnowledgeDocRepository) UpdateVectorStatus(id uint64, status string) error {
	return r.db.Model(&model.KnowledgeDocument{}).Where("id = ?", id).
		Update("vector_status", status).Error
}