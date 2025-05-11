package usecase

import (
	"chat_api/entity"
	"chat_api/pb"
	"chat_api/service/repository"
	"chat_api/utils/helper"
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type ChatUsecase interface {
	//write chat
	CreateChat(ctx context.Context, req *helper.CreateChatReq) (*entity.Chat, error)
	DeleteChat(ctx context.Context, chatId uint) (*entity.Chat, error)
	UpdateChat(ctx context.Context, chatId uint, message string) (*entity.Chat, error)

	//write group & member
	CreateGroup(ctx context.Context, name, desc string, userId uint) error
	UpdateGroup(ctx context.Context, name, desc string, adminId, groupId uint) error
	DeleteGroup(ctx context.Context, adminId, groupId uint) error
	AddMember(req []entity.GroupMember, adminId uint) error
	RemoveMember(req []uint, adminId uint) error
	ExitGroup(memberId uint) error
	UpdateRoleUser(memberId, adminId, groupId uint, role string) error

	//chat stream
	AddChatStream(stream pb.ChatService_ChatStreamingServer) string
	RemoveChatStream(clientID string)
	ChatBroadcast(chat *entity.Chat, action int)

	//status
	AddStatusStream(groupId, memberId uint, stream pb.ChatService_StatusStreamingServer) string
	RemoveStatusStream(clientID string)

	GetChatsByGroupID(groupID uint) ([]entity.Chat, error)
	GetMemberStatuses(groupId uint) ([]helper.MemberStatus, error)
	GetGroupMemberID(ctx context.Context, userID, groupID uint) (bool, uint, error)
	UpdateUnreadMessage(memberId uint) error
	GetListGroup(userId uint) ([]helper.GroupInfo, error)
}

type StreamStatus struct {
	Username string
	memberId uint
	GroupId  uint
	Stream   pb.ChatService_StatusStreamingServer
}

type chatUsecase struct {
	chatRepo            repository.ChatRepo
	mu                  sync.RWMutex
	streamChat          map[string]pb.ChatService_ChatStreamingServer
	streamStatusOnGroup map[string]*StreamStatus
}

func NewChatUsecase(r repository.ChatRepo) ChatUsecase {
	return &chatUsecase{
		chatRepo:            r,
		streamChat:          make(map[string]pb.ChatService_ChatStreamingServer),
		streamStatusOnGroup: make(map[string]*StreamStatus),
	}
}

// write chat
func (u *chatUsecase) CreateChat(ctx context.Context, req *helper.CreateChatReq) (*entity.Chat, error) {
	return u.chatRepo.CreateChat(ctx, req)
}

func (u *chatUsecase) DeleteChat(ctx context.Context, chatId uint) (*entity.Chat, error) {
	return u.chatRepo.DeleteChat(ctx, chatId)
}

func (u *chatUsecase) UpdateChat(ctx context.Context, chatId uint, message string) (*entity.Chat, error) {
	return u.chatRepo.UpdateChat(ctx, chatId, message)
}

// write group & member
func (u *chatUsecase) CreateGroup(ctx context.Context, name, desc string, userId uint) error {
	return u.chatRepo.CreateGroup(ctx, name, desc, userId)
}

func (u *chatUsecase) UpdateGroup(ctx context.Context, name, desc string, adminId, groupId uint) error {
	valid, err := u.chatRepo.IsMemberAdmin(adminId)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("you arent admin")
	}

	return u.chatRepo.UpdateGroup(ctx, name, desc, groupId)
}

func (u *chatUsecase) DeleteGroup(ctx context.Context, adminId uint, groupId uint) error {
	valid, err := u.chatRepo.IsMemberAdmin(adminId)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("you arent admin")
	}

	return u.chatRepo.DeleteGroup(ctx, groupId)

}

func (u *chatUsecase) AddMember(req []entity.GroupMember, adminId uint) error {
	valid, err := u.chatRepo.IsMemberAdmin(adminId)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("you arent admin")
	}

	return u.chatRepo.AddMember(req)
}

func (u *chatUsecase) RemoveMember(req []uint, adminId uint) error {
	valid, err := u.chatRepo.IsMemberAdmin(adminId)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("you arent admin")
	}

	return u.chatRepo.RemoveMember(req)
}

