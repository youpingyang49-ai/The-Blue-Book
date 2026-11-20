package service

import (
	"context"
	"errors"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
)

var ErrUserNotFound = errors.New("用户不存在")

type UserService struct {
	userRepo   *repository.UserRepo
	followRepo *repository.FollowRepo
}

func NewUserService(userRepo *repository.UserRepo, followRepo *repository.FollowRepo) *UserService {
	return &UserService{userRepo: userRepo, followRepo: followRepo}
}

type UserProfileDTO struct {
	ID             uint   `json:"id"`
	Nickname       string `json:"nickname"`
	AvatarURL      string `json:"avatar_url"`
	Bio            string `json:"bio"`
	FollowerCount  int64  `json:"follower_count"`
	FollowingCount int64  `json:"following_count"`
	IsFollowing    bool   `json:"is_following"`
}

// FollowUserDTO 用于「关注 / 粉丝」列表。
type FollowUserDTO struct {
	ID          uint   `json:"id"`
	Nickname    string `json:"nickname"`
	AvatarURL   string `json:"avatar_url"`
	Bio         string `json:"bio"`
	IsFollowing bool   `json:"is_following"`
}

func (s *UserService) toFollowDTOs(users []model.User, viewerID uint) ([]FollowUserDTO, error) {
	ids := make([]uint, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	followed, err := s.followRepo.FollowedIDs(viewerID, ids)
	if err != nil {
		followed = map[uint]bool{}
	}
	dtos := make([]FollowUserDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, FollowUserDTO{
			ID:          u.ID,
			Nickname:    u.Nickname,
			AvatarURL:   u.AvatarURL,
			Bio:         u.Bio,
			IsFollowing: followed[u.ID],
		})
	}
	return dtos, nil
}

func (s *UserService) Following(ctx context.Context, userID, viewerID uint, page, pageSize int) ([]FollowUserDTO, int64, error) {
	users, total, err := s.followRepo.ListFollowing(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	dtos, err := s.toFollowDTOs(users, viewerID)
	return dtos, total, err
}

func (s *UserService) Followers(ctx context.Context, userID, viewerID uint, page, pageSize int) ([]FollowUserDTO, int64, error) {
	users, total, err := s.followRepo.ListFollowers(userID, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	dtos, err := s.toFollowDTOs(users, viewerID)
	return dtos, total, err
}

// UserSearchDTO 用户搜索结果项。
type UserSearchDTO struct {
	ID            uint   `json:"id"`
	Nickname      string `json:"nickname"`
	AvatarURL     string `json:"avatar_url"`
	Bio           string `json:"bio"`
	FollowerCount int64  `json:"follower_count"`
	IsFollowing   bool   `json:"is_following"`
}

// SearchUsers 按昵称 / 简介搜索用户。
func (s *UserService) SearchUsers(ctx context.Context, q string, page, pageSize int, viewerID uint) ([]UserSearchDTO, int64, error) {
	users, total, err := s.userRepo.Search(q, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	ids := make([]uint, len(users))
	for i, u := range users {
		ids[i] = u.ID
	}
	followed, err := s.followRepo.FollowedIDs(viewerID, ids)
	if err != nil {
		followed = map[uint]bool{}
	}
	counts, err := s.followRepo.FollowerCounts(ids)
	if err != nil {
		counts = map[uint]int64{}
	}
	dtos := make([]UserSearchDTO, 0, len(users))
	for _, u := range users {
		dtos = append(dtos, UserSearchDTO{
			ID:            u.ID,
			Nickname:      u.Nickname,
			AvatarURL:     u.AvatarURL,
			Bio:           u.Bio,
			FollowerCount: counts[u.ID],
			IsFollowing:   followed[u.ID],
		})
	}
	return dtos, total, nil
}

func (s *UserService) Profile(ctx context.Context, userID, viewerID uint) (*UserProfileDTO, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	followerCount, _ := s.followRepo.FollowerCount(userID)
	followingCount, _ := s.followRepo.FollowingCount(userID)
	isFollowing := false
	if viewerID != 0 && viewerID != userID {
		isFollowing, _ = s.followRepo.IsFollowing(viewerID, userID)
	}
	return &UserProfileDTO{
		ID:             user.ID,
		Nickname:       user.Nickname,
		AvatarURL:      user.AvatarURL,
		Bio:            user.Bio,
		FollowerCount:  followerCount,
		FollowingCount: followingCount,
		IsFollowing:    isFollowing,
	}, nil
}
