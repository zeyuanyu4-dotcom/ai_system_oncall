package service

import (
	"fmt"
	"time"

	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/model"
	"ai_system_oncall/internal/repository"
)

type DashboardService struct {
	dashboardRepo  *repository.DashboardRepository
	projectMemberRepo *repository.ProjectMemberRepository
	projectRepo *repository.ProjectRepository
}

func NewDashboardService(
	dashboardRepo *repository.DashboardRepository,
	projectMemberRepo *repository.ProjectMemberRepository,
	projectRepo *repository.ProjectRepository,
) *DashboardService {
	return &DashboardService{
		dashboardRepo:     dashboardRepo,
		projectMemberRepo: projectMemberRepo,
		projectRepo:       projectRepo,
	}
}

// GetDashboardStats 获取看板统计数据
func (s *DashboardService) GetDashboardStats(userRole string, userID uint64) (*dto.DashboardStatsResponse, error) {
	var projectID *uint64
	var managedProjectIDs []uint64

	if userRole == "system_admin" {
		// 系统管理员看全局统计
		projectID = nil
	} else if userRole == "project_admin" {
		// 项目管理员看自己管理的项目
		var err error
		managedProjectIDs, err = s.dashboardRepo.GetManagedProjectIDs(userID)
		if err != nil || len(managedProjectIDs) == 0 {
			// 没有管理的项目，返回空数据而不是错误
			return &dto.DashboardStatsResponse{}, nil
		}
		// 如果只有一个项目，使用该项目ID
		if len(managedProjectIDs) == 1 {
			projectID = &managedProjectIDs[0]
		}
		// 多个项目时 projectID 为 nil，后续需要特殊处理
	} else {
		// 其他角色返回空数据
		return &dto.DashboardStatsResponse{}, nil
	}

	// 获取昨天的统计数据
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	stat, err := s.dashboardRepo.GetStatByDate(yesterday, projectID)
	if err != nil {
		// 如果没有昨天的统计，实时计算
		if userRole == "project_admin" && len(managedProjectIDs) > 1 {
			// 多个项目汇总
			stat, err = s.dashboardRepo.CalculateMultiProjectStats(yesterday, managedProjectIDs)
		} else {
			stat, err = s.dashboardRepo.CalculateDailyStats(yesterday, projectID)
		}
		if err != nil {
			stat = &model.StatDailyRecord{}
		}
	}

	// 获取问题类型分布（实时查询）
	// 对于项目管理员多项目的情况，需要特殊处理
	var typeStats []model.IssueTypeStat
	var serviceStats []model.ServiceIssueStat

	if userRole == "project_admin" && len(managedProjectIDs) > 1 {
		typeStats, _ = s.dashboardRepo.GetIssueTypeDistributionByProjects(managedProjectIDs)
		serviceStats, _ = s.dashboardRepo.GetServiceIssueRankingByProjects(managedProjectIDs, 10)
	} else {
		typeStats, _ = s.dashboardRepo.GetIssueTypeDistribution(projectID)
		serviceStats, _ = s.dashboardRepo.GetServiceIssueRanking(projectID, 10)
	}

	// 计算AI采纳率
	aiAdoptRate := float64(0)
	if stat.AIAnalysisCount > 0 {
		aiAdoptRate = float64(stat.AIAdoptedCount) / float64(stat.AIAnalysisCount) * 100
	}

	// 构建响应
	response := &dto.DashboardStatsResponse{
		TotalIssues:      stat.TotalIssues,
		NewIssues:        stat.NewIssues,
		ResolvedIssues:   stat.ResolvedIssues,
		UnresolvedIssues: stat.UnresolvedIssues,
		P0P1Issues:       stat.P0P1Issues,
		AvgResolveTime:   float64(stat.AvgResolveMinutes) / 60, // 转为小时
		RepeatedIssues:   stat.RepeatedIssues,
		AIAnalysisCount:  stat.AIAnalysisCount,
		AIAdoptedCount:   stat.AIAdoptedCount,
		AIAdoptRate:      aiAdoptRate,
	}

	// 问题类型分布（计算百分比）
	totalTypeCount := int64(0)
	for _, t := range typeStats {
		totalTypeCount += t.Count
	}
	for _, t := range typeStats {
		percent := "0%"
		if totalTypeCount > 0 {
			percent = fmt.Sprintf("%.1f%%", float64(t.Count)/float64(totalTypeCount)*100)
		}
		response.IssueTypeDistribution = append(response.IssueTypeDistribution, dto.IssueTypeDistItem{
			IssueType: t.IssueType,
			Count:     t.Count,
			Percent:   percent,
		})
	}

	// 服务排行
	for _, s := range serviceStats {
		response.ServiceIssueRanking = append(response.ServiceIssueRanking, dto.ServiceRankItem{
			ServiceID:   s.ServiceID,
			ServiceName: s.ServiceName,
			Count:       s.Count,
		})
	}

	// 系统管理员额外返回项目问题排行
	if userRole == "system_admin" {
		projectStats, _ := s.dashboardRepo.GetProjectIssueRanking(10)
		for _, p := range projectStats {
			response.ProjectIssueRanking = append(response.ProjectIssueRanking, dto.ProjectRankItem{
				ProjectID:   p.ProjectID,
				ProjectName: p.ProjectName,
				Count:       p.Count,
			})
		}
	}

	return response, nil
}

