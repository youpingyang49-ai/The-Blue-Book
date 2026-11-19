package service

import (
	"context"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
)

type FollowService struct {
	followRepo *repository.FollowRepo
	notifSvc   *NotificationService
}

func NewFollowService(followRepo *repository.FollowRepo, notifSvc *NotificationService) *FollowService {
	return &FollowService{followRepo: followRepo, notifSvc: notifSvc}
}

func (s *FollowService) Follow(ctx context.Context, followerID, followeeID uint) (bool, error) {
	followed, err := s.followRepo.Follow(followerID, followeeID)
	if err != nil {
		return false, err
	}
	if followed && s.notifSvc != nil {
		s.notifSvc.Notify(ctx, followeeID, followerID, "follow", 0, "")
	}
	return followed, nil
}

func (s *FollowService) Unfollow(ctx context.Context, followerID, followeeID uint) (bool, error) {
	return s.followRepo.Unfollow(followerID, followeeID)
}
