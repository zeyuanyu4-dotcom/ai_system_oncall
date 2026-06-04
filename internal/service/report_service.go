package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"
)

type ReportService struct {
	reportRepo *repository.ReportRepository
	issueRepo  *repository.IssueRepository
	serviceRepo *repository.ServiceRepository
	aiClient   AIAgentClient
}

type AIAgentClient interface {
	GenerateText(prompt string) (string, error)
}

func NewReportService(
	reportRepo *repository.ReportRepository,
	issueRepo *repository.IssueRepository,
	serviceRepo *repository.ServiceRepository,
	aiClient AIAgentClient,
) *ReportService {
	return &ReportService{
		reportRepo:  reportRepo,
		issueRepo:   issueRepo,
		serviceRepo: serviceRepo,
		aiClient:    aiClient,
	}
}

// GenerateDailyReport 生成日报
func (s *ReportService) GenerateDailyReport(creatorID uint64, date string, isAuto bool) (*dto.ReportResponse, error) {
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// Check if report already exists
	existing, _ := s.reportRepo.FindByTypeAndDate(model.ReportTypeDaily, date)
	if existing != nil && existing.ID > 0 {
		return nil, errors.New("当日报告已存在")
	}

	// Create report record with generating status
	report := &model.Report{
		ReportType: model.ReportTypeDaily,
		ReportDate: date,
		Title:      fmt.Sprintf("日报 %s", date),
		CreatorID:  creatorID,
		IsAuto:     isAuto,
		Status:     model.ReportStatusGenerating,
	}
	if err := s.reportRepo.Create(report); err != nil {
		return nil, err
	}

	// Generate report content
	content, err := s.generateDailyReportContent(date)
	if err != nil {
		s.reportRepo.UpdateStatus(report.ID, model.ReportStatusFailed, err.Error())
		return nil, err
	}

	// Update report with generated content
	report.Summary = content.Analysis
	contentBytes, _ := json.Marshal(content)
	report.Content = string(contentBytes)
	report.IssueCount = int(content.Stats.TotalIssues)
	report.Status = model.ReportStatusGenerated
	if err := s.reportRepo.Update(report); err != nil {
		return nil, err
	}

	return &dto.ReportResponse{
		ID:         report.ID,
		ReportType: report.ReportType,
		Status:     report.Status,
		Message:    "日报生成成功",
	}, nil
}

