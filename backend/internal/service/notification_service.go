package service

import (
	"context"
	"time"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
)

type NotificationService struct {
	notifRepo *repository.NotificationRepo
}

func NewNotificationService(notifRepo *repository.NotificationRepo) *NotificationService {
	return &NotificationService{notifRepo: notifRepo}
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
	_ = s.notifRepo.Create(&model.Notification{
		UserID:  recipientID,
		ActorID: actorID,
		Type:    typ,
		PostID:  postID,
		Content: content,
	})
}

func (s *NotificationService) List(ctx context.Context, userID uint, page, pageSize int) ([]NotificationDTO, error) {
	notifs, err := s.notifRepo.ListByUser(userID, page, pageSize)
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
