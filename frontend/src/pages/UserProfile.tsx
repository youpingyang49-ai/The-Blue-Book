import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import PostCard from "../components/PostCard";
import type { PostItem } from "../components/PostCard";
import TabBar from "../components/TabBar";

export default function UserProfile() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { token, user } = useAuth();
  const [profile, setProfile] = useState<any>(null);
  const [posts, setPosts] = useState<PostItem[]>([]);
  const [loading, setLoading] = useState(true);

  const loadProfile = async () => {
    try {
      const r = await api.getUser(Number(id));
      setProfile(r.user);
    } catch (e: any) {
      console.error(e);
    }
  };

  const loadPosts = async () => {
    try {
      const r = await api.userPosts(Number(id));
      setPosts(r.items || []);
    } catch (e: any) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadProfile();
    loadPosts();
  }, [id]);

  const toggleFollow = async () => {
    if (!token) {
      navigate("/login");
      return;
    }
    try {
      if (profile.is_following) await api.unfollow(profile.id);
      else await api.follow(profile.id);
      loadProfile();
    } catch (e: any) {
      console.error(e);
    }
  };

  if (!profile) {
    return <div className="empty">加载中...</div>;
  }

  const isSelf = user && user.id === profile.id;

  return (
    <div className="user-profile-page">
      <header className="topbar">
        <button className="back" onClick={() => navigate(-1)}>←</button>
        <span className="title">个人主页</span>
      </header>
      <div className="profile-head">
        <div className="avatar big">{(profile.nickname || "U").charAt(0)}</div>
        <div className="profile-info">
          <h2>{profile.nickname}</h2>
          <p className="bio">{profile.bio || "这个人很懒，什么都没写"}</p>
          <div className="profile-stats">
            <span>{profile.following_count} 关注</span>
            <span>{profile.follower_count} 粉丝</span>
          </div>
        </div>
        {!isSelf && (
          <button className={profile.is_following ? "follow-btn following" : "follow-btn"} onClick={toggleFollow}>
            {profile.is_following ? "已关注" : "+ 关注"}
          </button>
        )}
      </div>
      {loading ? (
        <p className="empty">加载中...</p>
      ) : posts.length === 0 ? (
        <p className="empty">还没有发布笔记</p>
      ) : (
        <div className="waterfall">
          {posts.map((p) => (
            <PostCard key={p.id} post={p} />
          ))}
        </div>
      )}
      <TabBar />
    </div>
  );
}
