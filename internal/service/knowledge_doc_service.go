package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"gorm.io/gorm"

	"ai_system_oncall/internal/constant"
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"
	"ai_system_oncall/internal/config"
)

// KnowledgeDocService 知识库文档服务
type KnowledgeDocService struct {
	docRepo     *repository.KnowledgeDocRepository
	versionRepo *repository.KnowledgeDocVersionRepository
	projectRepo *repository.ProjectRepository
	serviceRepo *repository.ServiceRepository
}

// NewKnowledgeDocService 创建知识库文档服务
func NewKnowledgeDocService(
	docRepo *repository.KnowledgeDocRepository,
	versionRepo *repository.KnowledgeDocVersionRepository,
	projectRepo *repository.ProjectRepository,
	serviceRepo *repository.ServiceRepository,
) *KnowledgeDocService {
	return &KnowledgeDocService{
		docRepo:     docRepo,
		versionRepo: versionRepo,
		projectRepo: projectRepo,
		serviceRepo: serviceRepo,
	}
}

// CreateDocument 创建文档
func (s *KnowledgeDocService) CreateDocument(creatorID uint64, req *dto.CreateKnowledgeDocRequest) (*dto.KnowledgeDocInfo, error) {
	// 验证文档类型
	if !constant.ValidateDocType(req.DocType) {
		return nil, errors.New("无效的文档类型")
	}

	// 如果指定了项目，验证项目存在
	if req.ProjectID != nil {
		_, err := s.projectRepo.FindByID(*req.ProjectID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("项目不存在")
			}
			return nil, err
		}
	}

	// 如果指定了服务，验证服务存在
	if req.ServiceID != nil {
		_, err := s.serviceRepo.FindByID(*req.ServiceID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, errors.New("服务不存在")
			}
			return nil, err
		}
	}

	now := time.Now()
	doc := &model.KnowledgeDocument{
		ProjectID:      req.ProjectID,
		ServiceID:      req.ServiceID,
		Title:          req.Title,
		Content:        req.Content,
		DocType:        req.DocType,
		CreatorID:      creatorID,
		UpdaterID:      creatorID,
		VectorStatus:   constant.VECTOR_STATUS_PENDING,
		CurrentVersion: 1,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.docRepo.Create(doc); err != nil {
		return nil, err
	}

	// 创建初始版本
	version := &model.KnowledgeDocVersion{
		DocumentID:    doc.ID,
		Version:       1,
		Title:         doc.Title,
		Content:       doc.Content,
		DocType:       doc.DocType,
		EditorID:      creatorID,
		ChangeSummary: "初始版本",
		CreatedAt:     now,
	}
	s.versionRepo.Create(version)

	return dto.ToKnowledgeDocInfo(doc), nil
}

// UpdateDocument 更新文档
func (s *KnowledgeDocService) UpdateDocument(id uint64, updaterID uint64, req *dto.UpdateKnowledgeDocRequest) (*dto.KnowledgeDocInfo, error) {
	// 验证文档类型
	if !constant.ValidateDocType(req.DocType) {
		return nil, errors.New("无效的文档类型")
	}

	doc, err := s.docRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文档不存在")
		}
		return nil, err
	}

	// 更新文档
	doc.Title = req.Title
	doc.Content = req.Content
	doc.DocType = req.DocType
	doc.UpdaterID = updaterID
	doc.UpdatedAt = time.Now()
	doc.CurrentVersion++

	if err := s.docRepo.Update(doc); err != nil {
		return nil, err
	}

	// 创建新版本
	summary := req.ChangeSummary
	if summary == "" {
		summary = "更新文档"
	}
	version := &model.KnowledgeDocVersion{
		DocumentID:    doc.ID,
		Version:       doc.CurrentVersion,
		Title:         doc.Title,
		Content:       doc.Content,
		DocType:       doc.DocType,
		EditorID:      updaterID,
		ChangeSummary: summary,
		CreatedAt:     time.Now(),
	}
	s.versionRepo.Create(version)

	// 重新加载关联数据
	doc, _ = s.docRepo.FindByID(id)
	return dto.ToKnowledgeDocInfo(doc), nil
}

// GetDocument 获取文档详情
func (s *KnowledgeDocService) GetDocument(id uint64) (*dto.KnowledgeDocInfo, error) {
	doc, err := s.docRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("文档不存在")
		}
		return nil, err
	}
	return dto.ToKnowledgeDocInfo(doc), nil
}

