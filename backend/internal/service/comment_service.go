package service

import (
	"context"
	"regexp"
	"time"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
)

type CommentService struct {
	commentRepo *repository.CommentRepo
	postRepo    *repository.PostRepo
	userRepo    *repository.UserRepo
	notifSvc    *NotificationService
}

func NewCommentService(commentRepo *repository.CommentRepo, postRepo *repository.PostRepo, userRepo *repository.UserRepo, notifSvc *NotificationService) *CommentService {
	return &CommentService{commentRepo: commentRepo, postRepo: postRepo, userRepo: userRepo, notifSvc: notifSvc}
}

type CommentDTO struct {
	ID          uint      `json:"id"`
	PostID      uint      `json:"post_id"`
	ParentID    uint      `json:"parent_id"`
	Content     string    `json:"content"`
	ReplyToName string    `json:"reply_to_name,omitempty"`
	Author      AuthorDTO `json:"author"`
	CreatedAt   time.Time `json:"created_at"`
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
	if s.notifSvc != nil {
		if parentID == 0 {
			if authorID, err := s.postRepo.FindAuthorID(postID); err == nil {
				s.notifSvc.Notify(ctx, authorID, userID, "comment", postID, content)
			}
		} else {
			if parent, err := s.commentRepo.FindByID(parentID); err == nil && parent != nil {
				s.notifSvc.Notify(ctx, parent.UserID, userID, "comment", postID, content)
			}
		}
	}
	s.notifyMentions(ctx, userID, postID, content)
	c, err := s.commentRepo.FindByID(comment.ID)
	if err != nil {
		return nil, err
	}
	return &CommentDTO{
		ID:        c.ID,
		PostID:    c.PostID,
		ParentID:  c.ParentID,
		Content:   c.Content,
		Author:    AuthorDTO{ID: c.User.ID, Nickname: c.User.Nickname},
		CreatedAt: c.CreatedAt,
	}, nil
}

func (s *CommentService) notifyMentions(ctx context.Context, actorID, postID uint, content string) {
	if s.notifSvc == nil {
		return
	}
	re := regexp.MustCompile(`@([^\s@]+)`)
	seen := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(content, -1) {
		name := m[1]
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		user, err := s.userRepo.FindByNickname(name)
		if err != nil || user == nil {
			continue
		}
		s.notifSvc.Notify(ctx, user.ID, actorID, "mention", postID, content)
	}
}

func (s *CommentService) List(ctx context.Context, postID uint) ([]CommentDTO, error) {
	comments, err := s.commentRepo.ListByPost(postID)
	if err != nil {
		return nil, err
	}
	nameMap := make(map[uint]string, len(comments))
	for _, c := range comments {
		nameMap[c.ID] = c.User.Nickname
	}
	dtos := make([]CommentDTO, 0, len(comments))
	for _, c := range comments {
		dto := CommentDTO{
			ID:        c.ID,
			PostID:    c.PostID,
			ParentID:  c.ParentID,
			Content:   c.Content,
			Author:    AuthorDTO{ID: c.User.ID, Nickname: c.User.Nickname},
			CreatedAt: c.CreatedAt,
		}
		if c.ParentID != 0 {
			dto.ReplyToName = nameMap[c.ParentID]
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}
