package service

import (
	"context"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
)

type FollowService struct {
	followRepo *repository.FollowRepo
}

func NewFollowService(followRepo *repository.FollowRepo) *FollowService {
	return &FollowService{followRepo: followRepo}
}

func (s *FollowService) Follow(ctx context.Context, followerID, followeeID uint) (bool, error) {
	return s.followRepo.Follow(followerID, followeeID)
}

func (s *FollowService) Unfollow(ctx context.Context, followerID, followeeID uint) (bool, error) {
	return s.followRepo.Unfollow(followerID, followeeID)
}
