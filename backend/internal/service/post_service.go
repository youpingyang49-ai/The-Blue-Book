package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/storage"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
)

var (
	ErrPostNotFound = errors.New("帖子不存在")
	ErrForbidden    = errors.New("无权限操作")
)

type PostService struct {
	postRepo   *repository.PostRepo
	followRepo *repository.FollowRepo
	topicRepo  *repository.TopicRepo
	notifSvc   *NotificationService
	storage    storage.Storage
}

func NewPostService(postRepo *repository.PostRepo, followRepo *repository.FollowRepo, topicRepo *repository.TopicRepo, notifSvc *NotificationService, st storage.Storage) *PostService {
	return &PostService{postRepo: postRepo, followRepo: followRepo, topicRepo: topicRepo, notifSvc: notifSvc, storage: st}
}

type AuthorDTO struct {
	ID        uint   `json:"id"`
	Nickname  string `json:"nickname"`
	AvatarURL string `json:"avatar_url"`
}

type PostDTO struct {
	ID            uint      `json:"id"`
	Title         string    `json:"title"`
	Content       string    `json:"content,omitempty"`
	CoverURL      string    `json:"cover_image_url"`
	ImageURLs     []string  `json:"image_urls,omitempty"`
	LikeCount     int64     `json:"like_count"`
	FavoriteCount int64     `json:"favorite_count"`
	CommentCount  int64     `json:"comment_count"`
	IsLiked       bool      `json:"is_liked"`
	IsFavorited   bool      `json:"is_favorited"`
	IsFollowing   bool      `json:"is_following"`
	Topics        []string  `json:"topics"`
	Author        AuthorDTO `json:"author"`
	CreatedAt     time.Time `json:"created_at"`
}

func (s *PostService) Create(ctx context.Context, userID uint, title, content string, keys []string, topics []string) (*model.Post, error) {
	post := &model.Post{
		UserID:  userID,
		Title:   title,
		Content: content,
	}
	for i, k := range keys {
		post.Images = append(post.Images, model.PostImage{Key: k, SortOrder: i})
	}
	if err := s.postRepo.Create(post); err != nil {
		return nil, err
	}
	for _, name := range topics {
		name = strings.TrimPrefix(strings.TrimSpace(name), "#")
		if name == "" {
			continue
		}
		topic, err := s.topicRepo.FindOrCreateByName(name)
		if err != nil {
			return nil, err
		}
		if err := s.topicRepo.Link(post.ID, topic.ID); err != nil {
			return nil, err
		}
		_ = s.topicRepo.IncrementCount(topic.ID)
	}
	return post, nil
}

func (s *PostService) Update(ctx context.Context, userID, postID uint, title, content string) error {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return err
	}
	if post == nil {
		return ErrPostNotFound
	}
	if post.UserID != userID {
		return ErrForbidden
	}
	return s.postRepo.Update(postID, title, content)
}

func (s *PostService) Delete(ctx context.Context, userID, postID uint) error {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return err
	}
	if post == nil {
		return ErrPostNotFound
	}
	if post.UserID != userID {
		return ErrForbidden
	}
	return s.postRepo.Delete(postID)
}

func (s *PostService) buildDTOs(ctx context.Context, posts []model.Post, viewerID uint) ([]PostDTO, error) {
	postIDs := make([]uint, len(posts))
	for i, p := range posts {
		postIDs[i] = p.ID
	}

	var liked map[uint]bool
	if viewerID != 0 {
		var err error
		liked, err = s.postRepo.LikedPostIDs(viewerID, postIDs)
		if err != nil {
			return nil, err
		}
	}

	dtos := make([]PostDTO, 0, len(posts))
	for _, p := range posts {
		dto := PostDTO{
			ID:            p.ID,
			Title:         p.Title,
			LikeCount:     p.LikeCount,
			FavoriteCount: p.FavoriteCount,
			CommentCount:  p.CommentCount,
			CreatedAt:     p.CreatedAt,
			Author: AuthorDTO{
				ID:       p.User.ID,
				Nickname: p.User.Nickname,
			},
		}
		if len(p.Images) > 0 {
			if cover, err := s.storage.PresignedGetURL(ctx, p.Images[0].Key); err == nil {
				dto.CoverURL = cover
			}
		}
		if liked != nil {
			dto.IsLiked = liked[p.ID]
		}
		dtos = append(dtos, dto)
	}
	return dtos, nil
}

func (s *PostService) Feed(ctx context.Context, sort string, page, pageSize int, viewerID uint) ([]PostDTO, error) {
	posts, err := s.postRepo.List(repository.ListParams{Sort: sort, Page: page, PageSize: pageSize})
	if err != nil {
		return nil, err
	}
	return s.buildDTOs(ctx, posts, viewerID)
}

