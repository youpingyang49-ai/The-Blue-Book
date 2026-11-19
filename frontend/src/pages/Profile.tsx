import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import AppShell from "../components/AppShell";
import Avatar from "../components/Avatar";
import MasonryFeed from "../components/MasonryFeed";
import type { PostItem } from "../components/PostCard";
import { EmptyState, ErrorState, FeedSkeleton } from "../components/States";
import { useToast } from "../components/Toast";

const TABS = [
  { key: "posts", label: "我的笔记", empty: "还没有发布过笔记", cta: "去发布" },
  { key: "likes", label: "我的点赞", empty: "还没有点过赞", cta: "去发现" },
  { key: "favorites", label: "我的收藏", empty: "还没有收藏任何笔记", cta: "去发现" },
];

function maskPhone(phone?: string) {
  if (!phone || phone.length < 7) return phone || "";
  return `${phone.slice(0, 3)}****${phone.slice(-4)}`;
}

export default function Profile() {
  const { token, user, setAuth, logout } = useAuth();
  const navigate = useNavigate();
  const toast = useToast();
  const [tab, setTab] = useState("posts");
  const [profile, setProfile] = useState<any>(null);
  const [items, setItems] = useState<PostItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");

  useEffect(() => {
    if (!token) return;
    (async () => {
      try {
        const me = await api.me();
        const fresh = me.user;
        if (fresh) {
          setAuth(token, fresh);
          const detail = await api.getUser(fresh.id);
          setProfile(detail.user);
        }
      } catch {
        /* 保留本地缓存的用户信息 */
      }
    })();
  }, [token, setAuth]);

  const load = useCallback(
    async (key: string) => {
      if (!token) return;
      setLoading(true);
      setError("");
      try {
        const result =
          key === "posts" ? await api.myPosts() : key === "likes" ? await api.myLikes() : await api.myFavorites();
        setItems(result.items || []);
      } catch (err: any) {
        setError(err?.message || "加载失败");
        setItems([]);
      } finally {
        setLoading(false);
      }
    },
    [token]
  );

  useEffect(() => {
    load(tab);
  }, [tab, load]);

  const currentTab = TABS.find((item) => item.key === tab) || TABS[0];
  const nickname = profile?.nickname || user?.nickname || "未命名用户";
  const avatar = profile?.avatar_url || user?.avatar_url;
  const bio = profile?.bio || "还没有填写简介，记录一点生活的碎片吧";

  if (!token) {
    return (
      <AppShell>
        <EmptyState
          icon="user"
          title="还没有登录"
          description="登录后可以管理你的笔记、点赞和收藏"
          action={
            <button className="btn btn-primary btn-sm" onClick={() => navigate("/login?redirect=/profile")}>
              去登录
            </button>
          }
        />
      </AppShell>
    );
  }

  return (
    <AppShell>
      <section className="profile-hero fade-up">
        <div className="profile-hero-main">
          <Avatar src={avatar} name={nickname} size="xl" />
          <div className="profile-identity">
            <h1 className="profile-name">
              {nickname}
              <span className="profile-own-badge">我</span>
            </h1>
            <p className="profile-bio">{bio}</p>
            <div className="profile-stats">
              <span className="profile-stat">
                <b>{profile?.following_count ?? 0}</b> 关注
              </span>
              <span className="profile-stat">
                <b>{profile?.follower_count ?? 0}</b> 粉丝
              </span>
              <span className="profile-stat muted">{maskPhone(user?.phone)}</span>
            </div>
          </div>
        </div>
        <div className="profile-actions">
          <button type="button" className="btn btn-primary btn-sm" onClick={() => navigate("/create")}>
            发布笔记
          </button>
          <button
            type="button"
            className="btn btn-outline btn-sm"
            onClick={() => {
              logout();
              toast("已退出登录");
              navigate("/login");
            }}
          >
            退出登录
          </button>
        </div>
      </section>

      <div className="profile-tabs">
        <div className="segmented">
          {TABS.map((item) => (
            <button
              key={item.key}
              type="button"
              className={tab === item.key ? "active" : ""}
              onClick={() => setTab(item.key)}
            >
              {item.label}
            </button>
          ))}
        </div>
        <span className="muted" style={{ fontSize: 12.5 }}>
          {loading ? "加载中…" : `${items.length} 篇`}
        </span>
      </div>

      {loading ? (
        <FeedSkeleton count={8} />
      ) : error ? (
        <ErrorState message={error} onRetry={() => load(tab)} />
      ) : items.length === 0 ? (
        <EmptyState
          icon={tab === "posts" ? "edit" : "heart"}
          title={currentTab.empty}
          description={tab === "posts" ? "把值得记录的生活分享出来" : "遇到喜欢的笔记，随手收进这里"}
          action={
            <button
              className="btn btn-primary btn-sm"
              onClick={() => navigate(tab === "posts" ? "/create" : "/search")}
            >
              {currentTab.cta}
            </button>
          }
        />
      ) : (
        <MasonryFeed items={items} />
      )}
    </AppShell>
  );
}
