package helper

import (
	"chat_api/entity"
	"chat_api/pb"
)

// parsing data
func ParsingPbToRegister(req *pb.RegisterRequest) *entity.User {
	return &entity.User{
		Email:    req.Email,
		Password: req.Password,
		Username: req.Username,
	}
}

// parsing data
func ParsingPbToLogin(req *pb.LoginRequest) *entity.User {
	return &entity.User{
		Email:    req.Email,
		Password: req.Password,
	}
}

// chat req
// for chat req
type CreateChatReq struct {
	MemberId      uint
	Message       string
	GroupId       uint
	AnyStatusUser []StatusUser
}

// struct for parsing pb to create chat
type StatusUser struct {
	MemberId uint
	Status   string
}

func ParsingPbToCreateChat(req *pb.CreateChatRequest, memberId uint) *CreateChatReq {
	var anyUserStatus []StatusUser

	for _, s := range req.Status {
		anyUserStatus = append(anyUserStatus, StatusUser{
			MemberId: uint(s.MemberId),
			Status:   s.Status,
		})
	}
	return &CreateChatReq{
		MemberId:      memberId,
		Message:       req.Message,
		GroupId:       uint(req.GroupId),
		AnyStatusUser: anyUserStatus,
	}
}

func ParsingPBtoAddMember(req []*pb.ListUserId, groupId uint) []entity.GroupMember {
	var newMembers []entity.GroupMember
	for _, newMember := range req {
		newMembers = append(newMembers, entity.GroupMember{
			GroupID: groupId,
			UserID:  uint(newMember.UserId),
			Role:    "member",
		})
	}

	return newMembers
}

func ParsingPBtoRemoveMember(req []*pb.ListUserId) []uint {
	var deleteMembers []uint
	for _, deleteMember := range req {
		deleteMembers = append(deleteMembers, uint(deleteMember.UserId))
	}

	return deleteMembers
}

type MemberStatus struct {
	MemberID uint
	Username string
	Status   string
}

// convert entity to response
func ConvertChatToPbResponse(chat *entity.Chat) []*pb.AnyUserStatus {
	var readStatus []*pb.AnyUserStatus
	for _, rs := range chat.ReadStatus {
		status := "unread"
		if rs.IsRead {
			status = "read"
		}

		// Set nilai ke dalam readStatus
		readStatus = append(readStatus, &pb.AnyUserStatus{
			MemberId: uint64(rs.GroupMemberID),
			Status:   status,
		})
	}
	return readStatus
}

func ParsingDtoGroupToPB(req []GroupInfo) []*pb.GroupInfo {
	var response []*pb.GroupInfo
	for _, group := range req {
		response = append(response, &pb.GroupInfo{
			Id:          uint64(group.GroupID),
			Name:        group.Name,
			LastMessage: group.LastMessage,
		})
	}

	return response
}

// dto
type MemberChat struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
}

// dto
type GroupInfo struct {
	GroupMemberID uint   `json:"group_member_id"`
	GroupID       uint   `json:"group_id"`
	Name          string `json:"name"`
	LastMessage   string `json:"last_message"`
	UnreadCount   int    `json:"unread_count"`
}
