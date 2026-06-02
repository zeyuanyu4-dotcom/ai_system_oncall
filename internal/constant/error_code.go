package constant

// Error codes
const (
	CodeSuccess         = 0
	CodeUnauthorized    = 10001
	CodeForbidden       = 10002
	CodeBadRequest      = 10003
	CodeNotFound        = 10004
	CodeInvalidStatus   = 10005
	CodeNotInProject    = 10006
	CodeAlreadyExists   = 10007
	CodeDatabaseError   = 20001
	CodeInternalError   = 30001
)

// Error messages
var ErrorMessages = map[int]string{
	CodeSuccess:       "success",
	CodeUnauthorized:  "未登录或Token无效",
	CodeForbidden:     "无权限",
	CodeBadRequest:    "参数错误",
	CodeNotFound:      "数据不存在",
	CodeInvalidStatus: "状态流转非法",
	CodeNotInProject:  "用户不属于该项目",
	CodeAlreadyExists: "资源已存在",
	CodeDatabaseError: "数据库错误",
	CodeInternalError: "系统内部错误",
}

func GetErrorMessage(code int) string {
	if msg, ok := ErrorMessages[code]; ok {
		return msg
	}
	return "未知错误"
}
