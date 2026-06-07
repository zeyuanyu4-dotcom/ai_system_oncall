package repository

import (
	"context"
	"encoding/json"
	"time"

	"ai_system_oncall/internal/cache"
	"ai_system_oncall/internal/model"

	"gorm.io/gorm"
)

type ReportRepository struct {
	db *gorm.DB
	sf *cache.SingleflightCache
}

func NewReportRepository(db *gorm.DB) *ReportRepository {
	return &ReportRepository{
		db: db,
		sf: cache.GetSingleflightCache(),
	}
}

// Create creates a new report
func (r *ReportRepository) Create(report *model.Report) error {
	if err := r.db.Create(report).Error; err != nil {
		return err
	}
	// 失效报告列表缓存
	ctx := context.Background()
	r.sf.InvalidateByPattern(ctx, "report:list:*")
	return nil
}

// Update updates a report
func (r *ReportRepository) Update(report *model.Report) error {
	if err := r.db.Save(report).Error; err != nil {
		return err
	}
	// 失效报告详情和列表缓存
	ctx := context.Background()
	r.sf.Invalidate(ctx, cache.KeyReportDetail, report.ID)
	r.sf.InvalidateByPattern(ctx, "report:list:*")
	if report.ReportType == model.ReportTypeDaily && report.ReportDate != "" {
		r.sf.Invalidate(ctx, cache.KeyDailyReport, report.ReportDate)
	}
	if report.ReportType == model.ReportTypeWeekly && report.ReportWeek != "" {
		r.sf.Invalidate(ctx, cache.KeyWeeklyReport, report.ReportWeek)
	}
	return nil
}

// FindByID finds a report by ID
func (r *ReportRepository) FindByID(id uint64) (*model.Report, error) {
	ctx := context.Background()
	var report model.Report

	err := r.sf.GetWithLoad(ctx, cache.KeyReportDetail, &report, []interface{}{id}, func() (interface{}, error) {
		var rp model.Report
		if err := r.db.Preload("Creator").First(&rp, id).Error; err != nil {
			return nil, err
		}
		return &rp, nil
	})

	if err != nil {
		return nil, err
	}
	return &report, nil
}

// FindByTypeAndDate finds a report by type and date
func (r *ReportRepository) FindByTypeAndDate(reportType, date string) (*model.Report, error) {
	ctx := context.Background()
	var report model.Report

	err := r.sf.GetWithLoad(ctx, cache.KeyDailyReport, &report, []interface{}{date}, func() (interface{}, error) {
		var rp model.Report
		if err := r.db.Where("report_type = ? AND report_date = ?", reportType, date).First(&rp).Error; err != nil {
			return nil, err
		}
		return &rp, nil
	})

	if err != nil {
		return nil, err
	}
	return &report, nil
}

// FindByTypeAndWeek finds a report by type and week
func (r *ReportRepository) FindByTypeAndWeek(reportType, week string) (*model.Report, error) {
	ctx := context.Background()
	var report model.Report

	err := r.sf.GetWithLoad(ctx, cache.KeyWeeklyReport, &report, []interface{}{week}, func() (interface{}, error) {
		var rp model.Report
		if err := r.db.Where("report_type = ? AND report_week = ?", reportType, week).First(&rp).Error; err != nil {
			return nil, err
		}
		return &rp, nil
	})

	if err != nil {
		return nil, err
	}
	return &report, nil
}

// FindByType finds reports by type with pagination
func (r *ReportRepository) FindByType(reportType string, page, pageSize int) ([]model.Report, int64, error) {
	var reports []model.Report
	var total int64

	query := r.db.Model(&model.Report{}).Where("report_type = ?", reportType)
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Preload("Creator").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reports).Error
	return reports, total, err
}

// FindAll finds all reports with pagination
func (r *ReportRepository) FindAll(page, pageSize int) ([]model.Report, int64, error) {
	var reports []model.Report
	var total int64

	err := r.db.Model(&model.Report{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = r.db.Preload("Creator").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reports).Error
	return reports, total, err
}

// FindByDateRange finds reports by date range
func (r *ReportRepository) FindByDateRange(startDate, endDate string, page, pageSize int) ([]model.Report, int64, error) {
	var reports []model.Report
	var total int64

	query := r.db.Where("report_date >= ? AND report_date <= ?", startDate, endDate)
	err := query.Model(&model.Report{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Preload("Creator").Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&reports).Error
	return reports, total, err
}

// UpdateStatus updates report status
func (r *ReportRepository) UpdateStatus(id uint64, status, errorMsg string) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}
	return r.db.Model(&model.Report{}).Where("id = ?", id).Updates(updates).Error
}

