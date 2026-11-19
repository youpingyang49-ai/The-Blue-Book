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
	notifRepo *repository.NotificationRepo
	mq        *mq.Client
}

func NewNotificationService(notifRepo *repository.NotificationRepo, mqClient *mq.Client) *NotificationService {
	return &NotificationService{notifRepo: notifRepo, mq: mqClient}
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

func (s *NotificationService) MarkAllRead(ctx context.Context, userID uint) error {
	return s.notifRepo.MarkAllRead(userID)
}
