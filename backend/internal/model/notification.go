package model

import "time"

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	ActorID   uint      `gorm:"not null" json:"actor_id"`
	Type      string    `gorm:"size:20;not null" json:"type"`
	PostID    uint      `gorm:"index;default:0" json:"post_id"`
	Content   string    `gorm:"size:255" json:"content"`
	IsRead    bool      `gorm:"default:false" json:"read"`
	CreatedAt time.Time `json:"created_at"`

	Actor User `gorm:"foreignKey:ActorID" json:"actor,omitempty"`
}
