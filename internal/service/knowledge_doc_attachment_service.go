package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"
)

// KnowledgeDocAttachmentService 文档附件服务
type KnowledgeDocAttachmentService struct {
	attachmentRepo *repository.KnowledgeDocAttachmentRepository
	docRepo        *repository.KnowledgeDocRepository
}

// NewKnowledgeDocAttachmentService 创建文档附件服务
func NewKnowledgeDocAttachmentService(
	attachmentRepo *repository.KnowledgeDocAttachmentRepository,
	docRepo *repository.KnowledgeDocRepository,
) *KnowledgeDocAttachmentService {
	return &KnowledgeDocAttachmentService{
		attachmentRepo: attachmentRepo,
		docRepo:        docRepo,
	}
}

// UploadAttachment 上传附件
func (s *KnowledgeDocAttachmentService) UploadAttachment(documentID uint64, creatorID uint64, file *multipart.FileHeader) (*dto.KnowledgeDocAttachmentInfo, error) {
	// 验证文档存在
	_, err := s.docRepo.FindByID(documentID)
	if err != nil {
		return nil, errors.New("文档不存在")
	}

	// 创建上传目录
	uploadDir := filepath.Join("uploads", "knowledge-docs", fmt.Sprintf("%d", documentID))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, fmt.Errorf("创建上传目录失败: %w", err)
	}

	// 生成唯一文件名
	ext := filepath.Ext(file.Filename)
	newFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(uploadDir, newFilename)

	// 保存文件
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	// 创建附件记录
	attachment := &model.KnowledgeDocAttachment{
		DocumentID: documentID,
		FileName:   file.Filename,
		FilePath:   filePath,
		FileSize:   int(file.Size),
		FileType:   file.Header.Get("Content-Type"),
		CreatorID:  creatorID,
		CreatedAt:  time.Now(),
	}

	if err := s.attachmentRepo.Create(attachment); err != nil {
		return nil, err
	}

	return dto.ToKnowledgeDocAttachmentInfo(attachment), nil
}

// GetAttachments 获取附件列表
func (s *KnowledgeDocAttachmentService) GetAttachments(documentID uint64) ([]*dto.KnowledgeDocAttachmentInfo, error) {
	attachments, err := s.attachmentRepo.ListByDocumentID(documentID)
	if err != nil {
		return nil, err
	}

	list := make([]*dto.KnowledgeDocAttachmentInfo, 0, len(attachments))
	for _, a := range attachments {
		list = append(list, dto.ToKnowledgeDocAttachmentInfo(&a))
	}

	return list, nil
}

// DeleteAttachment 删除附件
func (s *KnowledgeDocAttachmentService) DeleteAttachment(attachmentID uint64) error {
	attachment, err := s.attachmentRepo.FindByID(attachmentID)
	if err != nil {
		return errors.New("附件不存在")
	}

	// 删除物理文件
	if err := os.Remove(attachment.FilePath); err != nil && !os.IsNotExist(err) {
		// 文件不存在不影响删除数据库记录
	}

	return s.attachmentRepo.Delete(attachmentID)
}

// ParseAttachmentToContent 解析附件生成文档内容
func (s *KnowledgeDocAttachmentService) ParseAttachmentToContent(file *multipart.FileHeader) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	// 根据文件扩展名生成不同的内容格式
	ext := strings.ToLower(filepath.Ext(file.Filename))
	
	switch ext {
	case ".md":
		return string(content), nil
	case ".txt":
		return fmt.Sprintf("```\n%s\n```", string(content)), nil
	case ".log":
		return fmt.Sprintf("```\n%s\n```", string(content)), nil
	case ".json":
		return fmt.Sprintf("```json\n%s\n```", string(content)), nil
	case ".yaml", ".yml":
		return fmt.Sprintf("```yaml\n%s\n```", string(content)), nil
	default:
		return string(content), nil
	}
}

// GetAttachmentPath 获取附件文件路径（用于下载）
func (s *KnowledgeDocAttachmentService) GetAttachmentPath(attachmentID uint64) (string, string, error) {
	attachment, err := s.attachmentRepo.FindByID(attachmentID)
	if err != nil {
		return "", "", errors.New("附件不存在")
	}

	// 检查文件是否存在
	if _, err := os.Stat(attachment.FilePath); os.IsNotExist(err) {
		return "", "", errors.New("文件不存在")
	}

	return attachment.FilePath, attachment.FileName, nil
}

// ReadAttachmentContent 读取附件内容（供 Agent 读取）
func (s *KnowledgeDocAttachmentService) ReadAttachmentContent(attachmentID uint64) (string, string, error) {
	attachment, err := s.attachmentRepo.FindByID(attachmentID)
	if err != nil {
		return "", "", errors.New("附件不存在")
	}

	// 检查文件是否存在
	if _, err := os.Stat(attachment.FilePath); os.IsNotExist(err) {
		return "", "", errors.New("文件不存在")
	}

	// 读取文件内容
	content, err := os.ReadFile(attachment.FilePath)
	if err != nil {
		return "", "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 获取文件类型
	ext := strings.ToLower(filepath.Ext(attachment.FileName))
	fileType := "text/plain"
	switch ext {
	case ".md":
		fileType = "text/markdown"
	case ".txt":
		fileType = "text/plain"
	case ".json":
		fileType = "application/json"
	case ".yaml", ".yml":
		fileType = "application/yaml"
	case ".pdf":
		fileType = "application/pdf"
	case ".doc", ".docx":
		fileType = "application/msword"
	case ".xls", ".xlsx":
		fileType = "application/vnd.ms-excel"
	default:
		fileType = attachment.FileType
	}

	return string(content), fileType, nil
}