package service

import (
	"context"
	"log/slog"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/pkg/rank"
	"github.com/youpingyang49-ai/The-Blue-Book/backend/internal/repository"
)

const (
	// rankRefreshInterval 榜单重算周期。热度公式与时间有关，定时重算保证排序不会长期漂移。
	rankRefreshInterval = 5 * time.Minute
	// rankRefreshSize 每次重算的榜单头部规模（也是话题榜 / 创作者榜的聚合范围）。
	rankRefreshSize = 500
	// 启动回填候选：最近 30 天 + 历史互动最高的笔记。
	rankRecentDays       = 30
	rankRecentCandidates = 2000
	rankSignalCandidates = 1000
	// recommendPoolSize 推荐流参与个性化重排的候选池。
	recommendPoolSize = 500
	// 每条命中的兴趣话题提升 18% 热度，最多累计 4 个话题。
	interestBoostPerTopic = 0.18
	interestBoostMax      = 4
)

// RankingService 维护 Redis 中的三个 ZSet 榜单，并提供 feed 读取入口。
// MySQL 是事实来源，Redis 只是可重建的物化排名；Redis 故障时调用方回退到 SQL。
type RankingService struct {
	postRepo  *repository.PostRepo
	topicRepo *repository.TopicRepo
	rankRepo  *repository.RankRepo

	mu  sync.Mutex // 串行化话题榜 / 创作者榜的临时键替换
	now func() time.Time
}

func NewRankingService(postRepo *repository.PostRepo, topicRepo *repository.TopicRepo, rankRepo *repository.RankRepo) *RankingService {
	return &RankingService{
		postRepo:  postRepo,
		topicRepo: topicRepo,
		rankRepo:  rankRepo,
		now:       time.Now,
	}
}

// RankedEntry 榜单中的一项，Score 为 Redis 中的原始热度。
type RankedEntry struct {
	ID    uint
	Score float64
	Rank  int
}

// BoardSnapshot 三个榜单的原始读数。
type BoardSnapshot struct {
	Posts       []RankedEntry
	Topics      []RankedEntry
	Creators    []RankedEntry
	GeneratedAt time.Time
	Source      string
}

// Start 后台回填并定时重算榜单；ctx 结束时退出。
func (s *RankingService) Start(ctx context.Context) {
	if err := s.Backfill(ctx); err != nil {
		slog.Warn("ranking backfill failed", "error", err)
	}
	ticker := time.NewTicker(rankRefreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Refresh(ctx); err != nil {
				slog.Warn("ranking refresh failed", "error", err)
			}
		}
	}
}

// Backfill 在榜单为空时用数据库回填，否则只做一次纠偏重算。
func (s *RankingService) Backfill(ctx context.Context) error {
	card, err := s.rankRepo.Card(ctx, rank.KeyPostHeat)
	if err != nil {
		return err
	}
	if card > 0 {
		return s.Refresh(ctx)
	}
	candidates, err := s.postRepo.RankCandidates(rankRecentDays, rankRecentCandidates, rankSignalCandidates)
	if err != nil {
		return err
	}
	scores := make(map[uint]float64, len(candidates))
	now := s.now()
	for _, row := range candidates {
		scores[row.ID] = rank.Heat(rank.Metrics{
			Likes:     row.LikeCount,
			Favorites: row.FavoriteCount,
			Comments:  row.CommentCount,
			CreatedAt: row.CreatedAt,
		}, now)
	}
	if err := s.writePostScores(ctx, scores); err != nil {
		return err
	}
	return s.rebuildBoards(ctx, candidates, scores)
}

// Refresh 重算榜单头部热度，并据此重建话题榜与创作者榜。
func (s *RankingService) Refresh(ctx context.Context) error {
	entries, err := s.rankRepo.Top(ctx, rank.KeyPostHeat, 0, rankRefreshSize)
	if err != nil {
		return err
	}
	ids := parseEntries(entries)
	if len(ids) == 0 {
		return s.Backfill(ctx)
	}

	rows, err := s.postRepo.RankSnapshots(ids)
	if err != nil {
		return err
	}
	found := make(map[uint]struct{}, len(rows))
	scores := make(map[uint]float64, len(rows))
	now := s.now()
	for _, row := range rows {
		found[row.ID] = struct{}{}
		scores[row.ID] = rank.Heat(rank.Metrics{
			Likes:     row.LikeCount,
			Favorites: row.FavoriteCount,
			Comments:  row.CommentCount,
			CreatedAt: row.CreatedAt,
		}, now)
	}
	// 数据库里已经不存在的笔记（被删除）要移出榜单。
	for _, id := range ids {
		if _, ok := found[id]; !ok {
			if err := s.rankRepo.Remove(ctx, []string{rank.KeyPostHeat}, rank.PostMember(id)); err != nil {
				return err
			}
		}
	}
	if err := s.writePostScores(ctx, scores); err != nil {
		return err
	}
	return s.rebuildBoards(ctx, rows, scores)
}