func (u *chatUsecase) ExitGroup(memberId uint) error {
	return u.chatRepo.ExitGroup(memberId)
}

func (u *chatUsecase) UpdateRoleUser(memberId, adminId, groupId uint, role string) error {
	valid, err := u.chatRepo.IsMemberAdmin(adminId)
	if err != nil {
		return err
	}
	if !valid {
		return fmt.Errorf("you arent admin")
	}

	return u.chatRepo.UpdateRoleUser(memberId, role)
}

// n
func (u *chatUsecase) GetGroupMemberID(ctx context.Context, userID, groupID uint) (bool, uint, error) {
	return u.chatRepo.GetGroupMemberID(ctx, userID, groupID)
}

func (u *chatUsecase) GetChatsByGroupID(chatGroupID uint) ([]entity.Chat, error) {
	return u.chatRepo.GetChatsByGroupID(chatGroupID)
}

// chat stream
func (u *chatUsecase) AddChatStream(stream pb.ChatService_ChatStreamingServer) string {
	clientID := fmt.Sprintf("%d", time.Now().UnixNano())

	u.mu.Lock()
	defer u.mu.Unlock()

	u.streamChat[clientID] = stream

	return clientID
}

func (u *chatUsecase) RemoveChatStream(clientID string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	delete(u.streamChat, clientID)
}

func (u *chatUsecase) ChatBroadcast(chat *entity.Chat, action int) {
	u.mu.Lock()
	defer u.mu.Unlock()
	result := helper.ConvertChatToPbResponse(chat)
	for clientId, stream := range u.streamChat {
		err := stream.Send(&pb.ChatStreamingResponse{
			Member:     uint64(*chat.GroupMemberID),
			Username:   chat.GroupMember.User.Username,
			Message:    chat.Message,
			Timestamp:  chat.CreatedAt.Format(time.RFC3339),
			Action:     pb.Action(action),
			ReadStatus: result,
		})

		if err != nil {
			log.Printf("error sending to client %s: %v, removing client", clientId, err)
			delete(u.streamChat, clientId)
		}
	}
}

// user status stream
func (u *chatUsecase) AddStatusStream(groupId, memberId uint, stream pb.ChatService_StatusStreamingServer) string {
	clientID := fmt.Sprintf("%d", time.Now().UnixNano())

	u.mu.Lock()
	defer u.mu.Unlock()

	u.streamStatusOnGroup[clientID] = &StreamStatus{
		GroupId:  groupId,
		memberId: memberId,
		Stream:   stream,
	}
	return clientID
}

func (u *chatUsecase) RemoveStatusStream(clientID string) {
	u.mu.Lock()
	defer u.mu.Unlock()

	delete(u.streamStatusOnGroup, clientID)
}

// n
func (u *chatUsecase) GetMemberGroup(groupId uint) ([]helper.MemberChat, error) {
	return u.chatRepo.GetMemberGroup(groupId)
}

func (u *chatUsecase) GetMemberStatuses(groupId uint) ([]helper.MemberStatus, error) {
	memberIDs, err := u.chatRepo.GetMemberGroup(groupId)
	if err != nil {
		return nil, err
	}
	u.mu.RLock()
	defer u.mu.RUnlock()

	onlineMemberMap := make(map[uint]bool)
	for _, client := range u.streamStatusOnGroup {
		if client.GroupId == groupId {
			onlineMemberMap[client.memberId] = true
		}
	}

	var result []helper.MemberStatus
	for _, member := range memberIDs {
		status := "offline"
		if onlineMemberMap[member.ID] {
			status = "online"
		}
		result = append(result, helper.MemberStatus{
			MemberID: member.ID,
			Username: member.Username,
			Status:   status,
		})
	}

	return result, nil
}

func (u *chatUsecase) UpdateUnreadMessage(memberId uint) error {
	return u.chatRepo.UpdateUnreadMessage(memberId)
}

func (u *chatUsecase) GetListGroup(userId uint) ([]helper.GroupInfo, error) {
	return u.chatRepo.GetListGroup(userId)
}
