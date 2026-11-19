import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { api } from "../api/client";
import AppShell from "../components/AppShell";
import MasonryFeed from "../components/MasonryFeed";
import type { PostItem } from "../components/PostCard";
import { EmptyState, ErrorState, FeedSkeleton, LoadingRow } from "../components/States";

const PAGE_SIZE = 20;
const SUGGESTIONS = ["美食", "旅行", "穿搭", "家居", "摄影", "读书", "健身", "咖啡", "手作", "城市漫步"];

export default function Search() {
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const keyword = (params.get("q") || "").trim();

  const [items, setItems] = useState<PostItem[]>([]);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(false);
  const [error, setError] = useState("");

  // 「发现」默认内容（无关键词时展示热门笔记）
  const [hotItems, setHotItems] = useState<PostItem[]>([]);
  const [hotLoading, setHotLoading] = useState(true);
  const [hotError, setHotError] = useState("");
  const requestId = useRef(0);

  const run = useCallback(async (query: string, nextPage: number, append: boolean) => {
    const myRequest = ++requestId.current;
    if (!query) {
      setItems([]);
      setHasMore(false);
      return;
    }
    if (append) setLoadingMore(true);
    else setLoading(true);
    setError("");
    try {
      const result = await api.search(query, nextPage);
      if (myRequest !== requestId.current) return;
      const list: PostItem[] = result.items || [];
      setItems((prev) => (append ? [...prev, ...list] : list));
      setPage(nextPage);
      setHasMore(list.length === PAGE_SIZE);
    } catch (err: any) {
      if (myRequest !== requestId.current) return;
      setError(err?.message || "搜索失败");
      if (!append) setItems([]);
    } finally {
      if (myRequest === requestId.current) {
        setLoading(false);
        setLoadingMore(false);
      }
    }
  }, []);

  useEffect(() => {
    run(keyword, 1, false);
  }, [keyword, run]);

  const loadHot = useCallback(async () => {
    setHotLoading(true);
    setHotError("");
    try {
      const result = await api.feed("hot", 1);
      setHotItems(result.items || []);
    } catch (err: any) {
      setHotError(err?.message || "内容加载失败");
    } finally {
      setHotLoading(false);
    }
  }, []);

  useEffect(() => {
    if (!keyword) loadHot();
  }, [keyword, loadHot]);

  return (
    <AppShell>
      {keyword ? (
        <>
          <div className="page-head">
            <h1>“{keyword}” 的搜索结果</h1>
            <p>{loading ? "正在搜索…" : `找到 ${items.length} 篇相关笔记`}</p>
          </div>

          {loading ? (
            <FeedSkeleton count={8} />
          ) : error ? (
            <ErrorState message={error} onRetry={() => run(keyword, 1, false)} />
          ) : items.length === 0 ? (
            <EmptyState
              icon="search"
              title="没有找到相关笔记"
              description="换个关键词试试，或者去首页看看大家最近在分享什么"
              action={
                <button className="btn btn-ghost btn-sm" onClick={() => navigate("/")}>
                  回到首页
                </button>
              }
            />
          ) : (
            <>
              <MasonryFeed items={items} />
              {loadingMore ? <LoadingRow text="正在加载更多…" /> : null}
              {hasMore ? (
                <div className="load-more">
                  <button
                    type="button"
                    className="btn btn-outline btn-sm"
                    onClick={() => run(keyword, page + 1, true)}
                    disabled={loadingMore}
                  >
                    加载更多
                  </button>
                </div>
              ) : (
                <div className="load-more">
                  <span className="muted" style={{ fontSize: 12.5 }}>
                    已展示全部 {items.length} 篇笔记
                  </span>
                </div>
              )}
            </>
          )}
        </>
      ) : (
        <>
          <div className="page-head">
            <h1>发现</h1>
            <p>搜索你感兴趣的内容，或者直接看看大家最近在分享什么</p>
          </div>

          <div className="chip-row" style={{ flexWrap: "wrap", paddingTop: 10, paddingBottom: 16 }}>
            {SUGGESTIONS.map((suggestion) => (
              <button
                key={suggestion}
                type="button"
                className="chip"
                onClick={() => navigate(`/search?q=${encodeURIComponent(suggestion)}`)}
              >
                {suggestion}
              </button>
            ))}
          </div>

          <div className="discover-head">
            <h2>热门笔记</h2>
            <span className="muted">按热度与新鲜度排序</span>
          </div>

          {hotLoading ? (
            <FeedSkeleton count={10} />
          ) : hotError ? (
            <ErrorState message={hotError} onRetry={loadHot} />
          ) : hotItems.length === 0 ? (
            <EmptyState icon="compass" title="暂时没有内容" description="成为第一个发布笔记的人吧" />
          ) : (
            <MasonryFeed items={hotItems} />
          )}
        </>
      )}
    </AppShell>
  );
}
