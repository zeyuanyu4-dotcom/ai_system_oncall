// Package grpcclient 集中管理 Go 对 Python Agent 的 gRPC 客户端。
package grpcclient

import (
	"context"
	"fmt"
	"time"

	agentv1 "ai_system_oncall/api/proto/agent/v1"
	capabilityv1 "ai_system_oncall/api/proto/capability/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// AgentClient 调用 Python Agent 的 AgentService
type AgentClient struct {
	conn     *grpc.ClientConn
	agent    agentv1.AgentServiceClient
	timeout  time.Duration
	grpcAddr string
}

// NewAgentClient 创建 AgentService 客户端
func NewAgentClient(grpcAddr string, timeoutSec int) (*AgentClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to agent gRPC %s failed: %w", grpcAddr, err)
	}
	return &AgentClient{
		conn:     conn,
		agent:    agentv1.NewAgentServiceClient(conn),
		timeout:  time.Duration(timeoutSec) * time.Second,
		grpcAddr: grpcAddr,
	}, nil
}

// Close 释放连接
func (c *AgentClient) Close() error {
	return c.conn.Close()
}

// RunAgent 发起 ServerStreaming：Agent 分析 + 进度事件
// userToken 会注入 gRPC metadata（Authorization: Bearer xxx）
func (c *AgentClient) RunAgent(ctx context.Context, req *agentv1.RunAgentRequest, userToken string) (agentv1.AgentService_RunAgentClient, error) {
	md := metadata.New(nil)
	if userToken != "" {
		md = metadata.New(map[string]string{"authorization": "Bearer " + userToken})
	}
	ctx = metadata.NewOutgoingContext(ctx, md)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.agent.RunAgent(ctx, req)
}

// CapabilityClient 调用 Python Agent 的 CapabilityService（生成文本、向量化）
type CapabilityClient struct {
	conn      *grpc.ClientConn
	cap       capabilityv1.CapabilityServiceClient
	grpcAddr  string
	timeout   time.Duration
}

// NewCapabilityClient 创建 CapabilityService 客户端
func NewCapabilityClient(grpcAddr string, timeoutSec int) (*CapabilityClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("connect to capability gRPC %s failed: %w", grpcAddr, err)
	}
	return &CapabilityClient{
		conn:    conn,
		cap:     capabilityv1.NewCapabilityServiceClient(conn),
		timeout: time.Duration(timeoutSec) * time.Second,
		grpcAddr: grpcAddr,
	}, nil
}

// Close 释放连接
func (c *CapabilityClient) Close() error {
	return c.conn.Close()
}

// GenerateText 调用 CapabilityService.GenerateText
func (c *CapabilityClient) GenerateText(ctx context.Context, prompt, model string, maxTokens int) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	resp, err := c.cap.GenerateText(ctx, &capabilityv1.GenerateTextRequest{
		Prompt:    prompt,
		Model:     model,
		MaxTokens: int32(maxTokens),
	})
	if err != nil {
		return "", fmt.Errorf("capability GenerateText failed: %w", err)
	}
	return resp.GetText(), nil
}

// VectorizeText 调用 CapabilityService.VectorizeText
func (c *CapabilityClient) VectorizeText(ctx context.Context, texts []string, collection string) (*capabilityv1.VectorizeTextResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	return c.cap.VectorizeText(ctx, &capabilityv1.VectorizeTextRequest{
		Texts:      texts,
		Collection: collection,
	})
}
