package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/mq"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
)

const NotificationQueue = "notifications"

type notificationPayload struct {
	RecipientID uint   `json:"recipient_id"`
	ActorID     uint   `json:"actor_id"`
	Type        string `json:"type"`
	PostID      uint   `json:"post_id"`
	Content     string `json:"content"`
}

type NotificationService struct {
	notifRepo   *repository.NotificationRepo
	messageRepo *repository.MessageRepo
	mq          *mq.Client
}

func NewNotificationService(notifRepo *repository.NotificationRepo, messageRepo *repository.MessageRepo, mqClient *mq.Client) *NotificationService {
	return &NotificationService{notifRepo: notifRepo, messageRepo: messageRepo, mq: mqClient}
}

// UnreadStats 未读统计：按通知类型拆分，并包含私信未读。
type UnreadStats struct {
	Count  int64            `json:"count"`
	ByType map[string]int64 `json:"by_type"`
	DM     int64            `json:"dm"`
	Total  int64            `json:"total"`
}

type NotificationDTO struct {
	ID        uint      `json:"id"`
	Type      string    `json:"type"`
	PostID    uint      `json:"post_id"`
	Content   string    `json:"content"`
	Read      bool      `json:"read"`
	Actor     AuthorDTO `json:"actor"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *NotificationService) Notify(ctx context.Context, recipientID, actorID uint, typ string, postID uint, content string) {
	if recipientID == 0 || actorID == 0 || recipientID == actorID {
		return
	}
	if s.mq != nil {
		if payload, err := json.Marshal(notificationPayload{
			RecipientID: recipientID,
			ActorID:     actorID,
			Type:        typ,
			PostID:      postID,
			Content:     content,
		}); err == nil {
			_ = s.mq.Publish(NotificationQueue, payload)
			return
		}
	}
	// 兜底：未连接 MQ 时同步写入
	_ = s.notifRepo.Create(&model.Notification{
		UserID: recipientID, ActorID: actorID, Type: typ, PostID: postID, Content: content,
	})
}

func (s *NotificationService) ProcessNotification(body []byte) error {
	var p notificationPayload
	if err := json.Unmarshal(body, &p); err != nil {
		return err
	}
	return s.notifRepo.Create(&model.Notification{
		UserID: p.RecipientID, ActorID: p.ActorID, Type: p.Type, PostID: p.PostID, Content: p.Content,
	})
}

func (s *NotificationService) List(ctx context.Context, userID uint, typ string, page, pageSize int) ([]NotificationDTO, error) {
	notifs, err := s.notifRepo.ListByUser(userID, typ, page, pageSize)
	if err != nil {
		return nil, err
	}
	dtos := make([]NotificationDTO, 0, len(notifs))
	for _, n := range notifs {
		dtos = append(dtos, NotificationDTO{
			ID:        n.ID,
			Type:      n.Type,
			PostID:    n.PostID,
			Content:   n.Content,
			Read:      n.IsRead,
			Actor:     AuthorDTO{ID: n.Actor.ID, Nickname: n.Actor.Nickname, AvatarURL: n.Actor.AvatarURL},
			CreatedAt: n.CreatedAt,
		})
	}
	return dtos, nil
}

func (s *NotificationService) UnreadCount(ctx context.Context, userID uint) (int64, error) {
	return s.notifRepo.UnreadCount(userID)
}

// Stats 返回按类型拆分的未读数与私信未读数。
func (s *NotificationService) Stats(ctx context.Context, userID uint) (*UnreadStats, error) {
	byType, err := s.notifRepo.UnreadCountByType(userID)
	if err != nil {
		return nil, err
	}
	var count int64
	for _, v := range byType {
		count += v
	}
	var dm int64
	if s.messageRepo != nil {
		dm, err = s.messageRepo.UnreadCount(userID)
		if err != nil {
			return nil, err
		}
	}
	return &UnreadStats{Count: count, ByType: byType, DM: dm, Total: count + dm}, nil
}

// MarkRead 按类型标记已读；types 为空时全部标记。
func (s *NotificationService) MarkRead(ctx context.Context, userID uint, types []string) error {
	return s.notifRepo.MarkReadByTypes(userID, types)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uint) error {
	return s.notifRepo.MarkAllRead(userID)
}
