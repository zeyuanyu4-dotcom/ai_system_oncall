package router

import (
	"ai_system_oncall/internal/client"
	"ai_system_oncall/internal/config"
	"ai_system_oncall/internal/database"
	"ai_system_oncall/internal/handler"
	"ai_system_oncall/internal/middleware"
	"ai_system_oncall/internal/repository"
	"ai_system_oncall/internal/service"

	"github.com/gin-gonic/gin"
)

// SetupRouter initializes all routes
func SetupRouter() *gin.Engine {
	r := gin.New()

	// Middlewares
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())
	r.Use(middleware.CORS())

	// Serve static files for frontend
	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/index.html", "./web/index.html")

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

	// Initialize AI client
	var aiClient *client.AIClient
	if config.GetConfig() != nil && config.GetConfig().AI.Enabled {
		aiClient = client.NewAIClient(&config.GetConfig().AI)
	}
	aiAnalysisTaskService := service.NewAIAnalysisTaskService(aiAnalysisTaskRepo, issueRepo, aiClient)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService)
	userHandler := handler.NewUserHandler(userService)
	projectHandler := handler.NewProjectHandler(projectService)
	projectMemberHandler := handler.NewProjectMemberHandler(projectMemberService)
	serviceHandler := handler.NewServiceHandler(serviceService)
	serviceAPIHandler := handler.NewServiceAPIHandler(serviceAPIService)
	serviceDependencyHandler := handler.NewServiceDependencyHandler(serviceDependencyService)
	issueHandler := handler.NewIssueHandler(issueService)
	statusHandler := handler.NewStatusHandler(statusService)
	commentHandler := handler.NewCommentHandler(commentService)
	simulatedLogHandler := handler.NewSimulatedLogHandler(simulatedLogService)
	knowledgeDocHandler := handler.NewKnowledgeDocHandler(knowledgeDocService)
	knowledgeDocAttachmentHandler := handler.NewKnowledgeDocAttachmentHandler(knowledgeDocAttachmentService)
	aiHandler := handler.NewAIHandler(issueService, aiClient)
	aiTaskHandler := handler.NewAIAnalysisTaskHandler(aiAnalysisTaskService)

	// Public routes (no auth required)
	public := r.Group("/api")
	{
		public.POST("/auth/register", authHandler.Register)
		public.POST("/auth/login", authHandler.Login)
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
		protected.POST("/knowledge-docs/:id/vectorize", knowledgeDocHandler.TriggerVectorization)
		protected.GET("/knowledge-docs/:id/versions", knowledgeDocHandler.GetVersions)
		// Knowledge Documents by project
		protected.GET("/projects/:id/knowledge-docs", knowledgeDocHandler.GetDocumentsByProject)
		// Knowledge Documents by service
		protected.GET("/services/:id/knowledge-docs", knowledgeDocHandler.GetDocumentsByService)
		// Knowledge Document Attachments
		protected.POST("/knowledge-docs/:id/attachments", knowledgeDocAttachmentHandler.UploadAttachment)
		protected.GET("/knowledge-docs/:id/attachments", knowledgeDocAttachmentHandler.GetAttachments)
		protected.GET("/knowledge-docs/:id/attachments/:aid", knowledgeDocAttachmentHandler.DownloadAttachment)
		protected.GET("/knowledge-docs/:id/attachments/:aid/content", knowledgeDocAttachmentHandler.GetAttachmentContent)
		protected.DELETE("/knowledge-docs/:id/attachments/:aid", knowledgeDocAttachmentHandler.DeleteAttachment)
		protected.POST("/knowledge-docs/parse-attachment", knowledgeDocAttachmentHandler.ParseAttachmentToContent)

		// AI Analysis
		protected.POST("/issues/:id/ai-analysis", aiHandler.AnalyzeIssue)
		protected.GET("/ai/status", aiHandler.AIAnalysisStatus)

		// AI Agent Analysis Tasks
		protected.POST("/issues/:id/agent-tasks", aiTaskHandler.CreateTask)
		protected.GET("/issues/:id/agent-tasks", aiTaskHandler.GetIssueTasks)
		protected.GET("/agent-tasks/:task_id", aiTaskHandler.GetTask)
		protected.POST("/agent-tasks/:task_id/cancel", aiTaskHandler.CancelTask)
	}

	return r
}
