package model

import (
	"time"

	"gorm.io/gorm"
)

type Comment struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PostID    uint           `gorm:"index;not null" json:"post_id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	ParentID  uint           `gorm:"index;default:0" json:"parent_id"`
	Content   string         `gorm:"size:1000;not null" json:"content"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
