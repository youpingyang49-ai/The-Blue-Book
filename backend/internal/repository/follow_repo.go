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

// ListFollowing 返回 userID 关注的人（按关注时间倒序）。
func (r *FollowRepo) ListFollowing(userID uint, page, pageSize int) ([]model.User, int64, error) {
	var total int64
	if err := r.db.Model(&model.User{}).
		Joins("JOIN follows ON follows.followee_id = users.id").
		Where("follows.follower_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	err := r.db.Model(&model.User{}).
		Joins("JOIN follows ON follows.followee_id = users.id").
		Where("follows.follower_id = ?", userID).
		Order("follows.created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&users).Error
	return users, total, err
}

// ListFollowers 返回关注 userID 的人（按关注时间倒序）。
func (r *FollowRepo) ListFollowers(userID uint, page, pageSize int) ([]model.User, int64, error) {
	var total int64
	if err := r.db.Model(&model.User{}).
		Joins("JOIN follows ON follows.follower_id = users.id").
		Where("follows.followee_id = ?", userID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	err := r.db.Model(&model.User{}).
		Joins("JOIN follows ON follows.follower_id = users.id").
		Where("follows.followee_id = ?", userID).
		Order("follows.created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&users).Error
	return users, total, err
}

// FollowedIDs 批量查询 viewer 是否关注了给定用户，避免列表 N+1。
func (r *FollowRepo) FollowedIDs(viewerID uint, userIDs []uint) (map[uint]bool, error) {
	set := make(map[uint]bool, len(userIDs))
	if viewerID == 0 || len(userIDs) == 0 {
		return set, nil
	}
	var follows []model.Follow
	if err := r.db.Where("follower_id = ? AND followee_id IN ?", viewerID, userIDs).Find(&follows).Error; err != nil {
		return set, err
	}
	for _, f := range follows {
		set[f.FolloweeID] = true
	}
	return set, nil
}

// FollowerCounts 批量统计粉丝数，供用户搜索列表展示。
func (r *FollowRepo) FollowerCounts(userIDs []uint) (map[uint]int64, error) {
	out := make(map[uint]int64, len(userIDs))
	if len(userIDs) == 0 {
		return out, nil
	}
	type row struct {
		FolloweeID uint
		Total      int64
	}
	var rows []row
	err := r.db.Model(&model.Follow{}).
		Select("followee_id, COUNT(*) AS total").
		Where("followee_id IN ?", userIDs).
		Group("followee_id").
		Scan(&rows).Error
	if err != nil {
		return out, err
	}
	for _, rw := range rows {
		out[rw.FolloweeID] = rw.Total
	}
	return out, nil
}