// generateDailyReportContent 生成日报内容
func (s *ReportService) generateDailyReportContent(date string) (*dto.DailyReportContent, error) {
	// Get statistics
	stats, err := s.reportRepo.GetDailyStats(date)
	if err != nil {
		return nil, err
	}

	// Get issues
	issues, err := s.reportRepo.GetIssuesForDateRange(date, date)
	if err != nil {
		return nil, err
	}

	// Convert issues to DTO
	issueDTOs := make([]dto.IssueBriefDTO, len(issues))
	for i, issue := range issues {
		serviceName := ""
		if issue.Service != nil {
			serviceName = issue.Service.Name
		}
		issueDTOs[i] = dto.IssueBriefDTO{
			ID:          issue.ID,
			IssueNo:     issue.IssueNo,
			Title:       issue.Title,
			Priority:    issue.Priority,
			Status:      issue.Status,
			ServiceName: serviceName,
			CreatedAt:   issue.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	// Generate AI analysis
	analysisPrompt := fmt.Sprintf(`请分析以下当日告警统计数据，并生成简短的日报分析内容（约100字）：

统计数据：
- 总告警数: %d
- 待处理告警: %d
- 已解决告警: %d
- P0级告警: %d
- P1级告警: %d
- P2级告警: %d

请用简洁的语言总结当日告警情况，包括主要问题类型和处理建议。`, stats.TotalIssues, stats.PendingIssues, stats.ResolvedIssues, stats.P0Issues, stats.P1Issues, stats.P2Issues)

	analysis, _ := s.aiClient.GenerateText(analysisPrompt)
	if analysis == "" {
		analysis = fmt.Sprintf("今日共收到%d个告警，其中P0级%d个，P1级%d个。", stats.TotalIssues, stats.P0Issues, stats.P1Issues)
	}

	content := &dto.DailyReportContent{
		Date:     date,
		Stats:    dto.DailyReportStatsDTO{
			TotalIssues:     stats.TotalIssues,
			PendingIssues:  stats.PendingIssues,
			ResolvedIssues: stats.ResolvedIssues,
			P0Issues:       stats.P0Issues,
			P1Issues:       stats.P1Issues,
			P2Issues:       stats.P2Issues,
		},
		Issues:   issueDTOs,
		Analysis: analysis,
	}

	return content, nil
}

// GenerateWeeklyReport 生成周报
func (s *ReportService) GenerateWeeklyReport(creatorID uint64, weekStart string, isAuto bool) (*dto.ReportResponse, error) {
	if weekStart == "" {
		weekStart = s.getWeekStart(time.Now())
	}
	weekEnd := s.getWeekEnd(weekStart)
	week := s.getWeekString(weekStart)

	// Check if report already exists
	existing, _ := s.reportRepo.FindByTypeAndWeek(model.ReportTypeWeekly, week)
	if existing != nil && existing.ID > 0 {
		return nil, errors.New("本周报告已存在")
	}

	// Create report record
	report := &model.Report{
		ReportType: model.ReportTypeWeekly,
		ReportDate: weekStart,
		ReportWeek: week,
		Title:      fmt.Sprintf("周报 %s", week),
		CreatorID:  creatorID,
		IsAuto:     isAuto,
		Status:     model.ReportStatusGenerating,
	}
	if err := s.reportRepo.Create(report); err != nil {
		return nil, err
	}

	// Generate report content
	content, err := s.generateWeeklyReportContent(weekStart, weekEnd)
	if err != nil {
		s.reportRepo.UpdateStatus(report.ID, model.ReportStatusFailed, err.Error())
		return nil, err
	}

	// Update report
	report.Summary = content.Analysis
	contentBytes, _ := json.Marshal(content)
	report.Content = string(contentBytes)
	report.IssueCount = int(content.Stats.TotalIssues)
	report.Status = model.ReportStatusGenerated
	if err := s.reportRepo.Update(report); err != nil {
		return nil, err
	}

	return &dto.ReportResponse{
		ID:         report.ID,
		ReportType: report.ReportType,
		Status:     report.Status,
		Message:    "周报生成成功",
	}, nil
}

// generateWeeklyReportContent 生成周报内容
func (s *ReportService) generateWeeklyReportContent(weekStart, weekEnd string) (*dto.WeeklyReportContent, error) {
	// 汇总日报数据
	stats, dailySummaries, err := s.reportRepo.AggregateDailyReports(weekStart, weekEnd)
	if err != nil {
		return nil, err
	}

	// 如果没有日报，则直接查询 issues 表
	if stats.TotalIssues == 0 {
		stats, err = s.reportRepo.GetWeeklyStats(weekStart, weekEnd)
		if err != nil {
			return nil, err
		}
	}

	// Get issues
	issues, err := s.reportRepo.GetIssuesForDateRange(weekStart, weekEnd)
	if err != nil {
		return nil, err
	}

	// Convert issues to DTO
	issueDTOs := make([]dto.IssueBriefDTO, len(issues))
	for i, issue := range issues {
		serviceName := ""
		if issue.Service != nil {
			serviceName = issue.Service.Name
		}
		issueDTOs[i] = dto.IssueBriefDTO{
			ID:          issue.ID,
			IssueNo:     issue.IssueNo,
			Title:       issue.Title,
			Priority:    issue.Priority,
			Status:      issue.Status,
			ServiceName: serviceName,
			CreatedAt:   issue.CreatedAt.Format("2006-01-02 15:04"),
		}
	}

	// Calculate top services
	serviceCount := make(map[string]int)
	for _, issue := range issues {
		if issue.Service != nil {
			serviceCount[issue.Service.Name]++
		}
	}
	var topServices []dto.ServiceStatDTO
	for name, count := range serviceCount {
		topServices = append(topServices, dto.ServiceStatDTO{
			ServiceName: name,
			IssueCount:  count,
		})
	}

	// 转换日报摘要
	dailyDTOs := make([]dto.DailyReportSummaryDTO, len(dailySummaries))
	for i, d := range dailySummaries {
		dailyDTOs[i] = dto.DailyReportSummaryDTO{
			Date:           d.Date,
			TotalIssues:    d.TotalIssues,
			P0Issues:       d.P0Issues,
			P1Issues:       d.P1Issues,
			P2Issues:       d.P2Issues,
			ResolvedIssues: d.ResolvedIssues,
		}
	}

	// Generate AI analysis
	analysisPrompt := fmt.Sprintf(`请分析以下本周告警统计数据，并生成简短的周报分析内容（约150字）：

统计数据：
- 总告警数: %d
- 已解决告警: %d
- 未解决告警: %d
- 严重告警(P0/P1): %d

请总结本周告警的整体情况，包括主要问题类型、服务分布和改进建议。`, stats.TotalIssues, stats.ResolvedIssues, stats.UnresolvedIssues, stats.CriticalIssues)

	analysis, _ := s.aiClient.GenerateText(analysisPrompt)
	if analysis == "" {
		analysis = fmt.Sprintf("本周共收到%d个告警，已解决%d个，未解决%d个。严重告警%d个。", stats.TotalIssues, stats.ResolvedIssues, stats.UnresolvedIssues, stats.CriticalIssues)
	}

	content := &dto.WeeklyReportContent{
		WeekStart:      weekStart,
		WeekEnd:        weekEnd,
		Stats:          dto.WeeklyReportStatsDTO{
			TotalIssues:      stats.TotalIssues,
			ResolvedIssues:  stats.ResolvedIssues,
			UnresolvedIssues: stats.UnresolvedIssues,
			CriticalIssues:   stats.CriticalIssues,
			Trend:            stats.Trend,
		},
		DailySummaries: dailyDTOs,
		Issues:         issueDTOs,
		TopServices:    topServices,
		Analysis:       analysis,
	}

	return content, nil
}

// GenerateIncidentReview 生成事件复盘
func (s *ReportService) GenerateIncidentReview(creatorID uint64, issueIDs []uint64) (*dto.ReportResponse, error) {
	if len(issueIDs) == 0 {
		return nil, errors.New("请选择至少一个问题单")
	}

	// Create report record
	report := &model.Report{
		ReportType: model.ReportTypeIncident,
		Title:      fmt.Sprintf("事件复盘 %d个问题", len(issueIDs)),
		CreatorID:  creatorID,
		Status:     model.ReportStatusGenerating,
	}
	if err := s.reportRepo.Create(report); err != nil {
		return nil, err
	}

	// Generate incident review content
	content, err := s.generateIncidentReviewContent(issueIDs)
	if err != nil {
		s.reportRepo.UpdateStatus(report.ID, model.ReportStatusFailed, err.Error())
		return nil, err
	}

	// Update report
	report.Summary = content.Summary
	contentBytes, _ := json.Marshal(content)
	report.Content = string(contentBytes)
	report.IssueCount = content.IssueCount
	report.Status = model.ReportStatusGenerated
	if err := s.reportRepo.Update(report); err != nil {
		return nil, err
	}

	return &dto.ReportResponse{
		ID:         report.ID,
		ReportType: report.ReportType,
		Status:     report.Status,
		Message:    "事件复盘生成成功",
	}, nil
}

// generateIncidentReviewContent 生成事件复盘内容
func (s *ReportService) generateIncidentReviewContent(issueIDs []uint64) (*dto.IncidentReviewContent, error) {
	var reviews []dto.IncidentReviewDTO

	for _, issueID := range issueIDs {
		issue, err := s.issueRepo.FindByID(issueID)
		if err != nil {
			continue
		}

		// Generate incident analysis for this issue
		analysisPrompt := fmt.Sprintf(`请分析以下问题单，生成事件复盘内容：

问题单信息：
- 编号: %s
- 标题: %s
- 优先级: %s
- 状态: %s
- 创建时间: %s
- AI分析: %s
- 根因分析: %s
- 解决方案: %s

请以JSON格式输出复盘内容，包括：
- root_cause: 根因分析
- affected_services: 受影响服务列表
- impact: 影响范围
- resolution: 解决方案总结
- lessons_learned: 经验教训（数组）
- action_items: 后续行动项（数组）`, issue.IssueNo, issue.Title, issue.Priority, issue.Status, issue.CreatedAt.Format("2006-01-02 15:04"), issue.AISummary, issue.RootCause, issue.Solution)

		analysis, _ := s.aiClient.GenerateText(analysisPrompt)

		serviceName := ""
		if issue.Service != nil {
			serviceName = issue.Service.Name
		}

		duration := 0
		if issue.ResolvedAt != nil {
			duration = int(issue.ResolvedAt.Sub(issue.CreatedAt).Minutes())
		}

		review := dto.IncidentReviewDTO{
			IssueID:          issue.ID,
			IssueNo:          issue.IssueNo,
			Title:            issue.Title,
			Priority:         issue.Priority,
			IncidentTime:     issue.CreatedAt.Format("2006-01-02 15:04"),
			Duration:         duration,
			AffectedServices: []string{serviceName},
			Impact:           issue.ImpactScope,
			RootCause:        issue.RootCause,
			Resolution:       issue.Solution,
		}

		if analysis != "" {
			// Try to parse JSON response
			var parsed map[string]interface{}
			if json.Unmarshal([]byte(analysis), &parsed) == nil {
				if v, ok := parsed["root_cause"].(string); ok {
					review.RootCause = v
				}
				if v, ok := parsed["impact"].(string); ok {
					review.Impact = v
				}
				if v, ok := parsed["resolution"].(string); ok {
					review.Resolution = v
				}
				if v, ok := parsed["affected_services"].([]interface{}); ok {
					review.AffectedServices = make([]string, len(v))
					for i, s := range v {
						review.AffectedServices[i] = fmt.Sprintf("%v", s)
					}
				}
				if v, ok := parsed["lessons_learned"].([]interface{}); ok {
					review.LessonsLearned = make([]string, len(v))
					for i, s := range v {
						review.LessonsLearned[i] = fmt.Sprintf("%v", s)
					}
				}
				if v, ok := parsed["action_items"].([]interface{}); ok {
					review.ActionItems = make([]string, len(v))
					for i, s := range v {
						review.ActionItems[i] = fmt.Sprintf("%v", s)
					}
				}
			}
		}

		reviews = append(reviews, review)
	}

	// Generate summary
	summary := fmt.Sprintf("本次复盘共涉及%d个问题单，包括%d个P0/P1级严重问题。", len(reviews), len(reviews))

	content := &dto.IncidentReviewContent{
		IssueCount: len(reviews),
		Reviews:    reviews,
		Summary:    summary,
	}

	return content, nil
}

// GetReport 获取报告
func (s *ReportService) GetReport(id uint64) (*dto.ReportInfo, error) {
	report, err := s.reportRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	return s.convertToReportInfo(report), nil
}

// GetDailyReport 获取指定日期的日报
func (s *ReportService) GetDailyReport(date string) (*dto.ReportInfo, error) {
	report, err := s.reportRepo.FindByTypeAndDate(model.ReportTypeDaily, date)
	if err != nil {
		return nil, err
	}

	return s.convertToReportInfo(report), nil
}

// GetWeeklyReport 获取指定周的周报
func (s *ReportService) GetWeeklyReport(weekStart string) (*dto.ReportInfo, error) {
	week := s.getWeekString(weekStart)
	report, err := s.reportRepo.FindByTypeAndWeek(model.ReportTypeWeekly, week)
	if err != nil {
		return nil, err
	}

	return s.convertToReportInfo(report), nil
}

// ListReports 列出报告
func (s *ReportService) ListReports(req *dto.ReportListRequest) (*dto.ReportListResponse, error) {
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	var reports []model.Report
	var total int64
	var err error

	if req.ReportType != "" {
		reports, total, err = s.reportRepo.FindByType(req.ReportType, page, pageSize)
	} else {
		reports, total, err = s.reportRepo.FindAll(page, pageSize)
	}
	if err != nil {
		return nil, err
	}

	reportInfos := make([]dto.ReportInfo, len(reports))
	for i, r := range reports {
		reportInfos[i] = *s.convertToReportInfo(&r)
	}

	return &dto.ReportListResponse{
		Reports:  reportInfos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// convertToReportInfo converts model.Report to dto.ReportInfo
func (s *ReportService) convertToReportInfo(report *model.Report) *dto.ReportInfo {
	info := &dto.ReportInfo{
		ID:         report.ID,
		ReportType: report.ReportType,
		ReportDate: report.ReportDate,
		ReportWeek: report.ReportWeek,
		Title:      report.Title,
		Summary:    report.Summary,
		Content:    report.Content,
		IssueCount: report.IssueCount,
		IsAuto:     report.IsAuto,
		Status:     report.Status,
		CreatorID:  report.CreatorID,
		CreatedAt:  report.CreatedAt,
		UpdatedAt:  report.UpdatedAt,
	}

	if report.Creator != nil {
		info.Creator = &dto.UserInfo{
			ID:       report.Creator.ID,
			Username: report.Creator.Username,
			Email:    report.Creator.Email,
		}
	}

	return info
}

// getWeekStart 获取指定日期所在周周一
func (s *ReportService) getWeekStart(t time.Time) string {
	weekday := int(t.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	weekStart := t.AddDate(0, 0, -weekday+1)
	return weekStart.Format("2006-01-02")
}

// getWeekEnd 获取指定日期所在周周日
func (s *ReportService) getWeekEnd(weekStart string) string {
	t, _ := time.Parse("2006-01-02", weekStart)
	weekEnd := t.AddDate(0, 0, 6)
	return weekEnd.Format("2006-01-02")
}

// getWeekString 获取周字符串 (2026-W23)
func (s *ReportService) getWeekString(weekStart string) string {
	t, _ := time.Parse("2006-01-02", weekStart)
	year, week := t.ISOWeek()
	return fmt.Sprintf("%d-W%02d", year, week)
}