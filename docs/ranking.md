# 排行榜：用 Redis ZSet 做热度排名

推荐流、热门流和「热榜」页共用一套排名数据。MySQL 仍是事实来源，Redis 只是**可重建的物化排名**：榜单丢了可以从数据库回填，Redis 挂了接口会自动回退 SQL 排序。

## 1. 为什么需要 ZSet

首页推荐原先在 MySQL 里用 `ORDER BY` 现算热度分数，翻页时每页都要做一次全表排序，而且分数表达式无法走索引。把「笔记 → 热度」放进 ZSet 后：

- 排名的写入是 `ZADD` / 删除是 `ZREM`，复杂度 `O(log N)`；
- 读第 N 页是 `ZREVRANGE`，`O(log N + pageSize)`，不再扫表；
- 同分时按 member 字典序稳定排序，翻页不会重复或漏项。

## 2. 数据结构

| Key | 类型 | member | score |
| --- | --- | --- | --- |
| `bluebook:rank:post:heat` | ZSet | 12 位补零的笔记 ID（如 `000000000104`） | 热度 |
| `bluebook:rank:topic:heat` | ZSet | 话题 ID | 该话题下上榜笔记热度之和 |
| `bluebook:rank:creator:heat` | ZSet | 用户 ID | 该作者上榜笔记热度之和 |

member 补零是为了让同分时的字典序与数值序一致：`"2" > "10"`，但 `"000000000002" < "000000000010"`。

笔记榜最多保留 `PostBoardCap = 5000` 条，写入后执行 `ZREMRANGEBYRANK` 裁掉尾部，避免 ZSet 无限增长。

## 3. 热度公式

```text
signal = 3·ln(赞+1) + 2·ln(藏+1) + 2·ln(评+1) + 2
heat   = signal / (发布小时数 + 2)^0.8
```

实现分别在：

- Go：`backend/internal/pkg/rank/rank.go`
- SQL 降级：`backend/internal/repository/post_repo.go` 的 `hotOrderSQL`

两处常量保持一致，Redis 故障切换时排序基本不变。

## 4. 写入路径

| 触发 | 动作 |
| --- | --- |
| 发布笔记 | `TouchPost`：按新笔记的初始热度 `ZADD` |
| 点赞 / 取消点赞 | 事务提交后 `TouchPost`（重新读取三计数并重算） |
| 收藏 / 取消收藏 | 同上 |
| 发表评论 | `CommentService` 在评论计数 +1 后 `TouchPost` |
| 删除笔记 | `RemovePost`：从笔记榜 `ZREM` |
| 每 5 分钟 | `Refresh`：取榜单前 500 名，按当前时间重算并 `ZADD`，同时重建话题榜与创作者榜 |

热度公式里的「发布小时数」随时间增长，所以分数需要定期纠偏。`Refresh` 只重算头部 500 条，复杂度与总笔记数无关；被删除的笔记如果在榜单里但数据库查不到，会在这一步被 `ZREM`。

## 5. 读路径

```text
GET /api/v1/posts?sort=hot        → ZREVRANGE 分页 → MySQL 按 ID 批量取详情 → 还原顺序
GET /api/v1/posts?sort=recommend  → 取前 500 名候选 → 兴趣话题加权 → 内存重排分页
GET /api/v1/rankings?limit=10     → 三个榜单头部 → 批量补全笔记 / 话题 / 用户信息
```

`sort=latest` 和 `sort=following` 仍走 MySQL，排名只服务「热度」与「推荐」语义。

### 推荐流的个性化

1. 从 ZSet 取前 500 名作为候选池；
2. 用 `InterestTopicIDs` 找出用户互动过（赞 / 藏 / 评）的笔记所带话题；
3. 每条候选按命中的兴趣话题数加权：`score × (1 + 0.18 × min(命中数, 4))`；
4. 按加权分降序、ID 降序稳定排序后分页。

未登录或没有互动信号时，推荐流与热门流一致（冷启动）。

## 6. 降级与一致性

- **Redis 不可用 / 榜单为空**：`HotPostIDs` / `RecommendPostIDs` 返回 `ok=false`，`PostService.Feed` 回退到原来的 SQL 排序；`/rankings` 回退为「按互动总量 / 话题笔记数 / 粉丝数」的数据库榜单，响应里 `source` 为 `mysql`，此时 `heat` 为 0。
- **进程重启**：`RankingService.Start` 先检查 `ZCARD`，为 0 时从数据库回填（最近 30 天 + 历史互动最高的笔记），否则做一次纠偏重算；随后每 5 分钟执行一次。
- **互动后**：单篇笔记立即更新；话题榜 / 创作者榜最多延迟 5 分钟（由 `Refresh` 聚合）。
- **删除笔记**：立即从笔记榜移除；话题 / 创作者聚合值最多残留 5 分钟。

## 7. 运维排查

```bash
# 榜单前 10（member + score）
redis-cli ZREVRANGE bluebook:rank:post:heat 0 9 WITHSCORES

# 某篇笔记的排名与分数
redis-cli ZREVRANK bluebook:rank:post:heat 000000000104
redis-cli ZSCORE  bluebook:rank:post:heat 000000000104

# 榜单规模
redis-cli ZCARD bluebook:rank:post:heat

# 手工重建：删掉三个 key 后重启服务，会从 MySQL 回填
redis-cli DEL bluebook:rank:post:heat bluebook:rank:topic:heat bluebook:rank:creator:heat
```

相关代码：

- `backend/internal/pkg/rank/rank.go`：键名、member 编码、热度公式
- `backend/internal/repository/rank_repo.go`：ZSet 读写
- `backend/internal/service/ranking_service.go`：回填、纠偏、聚合、推荐重排
- `backend/internal/handler/post.go` 的 `Rankings`
