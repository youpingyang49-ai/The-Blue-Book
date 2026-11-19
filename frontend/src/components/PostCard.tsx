import { useEffect, useState } from "react";
import type { MouseEvent } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import { useToast } from "./Toast";
import Avatar from "./Avatar";
import Icon from "./Icon";

export interface PostItem {
  id: number;
  title: string;
  content?: string;
  cover_image_url: string;
  image_urls?: string[];
  like_count: number;
  favorite_count?: number;
  comment_count?: number;
  is_liked: boolean;
  is_favorited?: boolean;
  topics?: string[] | null;
  author: { id: number; nickname: string; avatar_url: string };
  created_at?: string;
}

interface PostCardProps {
  post: PostItem;
  onMediaLoad?: () => void;
}

export default function PostCard({ post, onMediaLoad }: PostCardProps) {
  const navigate = useNavigate();
  const { token } = useAuth();
  const toast = useToast();
  const [loaded, setLoaded] = useState(false);
  const [liked, setLiked] = useState(post.is_liked);
  const [likeCount, setLikeCount] = useState(post.like_count);
  const [pending, setPending] = useState(false);

  useEffect(() => {
    setLiked(post.is_liked);
    setLikeCount(post.like_count);
  }, [post.id, post.is_liked, post.like_count]);

  const open = () => navigate(`/post/${post.id}`);

  const toggleLike = async (event: MouseEvent) => {
    event.stopPropagation();
    if (!token) {
      toast("登录后即可点赞", "error");
      navigate("/login");
      return;
    }
    if (pending) return;
    setPending(true);
    const nextLiked = !liked;
    setLiked(nextLiked);
    setLikeCount((count) => Math.max(0, count + (nextLiked ? 1 : -1)));
    try {
      if (nextLiked) await api.like(post.id);
      else await api.unlike(post.id);
    } catch (err: any) {
      setLiked(!nextLiked);
      setLikeCount((count) => Math.max(0, count + (nextLiked ? -1 : 1)));
      toast(err?.message || "操作失败，请重试", "error");
    } finally {
      setPending(false);
    }
  };

  const topics = (post.topics || []).filter(Boolean).slice(0, 2);

  return (
    <article className="post-card" onClick={open}>
      <div className="post-cover">
        {!loaded && <div className="post-cover-skeleton" />}
        <img
          src={post.cover_image_url}
          alt={post.title}
          loading="lazy"
          decoding="async"
          onLoad={() => {
            setLoaded(true);
            onMediaLoad?.();
          }}
          onError={(event) => {
            (event.currentTarget as HTMLImageElement).style.display = "none";
            setLoaded(true);
            onMediaLoad?.();
          }}
        />
      </div>
      <div className="post-body">
        <h3 className="post-title clamp-2">{post.title}</h3>
        {topics.length > 0 && (
          <div className="post-tags">
            {topics.map((topic) => (
              <span className="post-tag" key={topic}>
                #{topic}
              </span>
            ))}
          </div>
        )}
        <div className="post-meta">
          <Avatar src={post.author?.avatar_url} name={post.author?.nickname} size="xs" />
          <span className="nickname" onClick={(event) => event.stopPropagation()}>
            {post.author?.nickname}
          </span>
          <button
            type="button"
            className={liked ? "like-btn liked" : "like-btn"}
            onClick={toggleLike}
            aria-label={liked ? "取消点赞" : "点赞"}
          >
            <Icon name="heart" size={15} />
            {likeCount > 0 ? likeCount : "赞"}
          </button>
        </div>
      </div>
    </article>
  );
}
