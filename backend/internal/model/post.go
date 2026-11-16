package model

import (
	"time"

	"gorm.io/gorm"
)

type Post struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	UserID        uint           `gorm:"index;not null" json:"user_id"`
	Title         string         `gorm:"size:255" json:"title"`
	Content       string         `gorm:"type:text" json:"content"`
	LikeCount     int64          `gorm:"default:0" json:"like_count"`
	CommentCount  int64          `gorm:"default:0" json:"comment_count"`
	FavoriteCount int64          `gorm:"default:0" json:"favorite_count"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	User   User        `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Images []PostImage `gorm:"foreignKey:PostID" json:"images,omitempty"`
}

type PostImage struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	PostID    uint      `gorm:"index;not null" json:"post_id"`
	Key       string    `gorm:"size:512;not null" json:"key"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`
}

type Like struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"uniqueIndex:idx_user_post;not null" json:"user_id"`
	PostID    uint      `gorm:"uniqueIndex:idx_user_post;not null" json:"post_id"`
	CreatedAt time.Time `json:"created_at"`
}
