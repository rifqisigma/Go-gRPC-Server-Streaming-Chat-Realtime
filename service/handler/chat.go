package handler

import (
	"chat_api/pb"
	"chat_api/service/usecase"
	"chat_api/utils/helper"
	"chat_api/utils/interceptor"
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ChatServer struct {
	pb.UnimplementedChatServiceServer
	chatUsecase usecase.ChatUsecase
}

func NewChatHandler(chatUsecase usecase.ChatUsecase) *ChatServer {
	return &ChatServer{chatUsecase: chatUsecase}
}

// write chat
func (s *ChatServer) CreateChat(ctx context.Context, req *pb.CreateChatRequest) (*pb.CreateChatResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Validation error: %v", err)
	}

	claimsRaw := ctx.Value(interceptor.UserContextKey)
	claims, ok := claimsRaw.(*helper.JWTClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid token claims")
	}
	ismember, memberId, err := s.chatUsecase.GetGroupMemberID(ctx, claims.UserID, uint(req.GroupId))
	if err != nil {
		return &pb.CreateChatResponse{Message: "failed to send message"}, status.Error(codes.Internal, "server error")
	}
	if !ismember {
		return &pb.CreateChatResponse{Message: "unautorized"}, status.Error(codes.Unauthenticated, "you arent member")
	}

	chat := helper.ParsingPbToCreateChat(req, memberId)

	cb, err := s.chatUsecase.CreateChat(ctx, chat)
	if err != nil {
		return &pb.CreateChatResponse{Message: "failed to send message"}, status.Error(codes.Internal, "server error")
	}

	s.chatUsecase.ChatBroadcast(cb, 0)

	return &pb.CreateChatResponse{Message: "success send message"}, nil
}

func (s *ChatServer) DeleteChat(ctx context.Context, req *pb.DeleteChatRequest) (*pb.StatusResponse, error) {
	if err := req.Validate(); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Canceled, "invalid message")
	}

	claimsRaw := ctx.Value(interceptor.UserContextKey)
	claims, ok := claimsRaw.(*helper.JWTClaims)
	if !ok {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Unauthenticated, "invalid jwt")
	}

	ismember, _, err := s.chatUsecase.GetGroupMemberID(ctx, claims.UserID, uint(req.GroupId))
	if err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}
	if !ismember {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Unauthenticated, "you arent member")
	}

	chat, err := s.chatUsecase.DeleteChat(ctx, uint(req.ChatId))
	if err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}

	s.chatUsecase.ChatBroadcast(chat, 2)
	return &pb.StatusResponse{
		Status: true,
	}, nil
}

func (s *ChatServer) UpdateChat(ctx context.Context, req *pb.UpdateChatRequest) (*pb.StatusResponse, error) {
	if err := req.Validate(); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Canceled, "invalid message")
	}

	claimsRaw := ctx.Value(interceptor.UserContextKey)
	claims, ok := claimsRaw.(*helper.JWTClaims)
	if !ok {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Unauthenticated, "invalid jwt")
	}

	ismember, _, err := s.chatUsecase.GetGroupMemberID(ctx, claims.UserID, uint(req.GroupId))
	if err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}
	if !ismember {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Unauthenticated, "you arent member")
	}

	chat, err := s.chatUsecase.UpdateChat(ctx, uint(req.ChatId), req.Message)
	if err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}

	s.chatUsecase.ChatBroadcast(chat, 1)

	return &pb.StatusResponse{
		Status: true,
	}, nil
}

// write group & member
func (s *ChatServer) CreateGroup(ctx context.Context, req *pb.CreateGroupRequest) (*pb.StatusResponse, error) {
	if err := req.Validate(); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Canceled, "invalid message")
	}

	claimsRaw := ctx.Value(interceptor.UserContextKey)
	claims, ok := claimsRaw.(*helper.JWTClaims)
	if !ok {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Unauthenticated, "invalid jwt")
	}

	if err := s.chatUsecase.CreateGroup(ctx, req.Name, req.Desc, claims.UserID); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}

	return &pb.StatusResponse{
		Status: true,
	}, nil
}

