import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import AppShell from "../components/AppShell";
import Avatar from "../components/Avatar";
import MasonryFeed from "../components/MasonryFeed";
import type { PostItem } from "../components/PostCard";
import { EmptyState, ErrorState, FeedSkeleton, LoadingRow } from "../components/States";
import { useToast } from "../components/Toast";

const PAGE_SIZE = 20;
const SUGGESTIONS = ["美食", "旅行", "穿搭", "家居", "摄影", "读书", "健身", "咖啡", "手作", "城市漫步"];

interface UserItem {
  id: number;
  nickname: string;
  avatar_url: string;
  bio: string;
  follower_count: number;
  is_following: boolean;
}

export default function Search() {
  const [params, setParams] = useSearchParams();
  const navigate = useNavigate();
  const { token, user } = useAuth();
  const toast = useToast();

  const keyword = (params.get("q") || "").trim();
  const tab: "notes" | "users" = params.get("type") === "users" ? "users" : "notes";

  // 笔记结果
  const [items, setItems] = useState<PostItem[]>([]);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [hasMore, setHasMore] = useState(false);
  const [error, setError] = useState("");
  const notesRequestId = useRef(0);

  // 用户结果
  const [users, setUsers] = useState<UserItem[]>([]);
  const [userTotal, setUserTotal] = useState(0);
  const [userPage, setUserPage] = useState(1);
  const [userLoading, setUserLoading] = useState(false);
  const [userLoadingMore, setUserLoadingMore] = useState(false);
  const [userError, setUserError] = useState("");
  const [pendingFollowId, setPendingFollowId] = useState<number | null>(null);
  const usersRequestId = useRef(0);

  // 无关键词时展示的热门笔记
  const [hotItems, setHotItems] = useState<PostItem[]>([]);
  const [hotLoading, setHotLoading] = useState(true);
  const [hotError, setHotError] = useState("");

  const runNotes = useCallback(async (query: string, nextPage: number, append: boolean) => {
    const myRequest = ++notesRequestId.current;
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
      if (myRequest !== notesRequestId.current) return;
      const list: PostItem[] = result.items || [];
      setItems((prev) => (append ? [...prev, ...list] : list));
      setPage(nextPage);
      setHasMore(list.length === PAGE_SIZE);
    } catch (err: any) {
      if (myRequest !== notesRequestId.current) return;
      setError(err?.message || "搜索失败");
      if (!append) setItems([]);
    } finally {
      if (myRequest === notesRequestId.current) {
        setLoading(false);
        setLoadingMore(false);
      }
    }
  }, []);

  const runUsers = useCallback(async (query: string, nextPage: number, append: boolean) => {
    const myRequest = ++usersRequestId.current;
    if (!query) {
      setUsers([]);
      setUserTotal(0);
      return;
    }
    if (append) setUserLoadingMore(true);
    else setUserLoading(true);
    setUserError("");
    try {
      const result = await api.searchUsers(query, nextPage);
      if (myRequest !== usersRequestId.current) return;
      const list: UserItem[] = result.items || [];
      setUsers((prev) => (append ? [...prev, ...list] : list));
      setUserTotal(Number(result.total) || 0);
      setUserPage(nextPage);
    } catch (err: any) {
      if (myRequest !== usersRequestId.current) return;
      setUserError(err?.message || "搜索失败");
      if (!append) setUsers([]);
    } finally {
      if (myRequest === usersRequestId.current) {
        setUserLoading(false);
        setUserLoadingMore(false);
      }
    }
  }, []);

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
    if (!keyword) {
      loadHot();
      return;
    }
    if (tab === "notes") runNotes(keyword, 1, false);
    else runUsers(keyword, 1, false);
  }, [keyword, tab, runNotes, runUsers, loadHot]);

  const switchTab = (next: "notes" | "users") => {
    const nextParams = new URLSearchParams(params);
    if (next === "users") nextParams.set("type", "users");
    else nextParams.delete("type");
    setParams(nextParams, { replace: true });
  };

  const toggleFollow = async (target: UserItem) => {
    if (!token) {
      toast("登录后即可关注", "error");
      navigate(`/login?redirect=${encodeURIComponent(window.location.pathname + window.location.search)}`);
      return;
    }
    if (pendingFollowId) return;
    const next = !target.is_following;
    setPendingFollowId(target.id);
    setUsers((prev) =>
      prev.map((item) =>
        item.id === target.id
          ? { ...item, is_following: next, follower_count: Math.max(0, item.follower_count + (next ? 1 : -1)) }
          : item
      )
    );
    try {
      if (next) await api.follow(target.id);
      else await api.unfollow(target.id);
    } catch (err: any) {
      setUsers((prev) =>
        prev.map((item) =>
          item.id === target.id
            ? { ...item, is_following: !next, follower_count: Math.max(0, item.follower_count + (next ? -1 : 1)) }
            : item
        )
      );
      toast(err?.message || "操作失败", "error");
    } finally {
      setPendingFollowId(null);
    }
  };


  return (
    <AppShell>
      {keyword ? (
        <>
          <div className="page-head">
            <h1>“{keyword}” 的搜索结果</h1>
            <p>
              {tab === "notes"
                ? loading
                  ? "正在搜索笔记…"
                  : `找到 ${items.length} 篇相关笔记`
                : userLoading
                  ? "正在搜索用户…"
                  : `找到 ${userTotal} 位相关用户`}
            </p>
          </div>

          <div className="search-tabs">
            <div className="tabs">
              <button
                type="button"
                className={`tab ${tab === "notes" ? "active" : ""}`}
                onClick={() => switchTab("notes")}
              >
                笔记
              </button>
              <button
                type="button"
                className={`tab ${tab === "users" ? "active" : ""}`}
                onClick={() => switchTab("users")}
              >
                用户
              </button>
            </div>
          </div>

          {tab === "notes" ? (
            loading ? (
              <FeedSkeleton count={8} />
            ) : error ? (
              <ErrorState message={error} onRetry={() => runNotes(keyword, 1, false)} />
            ) : items.length === 0 ? (
              <EmptyState
                icon="search"
                title="没有找到相关笔记"
                description="换个关键词试试，或者切换到「用户」看看"
                action={
                  <button className="btn btn-ghost btn-sm" onClick={() => switchTab("users")}>
                    搜索用户
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
                      onClick={() => runNotes(keyword, page + 1, true)}
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
            )
          ) : userLoading ? (
            <LoadingRow text="正在搜索用户…" />
          ) : userError ? (
            <ErrorState message={userError} onRetry={() => runUsers(keyword, 1, false)} />
          ) : users.length === 0 ? (
            <EmptyState
              icon="user"
              title="没有找到相关用户"
              description="换个昵称或简介里的关键词试试"
              action={
                <button className="btn btn-ghost btn-sm" onClick={() => switchTab("notes")}>
                  搜索笔记
                </button>
              }
            />
          ) : (
            <>
              <div className="user-results">
                {users.map((item) => (
                  <div key={item.id} className="user-result" onClick={() => navigate(`/user/${item.id}`)}>
                    <Avatar src={item.avatar_url} name={item.nickname} size="lg" />
                    <div className="info">
                      <div className="name">{item.nickname}</div>
                      <div className="bio">{item.bio || "这个人很懒，什么都没有写"}</div>
                      <div className="meta">{item.follower_count} 粉丝</div>
                    </div>
                    {user?.id !== item.id ? (
                      <button
                        type="button"
                        className={item.is_following ? "btn btn-ghost btn-sm" : "btn btn-soft btn-sm"}
                        disabled={pendingFollowId === item.id}
                        onClick={(event) => {
                          event.stopPropagation();
                          toggleFollow(item);
                        }}
                      >
                        {item.is_following ? "已关注" : "+ 关注"}
                      </button>
                    ) : null}
                  </div>
                ))}
              </div>
              {userLoadingMore ? <LoadingRow text="正在加载更多…" /> : null}
              {users.length < userTotal ? (
                <div className="load-more">
                  <button
                    type="button"
                    className="btn btn-outline btn-sm"
                    onClick={() => runUsers(keyword, userPage + 1, true)}
                    disabled={userLoadingMore}
                  >
                    加载更多
                  </button>
                </div>
              ) : (
                <div className="load-more">
                  <span className="muted" style={{ fontSize: 12.5 }}>
                    已展示全部 {userTotal} 位用户
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
