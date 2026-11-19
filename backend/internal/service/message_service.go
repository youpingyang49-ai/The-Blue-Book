package service

import (
	"context"
	"time"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
)

type MessageService struct {
	messageRepo *repository.MessageRepo
}

func NewMessageService(messageRepo *repository.MessageRepo) *MessageService {
	return &MessageService{messageRepo: messageRepo}
}

type MessageDTO struct {
	ID         uint      `json:"id"`
	FromUserID uint      `json:"from_user_id"`
	ToUserID   uint      `json:"to_user_id"`
	Content    string    `json:"content"`
	Read       bool      `json:"read"`
	CreatedAt  time.Time `json:"created_at"`
}

type ConversationDTO struct {
	UserID      uint      `json:"user_id"`
	Nickname    string    `json:"nickname"`
	AvatarURL   string    `json:"avatar_url"`
	LastMessage string    `json:"last_message"`
	Unread      int64     `json:"unread"`
	CreatedAt   time.Time `json:"created_at"`
}

func (s *MessageService) Send(ctx context.Context, fromUserID, toUserID uint, content string) (*MessageDTO, error) {
	msg := &model.Message{FromUserID: fromUserID, ToUserID: toUserID, Content: content}
	if err := s.messageRepo.Create(msg); err != nil {
		return nil, err
	}
	return &MessageDTO{ID: msg.ID, FromUserID: fromUserID, ToUserID: toUserID, Content: content, CreatedAt: msg.CreatedAt}, nil
}

func (s *MessageService) Conversation(ctx context.Context, userID, otherID uint) ([]MessageDTO, error) {
	msgs, err := s.messageRepo.ListConversation(userID, otherID, 100)
	if err != nil {
		return nil, err
	}
	_ = s.messageRepo.MarkRead(userID, otherID)
	dtos := make([]MessageDTO, 0, len(msgs))
	for _, m := range msgs {
		dtos = append(dtos, MessageDTO{ID: m.ID, FromUserID: m.FromUserID, ToUserID: m.ToUserID, Content: m.Content, Read: m.IsRead, CreatedAt: m.CreatedAt})
	}
	return dtos, nil
}

func (s *MessageService) Conversations(ctx context.Context, userID uint) ([]ConversationDTO, error) {
	msgs, err := s.messageRepo.RecentMessages(userID, 200)
	if err != nil {
		return nil, err
	}
	type conv struct {
		user        model.User
		lastMessage string
		unread      int64
		createdAt   time.Time
	}
	m := map[uint]*conv{}
	var order []uint
	for _, msg := range msgs {
		other := msg.FromUserID
		if other == userID {
			other = msg.ToUserID
		}
		c, ok := m[other]
		if !ok {
			var u model.User
			if msg.FromUserID == other {
				u = msg.FromUser
			} else {
				u = msg.ToUser
			}
			c = &conv{user: u, lastMessage: msg.Content, createdAt: msg.CreatedAt}
			m[other] = c
			order = append(order, other)
		}
		if msg.ToUserID == userID && !msg.IsRead {
			c.unread++
		}
	}
	dtos := make([]ConversationDTO, 0, len(order))
	for _, id := range order {
		c := m[id]
		dtos = append(dtos, ConversationDTO{
			UserID:      c.user.ID,
			Nickname:    c.user.Nickname,
			AvatarURL:   c.user.AvatarURL,
			LastMessage: c.lastMessage,
			Unread:      c.unread,
			CreatedAt:   c.createdAt,
		})
	}
	return dtos, nil
}
