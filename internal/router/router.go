package router

import (
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

	// Initialize services
	authService := service.NewAuthService(userRepo)
	userService := service.NewUserService(userRepo)
	projectService := service.NewProjectService(projectRepo, projectMemberRepo, userRepo)
	projectMemberService := service.NewProjectMemberService(projectMemberRepo, projectRepo, userRepo)
	serviceService := service.NewServiceService(serviceRepo, projectRepo)
	serviceAPIService := service.NewServiceAPIService(serviceAPIRepo, serviceRepo)
	serviceDependencyService := service.NewServiceDependencyService(serviceDependencyRepo, serviceRepo)
	issueService := service.NewIssueService(issueRepo, projectRepo, projectMemberRepo, serviceRepo, statusLogRepo, operationLogRepo)
	statusService := service.NewStatusService(issueRepo, projectMemberRepo, statusLogRepo, operationLogRepo)
	commentService := service.NewCommentService(commentRepo, issueRepo, operationLogRepo)

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
	}

	return r
}
