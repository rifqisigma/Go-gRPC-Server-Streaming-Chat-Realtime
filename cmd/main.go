package main

import (
	"chat_api/cmd/database"
	"chat_api/pb"
	"chat_api/service/handler"
	"chat_api/service/repository"
	"chat_api/service/usecase"
	"chat_api/utils/interceptor"
	"log"
	"net"

	"google.golang.org/grpc"
)

func main() {

	database.ConnectDB()
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Interceptor khusus untuk chat
	streamInterceptor := interceptor.StreamInterceptor()
	unaryInterceptor := interceptor.UnaryInterceptor()

	opts := []grpc.ServerOption{
		grpc.StreamInterceptor(streamInterceptor),
		grpc.UnaryInterceptor(unaryInterceptor),
	}

	grpcServer := grpc.NewServer(opts...)

	authRepo := repository.NewAuthRepo(database.DB)
	authUsecase := usecase.NewAuthUsecase(authRepo)
	authHandler := handler.NewAuthHandler(authUsecase)
	pb.RegisterAuthServer(grpcServer, authHandler)

	chatRepo := repository.NewChatRepo(database.DB)
	chatUsecase := usecase.NewChatUsecase(chatRepo)
	chatHandler := handler.NewChatHandler(chatUsecase)
	pb.RegisterChatServiceServer(grpcServer, chatHandler)

	log.Println("gRPC server running at :50051")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
