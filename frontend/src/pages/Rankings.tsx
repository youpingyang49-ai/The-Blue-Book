import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import AppShell from "../components/AppShell";
import Avatar from "../components/Avatar";
import Icon from "../components/Icon";
import type { PostItem } from "../components/PostCard";
import { EmptyState, ErrorState, PageLoading } from "../components/States";

interface RankedPost {
  rank: number;
  heat: number;
  post: PostItem;
}

interface RankedTopic {
  rank: number;
  heat: number;
  id: number;
  name: string;
  post_count: number;
}

interface RankedCreator {
  rank: number;
  heat: number;
  id: number;
  nickname: string;
  avatar_url: string;
  bio: string;
  follower_count: number;
}

interface RankingsData {
  posts: RankedPost[];
  topics: RankedTopic[];
  creators: RankedCreator[];
  source: string;
  generated_at: string;
}

const LIMIT = 10;

function heatText(heat?: number) {
  return heat && heat > 0 ? String(heat) : "—";
}

function updatedText(value?: string) {
  if (!value) return "";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toLocaleTimeString("zh-CN", { hour: "2-digit", minute: "2-digit" });
}

export default function Rankings() {
  const navigate = useNavigate();
  const [data, setData] = useState<RankingsData | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  const load = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result: RankingsData = await api.rankings(LIMIT);
      setData({
        posts: result?.posts || [],
        topics: result?.topics || [],
        creators: result?.creators || [],
        source: result?.source || "mysql",
        generated_at: result?.generated_at || "",
      });
    } catch (err: any) {
      setError(err?.message || "榜单加载失败");
      setData(null);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    load();
  }, [load]);

  const fromRedis = data?.source === "redis";
  const empty = !data || (data.posts.length === 0 && data.topics.length === 0 && data.creators.length === 0);

  return (
    <AppShell>
      <div className="page-head rank-head">
        <div>
          <h1>热榜</h1>
          <p>
            {fromRedis ? "Redis ZSet 实时排名 · 每 5 分钟重算热度" : "数据库排序（Redis 暂不可用，热度暂不展示）"}
          </p>
        </div>
        {data?.generated_at ? <span className="muted rank-updated">更新于 {updatedText(data.generated_at)}</span> : null}
      </div>

      {loading ? (
        <PageLoading text="正在读取榜单…" />
      ) : error ? (
        <ErrorState message={error} onRetry={load} />
      ) : empty ? (
        <EmptyState
          icon="flame"
          title="榜单还是空的"
          description="等第一波笔记和互动出现，这里就会热闹起来"
          action={
            <button className="btn btn-primary btn-sm" onClick={() => navigate("/create")}>
              发布笔记
            </button>
          }
        />
      ) : (
        <div className="rank-grid">
          <section className="rank-panel rank-posts">
            <header className="rank-panel-head">
              <h2>
                <Icon name="flame" size={16} /> 笔记热榜
              </h2>
              <span className="muted">热度</span>
            </header>
            {data!.posts.length === 0 ? (
              <p className="rank-empty muted">还没有上榜笔记</p>
            ) : (
              data!.posts.map((item) => {
                const title = (item.post.title || "").trim() || "无标题";
                return (
                  <button
                    key={item.post.id}
                    type="button"
                    className="rank-post"
                    onClick={() => navigate(`/post/${item.post.id}`)}
                  >
                    <span className={`rank-no ${item.rank <= 3 ? `rank-no-top rank-no-${item.rank}` : ""}`}>
                      {item.rank}
                    </span>
                    <img
                      className="rank-cover"
                      src={item.post.cover_image_url}
                      alt=""
                      loading="lazy"
                      onError={(event) => {
                        (event.currentTarget as HTMLImageElement).style.opacity = "0";
                      }}
                    />
                    <span className="rank-post-body">
                      <span className="rank-post-title clamp-2">{title}</span>
                      <span className="rank-post-meta">
                        {item.post.author?.nickname} · {item.post.like_count} 赞
                      </span>
                    </span>
                    <span className="rank-heat">
                      <span className="rank-heat-bar">
                        <i style={{ width: `${Math.max(item.heat, 2)}%` }} />
                      </span>
                      <span className="rank-heat-value">{heatText(item.heat)}</span>
                    </span>
                  </button>
                );
              })
            )}
          </section>

          <div className="rank-side">
            <section className="rank-panel">
              <header className="rank-panel-head">
                <h2>热门话题</h2>
              </header>
              {data!.topics.length === 0 ? (
                <p className="rank-empty muted">暂无话题</p>
              ) : (
                data!.topics.map((topic) => (
                  <button
                    key={topic.id}
                    type="button"
                    className="rank-topic"
                    onClick={() => navigate(`/search?q=${encodeURIComponent(topic.name)}`)}
                  >
                    <span className={`rank-no ${topic.rank <= 3 ? `rank-no-top rank-no-${topic.rank}` : ""}`}>
                      {topic.rank}
                    </span>
                    <span className="rank-topic-name">#{topic.name}</span>
                    <span className="rank-topic-count muted">{topic.post_count} 篇</span>
                    <span className="rank-heat-value">{heatText(topic.heat)}</span>
                  </button>
                ))
              )}
            </section>

            <section className="rank-panel">
              <header className="rank-panel-head">
                <h2>创作者榜</h2>
              </header>
              {data!.creators.length === 0 ? (
                <p className="rank-empty muted">暂无创作者</p>
              ) : (
                data!.creators.map((creator) => (
                  <button
                    key={creator.id}
                    type="button"
                    className="rank-creator"
                    onClick={() => navigate(`/user/${creator.id}`)}
                  >
                    <span className={`rank-no ${creator.rank <= 3 ? `rank-no-top rank-no-${creator.rank}` : ""}`}>
                      {creator.rank}
                    </span>
                    <Avatar src={creator.avatar_url} name={creator.nickname} size="xs" />
                    <span className="rank-creator-body">
                      <span className="rank-creator-name">{creator.nickname}</span>
                      <span className="rank-post-meta">{creator.follower_count} 粉丝</span>
                    </span>
                    <span className="rank-heat-value">{heatText(creator.heat)}</span>
                  </button>
                ))
              )}
            </section>
          </div>
        </div>
      )}
    </AppShell>
  );
}
