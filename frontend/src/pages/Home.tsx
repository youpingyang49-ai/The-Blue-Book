import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import PostCard from "../components/PostCard";
import type { PostItem } from "../components/PostCard";
import TabBar from "../components/TabBar";

export default function Home() {
  const [items, setItems] = useState<PostItem[]>([]);
  const [sort, setSort] = useState("recommend");
  const [loading, setLoading] = useState(true);
  const [unread, setUnread] = useState(0);
  const navigate = useNavigate();
  const { token } = useAuth();

  const load = async (s: string) => {
    setLoading(true);
    try {
      const r = await api.feed(s, 1);
      setItems(r.items || []);
    } catch (e: any) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const loadUnread = async () => {
    if (!token) return;
    try {
      const r = await api.unreadCount();
      setUnread(r.count || 0);
    } catch (e) {
      console.error(e);
    }
  };

  useEffect(() => {
    load(sort);
  }, [sort]);

  useEffect(() => {
    loadUnread();
  }, [token]);

  const needLoginForFollowing = sort === "following" && !token;

  return (
    <div className="home">
      <header className="topbar">
        <div className="search-box" onClick={() => navigate("/search")}>🔍 搜索笔记</div>
        <button className="bell" onClick={() => navigate("/notifications")}>
          🔔
          {unread > 0 && <span className="badge">{unread > 99 ? "99+" : unread}</span>}
        </button>
      </header>
      <div className="sort-tabs">
        <button className={sort === "recommend" ? "active" : ""} onClick={() => setSort("recommend")}>推荐</button>
        <button className={sort === "latest" ? "active" : ""} onClick={() => setSort("latest")}>最新</button>
        <button className={sort === "hot" ? "active" : ""} onClick={() => setSort("hot")}>热门</button>
        <button className={sort === "following" ? "active" : ""} onClick={() => setSort("following")}>关注</button>
      </div>
      {loading ? (
        <p className="empty">加载中...</p>
      ) : needLoginForFollowing ? (
        <p className="empty">登录后查看关注内容</p>
      ) : items.length === 0 ? (
        <p className="empty">还没有帖子，去发布第一篇吧</p>
      ) : (
        <div className="waterfall">
          {items.map((p) => (
            <PostCard key={p.id} post={p} />
          ))}
        </div>
      )}
      <TabBar />
    </div>
  );
}
