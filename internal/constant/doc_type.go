package constant

const (
	// DocType 文档类型
	DOC_TYPE_TROUBLESHOOTING_MANUAL = "troubleshooting_manual" // 故障排查手册
	DOC_TYPE_FAQ                    = "faq"                   // 常见问题处理方案
	DOC_TYPE_DEPLOYMENT_GUIDE       = "deployment_guide"      // 服务部署说明
	DOC_TYPE_API_DOC                = "api_doc"               // 接口说明
	DOC_TYPE_DB_DOC                 = "db_doc"                // 数据库说明
	DOC_TYPE_RELEASE_PROCESS        = "release_process"       // 发布流程规范
	DOC_TYPE_INCIDENT_REVIEW        = "incident_review"       // 历史故障复盘
	DOC_TYPE_ENVIRONMENT_GUIDE      = "environment_guide"     // 环境使用说明
)

var docTypes = []string{
	DOC_TYPE_TROUBLESHOOTING_MANUAL,
	DOC_TYPE_FAQ,
	DOC_TYPE_DEPLOYMENT_GUIDE,
	DOC_TYPE_API_DOC,
	DOC_TYPE_DB_DOC,
	DOC_TYPE_RELEASE_PROCESS,
	DOC_TYPE_INCIDENT_REVIEW,
	DOC_TYPE_ENVIRONMENT_GUIDE,
}

// GetDocTypeName 获取文档类型的中文名称
func GetDocTypeName(docType string) string {
	switch docType {
	case DOC_TYPE_TROUBLESHOOTING_MANUAL:
		return "故障排查手册"
	case DOC_TYPE_FAQ:
		return "常见问题处理方案"
	case DOC_TYPE_DEPLOYMENT_GUIDE:
		return "服务部署说明"
	case DOC_TYPE_API_DOC:
		return "接口说明"
	case DOC_TYPE_DB_DOC:
		return "数据库说明"
	case DOC_TYPE_RELEASE_PROCESS:
		return "发布流程规范"
	case DOC_TYPE_INCIDENT_REVIEW:
		return "历史故障复盘"
	case DOC_TYPE_ENVIRONMENT_GUIDE:
		return "环境使用说明"
	default:
		return docType
	}
}

// GetAllDocTypes 获取所有文档类型
func GetAllDocTypes() []string {
	return docTypes
}

// ValidateDocType 验证文档类型是否有效
func ValidateDocType(docType string) bool {
	for _, t := range docTypes {
		if t == docType {
			return true
		}
	}
	return false
}