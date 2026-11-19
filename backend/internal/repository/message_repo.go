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

func (r *MessageRepo) MarkUnread(userID, otherID uint) error {
	return r.db.Model(&model.Message{}).
		Where("from_user_id = ? AND to_user_id = ? AND is_read = ?", otherID, userID, true).
		Update("is_read", false).Error
}

func (r *MessageRepo) DeleteConversation(userID, otherID uint) error {
	return r.db.Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)", userID, otherID, otherID, userID).
		Delete(&model.Message{}).Error
}

func (r *MessageRepo) HideConversation(userID, otherID uint) error {
	h := model.ConversationHide{UserID: userID, OtherID: otherID}
	return r.db.Where("user_id = ? AND other_id = ?", userID, otherID).FirstOrCreate(&h).Error
}

func (r *MessageRepo) HiddenOtherIDs(userID uint) ([]uint, error) {
	var hides []model.ConversationHide
	if err := r.db.Where("user_id = ?", userID).Find(&hides).Error; err != nil {
		return nil, err
	}
	ids := make([]uint, 0, len(hides))
	for _, h := range hides {
		ids = append(ids, h.OtherID)
	}
	return ids, nil
}