// TouchPost 在点赞 / 收藏 / 评论后增量更新单篇笔记热度。
func (s *RankingService) TouchPost(ctx context.Context, postID uint) error {
	rows, err := s.postRepo.RankSnapshots([]uint{postID})
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return s.RemovePost(ctx, postID)
	}
	row := rows[0]
	score := rank.Heat(rank.Metrics{
		Likes:     row.LikeCount,
		Favorites: row.FavoriteCount,
		Comments:  row.CommentCount,
		CreatedAt: row.CreatedAt,
	}, s.now())
	return s.rankRepo.SetScores(ctx, rank.KeyPostHeat, map[string]float64{
		rank.PostMember(row.ID): score,
	})
}

// RemovePost 从笔记榜移除（笔记删除或失去全部互动时）。
func (s *RankingService) RemovePost(ctx context.Context, postID uint) error {
	return s.rankRepo.Remove(ctx, []string{rank.KeyPostHeat}, rank.PostMember(postID))
}

// HotPostIDs 读取热度榜分页。ok=false 表示 Redis 不可用或榜单为空，调用方应回退数据库。
func (s *RankingService) HotPostIDs(ctx context.Context, offset, limit int) ([]uint, bool) {
	if limit <= 0 {
		return nil, false
	}
	entries, err := s.rankRepo.Top(ctx, rank.KeyPostHeat, int64(offset), int64(limit))
	if err != nil {
		slog.Warn("ranking read failed", "key", rank.KeyPostHeat, "error", err)
		return nil, false
	}
	ids := parseEntries(entries)
	if len(ids) == 0 {
		if offset > 0 {
			// 已经翻到榜单末尾：返回空页而不是回退 SQL（避免排序口径突变）。
			return []uint{}, true
		}
		return nil, false
	}
	return ids, true
}

// RecommendPostIDs 在热度榜候选池上叠加兴趣话题权重后分页。
// 未登录 / 无兴趣信号时与热门榜一致（冷启动）。
func (s *RankingService) RecommendPostIDs(ctx context.Context, viewerID uint, offset, limit int) ([]uint, bool) {
	if limit <= 0 {
		return nil, false
	}
	if offset < 0 {
		offset = 0
	}
	entries, err := s.rankRepo.Top(ctx, rank.KeyPostHeat, 0, recommendPoolSize)
	if err != nil {
		slog.Warn("ranking read failed", "key", rank.KeyPostHeat, "error", err)
		return nil, false
	}
	candidates := rankCandidates(entries)
	if len(candidates) == 0 {
		return nil, false
	}

	if viewerID != 0 {
		interest, err := s.postRepo.InterestTopicIDs(viewerID)
		if err != nil {
			interest = nil
		}
		if len(interest) > 0 {
			ids := make([]uint, len(candidates))
			for i, c := range candidates {
				ids[i] = c.id
			}
			matches, err := s.topicRepo.TopicMatchCounts(ids, interest)
			if err == nil {
				for i := range candidates {
					hit := matches[candidates[i].id]
					if hit > interestBoostMax {
						hit = interestBoostMax
					}
					candidates[i].score *= 1 + interestBoostPerTopic*float64(hit)
				}
				sort.SliceStable(candidates, func(i, j int) bool {
					if candidates[i].score != candidates[j].score {
						return candidates[i].score > candidates[j].score
					}
					return candidates[i].id > candidates[j].id
				})
			}
		}
	}

	if offset >= len(candidates) {
		return []uint{}, true
	}
	end := offset + limit
	if end > len(candidates) {
		end = len(candidates)
	}
	ids := make([]uint, 0, end-offset)
	for _, c := range candidates[offset:end] {
		ids = append(ids, c.id)
	}
	return ids, true
}

