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

  return (
    <AppShell>
      <div className="page-head">
        <h1>{keyword ? `“${keyword}” 的搜索结果` : "发现"}</h1>
        <p>
          {keyword
            ? loading
              ? "正在搜索…"
              : `找到 ${items.length} 篇相关笔记`
            : "输入关键词，搜索你感兴趣的笔记、内容和话题"}
        </p>
      </div>

      {!keyword ? (
        <div style={{ paddingTop: 12 }}>
          <div className="chip-row" style={{ flexWrap: "wrap" }}>
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
        </div>
      ) : loading ? (
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
    </AppShell>
  );
}