func (s *PostService) Search(ctx context.Context, q string, page, pageSize int, viewerID uint) ([]PostDTO, error) {
	posts, err := s.postRepo.Search(q, page, pageSize)
	if err != nil {
		return nil, err
	}
	return s.buildDTOs(ctx, posts, viewerID)
}

func (s *PostService) MyPosts(ctx context.Context, userID uint, page, pageSize int) ([]PostDTO, error) {
	posts, err := s.postRepo.ListByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return s.buildDTOs(ctx, posts, userID)
}

func (s *PostService) MyLikes(ctx context.Context, userID uint, page, pageSize int) ([]PostDTO, error) {
	posts, err := s.postRepo.ListLikedByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return s.buildDTOs(ctx, posts, userID)
}

func (s *PostService) MyFavorites(ctx context.Context, userID uint, page, pageSize int) ([]PostDTO, error) {
	posts, err := s.postRepo.ListFavoritedByUser(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return s.buildDTOs(ctx, posts, userID)
}

func (s *PostService) FollowingFeed(ctx context.Context, userID uint, page, pageSize int) ([]PostDTO, error) {
	posts, err := s.postRepo.ListByFollowees(userID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return s.buildDTOs(ctx, posts, userID)
}

func (s *PostService) UserPosts(ctx context.Context, targetUserID uint, page, pageSize int, viewerID uint) ([]PostDTO, error) {
	posts, err := s.postRepo.ListByUser(targetUserID, page, pageSize)
	if err != nil {
		return nil, err
	}
	return s.buildDTOs(ctx, posts, viewerID)
}

func (s *PostService) Detail(ctx context.Context, postID, viewerID uint) (*PostDTO, error) {
	post, err := s.postRepo.FindByID(postID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrPostNotFound
	}

	dto := PostDTO{
		ID:            post.ID,
		Title:         post.Title,
		Content:       post.Content,
		LikeCount:     post.LikeCount,
		FavoriteCount: post.FavoriteCount,
		CommentCount:  post.CommentCount,
		CreatedAt:     post.CreatedAt,
		Author: AuthorDTO{
			ID:       post.User.ID,
			Nickname: post.User.Nickname,
		},
	}
	for _, img := range post.Images {
		if u, err := s.storage.PresignedGetURL(ctx, img.Key); err == nil {
			dto.ImageURLs = append(dto.ImageURLs, u)
		}
	}
	if len(dto.ImageURLs) > 0 {
		dto.CoverURL = dto.ImageURLs[0]
	}
	if topics, err := s.topicRepo.ListByPost(post.ID); err == nil {
		dto.Topics = make([]string, 0, len(topics))
		for _, t := range topics {
			dto.Topics = append(dto.Topics, t.Name)
		}
	}
	if viewerID != 0 {
		liked, _ := s.postRepo.LikedPostIDs(viewerID, []uint{postID})
		dto.IsLiked = liked[postID]
		faved, _ := s.postRepo.FavoritedPostIDs(viewerID, []uint{postID})
		dto.IsFavorited = faved[postID]
		following, _ := s.followRepo.IsFollowing(viewerID, post.UserID)
		dto.IsFollowing = following
	}
	return &dto, nil
}

func (s *PostService) Like(ctx context.Context, userID, postID uint) (bool, error) {
	liked, err := s.postRepo.Like(userID, postID)
	if err != nil {
		return false, err
	}
	if liked {
		s.notifyPostAction(ctx, userID, postID, "like", "")
	}
	return liked, nil
}

func (s *PostService) Unlike(ctx context.Context, userID, postID uint) (bool, error) {
	return s.postRepo.Unlike(userID, postID)
}

func (s *PostService) Favorite(ctx context.Context, userID, postID uint) (bool, error) {
	favorited, err := s.postRepo.Favorite(userID, postID)
	if err != nil {
		return false, err
	}
	if favorited {
		s.notifyPostAction(ctx, userID, postID, "favorite", "")
	}
	return favorited, nil
}

func (s *PostService) notifyPostAction(ctx context.Context, actorID, postID uint, typ, content string) {
	if s.notifSvc == nil {
		return
	}
	post, err := s.postRepo.FindByID(postID)
	if err != nil || post == nil {
		return
	}
	s.notifSvc.Notify(ctx, post.UserID, actorID, typ, postID, content)
}

func (s *PostService) Unfavorite(ctx context.Context, userID, postID uint) (bool, error) {
	return s.postRepo.Unfavorite(userID, postID)
}
