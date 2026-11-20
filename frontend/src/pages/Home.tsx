import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import AppShell from "../components/AppShell";
import FeedTabs from "../components/FeedTabs";
import MasonryFeed from "../components/MasonryFeed";
import type { PostItem } from "../components/PostCard";
import { EmptyState, ErrorState, FeedSkeleton, LoadingRow } from "../components/States";
import Icon from "../components/Icon";

const PAGE_SIZE = 20;

const SORT_HINT: Record<string, string> = {
  recommend: "根据你的兴趣推荐",
  latest: "刚刚发布的新鲜内容",
  hot: "大家都在看的热门笔记",
  following: "你关注的作者",
};

export default function Home() {
  const { token } = useAuth();
  const navigate = useNavigate();
  const [sort, setSort] = useState("recommend");
  const [items, setItems] = useState<PostItem[]>([]);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(true);
  const [error, setError] = useState("");
  const requestId = useRef(0);
  const sentinelRef = useRef<HTMLDivElement | null>(null);

  const fetchPage = useCallback(
    async (nextSort: string, nextPage: number, append: boolean) => {
      const myRequest = ++requestId.current;
      if (append) setLoadingMore(true);
      else setLoading(true);
      setError("");
      try {
        const result = await api.feed(nextSort, nextPage);
        if (myRequest !== requestId.current) return;
        const list: PostItem[] = result.items || [];
        setItems((prev) => (append ? [...prev, ...list] : list));
        setPage(nextPage);
        setHasMore(list.length === PAGE_SIZE);
      } catch (err: any) {
        if (myRequest !== requestId.current) return;
        setError(err?.message || "内容加载失败");
        if (!append) setItems([]);
      } finally {
        if (myRequest === requestId.current) {
          setLoading(false);
          setLoadingMore(false);
        }
      }
    },
    []
  );

  useEffect(() => {
    fetchPage(sort, 1, false);
  }, [sort, fetchPage]);

  const loadMore = useCallback(() => {
    if (loading || loadingMore || !hasMore) return;
    fetchPage(sort, page + 1, true);
  }, [fetchPage, hasMore, loading, loadingMore, page, sort]);

  useEffect(() => {
    const node = sentinelRef.current;
    if (!node) return;
    if (typeof IntersectionObserver === "undefined") return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries[0]?.isIntersecting) loadMore();
      },
      { rootMargin: "600px 0px" }
    );
    observer.observe(node);
    return () => observer.disconnect();
  }, [loadMore]);

  const topics = useMemo(() => {
    const set = new Set<string>();
    items.forEach((item) => (item.topics || []).forEach((topic) => set.add(topic)));
    return Array.from(set).slice(0, 12);
  }, [items]);

  // 话题 chips 优先用 Redis 话题榜；接口不可用时退回当前 feed 里出现的话题。
  const [rankTopics, setRankTopics] = useState<string[]>([]);
  useEffect(() => {
    let alive = true;
    api
      .rankings(12)
      .then((result: any) => {
        if (!alive) return;
        const names = (result?.topics || []).map((topic: any) => topic?.name).filter(Boolean);
        setRankTopics(names);
      })
      .catch(() => undefined);
    return () => {
      alive = false;
    };
  }, []);

  const topicChips = rankTopics.length > 0 ? rankTopics : topics;

  const needsLogin = sort === "following" && !token;

  return (
    <AppShell>
      <div className="feed-toolbar">
        <FeedTabs value={sort} onChange={setSort} />
        <span className="spacer" />
        <button type="button" className="btn btn-ghost btn-sm" onClick={() => navigate("/rankings")}>
          <Icon name="flame" size={14} />
          热榜
        </button>
        <span className="muted" style={{ fontSize: 12.5 }}>
          {SORT_HINT[sort]}
        </span>
      </div>

      {topicChips.length > 0 && !needsLogin ? (
        <div className="topic-row">
          <div className="chip-row">
            {topicChips.map((topic) => (
              <button
                key={topic}
                type="button"
                className="chip"
                onClick={() => navigate(`/search?q=${encodeURIComponent(topic)}`)}
              >
                <Icon name="search" size={13} />
                {topic}
              </button>
            ))}
          </div>
        </div>
      ) : null}

      <section className="feed-section">
        {loading ? (
          <FeedSkeleton count={10} />
        ) : error ? (
          <ErrorState message={error} onRetry={() => fetchPage(sort, 1, false)} />
        ) : needsLogin ? (
          <EmptyState
            icon="user"
            title="登录后查看关注内容"
            description="关注喜欢的作者，第一时间看到他们的新笔记"
            action={
              <button className="btn btn-primary btn-sm" onClick={() => navigate("/login?redirect=/")}>
                去登录
              </button>
            }
          />
        ) : items.length === 0 ? (
          <EmptyState
            icon="compass"
            title={sort === "following" ? "还没有关注任何人" : "这里还没有内容"}
            description={sort === "following" ? "去发现页逛逛，关注感兴趣的作者吧" : "成为第一个发布笔记的人"}
            action={
              <button className="btn btn-primary btn-sm" onClick={() => navigate(sort === "following" ? "/search" : "/create")}>
                {sort === "following" ? "去发现" : "发布笔记"}
              </button>
            }
          />
        ) : (
          <>
            <MasonryFeed items={items} />
            <div ref={sentinelRef} aria-hidden="true" />
            {loadingMore ? <LoadingRow text="正在加载更多笔记…" /> : null}
            {!hasMore && items.length > 0 ? (
              <div className="load-more">
                <span className="muted" style={{ fontSize: 12.5 }}>
                  已经到底啦 · 共 {items.length} 篇笔记
                </span>
              </div>
            ) : null}
          </>
        )}
      </section>
    </AppShell>
  );
}
