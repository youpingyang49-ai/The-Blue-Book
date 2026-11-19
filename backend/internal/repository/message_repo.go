package repository

import (
	"gorm.io/gorm"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
)

type MessageRepo struct {
	db *gorm.DB
}

func NewMessageRepo(db *gorm.DB) *MessageRepo {
	return &MessageRepo{db: db}
}

func (r *MessageRepo) Create(msg *model.Message) error {
	return r.db.Create(msg).Error
}

func (r *MessageRepo) ListConversation(userID, otherID uint, limit int) ([]model.Message, error) {
	var msgs []model.Message
	err := r.db.Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)", userID, otherID, otherID, userID).
		Preload("FromUser").Preload("ToUser").
		Order("id DESC").Limit(limit).
		Find(&msgs).Error
	if err != nil {
		return nil, err
	}
	for i, j := 0, len(msgs)-1; i < j; i, j = i+1, j-1 {
		msgs[i], msgs[j] = msgs[j], msgs[i]
	}
	return msgs, nil
}

func (r *MessageRepo) RecentMessages(userID uint, limit int) ([]model.Message, error) {
	var msgs []model.Message
	err := r.db.Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Preload("FromUser").Preload("ToUser").
		Order("id DESC").Limit(limit).
		Find(&msgs).Error
	return msgs, err
}

func (r *MessageRepo) MarkRead(userID, otherID uint) error {
	return r.db.Model(&model.Message{}).
		Where("from_user_id = ? AND to_user_id = ? AND is_read = ?", otherID, userID, false).
		Update("is_read", true).Error
}
