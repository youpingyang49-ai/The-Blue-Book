package model

import "time"

type Follow struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	FollowerID uint      `gorm:"uniqueIndex:idx_follow;not null" json:"follower_id"`
	FolloweeID uint      `gorm:"uniqueIndex:idx_follow;not null" json:"followee_id"`
	CreatedAt  time.Time `json:"created_at"`
}