func (s *ChatServer) DeleteGroup(ctx context.Context, req *pb.DeleteGroupRequest) (*pb.StatusResponse, error) {
	if err := req.Validate(); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Canceled, "invalid message")
	}

	claimsRaw := ctx.Value(interceptor.UserContextKey)
	claims, ok := claimsRaw.(*helper.JWTClaims)
	if !ok {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Unauthenticated, "invalid jwt")
	}

	ismember, memberId, err := s.chatUsecase.GetGroupMemberID(ctx, claims.UserID, uint(req.GroupId))
	if err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}
	if !ismember {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Unauthenticated, "you arent member")
	}

	if err := s.chatUsecase.DeleteGroup(ctx, memberId, uint(req.GroupId)); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}

	return &pb.StatusResponse{
		Status: true,
	}, nil
}

func (s *ChatServer) UpdateGroup(ctx context.Context, req *pb.UpdateGroupRequest) (*pb.StatusResponse, error) {
	if err := req.Validate(); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Canceled, "invalid message")
	}

	claimsRaw := ctx.Value(interceptor.UserContextKey)
	claims, ok := claimsRaw.(*helper.JWTClaims)
	if !ok {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, "invalid jwt")
	}

	ismember, memberId, err := s.chatUsecase.GetGroupMemberID(ctx, claims.UserID, uint(req.GroupId))
	if err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}
	if !ismember {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Unauthenticated, "you arent member")
	}

	if err := s.chatUsecase.UpdateGroup(ctx, req.Name, req.Desc, memberId, uint(req.GroupId)); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}

	return &pb.StatusResponse{
		Status: true,
	}, nil
}

func (s *ChatServer) AddMember(ctx context.Context, req *pb.AddMemberRequest) (*pb.StatusResponse, error) {
	if err := req.Validate(); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Canceled, "invalid message")
	}

	claimsRaw := ctx.Value(interceptor.UserContextKey)
	claims, ok := claimsRaw.(*helper.JWTClaims)
	if !ok {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, "invalid jwt")
	}

	ismember, memberId, err := s.chatUsecase.GetGroupMemberID(ctx, claims.UserID, uint(req.GroupId))
	if err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}
	if !ismember {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Unauthenticated, "you arent member")
	}

	fmt.Println(req.ListUserId)
	newMembers := helper.ParsingPBtoAddMember(req.ListUserId, uint(req.GroupId))
	if err := s.chatUsecase.AddMember(newMembers, memberId); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}

	return &pb.StatusResponse{
		Status: true,
	}, nil
}

func (s *ChatServer) RemoveMember(ctx context.Context, req *pb.RemoveMemberRequest) (*pb.StatusResponse, error) {
	if err := req.Validate(); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Canceled, "invalid message")
	}

	claimsRaw := ctx.Value(interceptor.UserContextKey)
	claims, ok := claimsRaw.(*helper.JWTClaims)
	if !ok {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, "invalid jwt")
	}

	ismember, memberId, err := s.chatUsecase.GetGroupMemberID(ctx, claims.UserID, uint(req.GroupId))
	if err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}
	if !ismember {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Unauthenticated, "you arent member")
	}

	deleteMembers := helper.ParsingPBtoRemoveMember(req.ListMemberId)
	if err := s.chatUsecase.RemoveMember(deleteMembers, memberId); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}
	return &pb.StatusResponse{
		Status: true,
	}, nil

}

func (s *ChatServer) ExitGroup(ctx context.Context, req *pb.ExitGroupRequest) (*pb.StatusResponse, error) {
	claimsRaw := ctx.Value(interceptor.UserContextKey)
	claims, ok := claimsRaw.(*helper.JWTClaims)
	if !ok {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, "invalid jwt")
	}

	ismember, memberId, err := s.chatUsecase.GetGroupMemberID(ctx, claims.UserID, uint(req.GroupId))
	if err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}
	if !ismember {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Unauthenticated, "you arent member")
	}

	if err := s.chatUsecase.ExitGroup(memberId); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}

	return &pb.StatusResponse{
		Status: true,
	}, nil
}

