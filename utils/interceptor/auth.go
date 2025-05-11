package interceptor

import (
	"chat_api/utils/helper"
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type key int

var UserContextKey key = 0

type wrappedStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedStream) Context() context.Context {
	return w.ctx
}

var (
	allowedUnaryMethods = map[string]bool{
		"/pb.ChatService/CreateChat":     true,
		"/pb.ChatService/DeleteChat":     true,
		"/pb.ChatService/UpdateChat":     true,
		"/pb.ChatService/CreateGroup":    true,
		"/pb.ChatService/DeleteGroup":    true,
		"/pb.ChatService/UpdateGroup":    true,
		"/pb.ChatService/AddMember":      true,
		"/pb.ChatService/RemoveMember":   true,
		"/pb.ChatService/ExitGroup":      true,
		"/pb.ChatService/UpdateRoleUser": true,

		"/pb.ChatService/GetListGroup": true,
	}

	allowedStreamMethods = map[string]bool{
		"/pb.ChatService/ChatStreaming":   true,
		"/pb.ChatService/StatusStreaming": true,
	}
)

func StreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {

		if !allowedStreamMethods[info.FullMethod] {
			return handler(srv, ss)
		}

		ctx := ss.Context()

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Unauthenticated, "metadata tidak ditemukan")
		}

		authHeaders := md["authorization"]
		if len(authHeaders) == 0 {
			return status.Error(codes.Unauthenticated, "authorization header tidak ditemukan")
		}

		tokenString := strings.TrimPrefix(authHeaders[0], "Bearer ")
		if tokenString == authHeaders[0] {
			return status.Error(codes.Unauthenticated, "format token tidak valid (harus Bearer ...)")
		}

		claims, err := helper.ParseJWT(tokenString)
		if err != nil {
			return status.Error(codes.Unauthenticated, "invalid token ")
		}

		newCtx := context.WithValue(ctx, UserContextKey, claims)

		wrapped := &wrappedStream{ServerStream: ss, ctx: newCtx}

		return handler(srv, wrapped)
	}
}

func UnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

		if !allowedUnaryMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Errorf(codes.Unauthenticated, "missing metadata")
		}
		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, status.Errorf(codes.Unauthenticated, "missing authorization token")
		}

		tokenString := strings.TrimPrefix(authHeader[0], "Bearer ")
		if tokenString == "" {
			return nil, status.Errorf(codes.Unauthenticated, "missing token")
		}

		claims, err := helper.ParseJWT(tokenString)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "invalid token: %v", err)
		}

		newCtx := context.WithValue(ctx, UserContextKey, claims)

		return handler(newCtx, req)
	}
}
