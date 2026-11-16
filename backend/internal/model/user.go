package model

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Phone     string         `gorm:"uniqueIndex;size:20;not null" json:"phone"`
	Nickname  string         `gorm:"size:64" json:"nickname"`
	AvatarURL string         `gorm:"size:255" json:"avatar_url"`
	Bio       string         `gorm:"size:255" json:"bio"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}
