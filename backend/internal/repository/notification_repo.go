package repository

import (
	"strings"

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

func (r *NotificationRepo) ListByUser(userID uint, typ string, page, pageSize int) ([]model.Notification, error) {
	q := r.db.Where("user_id = ?", userID)
	if typ != "" {
		q = q.Where("type IN ?", strings.Split(typ, ","))
	}
	var notifs []model.Notification
	err := q.Preload("Actor").
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

// UnreadCountByType 按通知类型统计未读数（点赞 / 收藏 / 评论 / 关注 / @）。
func (r *NotificationRepo) UnreadCountByType(userID uint) (map[string]int64, error) {
	type row struct {
		Type  string
		Total int64
	}
	var rows []row
	err := r.db.Model(&model.Notification{}).
		Select("type, COUNT(*) AS total").
		Where("user_id = ? AND is_read = ?", userID, false).
		Group("type").
		Scan(&rows).Error
	out := make(map[string]int64, len(rows))
	if err != nil {
		return out, err
	}
	for _, rw := range rows {
		out[rw.Type] = rw.Total
	}
	return out, nil
}

// MarkReadByTypes 按类型标记已读；types 为空时全部标记。
func (r *NotificationRepo) MarkReadByTypes(userID uint, types []string) error {
	q := r.db.Model(&model.Notification{}).Where("user_id = ? AND is_read = ?", userID, false)
	if len(types) > 0 {
		q = q.Where("type IN ?", types)
	}
	return q.Update("is_read", true).Error
}

func (r *NotificationRepo) MarkAllRead(userID uint) error {
	return r.db.Model(&model.Notification{}).Where("user_id = ?", userID).Update("is_read", true).Error
}
