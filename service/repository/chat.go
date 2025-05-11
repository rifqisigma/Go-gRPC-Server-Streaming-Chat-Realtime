package repository

import (
	"chat_api/entity"
	"chat_api/utils/helper"
	"context"
	"errors"

	"gorm.io/gorm"
)

type ChatRepo interface {
	//write chat
	CreateChat(ctx context.Context, req *helper.CreateChatReq) (*entity.Chat, error)
	DeleteChat(ctx context.Context, chatId uint) (*entity.Chat, error)
	UpdateChat(ctx context.Context, chatId uint, message string) (*entity.Chat, error)

	//write group & member
	CreateGroup(ctx context.Context, name, desc string, userId uint) error
	UpdateGroup(ctx context.Context, name, desc string, groupId uint) error
	DeleteGroup(ctx context.Context, groupId uint) error
	AddMember(req []entity.GroupMember) error
	RemoveMember(req []uint) error
	ExitGroup(memberId uint) error
	UpdateRoleUser(memberId uint, role string) error

	IsMemberAdmin(memberId uint) (bool, error)
	GetGroupMemberID(ctx context.Context, userID, groupID uint) (bool, uint, error)
	GetChatsByGroupID(groupID uint) ([]entity.Chat, error)
	GetMemberGroup(groupId uint) ([]helper.MemberChat, error)
	UpdateUnreadMessage(memberId uint) error
	GetListGroup(userId uint) ([]helper.GroupInfo, error)
}

type chatRepo struct {
	db *gorm.DB
}

func NewChatRepo(db *gorm.DB) ChatRepo {
	return &chatRepo{db}
}

// write chat
func (r *chatRepo) CreateChat(ctx context.Context, req *helper.CreateChatReq) (*entity.Chat, error) {
	tx := r.db.WithContext(ctx).Begin()

	chat := entity.Chat{
		GroupMemberID: &req.MemberId,
		Message:       req.Message,
		GroupID:       req.GroupId,
	}
	if err := tx.Create(&chat).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	var chatRead []entity.ChatRead
	for _, cr := range req.AnyStatusUser {
		if cr.MemberId == req.MemberId {
			continue
		}
		status := cr.Status == "online"
		chatRead = append(chatRead, entity.ChatRead{
			ChatID:        chat.ID,
			GroupMemberID: cr.MemberId,
			IsRead:        status,
		})
	}

	if len(chatRead) > 0 {
		if err := tx.CreateInBatches(&chatRead, 10).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Model(&entity.ChatGroup{}).
		Where("id = ?", req.GroupId).
		Update("last_message", req.Message).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &chat, nil

}

func (r *chatRepo) DeleteChat(ctx context.Context, chatId uint) (*entity.Chat, error) {
	var chat entity.Chat
	if err := r.db.Preload("GroupMember.User").Preload("ReadStatus").Where("id = ?", chatId).First(&chat).Error; err != nil {
		return nil, err
	}

	if err := r.db.Model(&entity.Chat{}).WithContext(ctx).Where("id = ?", chatId).Delete(&entity.Chat{}).Error; err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *chatRepo) UpdateChat(ctx context.Context, chatId uint, message string) (*entity.Chat, error) {

	if err := r.db.Model(&entity.Chat{}).WithContext(ctx).Where("id = ?", chatId).Update("message", message).Error; err != nil {
		return nil, err
	}

	var chat entity.Chat
	if err := r.db.Preload("GroupMember.User").Preload("ReadStatus").Where("id = ?", chatId).First(&chat).Error; err != nil {
		return nil, err
	}
	return &chat, nil
}

// write group & member
func (r *chatRepo) CreateGroup(ctx context.Context, name, desc string, userId uint) error {
	tx := r.db.WithContext(ctx).Begin()

	newGroup := entity.ChatGroup{
		Name:        name,
		Description: desc,
	}
	if err := tx.Create(&newGroup).Error; err != nil {
		tx.Rollback()
		return err
	}

	admin := entity.GroupMember{
		GroupID: newGroup.ID,
		UserID:  userId,
		Role:    "admin",
	}
	if err := tx.Create(&admin).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (r *chatRepo) UpdateGroup(ctx context.Context, name, desc string, groupId uint) error {
	return r.db.Model(&entity.ChatGroup{}).Where("id = ?", groupId).Updates(map[string]interface{}{
		"name":        name,
		"description": desc,
	}).Error
}

func (r *chatRepo) DeleteGroup(ctx context.Context, groupId uint) error {
	return r.db.Model(&entity.ChatGroup{}).WithContext(ctx).Where("id = ?", groupId).Delete(&entity.ChatGroup{}).Error
}

func (r *chatRepo) AddMember(req []entity.GroupMember) error {
	return r.db.Model(&entity.GroupMember{}).CreateInBatches(req, 2).Error
}

func (r *chatRepo) RemoveMember(req []uint) error {
	return r.db.Model(&entity.GroupMember{}).Where("id IN ?", req).Delete(&entity.GroupMember{}).Error
}

func (r *chatRepo) ExitGroup(memberId uint) error {
	return r.db.Model(&entity.GroupMember{}).Where("id = ?", memberId).Delete(&entity.GroupMember{}).Error
}

func (r *chatRepo) UpdateRoleUser(memberId uint, role string) error {
	return r.db.Model(&entity.GroupMember{}).Where("id = ?", memberId).Update("role", role).Error
}

func (r *chatRepo) IsMemberAdmin(memberId uint) (bool, error) {
	var count int64
	if err := r.db.Model(&entity.GroupMember{}).Where("id = ? AND role = ?", memberId, "admin").Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}
func (r *chatRepo) GetGroupMemberID(ctx context.Context, userID, groupID uint) (bool, uint, error) {
	var member entity.GroupMember
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		First(&member).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, 0, err
		}
	}
	return true, member.ID, nil
}

