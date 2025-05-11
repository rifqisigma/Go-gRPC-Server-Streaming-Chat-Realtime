package handler

import (
	"chat_api/pb"
	"chat_api/service/usecase"
	"chat_api/utils/helper"
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthServer struct {
	pb.UnimplementedAuthServer
	authUsecase usecase.AuthUsecase
}

func NewAuthHandler(authUsecase usecase.AuthUsecase) *AuthServer {
	return &AuthServer{authUsecase: authUsecase}
}

func (s *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Validation error: %v", err)
	}
	user := helper.ParsingPbToLogin(req)

	token, err := s.authUsecase.Login(ctx, user)
	if err != nil {
		return nil, err
	}

	return &pb.LoginResponse{
		Token: token,
	}, nil

}

func (s *AuthServer) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Validation error: %v", err)
	}

	fmt.Println(req)
	user := helper.ParsingPbToRegister(req)

	fmt.Println(user)
	if err := s.authUsecase.Register(ctx, user); err != nil {
		return &pb.RegisterResponse{
			Message: "failed to register",
		}, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterResponse{
		Message: "berhasil register",
	}, nil
}
