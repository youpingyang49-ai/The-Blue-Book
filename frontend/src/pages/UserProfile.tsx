import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import AppShell from "../components/AppShell";
import Avatar from "../components/Avatar";
import MasonryFeed from "../components/MasonryFeed";
import type { PostItem } from "../components/PostCard";
import { EmptyState, ErrorState, FeedSkeleton, PageLoading } from "../components/States";
import { useToast } from "../components/Toast";

export default function UserProfile() {
  const { id } = useParams();
  const userId = Number(id);
  const navigate = useNavigate();
  const { token, user } = useAuth();
  const toast = useToast();

  const [profile, setProfile] = useState<any>(null);
  const [posts, setPosts] = useState<PostItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [following, setFollowing] = useState(false);

  const loadProfile = useCallback(async () => {
    try {
      const result = await api.getUser(userId);
      setProfile(result.user);
      setFollowing(Boolean(result.user?.is_following));
    } catch (err: any) {
      setError(err?.message || "用户不存在");
    }
  }, [userId]);

  const loadPosts = useCallback(async () => {
    try {
      const result = await api.userPosts(userId);
      setPosts(result.items || []);
    } catch {
      setPosts([]);
    } finally {
      setLoading(false);
    }
  }, [userId]);

  useEffect(() => {
    setLoading(true);
    setError("");
    loadProfile();
    loadPosts();
  }, [loadProfile, loadPosts]);

  const toggleFollow = async () => {
    if (!token) {
      toast("登录后即可关注", "error");
      navigate(`/login?redirect=/user/${userId}`);
      return;
    }
    const next = !following;
    setFollowing(next);
    setProfile((prev: any) =>
      prev ? { ...prev, follower_count: Math.max(0, (prev.follower_count || 0) + (next ? 1 : -1)) } : prev
    );
    try {
      if (next) await api.follow(userId);
      else await api.unfollow(userId);
      toast(next ? "已关注" : "已取消关注");
    } catch (err: any) {
      setFollowing(!next);
      toast(err?.message || "操作失败", "error");
    }
  };

  if (error) {
    return (
      <AppShell>
        <ErrorState message={error} onRetry={loadProfile} />
      </AppShell>
    );
  }

  if (!profile) {
    return (
      <AppShell>
        <PageLoading text="正在加载主页…" />
      </AppShell>
    );
  }

  const isSelf = Boolean(user && user.id === profile.id);

  return (
    <AppShell>
      <section className="profile-hero fade-up">
        <div className="profile-hero-main">
          <Avatar src={profile.avatar_url} name={profile.nickname} size="xl" />
          <div className="profile-identity">
            <h1 className="profile-name">{profile.nickname}</h1>
            <p className="profile-bio">{profile.bio || "这个人很懒，还没有写简介"}</p>
            <div className="profile-stats">
              <span className="profile-stat">
                <b>{profile.following_count ?? 0}</b> 关注
              </span>
              <span className="profile-stat">
                <b>{profile.follower_count ?? 0}</b> 粉丝
              </span>
            </div>
          </div>
        </div>
        <div className="profile-actions">
          {isSelf ? (
            <button type="button" className="btn btn-outline btn-sm" onClick={() => navigate("/profile")}>
              管理我的主页
            </button>
          ) : (
            <button
              type="button"
              className={following ? "btn btn-ghost btn-sm" : "btn btn-primary btn-sm"}
              onClick={toggleFollow}
            >
              {following ? "已关注" : "+ 关注"}
            </button>
          )}
        </div>
      </section>

      <div className="profile-tabs">
        <div className="tabs">
          <span className="tab active">笔记</span>
        </div>
        <span className="muted" style={{ fontSize: 12.5 }}>
          {loading ? "加载中…" : `${posts.length} 篇`}
        </span>
      </div>

      {loading ? (
        <FeedSkeleton count={8} />
      ) : posts.length === 0 ? (
        <EmptyState icon="edit" title="还没有发布笔记" description="关注这位作者，等待 TA 的第一篇分享" />
      ) : (
        <MasonryFeed items={posts} />
      )}
    </AppShell>
  );
}
