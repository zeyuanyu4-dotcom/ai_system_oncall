package grpcserver

import (
	"context"
	"strconv"

	toolingv1 "ai_system_oncall/api/proto/tooling/v1"
	"ai_system_oncall/internal/dto"
	"ai_system_oncall/internal/service"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ToolingServer 实现 proto/tooling/v1/tooling.proto 中的 ToolingService
// 由 Python Agent 内部工具通过 gRPC 调用。
type ToolingServer struct {
	toolingv1.UnimplementedToolingServiceServer

	svc      *service.ServiceService
	issueSvc *service.IssueService
	kbSvc    *service.KnowledgeDocService
	logSvc   *service.SimulatedLogService
}

// NewToolingServer 构造 ToolingService
func NewToolingServer(
	svc *service.ServiceService,
	issueSvc *service.IssueService,
	kbSvc *service.KnowledgeDocService,
	logSvc *service.SimulatedLogService,
) *ToolingServer {
	return &ToolingServer{svc: svc, issueSvc: issueSvc, kbSvc: kbSvc, logSvc: logSvc}
}

// ========== 服务信息 ==========

func (s *ToolingServer) GetService(ctx context.Context, req *toolingv1.GetServiceRequest) (*toolingv1.GetServiceResponse, error) {
	si, err := s.svc.GetService(req.GetServiceId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "service not found: %v", err)
	}
	return &toolingv1.GetServiceResponse{Service: toProtoService(si)}, nil
}

func (s *ToolingServer) ListServicesByProject(ctx context.Context, req *toolingv1.ListServicesByProjectRequest) (*toolingv1.ListServicesByProjectResponse, error) {
	resp, err := s.svc.ListServices(&dto.ServiceListRequest{
		ProjectID: req.GetProjectId(),
		Page:      1,
		PageSize:  200,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list services failed: %v", err)
	}
	out := make([]*toolingv1.Service, 0, len(resp.List))
	for _, it := range resp.List {
		out = append(out, toProtoService(it))
	}
	return &toolingv1.ListServicesByProjectResponse{Services: out}, nil
}

// ========== 历史问题 ==========

func (s *ToolingServer) SearchHistoryIssues(ctx context.Context, req *toolingv1.SearchHistoryIssuesRequest) (*toolingv1.SearchHistoryIssuesResponse, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	resp, err := s.issueSvc.SearchHistoryIssues(&dto.HistoryIssueQueryRequest{
		Keyword:   req.GetKeyword(),
		ProjectID: req.GetProjectId(), // 0 = all
		IssueType: req.GetIssueType(),
		Page:      page,
		PageSize:  pageSize,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search issues failed: %v", err)
	}
	items := make([]*toolingv1.Issue, 0, len(resp.List))
	for _, it := range resp.List {
		items = append(items, toProtoHistoryIssue(it))
	}
	return &toolingv1.SearchHistoryIssuesResponse{Items: items, Total: int32(resp.Total)}, nil
}

// ========== 知识库 ==========

func (s *ToolingServer) SearchKnowledgeDocs(ctx context.Context, req *toolingv1.SearchKnowledgeDocsRequest) (*toolingv1.SearchKnowledgeDocsResponse, error) {
	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	resp, err := s.kbSvc.SearchDocuments(req.GetKeyword(), page, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search docs failed: %v", err)
	}
	items := make([]*toolingv1.KnowledgeDoc, 0, len(resp.List))
	for _, it := range resp.List {
		items = append(items, toProtoKnowledgeDoc(it))
	}
	return &toolingv1.SearchKnowledgeDocsResponse{Items: items, Total: int32(resp.Total)}, nil
}

func (s *ToolingServer) GetKnowledgeDoc(ctx context.Context, req *toolingv1.GetKnowledgeDocRequest) (*toolingv1.GetKnowledgeDocResponse, error) {
	d, err := s.kbSvc.GetDocument(req.GetDocId())
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "doc not found: %v", err)
	}
	return &toolingv1.GetKnowledgeDocResponse{Doc: toProtoKnowledgeDoc(d)}, nil
}

// ========== 日志 ==========

func (s *ToolingServer) GetLogsByTraceID(ctx context.Context, req *toolingv1.GetLogsByTraceIDRequest) (*toolingv1.GetLogsByTraceIDResponse, error) {
	logs, err := s.logSvc.GetLogsByTraceID(req.GetTraceId())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get logs failed: %v", err)
	}
	entries := make([]*toolingv1.LogEntry, 0, len(logs))
	for _, l := range logs {
		entries = append(entries, toProtoLogEntry(l))
	}
	return &toolingv1.GetLogsByTraceIDResponse{Entries: entries}, nil
}