// Snapshot 读取三个榜单的头部，供 /rankings 接口使用。
func (s *RankingService) Snapshot(ctx context.Context, posts, topics, creators int) (*BoardSnapshot, error) {
	snap := &BoardSnapshot{
		GeneratedAt: s.now(),
		Source:      "redis",
	}
	postEntries, err := s.rankRepo.Top(ctx, rank.KeyPostHeat, 0, int64(posts))
	if err != nil {
		return nil, err
	}
	snap.Posts = toEntries(postEntries)

	topicEntries, err := s.rankRepo.Top(ctx, rank.KeyTopicHeat, 0, int64(topics))
	if err != nil {
		return snap, nil // 话题榜缺失不影响笔记榜
	}
	snap.Topics = toEntries(topicEntries)

	creatorEntries, err := s.rankRepo.Top(ctx, rank.KeyCreatorHeat, 0, int64(creators))
	if err != nil {
		return snap, nil
	}
	snap.Creators = toEntries(creatorEntries)
	return snap, nil
}

func (s *RankingService) writePostScores(ctx context.Context, scores map[uint]float64) error {
	if len(scores) == 0 {
		return nil
	}
	members := make(map[string]float64, len(scores))
	for id, score := range scores {
		members[rank.PostMember(id)] = score
	}
	if err := s.rankRepo.SetScores(ctx, rank.KeyPostHeat, members); err != nil {
		return err
	}
	return s.rankRepo.Trim(ctx, rank.KeyPostHeat, rank.PostBoardCap)
}

// rebuildBoards 用上榜笔记的热度重新聚合话题榜与创作者榜。
// 先写临时键再 RENAME，读侧不会看到半成品。
func (s *RankingService) rebuildBoards(ctx context.Context, rows []repository.RankSnapshot, scores map[uint]float64) error {
	if len(rows) == 0 {
		s.mu.Lock()
		defer s.mu.Unlock()
		if err := s.rankRepo.Replace(ctx, rank.KeyTopicHeat+":tmp", rank.KeyTopicHeat); err != nil {
			return err
		}
		return s.rankRepo.Replace(ctx, rank.KeyCreatorHeat+":tmp", rank.KeyCreatorHeat)
	}

	ids := make([]uint, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.ID)
	}
	links, err := s.topicRepo.PostTopicIDs(ids)
	if err != nil {
		return err
	}

	topicScores := make(map[string]float64)
	creatorScores := make(map[string]float64)
	for _, row := range rows {
		score, ok := scores[row.ID]
		if !ok || score <= 0 {
			continue
		}
		for _, topicID := range links[row.ID] {
			topicScores[strconv.FormatUint(uint64(topicID), 10)] += score
		}
		creatorScores[strconv.FormatUint(uint64(row.UserID), 10)] += score
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.rankRepo.SetScores(ctx, rank.KeyTopicHeat+":tmp", topicScores); err != nil {
		return err
	}
	if err := s.rankRepo.Replace(ctx, rank.KeyTopicHeat+":tmp", rank.KeyTopicHeat); err != nil {
		return err
	}
	if err := s.rankRepo.SetScores(ctx, rank.KeyCreatorHeat+":tmp", creatorScores); err != nil {
		return err
	}
	return s.rankRepo.Replace(ctx, rank.KeyCreatorHeat+":tmp", rank.KeyCreatorHeat)
}

type scoredID struct {
	id    uint
	score float64
}

func rankCandidates(entries []redis.Z) []scoredID {
	candidates := make([]scoredID, 0, len(entries))
	for _, entry := range entries {
		member, ok := entry.Member.(string)
		if !ok {
			continue
		}
		id, ok := rank.ParseMember(member)
		if !ok {
			continue
		}
		candidates = append(candidates, scoredID{id: id, score: entry.Score})
	}
	return candidates
}

func parseEntries(entries []redis.Z) []uint {
	ids := make([]uint, 0, len(entries))
	for _, entry := range entries {
		member, ok := entry.Member.(string)
		if !ok {
			continue
		}
		if id, ok := rank.ParseMember(member); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

func toEntries(entries []redis.Z) []RankedEntry {
	out := make([]RankedEntry, 0, len(entries))
	for _, entry := range entries {
		member, ok := entry.Member.(string)
		if !ok {
			continue
		}
		id, ok := rank.ParseMember(member)
		if !ok {
			continue
		}
		out = append(out, RankedEntry{ID: id, Score: entry.Score, Rank: len(out) + 1})
	}
	return out
}
