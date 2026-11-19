package model

import "time"

type Message struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FromUserID uint      `gorm:"index;not null" json:"from_user_id"`
	ToUserID   uint      `gorm:"index;not null" json:"to_user_id"`
	Content    string    `gorm:"size:1000;not null" json:"content"`
	IsRead     bool      `gorm:"default:false" json:"read"`
	CreatedAt  time.Time `json:"created_at"`

	FromUser User `gorm:"foreignKey:FromUserID" json:"from_user,omitempty"`
	ToUser   User `gorm:"foreignKey:ToUserID" json:"to_user,omitempty"`
}

type ConversationHide struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	UserID  uint `gorm:"uniqueIndex:idx_user_other;not null" json:"user_id"`
	OtherID uint `gorm:"uniqueIndex:idx_user_other;not null" json:"other_id"`
}