func (s *ToolingServer) GetServiceLogs(ctx context.Context, req *toolingv1.GetServiceLogsRequest) (*toolingv1.GetServiceLogsResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 50
	}
	resp, err := s.logSvc.GetLogsByService(req.GetServiceId(), 1, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "get service logs failed: %v", err)
	}
	entries := make([]*toolingv1.LogEntry, 0, len(resp.List))
	for _, l := range resp.List {
		entries = append(entries, toProtoLogEntry(l))
	}
	return &toolingv1.GetServiceLogsResponse{Entries: entries}, nil
}

func (s *ToolingServer) SearchLogs(ctx context.Context, req *toolingv1.SearchLogsRequest) (*toolingv1.SearchLogsResponse, error) {
	limit := int(req.GetLimit())
	if limit <= 0 {
		limit = 50
	}
	resp, err := s.logSvc.ListLogs(&dto.SimulatedLogListRequest{
		ProjectID: req.GetProjectId(),
		ServiceID: req.GetServiceId(),
		LogLevel:  req.GetLogLevel(),
		Keyword:   req.GetKeyword(),
		Page:      1,
		PageSize:  limit,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "search logs failed: %v", err)
	}
	entries := make([]*toolingv1.LogEntry, 0, len(resp.List))
	for _, l := range resp.List {
		entries = append(entries, toProtoLogEntry(l))
	}
	return &toolingv1.SearchLogsResponse{Entries: entries}, nil
}

// ========== 问题更新 ==========

func (s *ToolingServer) UpdateIssue(ctx context.Context, req *toolingv1.UpdateIssueRequest) (*toolingv1.UpdateIssueResponse, error) {
	// Agent 内部工具替用户做更新，operator 设为 0 表示"系统/Agent 代发"
	// 业务权限校验放在 service 层
	upd := &dto.UpdateIssueRequest{}
	for k, v := range req.GetFields() {
		switch k {
		case "priority":
			upd.Priority = v
		case "root_cause":
			upd.RootCause = v
		case "description":
			upd.Description = v
		case "environment":
			upd.Environment = v
		case "impact_scope":
			upd.ImpactScope = v
		case "service_id":
			if id, err := strconv.ParseUint(v, 10, 64); err == nil {
				upd.ServiceID = &id
			}
		default:
			// 忽略未知/不可改字段（status 走单独的 status service）
		}
	}
	updated, err := s.issueSvc.UpdateIssue(req.GetIssueId(), 0, upd)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "update issue failed: %v", err)
	}
	return &toolingv1.UpdateIssueResponse{Issue: toProtoIssue(updated)}, nil
}

// ========== 转换函数 ==========

func toProtoService(s *dto.ServiceInfo) *toolingv1.Service {
	if s == nil {
		return nil
	}
	return &toolingv1.Service{
		Id:        s.ID,
		ProjectId: s.ProjectID,
		Name:      s.Name,
		Language:  s.Language,
		Owner:     s.OwnerName,
		Status:    strconv.Itoa(int(s.Status)),
	}
}

func toProtoIssue(i *dto.IssueInfo) *toolingv1.Issue {
	if i == nil {
		return nil
	}
	return &toolingv1.Issue{
		Id:        i.ID,
		IssueNo:   i.IssueNo,
		Title:     i.Title,
		Status:    i.Status,
		Priority:  i.Priority,
		IssueType: i.IssueType,
	}
}

func toProtoHistoryIssue(i *dto.HistoryIssueInfo) *toolingv1.Issue {
	if i == nil {
		return nil
	}
	return &toolingv1.Issue{
		Id:        i.ID,
		IssueNo:   i.IssueNo,
		Title:     i.Title,
		Status:    i.Status,
		Priority:  i.Priority,
		IssueType: i.IssueType,
	}
}

func toProtoKnowledgeDoc(d *dto.KnowledgeDocInfo) *toolingv1.KnowledgeDoc {
	if d == nil {
		return nil
	}
	return &toolingv1.KnowledgeDoc{
		Id:      d.ID,
		Title:   d.Title,
		DocType: d.DocType,
		Content: d.Content,
	}
}

func toProtoLogEntry(l *dto.SimulatedLogInfo) *toolingv1.LogEntry {
	if l == nil {
		return nil
	}
	return &toolingv1.LogEntry{
		Timestamp: l.OccurredAt.Format("2006-01-02T15:04:05Z"),
		Level:     l.LogLevel,
		Message:   l.LogContent,
		Service:   l.ServiceName,
		TraceId:   l.TraceID,
	}
}
