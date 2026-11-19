package repository

import (
	"errors"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
)

type PostRepo struct {
	db *gorm.DB
}

func NewPostRepo(db *gorm.DB) *PostRepo {
	return &PostRepo{db: db}
}

func (r *PostRepo) Create(post *model.Post) error {
	return r.db.Create(post).Error
}

func (r *PostRepo) FindByID(id uint) (*model.Post, error) {
	var post model.Post
	err := r.db.Preload("User").Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).First(&post, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &post, nil
}

type ListParams struct {
	Sort     string
	Page     int
	PageSize int
}

func (r *PostRepo) List(params ListParams) ([]model.Post, error) {
	var posts []model.Post
	q := r.db.Preload("User").Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	})

	switch params.Sort {
	case "hot":
		q = q.Order("like_count DESC").Order("id DESC")
	default:
		q = q.Order("id DESC")
	}

	offset := (params.Page - 1) * params.PageSize
	err := q.Limit(params.PageSize).Offset(offset).Find(&posts).Error
	return posts, err
}

func (r *PostRepo) FindAuthorID(postID uint) (uint, error) {
	var post model.Post
	if err := r.db.Select("user_id").First(&post, postID).Error; err != nil {
		return 0, err
	}
	return post.UserID, nil
}

func (r *PostRepo) Like(userID, postID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Like{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.Like{UserID: userID, PostID: postID}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("like_count + 1")).Error
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *PostRepo) Unlike(userID, postID uint) (bool, error) {
	res := r.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Like{})
	if res.Error != nil {
		return false, res.Error
	}
	if res.RowsAffected == 0 {
		return false, nil
	}
	if err := r.db.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("like_count", gorm.Expr("GREATEST(like_count - 1, 0)")).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (r *PostRepo) LikedPostIDs(userID uint, postIDs []uint) (map[uint]bool, error) {
	set := make(map[uint]bool)
	if len(postIDs) == 0 {
		return set, nil
	}
	var likes []model.Like
	if err := r.db.Where("user_id = ? AND post_id IN ?", userID, postIDs).Find(&likes).Error; err != nil {
		return nil, err
	}
	for _, l := range likes {
		set[l.PostID] = true
	}
	return set, nil
}

func (r *PostRepo) Search(q string, page, pageSize int) ([]model.Post, error) {
	var posts []model.Post
	like := "%" + q + "%"
	subQuery := r.db.Model(&model.PostTopic{}).
		Select("post_topics.post_id").
		Joins("JOIN topics ON topics.id = post_topics.topic_id").
		Where("topics.name LIKE ?", like)
	err := r.db.Where("title LIKE ? OR content LIKE ? OR id IN (?)", like, like, subQuery).
		Preload("User").Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Order("id DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepo) Update(postID uint, title, content string) error {
	return r.db.Model(&model.Post{}).Where("id = ?", postID).Updates(map[string]interface{}{
		"title":   title,
		"content": content,
	}).Error
}

func (r *PostRepo) Delete(postID uint) error {
	return r.db.Delete(&model.Post{}, postID).Error
}

func (r *PostRepo) ListByUser(userID uint, page, pageSize int) ([]model.Post, error) {
	var posts []model.Post
	err := r.db.Where("user_id = ?", userID).
		Preload("User").Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Order("id DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepo) ListLikedByUser(userID uint, page, pageSize int) ([]model.Post, error) {
	var posts []model.Post
	err := r.db.Joins("JOIN likes ON likes.post_id = posts.id").
		Where("likes.user_id = ?", userID).
		Preload("User").Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Order("likes.created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepo) Favorite(userID, postID uint) (bool, error) {
	var count int64
	if err := r.db.Model(&model.Favorite{}).Where("user_id = ? AND post_id = ?", userID, postID).Count(&count).Error; err != nil {
		return false, err
	}
	if count > 0 {
		return false, nil
	}
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&model.Favorite{UserID: userID, PostID: postID}).Error; err != nil {
			return err
		}
		return tx.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("favorite_count", gorm.Expr("favorite_count + 1")).Error
	})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *PostRepo) Unfavorite(userID, postID uint) (bool, error) {
	res := r.db.Where("user_id = ? AND post_id = ?", userID, postID).Delete(&model.Favorite{})
	if res.Error != nil {
		return false, res.Error
	}
	if res.RowsAffected == 0 {
		return false, nil
	}
	if err := r.db.Model(&model.Post{}).Where("id = ?", postID).UpdateColumn("favorite_count", gorm.Expr("GREATEST(favorite_count - 1, 0)")).Error; err != nil {
		return false, err
	}
	return true, nil
}

