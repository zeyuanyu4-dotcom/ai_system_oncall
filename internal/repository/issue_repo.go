package repository

import (
	"ai_system_oncall/internal/model"
	"time"

	"gorm.io/gorm"
)

type IssueRepository struct {
	db *gorm.DB
}

func NewIssueRepository(db *gorm.DB) *IssueRepository {
	return &IssueRepository{db: db}
}

// Create creates a new issue
func (r *IssueRepository) Create(issue *model.Issue) error {
	return r.db.Create(issue).Error
}

// Update updates an issue
func (r *IssueRepository) Update(issue *model.Issue) error {
	return r.db.Save(issue).Error
}

// Delete soft deletes an issue
func (r *IssueRepository) Delete(id uint64) error {
	return r.db.Delete(&model.Issue{}, id).Error
}

// FindByID finds an issue by ID
func (r *IssueRepository) FindByID(id uint64) (*model.Issue, error) {
	var issue model.Issue
	err := r.db.Preload("Project").Preload("Service").Preload("Creator").Preload("Assignee").First(&issue, id).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// FindByIssueNo finds an issue by issue number
func (r *IssueRepository) FindByIssueNo(issueNo string) (*model.Issue, error) {
	var issue model.Issue
	err := r.db.Where("issue_no = ?", issueNo).First(&issue).Error
	if err != nil {
		return nil, err
	}
	return &issue, nil
}

// List lists issues with pagination and filters
func (r *IssueRepository) List(page, pageSize int, filters map[string]interface{}) ([]model.Issue, int64, error) {
	var issues []model.Issue
	var total int64

	query := r.db.Model(&model.Issue{})

	if projectID, ok := filters["project_id"]; ok && projectID != nil {
		query = query.Where("project_id = ?", projectID)
	}
	if serviceID, ok := filters["service_id"]; ok && serviceID != nil {
		query = query.Where("service_id = ?", serviceID)
	}
	if status, ok := filters["status"]; ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if issueType, ok := filters["issue_type"]; ok && issueType != "" {
		query = query.Where("issue_type = ?", issueType)
	}
	if priority, ok := filters["priority"]; ok && priority != "" {
		query = query.Where("priority = ?", priority)
	}
	if environment, ok := filters["environment"]; ok && environment != "" {
		query = query.Where("environment = ?", environment)
	}
	if creatorID, ok := filters["creator_id"]; ok && creatorID != nil {
		query = query.Where("creator_id = ?", creatorID)
	}
	if assigneeID, ok := filters["assignee_id"]; ok && assigneeID != nil {
		query = query.Where("assignee_id = ?", assigneeID)
	}
	if keyword, ok := filters["keyword"]; ok && keyword != "" {
		query = query.Where("title LIKE ? OR issue_no LIKE ?", "%"+keyword.(string)+"%", "%"+keyword.(string)+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Project").Preload("Service").Preload("Creator").Preload("Assignee").
		Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&issues).Error; err != nil {
		return nil, 0, err
	}

	return issues, total, nil
}

// UpdateStatus updates issue status
func (r *IssueRepository) UpdateStatus(id uint64, status string) error {
	return r.db.Model(&model.Issue{}).Where("id = ?", id).Update("status", status).Error
}

// UpdateAssignee updates issue assignee
func (r *IssueRepository) UpdateAssignee(id uint64, assigneeID uint64) error {
	return r.db.Model(&model.Issue{}).Where("id = ?", id).Update("assignee_id", assigneeID).Error
}

// UpdateResolution updates issue resolution info
func (r *IssueRepository) UpdateResolution(id uint64, rootCause, solution string, resolvedAt time.Time) error {
	return r.db.Model(&model.Issue{}).Where("id = ?", id).Updates(map[string]interface{}{
		"root_cause":  rootCause,
		"solution":    solution,
		"resolved_at": resolvedAt,
	}).Error
}

// UpdateClosedAt updates issue closed time
func (r *IssueRepository) UpdateClosedAt(id uint64, closedAt time.Time) error {
	return r.db.Model(&model.Issue{}).Where("id = ?", id).Update("closed_at", closedAt).Error
}

// UpdateAIAnalysis updates AI analysis result
func (r *IssueRepository) UpdateAIAnalysis(id uint64, summary, analysis string) error {
	return r.db.Model(&model.Issue{}).Where("id = ?", id).Updates(map[string]interface{}{
		"ai_summary":  summary,
		"ai_analysis": analysis,
	}).Error
}

// GetTodayIssueCount gets the count of issues created today for generating issue number
func (r *IssueRepository) GetTodayIssueCount() (int64, error) {
	var count int64
	today := time.Now().Format("2006-01-02")
	err := r.db.Model(&model.Issue{}).Where("DATE(created_at) = ?", today).Count(&count).Error
	return count, err
}

// ListByProjectID lists issues of a project
func (r *IssueRepository) ListByProjectID(projectID uint64) ([]model.Issue, error) {
	var issues []model.Issue
	err := r.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&issues).Error
	return issues, err
}

// SearchHistoryIssues 多字段搜索历史问题（已解决或已关闭）
func (r *IssueRepository) SearchHistoryIssues(keyword string, projectID uint64, issueType, environment string, page, pageSize int) ([]model.Issue, int64, error) {
	var issues []model.Issue
	var total int64

	query := r.db.Model(&model.Issue{}).
		Where("status IN ?", []string{"resolved", "closed"})

	// 多字段关键词搜索
	if keyword != "" {
		searchPattern := "%" + keyword + "%"
		query = query.Where(
			"title LIKE ? OR description LIKE ? OR error_message LIKE ? OR log_excerpt LIKE ? OR root_cause LIKE ? OR solution LIKE ?",
			searchPattern, searchPattern, searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}

	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	if issueType != "" {
		query = query.Where("issue_type = ?", issueType)
	}
	if environment != "" {
		query = query.Where("environment = ?", environment)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Preload("Project").Preload("Service").Preload("Creator").
		Offset(offset).Limit(pageSize).Order("resolved_at DESC, created_at DESC").Find(&issues).Error; err != nil {
		return nil, 0, err
	}

	return issues, total, nil
}

// FindSimilarIssues 查找相似问题（基于关键词匹配计算相似度）
func (r *IssueRepository) FindSimilarIssues(title, errorMessage, logExcerpt string, excludeID uint64, limit int) ([]model.Issue, error) {
	var issues []model.Issue

	// 收集所有关键词
	keywords := extractKeywords(title, errorMessage, logExcerpt)
	if len(keywords) == 0 {
		return issues, nil
	}

	// 构建查询：只查找已解决或已关闭的问题
	query := r.db.Model(&model.Issue{}).
		Where("status IN ?", []string{"resolved", "closed"}).
		Where("id != ?", excludeID)

	// 构建关键词匹配条件
	conditions := r.db.Model(&model.Issue{})
	for _, kw := range keywords {
		pattern := "%" + kw + "%"
		conditions = conditions.Or("title LIKE ?", pattern).
			Or("description LIKE ?", pattern).
			Or("error_message LIKE ?", pattern).
			Or("log_excerpt LIKE ?", pattern).
			Or("root_cause LIKE ?", pattern).
			Or("solution LIKE ?", pattern)
	}

	query = query.Where(conditions)

	if err := query.Preload("Project").Preload("Service").
		Limit(limit).Order("resolved_at DESC").Find(&issues).Error; err != nil {
		return nil, err
	}

	return issues, nil
}

// extractKeywords 从文本中提取关键词（简化版本）
func extractKeywords(texts ...string) []string {
	keywords := make(map[string]bool)
	for _, text := range texts {
		if text == "" {
			continue
		}
		// 简单分词：按空格、常见符号分割
		words := splitWords(text)
		for _, word := range words {
			if len(word) >= 2 { // 过滤太短的词
				keywords[word] = true
			}
		}
	}

	result := make([]string, 0, len(keywords))
	for kw := range keywords {
		result = append(result, kw)
	}
	// 最多取10个关键词
	if len(result) > 10 {
		result = result[:10]
	}
	return result
}

// splitWords 分割文本为单词
func splitWords(text string) []string {
	// 简单实现：按空格和常见符号分割
	delimiters := " \t\n\r,.;:!?()[]{}/\\\"'<>|@#$%^&*+=~`"
	words := make([]string, 0)
	start := 0

	for i, c := range text {
		isDelimiter := false
		for _, d := range delimiters {
			if rune(d) == c {
				isDelimiter = true
				break
			}
		}
		if isDelimiter {
			if i > start {
				words = append(words, text[start:i])
			}
			start = i + 1
		}
	}
	if start < len(text) {
		words = append(words, text[start:])
	}

	return words
}
