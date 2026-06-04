package repository

import (
	"ai_system_oncall/internal/model"
	"time"

	"gorm.io/gorm"
)

type DashboardRepository struct {
	db *gorm.DB
}

func NewDashboardRepository(db *gorm.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

// GetStatByDate 获取指定日期的统计数据
func (r *DashboardRepository) GetStatByDate(date string, projectID *uint64) (*model.StatDailyRecord, error) {
	var record model.StatDailyRecord
	query := r.db.Where("stat_date = ?", date)
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	} else {
		query = query.Where("project_id IS NULL")
	}
	err := query.First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}

// GetStatsByDateRange 获取日期范围内的统计数据
func (r *DashboardRepository) GetStatsByDateRange(startDate, endDate string, projectID *uint64) ([]model.StatDailyRecord, error) {
	var records []model.StatDailyRecord
	query := r.db.Where("stat_date >= ? AND stat_date <= ?", startDate, endDate)
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	} else {
		query = query.Where("project_id IS NULL")
	}
	err := query.Order("stat_date ASC").Find(&records).Error
	return records, err
}

// CreateStat 创建统计记录
func (r *DashboardRepository) CreateStat(record *model.StatDailyRecord) error {
	return r.db.Create(record).Error
}

// UpsertStat 创建或更新统计记录
func (r *DashboardRepository) UpsertStat(record *model.StatDailyRecord) error {
	return r.db.Save(record).Error
}

// CalculateDailyStats 计算指定日期的统计数据（实时查询）
func (r *DashboardRepository) CalculateDailyStats(date string, projectID *uint64) (*model.StatDailyRecord, error) {
	record := &model.StatDailyRecord{
		StatDate:  date,
		ProjectID: projectID,
	}

	// 解析日期
	statDate, _ := time.Parse("2006-01-02", date)

	// 构建项目过滤条件
	projectFilter := func(query *gorm.DB) *gorm.DB {
		if projectID != nil {
			return query.Where("project_id = ?", *projectID)
		}
		return query
	}

	// 1. 问题总数（所有问题，不限定时间）
	var totalCount int64
	projectFilter(r.db.Model(&model.Issue{})).Count(&totalCount)
	record.TotalIssues = totalCount

	// 2. 新增问题（指定日期创建的）
	var newCount int64
	startTime := statDate
	endTime := statDate.Add(24 * time.Hour)
	query := r.db.Model(&model.Issue{}).Where("created_at >= ? AND created_at < ?", startTime, endTime)
	projectFilter(query).Count(&newCount)
	record.NewIssues = newCount

	// 3. 未闭环问题（状态不是 resolved/closed）
	var unresolvedCount int64
	query = r.db.Model(&model.Issue{}).Where("status NOT IN ?", []string{"resolved", "closed"})
	projectFilter(query).Count(&unresolvedCount)
	record.UnresolvedIssues = unresolvedCount

	// 4. 已解决问题总数
	var resolvedCount int64
	query = r.db.Model(&model.Issue{}).Where("status IN ?", []string{"resolved", "closed"})
	projectFilter(query).Count(&resolvedCount)
	record.ResolvedIssues = resolvedCount

	// 5. 高优问题 P0+P1
	var p0p1Count int64
	query = r.db.Model(&model.Issue{}).Where("priority IN ?", []string{"P0", "P1"})
	projectFilter(query).Count(&p0p1Count)
	record.P0P1Issues = p0p1Count

	// 6. 平均处理时长（已解决问题）
	var avgMinutes float64
	query = r.db.Model(&model.Issue{}).
		Where("status IN ? AND resolved_at IS NOT NULL", []string{"resolved", "closed"})
	projectFilter(query).Select("AVG(TIMESTAMPDIFF(MINUTE, created_at, resolved_at))").Scan(&avgMinutes)
	record.AvgResolveMinutes = int64(avgMinutes)

	// 7. 重复问题（标题相同）
	var repeatedCount int64
	subQuery := r.db.Model(&model.Issue{}).
		Select("title, COUNT(*) as cnt").
		Group("title").
		Having("COUNT(*) > 1")
	if projectID != nil {
		subQuery = subQuery.Where("project_id = ?", *projectID)
	}
	subQuery.Scan(&repeatedCount)
	record.RepeatedIssues = repeatedCount

	// 8. AI分析次数
	var aiCount int64
	query = r.db.Model(&model.AIAnalysisTask{})
	if projectID != nil {
		query = query.Joins("JOIN issues ON issues.id = ai_analysis_tasks.issue_id").
			Where("issues.project_id = ?", *projectID)
	}
	query.Count(&aiCount)
	record.AIAnalysisCount = aiCount

	// 9. AI建议采纳数
	var aiAdoptedCount int64
	query = r.db.Model(&model.AIAnalysisTask{}).
		Joins("JOIN issues ON issues.id = ai_analysis_tasks.issue_id").
		Where("issues.status IN ?", []string{"resolved", "closed"})
	if projectID != nil {
		query = query.Where("issues.project_id = ?", *projectID)
	}
	query.Count(&aiAdoptedCount)
	record.AIAdoptedCount = aiAdoptedCount

	return record, nil
}

