package service

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"time"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/rank"
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
	userRepo   *repository.UserRepo
	notifSvc   *NotificationService
	storage    storage.Storage
	ranking    *RankingService
}

func NewPostService(postRepo *repository.PostRepo, followRepo *repository.FollowRepo, topicRepo *repository.TopicRepo, userRepo *repository.UserRepo, notifSvc *NotificationService, st storage.Storage, ranking *RankingService) *PostService {
	return &PostService{
		postRepo:   postRepo,
		followRepo: followRepo,
		topicRepo:  topicRepo,
		userRepo:   userRepo,
		notifSvc:   notifSvc,
		storage:    st,
		ranking:    ranking,
	}
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
	s.touchRank(ctx, post.ID)
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
	if err := s.postRepo.Delete(postID); err != nil {
		return err
	}
	if s.ranking != nil {
		if err := s.ranking.RemovePost(ctx, postID); err != nil {
			slog.Warn("ranking remove failed", "post_id", postID, "error", err)
		}
	}
	return nil
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

	topicMap, err := s.topicRepo.ListNamesByPostIDs(postIDs)
	if err != nil {
		topicMap = nil // 话题仅用于展示，取不到不影响主流程
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
			Topics:        topicMap[p.ID],
			Author: AuthorDTO{
				ID:        p.User.ID,
				Nickname:  p.User.Nickname,
				AvatarURL: p.User.AvatarURL,
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
	offset := (page - 1) * pageSize

	// 推荐 / 热门走 Redis ZSet 榜单；Redis 不可用或榜单为空时回退 MySQL。
	if s.ranking != nil {
		switch sort {
		case "hot":
			if ids, ok := s.ranking.HotPostIDs(ctx, offset, pageSize); ok {
				if posts, err := s.postRepo.ListByIDs(ids); err == nil {
					if ordered := orderPostsByIDs(posts, ids); len(ordered) > 0 {
						return s.buildDTOs(ctx, ordered, viewerID)
					}
				}
			}
		case "recommend":
			if ids, ok := s.ranking.RecommendPostIDs(ctx, viewerID, offset, pageSize); ok {
				if len(ids) == 0 {
					return []PostDTO{}, nil
				}
				if posts, err := s.postRepo.ListByIDs(ids); err == nil {
					if ordered := orderPostsByIDs(posts, ids); len(ordered) > 0 {
						return s.buildDTOs(ctx, ordered, viewerID)
					}
				}
			}
		}
	}

	var posts []model.Post
	var err error
	if sort == "recommend" {
		posts, err = s.postRepo.RecommendList(viewerID, page, pageSize)
	} else {
		posts, err = s.postRepo.List(repository.ListParams{Sort: sort, Page: page, PageSize: pageSize})
	}
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
			ID:        post.User.ID,
			Nickname:  post.User.Nickname,
			AvatarURL: post.User.AvatarURL,
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
		s.touchRank(ctx, postID)
		s.notifyPostAction(ctx, userID, postID, "like", "")
	}
	return liked, nil
}

func (s *PostService) Unlike(ctx context.Context, userID, postID uint) (bool, error) {
	unliked, err := s.postRepo.Unlike(userID, postID)
	if err != nil {
		return false, err
	}
	if unliked {
		s.touchRank(ctx, postID)
	}
	return unliked, nil
}

func (s *PostService) Favorite(ctx context.Context, userID, postID uint) (bool, error) {
	favorited, err := s.postRepo.Favorite(userID, postID)
	if err != nil {
		return false, err
	}
	if favorited {
		s.touchRank(ctx, postID)
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
	unfavorited, err := s.postRepo.Unfavorite(userID, postID)
	if err != nil {
		return false, err
	}
	if unfavorited {
		s.touchRank(ctx, postID)
	}
	return unfavorited, nil
}

// touchRank 在互动后刷新单篇笔记的榜单分数；失败只记日志，不影响主流程。
func (s *PostService) touchRank(ctx context.Context, postID uint) {
	if s.ranking == nil {
		return
	}
	if err := s.ranking.TouchPost(ctx, postID); err != nil {
		slog.Warn("ranking touch failed", "post_id", postID, "error", err)
	}
}

// ------------------------------- 排行榜接口 -------------------------------

type RankedPostDTO struct {
	Rank int     `json:"rank"`
	Heat int     `json:"heat"`
	Post PostDTO `json:"post"`
}

type RankedTopicDTO struct {
	Rank      int    `json:"rank"`
	Heat      int    `json:"heat"`
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	PostCount int64  `json:"post_count"`
}

type RankedCreatorDTO struct {
	Rank          int    `json:"rank"`
	Heat          int    `json:"heat"`
	ID            uint   `json:"id"`
	Nickname      string `json:"nickname"`
	AvatarURL     string `json:"avatar_url"`
	Bio           string `json:"bio"`
	FollowerCount int64  `json:"follower_count"`
}

type RankingsDTO struct {
	Posts       []RankedPostDTO    `json:"posts"`
	Topics      []RankedTopicDTO   `json:"topics"`
	Creators    []RankedCreatorDTO `json:"creators"`
	Source      string             `json:"source"`
	GeneratedAt time.Time          `json:"generated_at"`
}

// Rankings 返回笔记 / 话题 / 创作者三个榜单。Redis 可用时读 ZSet，否则回退 MySQL。
func (s *PostService) Rankings(ctx context.Context, viewerID uint, limit int) (*RankingsDTO, error) {
	if limit < 1 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}
	out := &RankingsDTO{
		Posts:       []RankedPostDTO{},
		Topics:      []RankedTopicDTO{},
		Creators:    []RankedCreatorDTO{},
		Source:      "mysql",
		GeneratedAt: time.Now(),
	}
	if s.ranking != nil {
		snapshot, err := s.ranking.Snapshot(ctx, limit, limit, limit)
		if err == nil && len(snapshot.Posts) > 0 {
			s.fillRankings(ctx, out, snapshot, viewerID)
			out.Source = "redis"
			out.GeneratedAt = snapshot.GeneratedAt
			return out, nil
		}
	}
	return s.rankingsFallback(ctx, out, viewerID, limit)
}

func (s *PostService) fillRankings(ctx context.Context, out *RankingsDTO, snapshot *BoardSnapshot, viewerID uint) {
	postIDs := make([]uint, 0, len(snapshot.Posts))
	for _, entry := range snapshot.Posts {
		postIDs = append(postIDs, entry.ID)
	}
	if posts, err := s.postRepo.ListByIDs(postIDs); err == nil {
		if dtos, err := s.buildDTOs(ctx, orderPostsByIDs(posts, postIDs), viewerID); err == nil {
			byID := make(map[uint]PostDTO, len(dtos))
			for _, dto := range dtos {
				byID[dto.ID] = dto
			}
			top := snapshot.Posts[0].Score
			for _, entry := range snapshot.Posts {
				dto, ok := byID[entry.ID]
				if !ok {
					continue
				}
				out.Posts = append(out.Posts, RankedPostDTO{
					Rank: len(out.Posts) + 1,
					Heat: rank.Normalize(entry.Score, top),
					Post: dto,
				})
			}
		}
	}

	topicIDs := make([]uint, 0, len(snapshot.Topics))
	for _, entry := range snapshot.Topics {
		topicIDs = append(topicIDs, entry.ID)
	}
	if len(topicIDs) > 0 {
		topics, err := s.topicRepo.FindByIDs(topicIDs)
		if err == nil {
			byID := make(map[uint]model.Topic, len(topics))
			for _, t := range topics {
				byID[t.ID] = t
			}
			top := snapshot.Topics[0].Score
			for _, entry := range snapshot.Topics {
				topic, ok := byID[entry.ID]
				if !ok {
					continue
				}
				out.Topics = append(out.Topics, RankedTopicDTO{
					Rank:      len(out.Topics) + 1,
					Heat:      rank.Normalize(entry.Score, top),
					ID:        topic.ID,
					Name:      topic.Name,
					PostCount: topic.PostCount,
				})
			}
		}
	}

	userIDs := make([]uint, 0, len(snapshot.Creators))
	for _, entry := range snapshot.Creators {
		userIDs = append(userIDs, entry.ID)
	}
	if len(userIDs) > 0 {
		users, err := s.userRepo.FindByIDs(userIDs)
		if err == nil {
			counts, _ := s.followRepo.FollowerCounts(userIDs)
			byID := make(map[uint]model.User, len(users))
			for _, u := range users {
				byID[u.ID] = u
			}
			top := snapshot.Creators[0].Score
			for _, entry := range snapshot.Creators {
				user, ok := byID[entry.ID]
				if !ok {
					continue
				}
				out.Creators = append(out.Creators, RankedCreatorDTO{
					Rank:          len(out.Creators) + 1,
					Heat:          rank.Normalize(entry.Score, top),
					ID:            user.ID,
					Nickname:      user.Nickname,
					AvatarURL:     user.AvatarURL,
					Bio:           user.Bio,
					FollowerCount: counts[user.ID],
				})
			}
		}
	}
}

func (s *PostService) rankingsFallback(ctx context.Context, out *RankingsDTO, viewerID uint, limit int) (*RankingsDTO, error) {
	posts, err := s.postRepo.TopBySignal(limit)
	if err != nil {
		return nil, err
	}
	dtos, err := s.buildDTOs(ctx, posts, viewerID)
	if err != nil {
		return nil, err
	}
	for i, dto := range dtos {
		out.Posts = append(out.Posts, RankedPostDTO{Rank: i + 1, Post: dto})
	}

	if topics, err := s.topicRepo.TopByPostCount(limit); err == nil {
		for i, topic := range topics {
			out.Topics = append(out.Topics, RankedTopicDTO{
				Rank:      i + 1,
				ID:        topic.ID,
				Name:      topic.Name,
				PostCount: topic.PostCount,
			})
		}
	}

	if users, err := s.userRepo.TopByFollowerCount(limit); err == nil {
		ids := make([]uint, len(users))
		for i, u := range users {
			ids[i] = u.ID
		}
		counts, _ := s.followRepo.FollowerCounts(ids)
		for i, user := range users {
			out.Creators = append(out.Creators, RankedCreatorDTO{
				Rank:          i + 1,
				ID:            user.ID,
				Nickname:      user.Nickname,
				AvatarURL:     user.AvatarURL,
				Bio:           user.Bio,
				FollowerCount: counts[user.ID],
			})
		}
	}
	return out, nil
}

// orderPostsByIDs 把数据库返回的笔记按 ID 列表顺序还原（榜单顺序）。
func orderPostsByIDs(posts []model.Post, ids []uint) []model.Post {
	byID := make(map[uint]model.Post, len(posts))
	for _, post := range posts {
		byID[post.ID] = post
	}
	ordered := make([]model.Post, 0, len(posts))
	for _, id := range ids {
		if post, ok := byID[id]; ok {
			ordered = append(ordered, post)
		}
	}
	return ordered
}
