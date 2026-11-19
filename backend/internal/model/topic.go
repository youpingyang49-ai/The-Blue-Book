package model

import "time"

type Topic struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;size:64;not null" json:"name"`
	PostCount int64     `gorm:"default:0" json:"post_count"`
	CreatedAt time.Time `json:"created_at"`
}

type PostTopic struct {
	ID      uint `gorm:"primaryKey" json:"id"`
	PostID  uint `gorm:"uniqueIndex:idx_post_topic;not null" json:"post_id"`
	TopicID uint `gorm:"uniqueIndex:idx_post_topic;not null" json:"topic_id"`
}
