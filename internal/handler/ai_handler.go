package handler

import (
	"encoding/json"
	"strconv"

	"ai_system_oncall/internal/client"
	"ai_system_oncall/internal/config"
	"ai_system_oncall/internal/response"
	"ai_system_oncall/internal/service"

	"github.com/gin-gonic/gin"
)

// AIHandler AI 分析处理器
type AIHandler struct {
	issueService *service.IssueService
	aiClient     *client.AIClient
}

// NewAIHandler 创建 AI 分析处理器
func NewAIHandler(issueService *service.IssueService, aiClient *client.AIClient) *AIHandler {
	return &AIHandler{
		issueService: issueService,
		aiClient:     aiClient,
	}
}

// AnalyzeIssue 分析问题
func (h *AIHandler) AnalyzeIssue(c *gin.Context) {
	// 检查 AI 服务是否启用
	if config.GetConfig() != nil && !config.GetConfig().AI.Enabled {
		response.Fail(c, 10001, "AI 服务未启用")
		return
	}

	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的问题ID")
		return
	}

	// 获取问题详情
	issueInfo, err := h.issueService.GetIssue(id)
	if err != nil {
		response.Fail(c, 10004, err.Error())
		return
	}

	// 构建分析请求
	req := &client.IssueAnalysisRequest{
		IssueID:      int(issueInfo.ID),
		IssueNo:      issueInfo.IssueNo,
		Title:        issueInfo.Title,
		Description:  issueInfo.Description,
		ErrorMessage: issueInfo.ErrorMessage,
		LogExcerpt:   issueInfo.LogExcerpt,
		Environment:  issueInfo.Environment,
		ProjectName:  issueInfo.ProjectName,
		ServiceName:  issueInfo.ServiceName,
		ImpactScope:  issueInfo.ImpactScope,
	}

	// 调用 AI 服务
	result, err := h.aiClient.AnalyzeIssue(req)
	if err != nil {
		response.Fail(c, 10001, "AI 分析失败: "+err.Error())
		return
	}

	// 将分析结果转换为 JSON 并保存到问题单的 ai_analysis 字段
	analysisJSON, err := json.Marshal(result)
	if err != nil {
		response.Fail(c, 10001, "保存分析结果失败")
		return
	}

	// 直接保存 JSON 到 ai_analysis 字段
	err = h.issueService.UpdateAIAnalysis(id, result.Summary, string(analysisJSON))
	if err != nil {
		// 保存失败不影响返回结果，只是记录日志
		// log.Printf("Failed to save AI analysis: %v", err)
	}

	response.Success(c, result)
}

// AIAnalysisStatus AI 分析状态
func (h *AIHandler) AIAnalysisStatus(c *gin.Context) {
	enabled := false
	healthy := false

	if config.GetConfig() != nil {
		enabled = config.GetConfig().AI.Enabled
	}

	if enabled && h.aiClient != nil {
		healthy = h.aiClient.HealthCheck() == nil
	}

	response.Success(c, gin.H{
		"enabled": enabled,
		"healthy": healthy,
	})
}
