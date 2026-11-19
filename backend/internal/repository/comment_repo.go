package repository

import (
	"gorm.io/gorm"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
)

type CommentRepo struct {
	db *gorm.DB
}

func NewCommentRepo(db *gorm.DB) *CommentRepo {
	return &CommentRepo{db: db}
}

func (r *CommentRepo) Create(comment *model.Comment) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(comment).Error; err != nil {
			return err
		}
		return tx.Model(&model.Post{}).Where("id = ?", comment.PostID).UpdateColumn("comment_count", gorm.Expr("comment_count + 1")).Error
	})
}

func (r *CommentRepo) FindByID(id uint) (*model.Comment, error) {
	var comment model.Comment
	if err := r.db.Preload("User").First(&comment, id).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *CommentRepo) ListByPost(postID uint, page, pageSize int) ([]model.Comment, error) {
	var comments []model.Comment
	err := r.db.Where("post_id = ? AND parent_id = 0", postID).
		Preload("User").
		Order("id ASC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&comments).Error
	return comments, err
}