// Delete deletes a report
func (r *ReportRepository) Delete(id uint64) error {
	// 先获取报告信息用于失效缓存
	var report model.Report
	if err := r.db.First(&report, id).Error; err != nil {
		return r.db.Delete(&model.Report{}, id).Error
	}

	if err := r.db.Delete(&model.Report{}, id).Error; err != nil {
		return err
	}

	// 失效所有相关缓存
	ctx := context.Background()
	r.sf.Invalidate(ctx, cache.KeyReportDetail, id)
	r.sf.InvalidateByPattern(ctx, "report:list:*")
	if report.ReportType == model.ReportTypeDaily && report.ReportDate != "" {
		r.sf.Invalidate(ctx, cache.KeyDailyReport, report.ReportDate)
	}
	if report.ReportType == model.ReportTypeWeekly && report.ReportWeek != "" {
		r.sf.Invalidate(ctx, cache.KeyWeeklyReport, report.ReportWeek)
	}
	return nil
}

// GetDailyStats gets daily issue statistics
func (r *ReportRepository) GetDailyStats(date string) (*model.DailyReportStats, error) {
	stats := &model.DailyReportStats{}

	startTime, _ := time.Parse("2006-01-02", date)
	endTime := startTime.Add(24 * time.Hour)

	// Total issues created on this date
	err := r.db.Model(&model.Issue{}).Where("created_at >= ? AND created_at < ?", startTime, endTime).Count(&stats.TotalIssues).Error
	if err != nil {
		return nil, err
	}

	// Pending issues (created on this date and still pending)
	err = r.db.Model(&model.Issue{}).Where("created_at >= ? AND created_at < ? AND status NOT IN ?", startTime, endTime, []string{"resolved", "closed"}).Count(&stats.PendingIssues).Error
	if err != nil {
		return nil, err
	}

	// Resolved issues
	err = r.db.Model(&model.Issue{}).Where("created_at >= ? AND created_at < ? AND status IN ?", startTime, endTime, []string{"resolved", "closed"}).Count(&stats.ResolvedIssues).Error
	if err != nil {
		return nil, err
	}

	// Priority breakdown
	err = r.db.Model(&model.Issue{}).Where("created_at >= ? AND created_at < ? AND priority = ?", startTime, endTime, "P0").Count(&stats.P0Issues).Error
	if err != nil {
		return nil, err
	}

	err = r.db.Model(&model.Issue{}).Where("created_at >= ? AND created_at < ? AND priority = ?", startTime, endTime, "P1").Count(&stats.P1Issues).Error
	if err != nil {
		return nil, err
	}

	err = r.db.Model(&model.Issue{}).Where("created_at >= ? AND created_at < ? AND priority = ?", startTime, endTime, "P2").Count(&stats.P2Issues).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetWeeklyStats gets weekly issue statistics
func (r *ReportRepository) GetWeeklyStats(weekStart, weekEnd string) (*model.WeeklyReportStats, error) {
	stats := &model.WeeklyReportStats{}

	startTime, _ := time.Parse("2006-01-02", weekStart)
	endTime, _ := time.Parse("2006-01-02", weekEnd)
	endTime = endTime.Add(24 * time.Hour)

	// Total issues in this week
	err := r.db.Model(&model.Issue{}).Where("created_at >= ? AND created_at < ?", startTime, endTime).Count(&stats.TotalIssues).Error
	if err != nil {
		return nil, err
	}

	// Resolved issues
	err = r.db.Model(&model.Issue{}).Where("created_at >= ? AND created_at < ? AND status IN ?", startTime, endTime, []string{"resolved", "closed"}).Count(&stats.ResolvedIssues).Error
	if err != nil {
		return nil, err
	}

	// Unresolved issues
	err = r.db.Model(&model.Issue{}).Where("created_at >= ? AND created_at < ? AND status NOT IN ?", startTime, endTime, []string{"resolved", "closed"}).Count(&stats.UnresolvedIssues).Error
	if err != nil {
		return nil, err
	}

	// Critical issues (P0 and P1)
	err = r.db.Model(&model.Issue{}).Where("created_at >= ? AND created_at < ? AND priority IN ?", startTime, endTime, []string{"P0", "P1"}).Count(&stats.CriticalIssues).Error
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetDailyReportsInWeek 获取一周内的日报列表
func (r *ReportRepository) GetDailyReportsInWeek(weekStart, weekEnd string) ([]model.Report, error) {
	var reports []model.Report

	startTime, _ := time.Parse("2006-01-02", weekStart)
	endTime, _ := time.Parse("2006-01-02", weekEnd)
	endTime = endTime.Add(24 * time.Hour)

	err := r.db.Where("report_type = ? AND created_at >= ? AND created_at < ?", model.ReportTypeDaily, startTime, endTime).
		Order("report_date ASC").
		Find(&reports).Error
	return reports, err
}

// AggregateDailyReports 汇总日报数据
func (r *ReportRepository) AggregateDailyReports(weekStart, weekEnd string) (*model.WeeklyReportStats, []DailyReportSummary, error) {
	reports, err := r.GetDailyReportsInWeek(weekStart, weekEnd)
	if err != nil {
		return nil, nil, err
	}

	stats := &model.WeeklyReportStats{}
	var dailySummaries []DailyReportSummary

	for _, report := range reports {
		// 解析日报内容
		var content struct {
			Stats struct {
				TotalIssues     int64 `json:"total_issues"`
				PendingIssues   int64 `json:"pending_issues"`
				ResolvedIssues  int64 `json:"resolved_issues"`
				P0Issues        int64 `json:"p0_issues"`
				P1Issues        int64 `json:"p1_issues"`
				P2Issues        int64 `json:"p2_issues"`
			} `json:"stats"`
		}

		if report.Content != "" {
			json.Unmarshal([]byte(report.Content), &content)
		}

		// 汇总统计
		stats.TotalIssues += content.Stats.TotalIssues
		stats.ResolvedIssues += content.Stats.ResolvedIssues
		stats.UnresolvedIssues += content.Stats.PendingIssues
		stats.CriticalIssues += content.Stats.P0Issues + content.Stats.P1Issues

		dailySummaries = append(dailySummaries, DailyReportSummary{
			Date:          report.ReportDate,
			TotalIssues:   content.Stats.TotalIssues,
			P0Issues:      content.Stats.P0Issues,
			P1Issues:      content.Stats.P1Issues,
			P2Issues:      content.Stats.P2Issues,
			ResolvedIssues: content.Stats.ResolvedIssues,
		})
	}

	return stats, dailySummaries, nil
}

// DailyReportSummary 日报摘要
type DailyReportSummary struct {
	Date           string `json:"date"`
	TotalIssues    int64  `json:"total_issues"`
	P0Issues       int64  `json:"p0_issues"`
	P1Issues       int64  `json:"p1_issues"`
	P2Issues       int64  `json:"p2_issues"`
	ResolvedIssues int64  `json:"resolved_issues"`
}

// GetIssuesForDateRange gets issues created in a date range
func (r *ReportRepository) GetIssuesForDateRange(startDate, endDate string) ([]model.Issue, error) {
	var issues []model.Issue

	startTime, _ := time.Parse("2006-01-02", startDate)
	endTime, _ := time.Parse("2006-01-02", endDate)
	endTime = endTime.Add(24 * time.Hour)

	err := r.db.Preload("Service").Preload("Project").
		Where("created_at >= ? AND created_at < ?", startTime, endTime).
		Order("created_at DESC").
		Find(&issues).Error
	return issues, err
}

// GetIssuesByPriority gets issues by priority for incident review
func (r *ReportRepository) GetIssuesByPriority(priorities []string, limit int) ([]model.Issue, error) {
	var issues []model.Issue

	query := r.db.Preload("Service").Preload("Project").Preload("Creator")
	if len(priorities) > 0 {
		query = query.Where("priority IN ?", priorities)
	}
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Order("created_at DESC").Find(&issues).Error
	return issues, err
}