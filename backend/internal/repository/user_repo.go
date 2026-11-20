package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByPhone(phone string) (*model.User, error) {
	var user model.User
	err := r.db.Where("phone = ?", phone).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepo) FindByNickname(nickname string) (*model.User, error) {
	var user model.User
	err := r.db.Where("nickname = ?", nickname).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Search 按昵称 / 简介模糊搜索用户。
func (r *UserRepo) Search(q string, page, pageSize int) ([]model.User, int64, error) {
	like := "%" + q + "%"
	var total int64
	if err := r.db.Model(&model.User{}).
		Where("nickname LIKE ? OR bio LIKE ?", like, like).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []model.User
	err := r.db.Where("nickname LIKE ? OR bio LIKE ?", like, like).
		Order("id ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&users).Error
	return users, total, err
}
