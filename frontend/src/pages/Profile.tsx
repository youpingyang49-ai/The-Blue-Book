import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import PostCard from "../components/PostCard";
import type { PostItem } from "../components/PostCard";
import TabBar from "../components/TabBar";

export default function Profile() {
  const { token, user, logout } = useAuth();
  const navigate = useNavigate();
  const [tab, setTab] = useState("posts");
  const [items, setItems] = useState<PostItem[]>([]);
  const [loading, setLoading] = useState(false);

  const load = async (t: string) => {
    if (!token) return;
    setLoading(true);
    try {
      const r = t === "posts" ? await api.myPosts() : await api.myLikes();
      setItems(r.items || []);
    } catch (e: any) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load(tab);
  }, [tab, token]);

  if (!token) {
    return (
      <div className="profile-page">
        <p className="empty">还没有登录</p>
        <button className="primary" onClick={() => navigate("/login")}>去登录</button>
        <TabBar />
      </div>
    );
  }

  return (
    <div className="profile-page">
      <div className="profile-head">
        <div className="avatar big">{(user?.nickname || "U").charAt(0)}</div>
        <div>
          <h2>{user?.nickname}</h2>
          <p style={{ color: "#999" }}>{user?.phone}</p>
        </div>
        <button className="logout" onClick={() => { logout(); navigate("/login"); }}>退出</button>
      </div>
      <div className="sort-tabs">
        <button className={tab === "posts" ? "active" : ""} onClick={() => setTab("posts")}>我的帖子</button>
        <button className={tab === "likes" ? "active" : ""} onClick={() => setTab("likes")}>我的点赞</button>
      </div>
      {loading ? (
        <p className="empty">加载中...</p>
      ) : items.length === 0 ? (
        <p className="empty">{tab === "posts" ? "还没有发过帖子" : "还没有点过赞"}</p>
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
