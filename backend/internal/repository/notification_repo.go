package repository

import (
	"gorm.io/gorm"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
)

type NotificationRepo struct {
	db *gorm.DB
}

func NewNotificationRepo(db *gorm.DB) *NotificationRepo {
	return &NotificationRepo{db: db}
}

func (r *NotificationRepo) Create(n *model.Notification) error {
	return r.db.Create(n).Error
}

func (r *NotificationRepo) ListByUser(userID uint, page, pageSize int) ([]model.Notification, error) {
	var notifs []model.Notification
	err := r.db.Where("user_id = ?", userID).
		Preload("Actor").
		Order("id DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&notifs).Error
	return notifs, err
}

func (r *NotificationRepo) UnreadCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, false).Count(&count).Error
	return count, err
}

func (r *NotificationRepo) MarkAllRead(userID uint) error {
	return r.db.Model(&model.Notification{}).Where("user_id = ?", userID).Update("is_read", true).Error
}