// GetTrendData 获取趋势图数据
func (s *DashboardService) GetTrendData(startDate, endDate string, projectID *uint64) (*dto.DashboardTrendResponse, error) {
	records, err := s.dashboardRepo.GetStatsByDateRange(startDate, endDate, projectID)
	if err != nil {
		return nil, err
	}

	response := &dto.DashboardTrendResponse{
		Dates:  make([]string, 0),
		Series: make([]dto.TrendSeriesItem, 0),
	}

	// 准备数据序列
	newIssuesSeries := dto.TrendSeriesItem{Name: "新增问题", Data: make([]int64, 0)}
	resolvedSeries := dto.TrendSeriesItem{Name: "已解决", Data: make([]int64, 0)}
	p0p1Series := dto.TrendSeriesItem{Name: "高优问题", Data: make([]int64, 0)}
	aiAnalysisSeries := dto.TrendSeriesItem{Name: "AI分析", Data: make([]int64, 0)}

	for _, r := range records {
		response.Dates = append(response.Dates, r.StatDate)
		newIssuesSeries.Data = append(newIssuesSeries.Data, r.NewIssues)
		resolvedSeries.Data = append(resolvedSeries.Data, r.ResolvedIssues)
		p0p1Series.Data = append(p0p1Series.Data, r.P0P1Issues)
		aiAnalysisSeries.Data = append(aiAnalysisSeries.Data, r.AIAnalysisCount)
	}

	response.Series = append(response.Series,
		newIssuesSeries,
		resolvedSeries,
		p0p1Series,
		aiAnalysisSeries,
	)

	return response, nil
}

// GenerateDailyStat 生成每日统计（定时任务调用）
func (s *DashboardService) GenerateDailyStat(date string) error {
	if date == "" {
		// 默认统计昨天
		date = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	}

	// 1. 生成全局统计
	globalStat, err := s.dashboardRepo.CalculateDailyStats(date, nil)
	if err != nil {
		return err
	}
	if err := s.dashboardRepo.UpsertStat(globalStat); err != nil {
		return err
	}

	// 2. 可选：为每个项目生成单独统计
	// 这里简化处理，只生成全局统计

	return nil
}

// GetUserProjectFilter 获取用户的项目过滤条件
func (s *DashboardService) GetUserProjectFilter(userRole string, userID uint64) (*uint64, error) {
	if userRole == "system_admin" {
		return nil, nil // 系统管理员看全部
	}

	if userRole == "project_admin" {
		projectIDs, err := s.dashboardRepo.GetManagedProjectIDs(userID)
		if err != nil || len(projectIDs) == 0 {
			return nil, nil
		}
		return &projectIDs[0], nil
	}

	return nil, nil
}
