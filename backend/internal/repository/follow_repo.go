package repository

import (
	"gorm.io/gorm"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
)

type FollowRepo struct {
	db *gorm.DB
}

func NewFollowRepo(db *gorm.DB) *FollowRepo {
	return &FollowRepo{db: db}
}

func (r *FollowRepo) Follow(followerID, followeeID uint) (bool, error) {
	if followerID == followeeID {
		return false, nil
	}
	var count int64
	if err := r.db.Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ?", followerID, followeeID).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}
	if err := r.db.Create(&model.Follow{FollowerID: followerID, FolloweeID: followeeID}).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (r *FollowRepo) Unfollow(followerID, followeeID uint) (bool, error) {
	res := r.db.Where("follower_id = ? AND followee_id = ?", followerID, followeeID).Delete(&model.Follow{})
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected > 0, nil
}

func (r *FollowRepo) IsFollowing(followerID, followeeID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Follow{}).Where("follower_id = ? AND followee_id = ?", followerID, followeeID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *FollowRepo) FollowerCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Follow{}).Where("followee_id = ?", userID).Count(&count).Error
	return count, err
}

func (r *FollowRepo) FollowingCount(userID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Follow{}).Where("follower_id = ?", userID).Count(&count).Error
	return count, err
}
