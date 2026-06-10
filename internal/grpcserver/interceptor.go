package grpcserver

import (
	"context"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// metadataKeyUserToken 是 gRPC metadata 中 JWT 所在的 key
const metadataKeyUserToken = "authorization"

// ctxKeyUserToken 是 gRPC context 中 JWT 所在的自定义 key
type ctxKey string

const userTokenCtxKey ctxKey = "user_token"

// JWTAuthInterceptor 拦截所有 gRPC 调用，提取 JWT 并放入 context。
// 注意：与 HTTP 中间件不同，gRPC 不在这里做签名校验（依赖方各自校验）。
// ToolingService 调用方是 Python Agent 内部的工具客户端，可信度高；
// 真实业务权限校验由各 RPC handler 内部借助 service 层做。
func JWTAuthInterceptor(log *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		token := extractToken(ctx)
		if token != "" {
			ctx = context.WithValue(ctx, userTokenCtxKey, token)
		}
		// 记录调用方身份（脱敏）
		if token != "" {
			log.Debug("gRPC unary call with token",
				zap.String("method", info.FullMethod),
				zap.String("token_prefix", token[:min(10, len(token))]+"..."))
		} else {
			log.Warn("gRPC unary call WITHOUT token",
				zap.String("method", info.FullMethod))
		}
		return handler(ctx, req)
	}
}

// JWTAuthStreamInterceptor 流式版本
func JWTAuthStreamInterceptor(log *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		token := extractToken(ss.Context())
		if token != "" {
			log.Debug("gRPC stream call with token",
				zap.String("method", info.FullMethod),
				zap.String("token_prefix", token[:min(10, len(token))]+"..."))
		} else {
			log.Warn("gRPC stream call WITHOUT token",
				zap.String("method", info.FullMethod))
		}
		return handler(srv, &tokenStream{ServerStream: ss, token: token})
	}
}

// UserTokenFromContext 从 gRPC context 拿用户 token
func UserTokenFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(userTokenCtxKey).(string); ok {
		return v
	}
	return ""
}

func extractToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get(metadataKeyUserToken)
	if len(vals) == 0 {
		return ""
	}
	v := vals[0]
	if strings.HasPrefix(v, "Bearer ") {
		return strings.TrimPrefix(v, "Bearer ")
	}
	return v
}

type tokenStream struct {
	grpc.ServerStream
	token string
}

// Context 注入 token
func (t *tokenStream) Context() context.Context {
	if t.token == "" {
		return t.ServerStream.Context()
	}
	return context.WithValue(t.ServerStream.Context(), userTokenCtxKey, t.token)
}

// min 工具函数（Go 1.21+ 内置，但为兼容老版本显式写）
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// RequireToken 内部 RPC 调用前断言 token 存在。
// 写在各 handler 入口，**不强依赖**——业务层可以按需跳过。
func RequireToken(ctx context.Context) (string, error) {
	tok := UserTokenFromContext(ctx)
	if tok == "" {
		return "", status.Error(codes.Unauthenticated, "missing user token in gRPC metadata")
	}
	return tok, nil
}
