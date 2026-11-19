package repository

import (
	"errors"

	"gorm.io/gorm"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/model"
)

type TopicRepo struct {
	db *gorm.DB
}

func NewTopicRepo(db *gorm.DB) *TopicRepo {
	return &TopicRepo{db: db}
}

func (r *TopicRepo) FindOrCreateByName(name string) (*model.Topic, error) {
	var topic model.Topic
	err := r.db.Where("name = ?", name).First(&topic).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		topic = model.Topic{Name: name}
		if err := r.db.Create(&topic).Error; err != nil {
			return nil, err
		}
		return &topic, nil
	}
	if err != nil {
		return nil, err
	}
	return &topic, nil
}

func (r *TopicRepo) Link(postID, topicID uint) error {
	return r.db.Create(&model.PostTopic{PostID: postID, TopicID: topicID}).Error
}

func (r *TopicRepo) IncrementCount(topicID uint) error {
	return r.db.Model(&model.Topic{}).Where("id = ?", topicID).UpdateColumn("post_count", gorm.Expr("post_count + 1")).Error
}

// ListNamesByPostIDs 批量取帖子的话题名，避免列表接口 N+1 查询。
func (r *TopicRepo) ListNamesByPostIDs(postIDs []uint) (map[uint][]string, error) {
	out := make(map[uint][]string, len(postIDs))
	if len(postIDs) == 0 {
		return out, nil
	}
	type row struct {
		PostID uint
		Name   string
	}
	var rows []row
	err := r.db.Table("post_topics").
		Select("post_topics.post_id AS post_id, topics.name AS name").
		Joins("JOIN topics ON topics.id = post_topics.topic_id").
		Where("post_topics.post_id IN ?", postIDs).
		Order("topics.id ASC").
		Scan(&rows).Error
	if err != nil {
		return out, err
	}
	for _, rw := range rows {
		out[rw.PostID] = append(out[rw.PostID], rw.Name)
	}
	return out, nil
}

func (r *TopicRepo) ListByPost(postID uint) ([]model.Topic, error) {
	var topics []model.Topic
	err := r.db.Joins("JOIN post_topics ON post_topics.topic_id = topics.id").
		Where("post_topics.post_id = ?", postID).
		Find(&topics).Error
	return topics, err
}