// GetIssueTypeDistribution 获取问题类型分布
func (r *DashboardRepository) GetIssueTypeDistribution(projectID *uint64) ([]model.IssueTypeStat, error) {
	var stats []model.IssueTypeStat
	query := r.db.Model(&model.Issue{}).
		Select("issue_type, COUNT(*) as count").
		Group("issue_type").
		Order("count DESC")
	if projectID != nil {
		query = query.Where("project_id = ?", *projectID)
	}
	err := query.Scan(&stats).Error
	return stats, err
}

// GetIssueTypeDistributionByProjects 获取多项目的问题类型分布
func (r *DashboardRepository) GetIssueTypeDistributionByProjects(projectIDs []uint64) ([]model.IssueTypeStat, error) {
	var stats []model.IssueTypeStat
	err := r.db.Model(&model.Issue{}).
		Select("issue_type, COUNT(*) as count").
		Where("project_id IN ?", projectIDs).
		Group("issue_type").
		Order("count DESC").
		Scan(&stats).Error
	return stats, err
}

// GetServiceIssueRanking 获取服务问题排行 TOP10
func (r *DashboardRepository) GetServiceIssueRanking(projectID *uint64, limit int) ([]model.ServiceIssueStat, error) {
	var stats []model.ServiceIssueStat
	query := r.db.Model(&model.Issue{}).
		Select("services.id as service_id, services.name as service_name, COUNT(*) as count").
		Joins("LEFT JOIN services ON services.id = issues.service_id").
		Group("services.id, services.name").
		Order("count DESC")
	if projectID != nil {
		query = query.Where("issues.project_id = ?", *projectID)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Scan(&stats).Error
	return stats, err
}

// GetServiceIssueRankingByProjects 获取多项目的服务问题排行
func (r *DashboardRepository) GetServiceIssueRankingByProjects(projectIDs []uint64, limit int) ([]model.ServiceIssueStat, error) {
	var stats []model.ServiceIssueStat
	query := r.db.Model(&model.Issue{}).
		Select("services.id as service_id, services.name as service_name, COUNT(*) as count").
		Joins("LEFT JOIN services ON services.id = issues.service_id").
		Where("issues.project_id IN ?", projectIDs).
		Group("services.id, services.name").
		Order("count DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Scan(&stats).Error
	return stats, err
}

// GetManagedProjectIDs 获取用户管理的项目ID列表
func (r *DashboardRepository) GetManagedProjectIDs(userID uint64) ([]uint64, error) {
	var projectIDs []uint64
	err := r.db.Model(&model.ProjectMember{}).
		Where("user_id = ? AND project_role IN ?", userID, []string{"admin", "owner"}).
		Pluck("project_id", &projectIDs).Error
	return projectIDs, err
}

// CalculateMultiProjectStats 计算多个项目的汇总统计数据
func (r *DashboardRepository) CalculateMultiProjectStats(date string, projectIDs []uint64) (*model.StatDailyRecord, error) {
	record := &model.StatDailyRecord{
		StatDate: date,
	}

	if len(projectIDs) == 0 {
		return record, nil
	}

	// 1. 问题总数
	var totalCount int64
	r.db.Model(&model.Issue{}).Where("project_id IN ?", projectIDs).Count(&totalCount)
	record.TotalIssues = totalCount

	// 2. 新增问题
	statDate, _ := time.Parse("2006-01-02", date)
	startTime := statDate
	endTime := statDate.Add(24 * time.Hour)
	var newCount int64
	r.db.Model(&model.Issue{}).Where("project_id IN ? AND created_at >= ? AND created_at < ?", projectIDs, startTime, endTime).Count(&newCount)
	record.NewIssues = newCount

	// 3. 未闭环问题
	var unresolvedCount int64
	r.db.Model(&model.Issue{}).Where("project_id IN ? AND status NOT IN ?", projectIDs, []string{"resolved", "closed"}).Count(&unresolvedCount)
	record.UnresolvedIssues = unresolvedCount

	// 4. 已解决问题
	var resolvedCount int64
	r.db.Model(&model.Issue{}).Where("project_id IN ? AND status IN ?", projectIDs, []string{"resolved", "closed"}).Count(&resolvedCount)
	record.ResolvedIssues = resolvedCount

	// 5. 高优问题
	var p0p1Count int64
	r.db.Model(&model.Issue{}).Where("project_id IN ? AND priority IN ?", projectIDs, []string{"P0", "P1"}).Count(&p0p1Count)
	record.P0P1Issues = p0p1Count

	// 6. 平均处理时长
	var avgMinutes float64
	r.db.Model(&model.Issue{}).
		Where("project_id IN ? AND status IN ? AND resolved_at IS NOT NULL", projectIDs, []string{"resolved", "closed"}).
		Select("AVG(TIMESTAMPDIFF(MINUTE, created_at, resolved_at))").Scan(&avgMinutes)
	record.AvgResolveMinutes = int64(avgMinutes)

	// 7. 重复问题
	var repeatedCount int64
	r.db.Model(&model.Issue{}).
		Select("title, COUNT(*) as cnt").
		Where("project_id IN ?", projectIDs).
		Group("title").
		Having("COUNT(*) > 1").
		Scan(&repeatedCount)
	record.RepeatedIssues = repeatedCount

	// 8. AI分析次数
	var aiCount int64
	r.db.Model(&model.AIAnalysisTask{}).
		Joins("JOIN issues ON issues.id = ai_analysis_tasks.issue_id").
		Where("issues.project_id IN ?", projectIDs).
		Count(&aiCount)
	record.AIAnalysisCount = aiCount

	// 9. AI建议采纳数
	var aiAdoptedCount int64
	r.db.Model(&model.AIAnalysisTask{}).
		Joins("JOIN issues ON issues.id = ai_analysis_tasks.issue_id").
		Where("issues.project_id IN ? AND issues.status IN ?", projectIDs, []string{"resolved", "closed"}).
		Count(&aiAdoptedCount)
	record.AIAdoptedCount = aiAdoptedCount

	return record, nil
}

// GetProjectIssueRanking 获取项目问题排行
func (r *DashboardRepository) GetProjectIssueRanking(limit int) ([]model.ProjectIssueStat, error) {
	var stats []model.ProjectIssueStat
	query := r.db.Model(&model.Issue{}).
		Select("projects.id as project_id, projects.name as project_name, COUNT(*) as count").
		Joins("LEFT JOIN projects ON projects.id = issues.project_id").
		Group("projects.id, projects.name").
		Order("count DESC")
	if limit > 0 {
		query = query.Limit(limit)
	}
	err := query.Scan(&stats).Error
	return stats, err
}
