package router

import (
	"context"

	"ai_system_oncall/internal/client"
	"ai_system_oncall/internal/config"
	"ai_system_oncall/internal/database"
	"ai_system_oncall/internal/grpcclient"
	"ai_system_oncall/internal/repository"
	"ai_system_oncall/internal/service"
)

// Services 集中管理所有初始化的 service 实例（供 HTTP handler + gRPC server 共用）
type Services struct {
	AuthService                 *service.AuthService
	UserService                 *service.UserService
	ProjectService              *service.ProjectService
	ProjectMemberService        *service.ProjectMemberService
	ServiceService              *service.ServiceService
	ServiceAPIService           *service.ServiceAPIService
	ServiceDependencyService    *service.ServiceDependencyService
	IssueService                *service.IssueService
	StatusService               *service.StatusService
	CommentService              *service.CommentService
	SimulatedLogService         *service.SimulatedLogService
	KnowledgeDocService         *service.KnowledgeDocService
	KnowledgeDocAttachmentService *service.KnowledgeDocAttachmentService
	AIAnalysisTaskService       *service.AIAnalysisTaskService
	ReportService               *service.ReportService
	DashboardService            *service.DashboardService
	AIClient                    *client.AIClient
}

// InitServices 初始化全部 service（依赖 database.Init 已就绪）
func InitServices() *Services {
	// Initialize repositories
	userRepo := repository.NewUserRepository(database.GetDB())
	projectRepo := repository.NewProjectRepository(database.GetDB())
	projectMemberRepo := repository.NewProjectMemberRepository(database.GetDB())
	serviceRepo := repository.NewServiceRepository(database.GetDB())
	serviceAPIRepo := repository.NewServiceAPIRepository(database.GetDB())
	serviceDependencyRepo := repository.NewServiceDependencyRepository(database.GetDB())
	issueRepo := repository.NewIssueRepository(database.GetDB())
	commentRepo := repository.NewCommentRepository(database.GetDB())
	statusLogRepo := repository.NewStatusLogRepository(database.GetDB())
	operationLogRepo := repository.NewOperationLogRepository(database.GetDB())
	simulatedLogRepo := repository.NewSimulatedLogRepository(database.GetDB())
	knowledgeDocRepo := repository.NewKnowledgeDocRepository(database.GetDB())
	knowledgeDocVersionRepo := repository.NewKnowledgeDocVersionRepository(database.GetDB())
	knowledgeDocAttachmentRepo := repository.NewKnowledgeDocAttachmentRepository(database.GetDB())
	aiAnalysisTaskRepo := repository.NewAIAnalysisTaskRepository(database.GetDB())
	reportRepo := repository.NewReportRepository(database.GetDB())
	dashboardRepo := repository.NewDashboardRepository(database.GetDB())

	// Initialize services
	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	projectService := service.NewProjectService(projectRepo, projectMemberRepo, userRepo)
	projectMemberService := service.NewProjectMemberService(projectMemberRepo, projectRepo, userRepo)
	serviceService := service.NewServiceService(serviceRepo, projectRepo)
	serviceAPIService := service.NewServiceAPIService(serviceAPIRepo, serviceRepo)
	serviceDependencyService := service.NewServiceDependencyService(serviceDependencyRepo, serviceRepo)
	issueService := service.NewIssueService(issueRepo, projectRepo, projectMemberRepo, serviceRepo, statusLogRepo, operationLogRepo)
	statusService := service.NewStatusService(issueRepo, projectMemberRepo, statusLogRepo, operationLogRepo, userRepo)
	commentService := service.NewCommentService(commentRepo, issueRepo, operationLogRepo)
	simulatedLogService := service.NewSimulatedLogService(simulatedLogRepo, projectRepo, serviceRepo, issueRepo, projectMemberRepo)
	knowledgeDocService := service.NewKnowledgeDocService(knowledgeDocRepo, knowledgeDocVersionRepo, projectRepo, serviceRepo)
	knowledgeDocAttachmentService := service.NewKnowledgeDocAttachmentService(knowledgeDocAttachmentRepo, knowledgeDocRepo)

	// Initialize AI client（HTTP 兜底）
	var aiClient *client.AIClient
	if config.GetConfig() != nil && config.GetConfig().AI.Enabled {
		aiClient = client.NewAIClient(&config.GetConfig().AI)
	}
	aiAnalysisTaskService := service.NewAIAnalysisTaskService(aiAnalysisTaskRepo, issueRepo, aiClient)
	dashboardService := service.NewDashboardService(dashboardRepo, projectMemberRepo, projectRepo)

	// 报告生成：优先 gRPC，否则 HTTP 兜底
	var reportAIClient service.AIAgentClient = aiClient
	if config.GetConfig() != nil && config.GetConfig().AI.GRPCAddr != "" {
		capClient, err := grpcclient.NewCapabilityClient(config.GetConfig().AI.GRPCAddr, config.GetConfig().AI.GRPCTimeout)
		if err == nil {
			reportAIClient = &grpcAIAdapter{cap: capClient}
		}
	}
	reportService := service.NewReportService(reportRepo, issueRepo, serviceRepo, reportAIClient)

	return &Services{
		AuthService:                    authService,
		UserService:                    userService,
		ProjectService:                 projectService,
		ProjectMemberService:           projectMemberService,
		ServiceService:                 serviceService,
		ServiceAPIService:              serviceAPIService,
		ServiceDependencyService:       serviceDependencyService,
		IssueService:                   issueService,
		StatusService:                  statusService,
		CommentService:                 commentService,
		SimulatedLogService:            simulatedLogService,
		KnowledgeDocService:            knowledgeDocService,
		KnowledgeDocAttachmentService:  knowledgeDocAttachmentService,
		AIAnalysisTaskService:          aiAnalysisTaskService,
		ReportService:                  reportService,
		DashboardService:               dashboardService,
		AIClient:                       aiClient,
	}
}

// grpcAIAdapter 让 grpcclient.CapabilityClient 满足 service.AIAgentClient 接口
type grpcAIAdapter struct {
	cap *grpcclient.CapabilityClient
}

func (a *grpcAIAdapter) GenerateText(prompt string) (string, error) {
	return a.cap.GenerateText(context.Background(), prompt, "", 2048)
}