func (r *chatRepo) GetChatsByGroupID(groupID uint) ([]entity.Chat, error) {
	var chats []entity.Chat
	if err := r.db.Preload("GroupMember.User").Preload("ReadStatus").Where("group_id = ?", groupID).Order("created_at ASC").Find(&chats).Error; err != nil {
		return nil, err
	}

	return chats, nil
}

func (r *chatRepo) GetMemberGroup(groupId uint) ([]helper.MemberChat, error) {
	var members []helper.MemberChat
	if err := r.db.Model(&entity.GroupMember{}).Select("group_members.id", "users.username").Joins("JOIN users ON users.id = group_members.user_id").Where("group_id =?", groupId).Scan(&members).Error; err != nil {
		return nil, err
	}
	return members, nil
}

func (r *chatRepo) UpdateUnreadMessage(memberId uint) error {
	err := r.db.Model(&entity.ChatRead{}).Where("group_member_id = ? AND is_read = ?", memberId, false).Updates(map[string]interface{}{
		"is_read": true,
	})
	if err.Error != nil {
		return err.Error
	}
	if err.RowsAffected == 0 {
		return nil
	}
	return nil
}

func (r *chatRepo) GetListGroup(userId uint) ([]helper.GroupInfo, error) {

	var groups []helper.GroupInfo
	if err := r.db.Model(&entity.GroupMember{}).
		Select("group_members.id AS group_member_id",
			"chat_groups.id AS group_id",
			"chat_groups.name",
			"chat_groups.last_message",
			"COUNT(chat_reads.id) AS unread_count").
		Joins("JOIN chat_groups ON chat_groups.id = group_members.group_id").
		Joins("LEFT JOIN chat_reads ON chat_reads.group_member_id = group_members.id AND chat_reads.is_read = false").
		Where("group_members.user_id = ?", userId).
		Group("group_members.id, chat_groups.id, chat_groups.name, chat_groups.last_message").
		Find(&groups).Error; err != nil {
		return nil, err
	}
	return groups, nil

}
