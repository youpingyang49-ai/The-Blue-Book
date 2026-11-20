// Package rank 定义排行榜的热度算法、Redis 键名与成员编码。
//
// 榜单本身存在 Redis ZSet 中，MySQL 始终是事实来源：
// Redis 提供 O(log N) 的排名读写，进程重启或 Redis 被清空后由数据库回填。
package rank

import (
	"fmt"
	"math"
	"strconv"
	"time"
)

const (
	// KeyPostHeat 笔记热度榜，member = 12 位补零的笔记 ID，score = 热度。
	KeyPostHeat = "bluebook:rank:post:heat"
	// KeyTopicHeat 话题热度榜，member = 话题 ID，score = 该话题下上榜笔记热度之和。
	KeyTopicHeat = "bluebook:rank:topic:heat"
	// KeyCreatorHeat 创作者榜，member = 用户 ID，score = 其上榜笔记热度之和。
	KeyCreatorHeat = "bluebook:rank:creator:heat"

	// PostBoardCap 热度榜最多保留的笔记数，防止 ZSet 无限增长。
	PostBoardCap = 5000
)

const (
	// 热度权重与时间衰减系数，与旧版 SQL 推荐公式保持一致：
	// (3*ln(赞+1) + 2*ln(藏+1) + 2*ln(评+1) + 2) / (发布小时数 + 2)^0.8
	likeWeight     = 3.0
	favoriteWeight = 2.0
	commentWeight  = 2.0
	baseSignal     = 2.0
	gravity        = 0.8
)

// Metrics 是计算热度所需的最小字段集。
type Metrics struct {
	Likes     int64
	Favorites int64
	Comments  int64
	CreatedAt time.Time
}

// Heat 计算笔记热度：互动量取对数压缩，再按发布时长做幂律衰减。
func Heat(m Metrics, now time.Time) float64 {
	ageHours := now.Sub(m.CreatedAt).Hours()
	if ageHours < 0 {
		ageHours = 0
	}
	signal := likeWeight*math.Log1p(float64(m.Likes)) +
		favoriteWeight*math.Log1p(float64(m.Favorites)) +
		commentWeight*math.Log1p(float64(m.Comments)) +
		baseSignal
	return signal / math.Pow(ageHours+2, gravity)
}

// PostMember 把笔记 ID 编码成 ZSet member。
// 补零保证同分时按 ID 字典序排序与数值序一致，翻页不会乱序。
func PostMember(id uint) string {
	return fmt.Sprintf("%012d", id)
}

// ParseMember 把 ZSet member 还原成 ID。
func ParseMember(member string) (uint, bool) {
	id, err := strconv.ParseUint(member, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}
	return uint(id), true
}

// Normalize 把原始热度换算成 0-100 的展示分，榜首恒为 100。
func Normalize(score, top float64) int {
	if score <= 0 || top <= 0 {
		return 0
	}
	value := int(math.Round(score / top * 100))
	if value < 1 {
		return 1
	}
	if value > 100 {
		return 100
	}
	return value
}
