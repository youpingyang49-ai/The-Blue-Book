package service

import (
	"context"
	"time"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
)

type CommentService struct {
	commentRepo *repository.CommentRepo
}

func NewCommentService(commentRepo *repository.CommentRepo) *CommentService {
	return &CommentService{commentRepo: commentRepo}
}

type CommentDTO struct {
	ID        uint      `json:"id"`
	PostID    uint      `json:"post_id"`
	Content   string    `json:"content"`
	Author    AuthorDTO `json:"author"`
	CreatedAt time.Time `json:"created_at"`
}

func (s *CommentService) Create(ctx context.Context, userID, postID, parentID uint, content string) (*CommentDTO, error) {
	comment := &model.Comment{
		PostID:   postID,
		UserID:   userID,
		ParentID: parentID,
		Content:  content,
	}
	if err := s.commentRepo.Create(comment); err != nil {
		return nil, err
	}
	c, err := s.commentRepo.FindByID(comment.ID)
	if err != nil {
		return nil, err
	}
	return &CommentDTO{
		ID:        c.ID,
		PostID:    c.PostID,
		Content:   c.Content,
		Author:    AuthorDTO{ID: c.User.ID, Nickname: c.User.Nickname},
		CreatedAt: c.CreatedAt,
	}, nil
}

func (s *CommentService) List(ctx context.Context, postID uint, page, pageSize int) ([]CommentDTO, error) {
	comments, err := s.commentRepo.ListByPost(postID, page, pageSize)
	if err != nil {
		return nil, err
	}
	dtos := make([]CommentDTO, 0, len(comments))
	for _, c := range comments {
		dtos = append(dtos, CommentDTO{
			ID:        c.ID,
			PostID:    c.PostID,
			Content:   c.Content,
			Author:    AuthorDTO{ID: c.User.ID, Nickname: c.User.Nickname},
			CreatedAt: c.CreatedAt,
		})
	}
	return dtos, nil
}
