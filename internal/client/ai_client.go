package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"ai_system_oncall/internal/config"
)

// AIClient AI 服务客户端
type AIClient struct {
	baseURL string
	timeout time.Duration
	client  *http.Client
}

// NewAIClient 创建 AI 客户端
func NewAIClient(cfg *config.AIConfig) *AIClient {
	if cfg == nil {
		cfg = &config.AIConfig{
			Enabled: false,
			BaseURL: "http://127.0.0.1:8001",
			Timeout: 60,
		}
	}

	return &AIClient{
		baseURL: cfg.BaseURL,
		timeout: time.Duration(cfg.Timeout) * time.Second,
		client: &http.Client{
			Timeout: time.Duration(cfg.Timeout+10) * time.Second,
		},
	}
}

// IssueAnalysisRequest 问题分析请求（基础分析）
type IssueAnalysisRequest struct {
	IssueID      int    `json:"issue_id"`
	IssueNo      string `json:"issue_no"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ErrorMessage string `json:"error_message"`
	LogExcerpt   string `json:"log_excerpt"`
	Environment  string `json:"environment"`
	ProjectName  string `json:"project_name"`
	ServiceName  string `json:"service_name"`
	ImpactScope  string `json:"impact_scope"`
}

// AgentAnalysisRequest Agent分析请求（深度分析）
type AgentAnalysisRequest struct {
	TaskID       int    `json:"task_id"`
	IssueID      int    `json:"issue_id"`
	IssueNo      string `json:"issue_no"`
	Title        string `json:"title"`
	Description  string `json:"description"`
	ErrorMessage string `json:"error_message"`
	LogExcerpt   string `json:"log_excerpt"`
	Environment  string `json:"environment"`
	ProjectID    uint64 `json:"project_id"`
	ProjectName  string `json:"project_name"`
	ServiceName  string `json:"service_name"`
	ImpactScope  string `json:"impact_scope"`
}

// KeyInfo 关键信息
type KeyInfo struct {
	ErrorCode        string                 `json:"error_code"`
	ErrorType        string                 `json:"error_type"`
	AffectedEndpoint string                 `json:"affected_endpoint"`
	ErrorTime        string                 `json:"error_time"`
	TraceID          string                 `json:"trace_id"`
	AdditionalInfo   map[string]interface{} `json:"additional_info"`
}

// AnalysisResult 基础分析结果
type AnalysisResult struct {
	Summary          string   `json:"summary"`
	IssueType        string   `json:"issue_type"`
	Environment      string   `json:"environment"`
	RelatedServices  []string `json:"related_services"`
	Priority         string   `json:"priority"`
	Confidence       float64  `json:"confidence"`
	KeyInfo          *KeyInfo `json:"key_info"`
	MissingInfo      []string `json:"missing_info"`
	Suggestions      []string `json:"suggestions"`
}

// ToolCallRecord 工具调用记录
type ToolCallRecord struct {
	Step       int    `json:"step"`
	ToolName   string `json:"tool_name"`
	Input      string `json:"input"`
	Output     string `json:"output"`
	Thought    string `json:"thought"`
	ExecutedAt string `json:"executed_at"`
	DurationMs int64  `json:"duration_ms"`
}

// EvidenceItem 证据项
type EvidenceItem struct {
	Source    string `json:"source"`
	Content   string `json:"content"`
	Relevance string `json:"relevance"`
}

// AgentResult Agent分析结果
type AgentResult struct {
	Summary         string           `json:"summary"`
	IssueType       string           `json:"issue_type"`
	RelatedServices []string         `json:"related_services"`
	SuspectedCause  string           `json:"suspected_cause"`
	Evidence        []EvidenceItem   `json:"evidence"`
	Suggestions     []string         `json:"suggestions"`
	MissingInfo     []string         `json:"missing_info"`
	NextSteps       []string         `json:"next_steps"`
	Confidence      float64          `json:"confidence"`
	ToolCalls       []ToolCallRecord `json:"tool_calls"`
}

// AnalysisResponse 分析响应
type AnalysisResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    *AnalysisResult `json:"data"`
}

// AgentAnalysisResponse Agent分析响应
type AgentAnalysisResponse struct {
	Code    int          `json:"code"`
	Message string       `json:"message"`
	Data    *AgentResult `json:"data"`
}

// AnalyzeIssue 基础分析
func (c *AIClient) AnalyzeIssue(req *IssueAnalysisRequest) (*AnalysisResult, error) {
	url := fmt.Sprintf("%s/api/analyze", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI service error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result AnalysisResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("AI analysis failed: %s", result.Message)
	}

	return result.Data, nil
}

// AgentAnalyze Agent深度分析
func (c *AIClient) AgentAnalyze(req *AgentAnalysisRequest, userToken string) (*AgentResult, error) {
	url := fmt.Sprintf("%s/api/agent/analyze", c.baseURL)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal request failed: %w", err)
	}

	// Agent分析可能需要更长时间
	client := &http.Client{Timeout: 180 * time.Second}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	// 透传用户 JWT，让 Python Agent 内部工具能以此身份访问受保护接口
	if userToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+userToken)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI agent service error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	var result AgentAnalysisResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response failed: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("Agent analysis failed: %s", result.Message)
	}

	return result.Data, nil
}

// HealthCheck 健康检查
func (c *AIClient) HealthCheck() error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("AI service unhealthy: status=%d", resp.StatusCode)
	}

	return nil
}

// GenerateText 生成文本（用于报告分析）
func (c *AIClient) GenerateText(prompt string) (string, error) {
	url := fmt.Sprintf("%s/api/generate", c.baseURL)

	reqBody := map[string]string{"prompt": prompt}
	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request failed: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request failed: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("AI service error: status=%d, body=%s", resp.StatusCode, string(respBody))
	}

	// 尝试解析为通用响应
	var result struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    string `json:"data"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("unmarshal response failed: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("AI generate failed: %s", result.Message)
	}

	return result.Data, nil
}
