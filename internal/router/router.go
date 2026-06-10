package router

import (
	"ai_system_oncall/internal/handler"
	"ai_system_oncall/internal/middleware"

	"github.com/gin-gonic/gin"
)

// SetupRouterWithServices 接收已初始化的 service 列表（供 HTTP + gRPC 共用）
func SetupRouterWithServices(svcs *Services) *gin.Engine {
	r := gin.New()

	// Middlewares
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// Serve static files for frontend
	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/index.html", "./web/index.html")

	// Initialize handlers (复用已初始化的 service)
	authHandler := handler.NewAuthHandler(svcs.AuthService)
	userHandler := handler.NewUserHandler(svcs.UserService)
	projectHandler := handler.NewProjectHandler(svcs.ProjectService)
	projectMemberHandler := handler.NewProjectMemberHandler(svcs.ProjectMemberService)
	serviceHandler := handler.NewServiceHandler(svcs.ServiceService)
	serviceAPIHandler := handler.NewServiceAPIHandler(svcs.ServiceAPIService)
	serviceDependencyHandler := handler.NewServiceDependencyHandler(svcs.ServiceDependencyService)
	issueHandler := handler.NewIssueHandler(svcs.IssueService)
	statusHandler := handler.NewStatusHandler(svcs.StatusService)
	commentHandler := handler.NewCommentHandler(svcs.CommentService)
	simulatedLogHandler := handler.NewSimulatedLogHandler(svcs.SimulatedLogService)
	knowledgeDocHandler := handler.NewKnowledgeDocHandler(svcs.KnowledgeDocService)
	knowledgeDocAttachmentHandler := handler.NewKnowledgeDocAttachmentHandler(svcs.KnowledgeDocAttachmentService)
	aiHandler := handler.NewAIHandler(svcs.IssueService, svcs.AIClient)
	aiTaskHandler := handler.NewAIAnalysisTaskHandler(svcs.AIAnalysisTaskService)
	reportHandler := handler.NewReportHandler(svcs.ReportService)
	dashboardHandler := handler.NewDashboardHandler(svcs.DashboardService)

	// Public routes (no auth required)
	public := r.Group("/api")
	{
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", middleware.LoginRateLimit(), authHandler.Login)
	}

	// Protected routes (auth required)
	protected := r.Group("/api")
	protected.Use(middleware.JWTAuth())
	{
		// Auth
		protected.GET("/auth/me", authHandler.GetCurrentUser)
		protected.POST("/auth/logout", authHandler.Logout)

		// Users
		protected.GET("/users", userHandler.ListUsers)
		protected.GET("/users/:id", userHandler.GetUser)
		protected.PUT("/users/:id", userHandler.UpdateUser)
		protected.PATCH("/users/:id/status", userHandler.UpdateUserStatus)
		protected.DELETE("/users/:id", userHandler.DeleteUser)

		// Projects
		protected.POST("/projects", projectHandler.CreateProject)
		protected.GET("/projects", projectHandler.ListProjects)
		protected.GET("/projects/:id", projectHandler.GetProject)
		protected.PUT("/projects/:id", projectHandler.UpdateProject)
		protected.DELETE("/projects/:id", projectHandler.DeleteProject)

		// Project Members
		protected.POST("/projects/:id/members", projectMemberHandler.AddMember)
		protected.GET("/projects/:id/members", projectMemberHandler.ListMembers)
		protected.PUT("/projects/:id/members/:user_id", projectMemberHandler.UpdateMemberRole)
		protected.DELETE("/projects/:id/members/:user_id", projectMemberHandler.RemoveMember)

		// Services (under project)
		protected.POST("/projects/:id/services", serviceHandler.CreateService)
		protected.GET("/projects/:id/services", serviceHandler.ListServices)
		// Services (all)
		protected.GET("/services", serviceHandler.ListAllServices)
		protected.GET("/services/:id", serviceHandler.GetService)
		protected.PUT("/services/:id", serviceHandler.UpdateService)
		protected.DELETE("/services/:id", serviceHandler.DeleteService)

		// Service APIs
		protected.POST("/services/:id/apis", serviceAPIHandler.CreateAPI)
		protected.GET("/services/:id/apis", serviceAPIHandler.ListAPIs)
		protected.PUT("/service-apis/:api_id", serviceAPIHandler.UpdateAPI)
		protected.DELETE("/service-apis/:api_id", serviceAPIHandler.DeleteAPI)

		// Service Dependencies
		protected.POST("/services/:id/dependencies", serviceDependencyHandler.CreateDependency)
		protected.GET("/services/:id/dependencies", serviceDependencyHandler.ListDependencies)
		protected.DELETE("/service-dependencies/:dependency_id", serviceDependencyHandler.DeleteDependency)

		// Issues
		protected.POST("/issues", issueHandler.CreateIssue)
		protected.GET("/issues", issueHandler.ListIssues)
		protected.GET("/issues/:id", issueHandler.GetIssue)
		protected.PUT("/issues/:id", issueHandler.UpdateIssue)
		protected.DELETE("/issues/:id", issueHandler.DeleteIssue)

		// Issue Status
		protected.PATCH("/issues/:id/assign", statusHandler.AssignIssue)
		protected.PATCH("/issues/:id/status", statusHandler.ChangeStatus)
		protected.GET("/issues/:id/status-logs", statusHandler.GetStatusLogs)

		// Comments
		protected.POST("/issues/:id/comments", commentHandler.CreateComment)
		protected.GET("/issues/:id/comments", commentHandler.ListComments)
		protected.DELETE("/comments/:id", commentHandler.DeleteComment)

		// Operation Logs
		protected.GET("/issues/:id/operation-logs", issueHandler.GetOperationLogs)

		// History Issue Query (历史问题查询)
		protected.GET("/issues/history/search", issueHandler.SearchHistoryIssues)
		protected.GET("/issues/:id/similar", issueHandler.GetSimilarIssues)

		// Simulated Logs (模拟日志)
		protected.POST("/logs", simulatedLogHandler.CreateLog)
		protected.POST("/logs/batch", simulatedLogHandler.BatchCreateLogs)
		protected.GET("/logs", simulatedLogHandler.ListLogs)
		protected.GET("/logs/:id", simulatedLogHandler.GetLog)
		protected.GET("/logs/trace/:trace_id", simulatedLogHandler.GetLogsByTraceID)
		protected.GET("/logs/issue/:id", simulatedLogHandler.GetLogsByIssue)
		protected.PATCH("/logs/:id/link-issue", simulatedLogHandler.LinkIssue)
		protected.DELETE("/logs/:id", simulatedLogHandler.DeleteLog)
		// Logs by service
		protected.GET("/services/:id/logs", simulatedLogHandler.GetLogsByService)

		// Knowledge Documents
		protected.POST("/knowledge-docs", knowledgeDocHandler.CreateDocument)
		protected.GET("/knowledge-docs", knowledgeDocHandler.ListDocuments)
		protected.GET("/knowledge-docs/search", knowledgeDocHandler.SearchDocuments)
		protected.GET("/knowledge-docs/by-type/:type", knowledgeDocHandler.GetDocumentsByType)
		protected.GET("/knowledge-docs/:id", knowledgeDocHandler.GetDocument)
		protected.PUT("/knowledge-docs/:id", knowledgeDocHandler.UpdateDocument)
		protected.DELETE("/knowledge-docs/:id", knowledgeDocHandler.DeleteDocument)
		protected.PUT("/knowledge-docs/:id/vector-status", knowledgeDocHandler.UpdateVectorStatus)
		protected.POST("/knowledge-docs/:id/vectorize", middleware.VectorizeRateLimit(), knowledgeDocHandler.TriggerVectorization)
		protected.GET("/knowledge-docs/:id/versions", knowledgeDocHandler.GetVersions)
		// Knowledge Documents by project
		protected.GET("/projects/:id/knowledge-docs", knowledgeDocHandler.GetDocumentsByProject)
		// Knowledge Documents by service
		protected.GET("/services/:id/knowledge-docs", knowledgeDocHandler.GetDocumentsByService)
		// Knowledge Document Attachments (with upload rate limit)
		protected.POST("/knowledge-docs/:id/attachments", middleware.UploadRateLimit(), knowledgeDocAttachmentHandler.UploadAttachment)
		protected.GET("/knowledge-docs/:id/attachments", knowledgeDocAttachmentHandler.GetAttachments)
		protected.GET("/knowledge-docs/:id/attachments/:aid", knowledgeDocAttachmentHandler.DownloadAttachment)
		protected.GET("/knowledge-docs/:id/attachments/:aid/content", knowledgeDocAttachmentHandler.GetAttachmentContent)
		protected.DELETE("/knowledge-docs/:id/attachments/:aid", knowledgeDocAttachmentHandler.DeleteAttachment)
		protected.POST("/knowledge-docs/parse-attachment", knowledgeDocAttachmentHandler.ParseAttachmentToContent)

		// AI Analysis (with rate limit)
		protected.POST("/issues/:id/ai-analysis", middleware.AIAnalysisRateLimit(), aiHandler.AnalyzeIssue)
		protected.GET("/ai/status", aiHandler.AIAnalysisStatus)

		// AI Agent Analysis Tasks (with rate limit)
		protected.POST("/issues/:id/agent-tasks", middleware.AIAnalysisRateLimit(), aiTaskHandler.CreateTask)
		protected.GET("/issues/:id/agent-tasks", aiTaskHandler.GetIssueTasks)
		protected.GET("/agent-tasks/:task_id", aiTaskHandler.GetTask)
		protected.POST("/agent-tasks/:task_id/cancel", aiTaskHandler.CancelTask)
		protected.POST("/agent-tasks/:task_id/progress", aiTaskHandler.UpdateTaskProgress)

		// Reports
		protected.POST("/reports/daily", reportHandler.GenerateDailyReport)
		protected.GET("/reports/daily/:date", reportHandler.GetDailyReport)
		protected.POST("/reports/daily/auto", reportHandler.GenerateDailyReportAuto)
		protected.POST("/reports/weekly", reportHandler.GenerateWeeklyReport)
		protected.GET("/reports/weekly/:week", reportHandler.GetWeeklyReport)
		protected.POST("/reports/incident", reportHandler.GenerateIncidentReview)
		protected.GET("/reports/incident/:id", reportHandler.GetIncidentReview)
		protected.GET("/reports", reportHandler.ListReports)
		protected.GET("/reports/:id", reportHandler.GetReport)

		// Dashboard (系统管理员和项目管理员可访问)
		protected.GET("/dashboard/stats", dashboardHandler.GetDashboardStats)
		protected.GET("/dashboard/trend", dashboardHandler.GetTrendData)
		protected.POST("/dashboard/generate-stat", dashboardHandler.GenerateDailyStat)
	}

	return r
}

// SetupRouter 兼容旧接口（自行初始化全部 service）
func SetupRouter() *gin.Engine {
	return SetupRouterWithServices(InitServices())
}