func (r *PostRepo) ListFavoritedByUser(userID uint, page, pageSize int) ([]model.Post, error) {
	var posts []model.Post
	err := r.db.Joins("JOIN favorites ON favorites.post_id = posts.id").
		Where("favorites.user_id = ?", userID).
		Preload("User").Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Order("favorites.created_at DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&posts).Error
	return posts, err
}

func (r *PostRepo) FavoritedPostIDs(userID uint, postIDs []uint) (map[uint]bool, error) {
	set := make(map[uint]bool)
	if len(postIDs) == 0 {
		return set, nil
	}
	var favs []model.Favorite
	if err := r.db.Where("user_id = ? AND post_id IN ?", userID, postIDs).Find(&favs).Error; err != nil {
		return nil, err
	}
	for _, f := range favs {
		set[f.PostID] = true
	}
	return set, nil
}

func (r *PostRepo) ListByFollowees(followerID uint, page, pageSize int) ([]model.Post, error) {
	var posts []model.Post
	err := r.db.Joins("JOIN follows ON follows.followee_id = posts.user_id").
		Where("follows.follower_id = ?", followerID).
		Preload("User").Preload("Images", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		Order("posts.id DESC").
		Limit(pageSize).Offset((page - 1) * pageSize).
		Find(&posts).Error
	return posts, err
}

// InterestTopicIDs 返回用户互动过（赞/藏/评）的帖子所带的话题 ID，作为个性化兴趣信号。
func (r *PostRepo) InterestTopicIDs(userID uint) ([]uint, error) {
	var ids []uint
	err := r.db.Raw(`
		SELECT DISTINCT pt.topic_id
		FROM post_topics pt
		WHERE pt.post_id IN (
			SELECT post_id FROM likes WHERE user_id = ?
			UNION
			SELECT post_id FROM favorites WHERE user_id = ?
			UNION
			SELECT post_id FROM comments WHERE user_id = ?
		)
	`, userID, userID, userID).Scan(&ids).Error
	return ids, err
}

// RecommendList 推荐排序：热度(对数压缩+时间衰减) + 冷启动兜底 + 兴趣话题加权。
func (r *PostRepo) RecommendList(viewerID uint, page, pageSize int) ([]model.Post, error) {
	interestIDs, err := r.InterestTopicIDs(viewerID)
	if err != nil {
		interestIDs = nil // 降级：无个性化
	}

	base := "(LOG(like_count + 1) * 3 + LOG(favorite_count + 1) * 2 + LOG(comment_count + 1) * 2 + 2) / POW(TIMESTAMPDIFF(HOUR, created_at, NOW()) + 2, 0.8)"

	q := r.db.Preload("User").Preload("Images", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	})

	if len(interestIDs) > 0 {
		ids := make([]string, len(interestIDs))
		for i, id := range interestIDs {
			ids[i] = strconv.FormatUint(uint64(id), 10)
		}
		inClause := strings.Join(ids, ",")
		boost := "(SELECT COUNT(*) FROM post_topics WHERE post_topics.post_id = posts.id AND post_topics.topic_id IN (" + inClause + ")) * 5"
		q = q.Order("(" + base + " + " + boost + ") DESC").Order("id DESC")
	} else {
		q = q.Order(base + " DESC").Order("id DESC")
	}

	var posts []model.Post
	offset := (page - 1) * pageSize
	err = q.Limit(pageSize).Offset(offset).Find(&posts).Error
	return posts, err
}
