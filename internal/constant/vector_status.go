package constant

const (
	// VectorStatus 向量化状态
	VECTOR_STATUS_PENDING    = "pending"    // 待处理
	VECTOR_STATUS_PROCESSING = "processing" // 处理中
	VECTOR_STATUS_COMPLETED  = "completed"  // 已完成
	VECTOR_STATUS_FAILED     = "failed"    // 失败
)

var vectorStatuses = []string{
	VECTOR_STATUS_PENDING,
	VECTOR_STATUS_PROCESSING,
	VECTOR_STATUS_COMPLETED,
	VECTOR_STATUS_FAILED,
}

// GetVectorStatusName 获取向量化状态的中文名称
func GetVectorStatusName(status string) string {
	switch status {
	case VECTOR_STATUS_PENDING:
		return "待处理"
	case VECTOR_STATUS_PROCESSING:
		return "处理中"
	case VECTOR_STATUS_COMPLETED:
		return "已完成"
	case VECTOR_STATUS_FAILED:
		return "失败"
	default:
		return status
	}
}

// GetAllVectorStatuses 获取所有向量化状态
func GetAllVectorStatuses() []string {
	return vectorStatuses
}

// ValidateVectorStatus 验证向量化状态是否有效
func ValidateVectorStatus(status string) bool {
	for _, s := range vectorStatuses {
		if s == status {
			return true
		}
	}
	return false
}