package entity

import (
	"time"
)

type User struct {
	ID         uint      `gorm:"primaryKey"`
	Username   string    `gorm:"unique"`
	Email      string    `gorm:"unique"`
	Password   string    `gorm:"not null"`
	IsVerified bool      `gorm:"default:false"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

// chat
type ChatGroup struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	LastMessage string
	UnreadCount int
	Members     []GroupMember `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
	Chats       []Chat        `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
}

type GroupMember struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"index"`
	User      User      `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE"`
	GroupID   uint      `gorm:"index"`
	Role      string    `gorm:"not null"`
	ChatGroup ChatGroup `gorm:"foreignKey:GroupID;constraint:OnDelete:CASCADE"`
}

type Chat struct {
	ID            uint         `gorm:"primaryKey"`
	GroupMemberID *uint        `gorm:"index"`
	GroupMember   *GroupMember `gorm:"foreignKey:GroupMemberID;constraint:OnDelete:SET NULL"`
	Message       string       `gorm:"not null"`
	GroupID       uint         `gorm:"index"`
	CreatedAt     time.Time    `gorm:"autoCreateTime"`
	ReadStatus    []ChatRead   `gorm:"foreignKey:ChatID;constraint:OnDelete:CASCADE"`
}

type ChatRead struct {
	ID            uint        `gorm:"primaryKey"`
	GroupMemberID uint        `gorm:"uniqueIndex:idx_chat"`
	GroupMember   GroupMember `gorm:"foreignKey:GroupMemberID;OnDelete:CASCADE"`
	ChatID        uint        `gorm:"uniqueIndex:idx_chat"`
	IsRead        bool        `gorm:"default:false"`
}