func (s *ChatServer) UpdateRoleUser(ctx context.Context, req *pb.UpdateRoleUserRequest) (*pb.StatusResponse, error) {
	if err := req.Validate(); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Canceled, "invalid message")
	}

	claimsRaw := ctx.Value(interceptor.UserContextKey)
	claims, ok := claimsRaw.(*helper.JWTClaims)
	if !ok {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, "invalid jwt")
	}

	ismember, memberId, err := s.chatUsecase.GetGroupMemberID(ctx, claims.UserID, uint(req.GroupId))
	if err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}
	if !ismember {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Unauthenticated, "you arent member")
	}

	if err := s.chatUsecase.UpdateRoleUser(memberId, memberId, uint(req.GroupId), req.Role); err != nil {
		return &pb.StatusResponse{
			Status: false,
		}, status.Error(codes.Internal, err.Error())
	}

	return &pb.StatusResponse{
		Status: true,
	}, nil
}

// chat stream
func (s *ChatServer) ChatStreaming(req *pb.ChatStreamingRequest, stream pb.ChatService_ChatStreamingServer) error {
	if err := req.Validate(); err != nil {
		return err
	}
	ctx := stream.Context()

	claims, ok := ctx.Value(interceptor.UserContextKey).(*helper.JWTClaims)
	if !ok {
		return status.Error(codes.Unauthenticated, "invalid jwt")
	}
	ismember, memberId, err := s.chatUsecase.GetGroupMemberID(ctx, claims.UserID, uint(req.GroupId))
	if err != nil {
		return status.Error(codes.Internal, "server error")
	}
	if !ismember {
		return status.Error(codes.Unauthenticated, "you arent member")
	}

	if err := s.chatUsecase.UpdateUnreadMessage(memberId); err != nil {
		return status.Error(codes.Internal, err.Error())
	}

	clientId := s.chatUsecase.AddChatStream(stream)
	defer s.chatUsecase.RemoveChatStream(clientId)

	chatHistory, err := s.chatUsecase.GetChatsByGroupID(uint(req.GroupId))
	if err != nil {
		return err
	}

	for _, chat := range chatHistory {

		select {
		case <-ctx.Done():
			return nil
		default:
			readStatus := helper.ConvertChatToPbResponse(&chat)
			err := stream.Send(&pb.ChatStreamingResponse{
				ReadStatus: readStatus,
				Member:     uint64(*chat.GroupMemberID),
				Username:   chat.GroupMember.User.Username,
				Message:    chat.Message,
				Timestamp:  chat.CreatedAt.Format(time.RFC3339),
			})
			if err != nil {
				return status.Errorf(codes.Internal, "error sending chat: %v", err)
			}
		}
	}

	<-ctx.Done()
	return nil

}

// user status stream
func (s *ChatServer) StatusStreaming(req *pb.StatusStreamingRequest, stream pb.ChatService_StatusStreamingServer) error {
	if err := req.Validate(); err != nil {
		return err
	}

	ctx := stream.Context()

	claims, ok := ctx.Value(interceptor.UserContextKey).(*helper.JWTClaims)
	if !ok {
		return status.Error(codes.Unauthenticated, "invalid jwt")
	}

	isMember, memberId, err := s.chatUsecase.GetGroupMemberID(ctx, claims.UserID, uint(req.GroupId))
	if err != nil || !isMember {
		return status.Error(codes.Unauthenticated, "you arent member")
	}
	clientID := s.chatUsecase.AddStatusStream(uint(req.GroupId), memberId, stream)
	defer s.chatUsecase.RemoveStatusStream(clientID)

	statuses, err := s.chatUsecase.GetMemberStatuses(uint(req.GroupId))
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return nil
	default:
		for _, status_user := range statuses {
			err := stream.Send(&pb.StatusStreamingResponse{
				Member:   uint64(status_user.MemberID),
				Username: status_user.Username,
				Status:   status_user.Status,
			})

			if err != nil {
				return status.Errorf(codes.Internal, "send error: %v", err)
			}
		}
	}

	<-ctx.Done()
	return nil
}

func (s *ChatServer) GetListGroup(ctx context.Context, req *emptypb.Empty) (*pb.GetListGroupResponse, error) {
	claimsRaw := ctx.Value(interceptor.UserContextKey)
	claims, ok := claimsRaw.(*helper.JWTClaims)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "invalid jwt")
	}

	groups, err := s.chatUsecase.GetListGroup(claims.UserID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	response := helper.ParsingDtoGroupToPB(groups)
	return &pb.GetListGroupResponse{Group: response}, nil
}