// ListDocuments 查询文档列表
func (s *KnowledgeDocService) ListDocuments(req *dto.KnowledgeDocListRequest) (*dto.KnowledgeDocListResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filters := map[string]interface{}{
		"project_id": req.ProjectID,
		"service_id": req.ServiceID,
		"doc_type":   req.DocType,
		"keyword":    req.Keyword,
	}

	docs, total, err := s.docRepo.List(page, pageSize, filters)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.KnowledgeDocInfo, 0, len(docs))
	for _, doc := range docs {
		list = append(list, dto.ToKnowledgeDocInfo(&doc))
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.KnowledgeDocListResponse{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// SearchDocuments 搜索文档
func (s *KnowledgeDocService) SearchDocuments(keyword string, page, pageSize int) (*dto.KnowledgeDocListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	docs, total, err := s.docRepo.Search(keyword, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.KnowledgeDocInfo, 0, len(docs))
	for _, doc := range docs {
		list = append(list, dto.ToKnowledgeDocInfo(&doc))
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.KnowledgeDocListResponse{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// DeleteDocument 删除文档
func (s *KnowledgeDocService) DeleteDocument(id uint64) error {
	return s.docRepo.Delete(id)
}

// GetDocumentsByProject 按项目查询
func (s *KnowledgeDocService) GetDocumentsByProject(projectID uint64, page, pageSize int) (*dto.KnowledgeDocListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	docs, total, err := s.docRepo.ListByProjectID(projectID, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.KnowledgeDocInfo, 0, len(docs))
	for _, doc := range docs {
		list = append(list, dto.ToKnowledgeDocInfo(&doc))
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.KnowledgeDocListResponse{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetDocumentsByService 按服务查询
func (s *KnowledgeDocService) GetDocumentsByService(serviceID uint64, page, pageSize int) (*dto.KnowledgeDocListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	docs, total, err := s.docRepo.ListByServiceID(serviceID, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.KnowledgeDocInfo, 0, len(docs))
	for _, doc := range docs {
		list = append(list, dto.ToKnowledgeDocInfo(&doc))
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.KnowledgeDocListResponse{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetDocumentsByType 按类型查询
func (s *KnowledgeDocService) GetDocumentsByType(docType string, page, pageSize int) (*dto.KnowledgeDocListResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	docs, total, err := s.docRepo.ListByDocType(docType, page, pageSize)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.KnowledgeDocInfo, 0, len(docs))
	for _, doc := range docs {
		list = append(list, dto.ToKnowledgeDocInfo(&doc))
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &dto.KnowledgeDocListResponse{
		List:       list,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateVectorStatus 更新向量化状态
func (s *KnowledgeDocService) UpdateVectorStatus(id uint64, status string) error {
	if !constant.ValidateVectorStatus(status) {
		return errors.New("无效的向量化状态")
	}

	_, err := s.docRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文档不存在")
		}
		return err
	}

	return s.docRepo.UpdateVectorStatus(id, status)
}

// GetVersions 获取文档版本历史
func (s *KnowledgeDocService) GetVersions(documentID uint64) ([]*dto.KnowledgeDocVersionInfo, error) {
	versions, err := s.versionRepo.ListByDocumentID(documentID)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.KnowledgeDocVersionInfo, 0, len(versions))
	for _, v := range versions {
		list = append(list, dto.ToKnowledgeDocVersionInfo(&v))
	}

	return list, nil
}

// TriggerVectorization 触发文档向量化
func (s *KnowledgeDocService) TriggerVectorization(docID uint64) error {
	doc, err := s.docRepo.FindByID(docID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("文档不存在")
		}
		return err
	}

	// 获取项目名称
	projectName := ""
	serviceName := ""
	if doc.Project != nil {
		projectName = doc.Project.Name
	}
	if doc.Service != nil {
		serviceName = doc.Service.Name
	}

	// 调用 Python 服务向量化
	err = s.vectorizeDocument(doc.ID, doc.Title, doc.DocType, projectName, serviceName, doc.Content)
	if err != nil {
		return err
	}

	// 更新向量化状态
	return s.docRepo.UpdateVectorStatus(docID, "completed")
}

// vectorizeDocument 调用 Python AI 服务向量化文档
func (s *KnowledgeDocService) vectorizeDocument(docID uint64, title, docType, projectName, serviceName, content string) error {
	baseURL := "http://127.0.0.1:8001"
	if config.GetConfig() != nil && config.GetConfig().AI.BaseURL != "" {
		baseURL = config.GetConfig().AI.BaseURL
	}

	url := fmt.Sprintf("%s/api/rag/vectorize/text", baseURL)

	reqBody := map[string]interface{}{
		"doc_id":       docID,
		"title":        title,
		"doc_type":     docType,
		"project_name": projectName,
		"service_name": serviceName,
		"content":      content,
		"heading_path": "",
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("调用向量化服务失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("向量化服务错误: %s", string(respBody))
	}

	// 解析响应体，检查 code 字段
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取向量化响应失败: %w", err)
	}

	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    any    `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return fmt.Errorf("解析向量化响应失败: %w", err)
	}

	if result.Code != 0 {
		return fmt.Errorf("向量化失败: %s (code=%d)", result.Message, result.Code)
	}

	return nil
}