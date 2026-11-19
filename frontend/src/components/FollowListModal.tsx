import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import Avatar from "./Avatar";
import Icon from "./Icon";
import { EmptyState, ErrorState, LoadingRow } from "./States";
import { useToast } from "./Toast";

interface FollowUser {
  id: number;
  nickname: string;
  avatar_url: string;
  bio: string;
  is_following: boolean;
}

interface Props {
  userId: number;
  initialTab: "following" | "followers";
  onClose: () => void;
  onChanged?: () => void;
}

export default function FollowListModal({ userId, initialTab, onClose, onChanged }: Props) {
  const navigate = useNavigate();
  const toast = useToast();
  const [tab, setTab] = useState<"following" | "followers">(initialTab);
  const [items, setItems] = useState<FollowUser[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState("");
  const [pendingId, setPendingId] = useState<number | null>(null);

  const load = useCallback(
    async (which: "following" | "followers", nextPage: number, append: boolean) => {
      if (append) setLoadingMore(true);
      else setLoading(true);
      setError("");
      try {
        const result =
          which === "following" ? await api.following(userId, nextPage) : await api.followers(userId, nextPage);
        const list: FollowUser[] = result.items || [];
        setItems((prev) => (append ? [...prev, ...list] : list));
        setTotal(Number(result.total) || 0);
        setPage(nextPage);
      } catch (err: any) {
        setError(err?.message || "加载失败");
        if (!append) setItems([]);
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [userId]
  );

  useEffect(() => {
    setItems([]);
    setTotal(0);
    load(tab, 1, false);
  }, [tab, load]);

  const toggleFollow = async (user: FollowUser) => {
    if (pendingId) return;
    const next = !user.is_following;
    setPendingId(user.id);
    setItems((prev) => prev.map((item) => (item.id === user.id ? { ...item, is_following: next } : item)));
    try {
      if (next) await api.follow(user.id);
      else await api.unfollow(user.id);
      onChanged?.();
    } catch (err: any) {
      setItems((prev) => prev.map((item) => (item.id === user.id ? { ...item, is_following: !next } : item)));
      toast(err?.message || "操作失败", "error");
    } finally {
      setPendingId(null);
    }
  };

  const hasMore = items.length < total;

  return (
    <div className="overlay" onClick={onClose}>
      <div className="follow-modal" onClick={(event) => event.stopPropagation()}>
        <div className="follow-modal-head">
          <div className="segmented">
            <button type="button" className={tab === "following" ? "active" : ""} onClick={() => setTab("following")}>
              关注
            </button>
            <button type="button" className={tab === "followers" ? "active" : ""} onClick={() => setTab("followers")}>
              粉丝
            </button>
          </div>
          <div className="follow-modal-actions">
            <span className="muted" style={{ fontSize: 12.5 }}>{total} 人</span>
            <button type="button" className="icon-btn" onClick={onClose} aria-label="关闭">
              <Icon name="close" size={18} />
            </button>
          </div>
        </div>

        <div className="follow-modal-body">
          {loading ? (
            <LoadingRow text="加载中…" />
          ) : error ? (
            <ErrorState message={error} onRetry={() => load(tab, 1, false)} />
          ) : items.length === 0 ? (
            <EmptyState
              icon="user"
              title={tab === "following" ? "还没有关注任何人" : "还没有粉丝"}
              description={tab === "following" ? "去发现页逛逛，关注喜欢的作者吧" : "发布优质笔记，会有人关注你的"}
            />
          ) : (
            <>
              {items.map((user) => (
                <div
                  key={user.id}
                  className="follow-user"
                  onClick={() => {
                    onClose();
                    navigate(`/user/${user.id}`);
                  }}
                >
                  <Avatar src={user.avatar_url} name={user.nickname} size="md" />
                  <div className="info">
                    <div className="name">{user.nickname}</div>
                    <div className="bio">{user.bio || "这个人很懒，还没有写简介"}</div>
                  </div>
                  <button
                    type="button"
                    className={user.is_following ? "btn btn-ghost btn-sm" : "btn btn-soft btn-sm"}
                    onClick={(event) => {
                      event.stopPropagation();
                      toggleFollow(user);
                    }}
                    disabled={pendingId === user.id}
                  >
                    {user.is_following ? "已关注" : "+ 关注"}
                  </button>
                </div>
              ))}
              {hasMore ? (
                <div style={{ display: "flex", justifyContent: "center", padding: "12px 0 4px" }}>
                  <button
                    type="button"
                    className="btn btn-outline btn-sm"
                    onClick={() => load(tab, page + 1, true)}
                    disabled={loadingMore}
                  >
                    {loadingMore ? "加载中…" : "加载更多"}
                  </button>
                </div>
              ) : null}
            </>
          )}
        </div>
      </div>
    </div>
  );
}
