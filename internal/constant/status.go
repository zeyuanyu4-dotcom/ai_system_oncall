package constant

// Issue status
const (
	StatusPendingAnalysis   = "pending_analysis"
	StatusPendingAssignment = "pending_assignment"
	StatusProcessing        = "processing"
	StatusPendingConfirmation = "pending_confirmation"
	StatusResolved          = "resolved"
	StatusClosed            = "closed"
	StatusReopened          = "reopened"
)

// Issue types
const (
	IssueTypeApiError        = "api_error"
	IssueTypeDataIssue       = "data_issue"
	IssueTypePermissionIssue = "permission_issue"
	IssueTypePerformanceIssue = "performance_issue"
	IssueTypeReleaseIssue    = "release_issue"
	IssueTypeConfigIssue     = "config_issue"
	IssueTypeJobIssue        = "job_issue"
	IssueTypeEnvironmentIssue = "environment_issue"
	IssueTypeOther           = "other"
)

// Priority levels
const (
	PriorityP0 = "P0"
	PriorityP1 = "P1"
	PriorityP2 = "P2"
	PriorityP3 = "P3"
)

// Environments
const (
	EnvDev     = "dev"
	EnvTest    = "test"
	EnvStaging = "staging"
	EnvProd    = "prod"
)

// Comment types
const (
	CommentTypeComment      = "comment"
	CommentTypeProcessRecord = "process_record"
	CommentTypeSolution     = "solution"
	CommentTypeAiAnalysis   = "ai_analysis"
	CommentTypeSystem       = "system"
)

// Visibility
const (
	VisibilityPublic   = "public"
	VisibilityInternal = "internal"
)

// Service types
const (
	ServiceTypeBackend   = "backend"
	ServiceTypeFrontend  = "frontend"
	ServiceTypeJob       = "job"
	ServiceTypeMiddleware = "middleware"
	ServiceTypeDatabase  = "database"
	ServiceTypeGateway   = "gateway"
)

// Dependency types
const (
	DependencyTypeHttp     = "http"
	DependencyTypeGrpc     = "grpc"
	DependencyTypeMysql    = "mysql"
	DependencyTypeRedis    = "redis"
	DependencyTypeMq       = "mq"
	DependencyTypeJob      = "job"
	DependencyTypeExternal = "external"
)

// Operation types
const (
	OperationCreateIssue   = "create_issue"
	OperationUpdateIssue   = "update_issue"
	OperationAssignIssue   = "assign_issue"
	OperationChangeStatus  = "change_status"
	OperationAddComment    = "add_comment"
	OperationUpdatePriority = "update_priority"
	OperationUpdateService = "update_service"
	OperationResolveIssue  = "resolve_issue"
	OperationCloseIssue    = "close_issue"
	OperationReopenIssue   = "reopen_issue"
)
