package response

// 错误码按模块分配
// 格式: 模块前缀 + 错误类型

const (
	// ========== 通用错误 10000-10999 ==========
	CodeSuccess         = 0     // 成功
	CodeUnknownError    = 10000 // 未知错误
	CodeInvalidParam    = 10001 // 参数错误
	CodeInvalidJSON     = 10002 // JSON格式错误
	CodeRequestTooLarge = 10003 // 请求体过大

	// ========== 用户/权限错误 11000-11999 ==========
	CodeUnauthorized    = 11001 // 未登录或token无效
	CodeTokenExpired    = 11002 // Token已过期
	CodeForbidden       = 11003 // 无权限访问
	CodeUserNotFound    = 11004 // 用户不存在
	CodeUserDisabled    = 11005 // 用户已禁用
	CodePasswordError   = 11006 // 密码错误
	CodeUserExists      = 11007 // 用户已存在

	// ========== 项目错误 12000-12999 ==========
	CodeProjectNotFound = 12001 // 项目不存在
	CodeProjectDisabled = 12002 // 项目已停用
	CodeNotProjectMember = 12003 // 不是项目成员
	CodeProjectExists   = 12004 // 项目已存在

	// ========== 服务错误 13000-13999 ==========
	CodeServiceNotFound = 13001 // 服务不存在
	CodeServiceDisabled = 13002 // 服务已下线
	CodeServiceExists   = 13003 // 服务已存在
	CodeAPINotFound     = 13004 // API不存在

	// ========== 问题错误 14000-14999 ==========
	CodeIssueNotFound   = 14001 // 问题不存在
	CodeIssueClosed     = 14002 // 问题已关闭
	CodeInvalidStatus   = 14003 // 无效的状态
	CodeInvalidPriority = 14004 // 无效的优先级
	CodeCommentNotFound = 14005 // 评论不存在

	// ========== 知识库错误 15000-15999 ==========
	CodeDocNotFound     = 15001 // 文档不存在
	CodeDocVersionNotFound = 15002 // 文档版本不存在
	CodeAttachmentNotFound = 15003 // 附件不存在
	CodeAttachmentTooLarge = 15004 // 附件过大
	CodeUnsupportedFileType = 15005 // 不支持的文件类型

	// ========== AI分析错误 16000-16999 ==========
	CodeAITaskNotFound  = 16001 // AI任务不存在
	CodeAITaskRunning   = 16002 // AI任务运行中
	CodeAITaskCancelled = 16003 // AI任务已取消
	CodeAITaskFailed    = 16004 // AI任务失败
	CodeAIServiceError  = 16005 // AI服务错误

	// ========== 报告错误 17000-17999 ==========
	CodeReportNotFound  = 17001 // 报告不存在
	CodeReportExists    = 17002 // 报告已存在
	CodeReportGenerating = 17003 // 报告生成中

	// ========== 看板错误 18000-18999 ==========
	CodeDashboardNoPermission = 18001 // 无权限查看看板
	CodeStatNotFound    = 18002 // 统计数据不存在

	// ========== 日志错误 19000-19999 ==========
	CodeLogNotFound     = 19001 // 日志不存在
	CodeLogParseError   = 19002 // 日志解析错误

	// ========== 系统错误 50000-59999 ==========
	CodeDBError         = 50001 // 数据库错误
	CodeCacheError      = 50002 // 缓存错误
	CodeInternalError   = 50003 // 内部服务错误
	CodeServiceUnavailable = 50004 // 服务不可用
	CodeTooManyRequests = 42900 // 请求过于频繁
)
