import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";

interface CommentItem {
  id: number;
  content: string;
  author: { id: number; nickname: string; avatar_url: string };
}

export default function PostDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { token, user } = useAuth();
  const [post, setPost] = useState<any>(null);
  const [comments, setComments] = useState<CommentItem[]>([]);
  const [commentText, setCommentText] = useState("");
  const [msg, setMsg] = useState("");
  const [showShare, setShowShare] = useState(false);

  const load = async () => {
    try {
      const r = await api.postDetail(Number(id));
      setPost(r.post);
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  const loadComments = async () => {
    try {
      const r = await api.listComments(Number(id));
      setComments(r.items || []);
    } catch (e: any) {
      console.error(e);
    }
  };

  useEffect(() => {
    load();
    loadComments();
  }, [id]);

  const needLogin = () => {
    if (!token) {
      navigate("/login");
      return true;
    }
    return false;
  };

  const toggleLike = async () => {
    if (needLogin()) return;
    try {
      if (post.is_liked) await api.unlike(post.id);
      else await api.like(post.id);
      load();
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  const toggleFavorite = async () => {
    if (needLogin()) return;
    try {
      if (post.is_favorited) await api.unfavorite(post.id);
      else await api.favorite(post.id);
      load();
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  const toggleFollow = async () => {
    if (needLogin()) return;
    try {
      if (post.is_following) await api.unfollow(post.author.id);
      else await api.follow(post.author.id);
      load();
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  const deletePost = async () => {
    if (!window.confirm("确定删除这篇笔记吗？")) return;
    try {
      await api.deletePost(post.id);
      navigate("/");
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  const copyLink = async () => {
    try {
      await navigator.clipboard.writeText(window.location.href);
      setMsg("链接已复制，去分享吧");
    } catch (e) {
      setMsg("复制失败");
    }
    setShowShare(false);
  };

  const saveImage = async () => {
    const url = post.image_urls && post.image_urls[0];
    if (!url) return;
    try {
      const resp = await fetch(url);
      const blob = await resp.blob();
      const a = document.createElement("a");
      a.href = URL.createObjectURL(blob);
      a.download = "cover.jpg";
      a.click();
      URL.revokeObjectURL(a.href);
    } catch (e) {
      setMsg("保存失败");
    }
    setShowShare(false);
  };

  const submitComment = async () => {
    const c = commentText.trim();
    if (!c) return;
    if (needLogin()) return;
    try {
      await api.createComment(Number(id), c);
      setCommentText("");
      loadComments();
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  if (!post) {
    return <div className="empty">{msg || "加载中..."}</div>;
  }

  const isSelf = user && user.id === post.author.id;

  return (
    <div className="detail-page">
      <header className="topbar detail-topbar">
        <button className="back" onClick={() => navigate(-1)}>←</button>
        <div className="detail-author" onClick={() => navigate("/user/" + post.author.id)}>
          <div className="avatar">{(post.author?.nickname || "U").charAt(0)}</div>
          <span className="nickname">{post.author?.nickname}</span>
        </div>
        {!isSelf && (
          <button className={post.is_following ? "follow-btn following" : "follow-btn"} onClick={toggleFollow}>
            {post.is_following ? "已关注" : "+ 关注"}
          </button>
        )}
        <button className="share-btn" onClick={() => setShowShare(true)}>分享</button>
      </header>
      <div className="detail-images">
        {(post.image_urls || []).map((u: string) => (
          <img key={u} src={u} alt="" />
        ))}
      </div>
      <div className="detail-body">
        <h2>{post.title}</h2>
        {post.topics && post.topics.length > 0 && (
          <div className="topic-tags">
            {post.topics.map((t: string) => (
              <span className="topic-tag" key={t}>#{t}</span>
            ))}
          </div>
        )}
        <p className="content">{post.content}</p>
        <div className="action-row">
          <button className={post.is_liked ? "like-btn liked" : "like-btn"} onClick={toggleLike}>
            {post.is_liked ? "❤ 已赞" : "🤍 点赞"} {post.like_count}
          </button>
          <button className={post.is_favorited ? "like-btn liked" : "like-btn"} onClick={toggleFavorite}>
            {post.is_favorited ? "★ 已收藏" : "☆ 收藏"} {post.favorite_count}
          </button>
        </div>
        {isSelf && (
          <div className="action-row">
            <button className="like-btn" onClick={() => navigate("/edit/" + post.id)}>编辑</button>
            <button className="like-btn" onClick={deletePost}>删除</button>
          </div>
        )}
        {msg && <p className="msg">{msg}</p>}
      </div>
      <div className="comments">
        <h3>评论 {comments.length}</h3>
        <div className="comment-input">
          <input
            placeholder={token ? "说点什么..." : "登录后评论"}
            value={commentText}
            onChange={(e) => setCommentText(e.target.value)}
            onKeyDown={(e) => {
              if (e.key === "Enter") submitComment();
            }}
          />
          <button className="primary small" onClick={submitComment}>发送</button>
        </div>
        {comments.length === 0 ? (
          <p className="empty">还没有评论</p>
        ) : (
          comments.map((c) => (
            <div className="comment-item" key={c.id}>
              <div className="avatar">{(c.author?.nickname || "U").charAt(0)}</div>
              <div className="comment-main">
                <span className="nickname">{c.author?.nickname}</span>
                <p>{c.content}</p>
              </div>
            </div>
          ))
        )}
      </div>
      {showShare && (
        <div className="share-overlay" onClick={() => setShowShare(false)}>
          <div className="share-sheet" onClick={(e) => e.stopPropagation()}>
            <button onClick={copyLink}>复制链接</button>
            <button onClick={saveImage}>保存封面图</button>
            <button className="cancel" onClick={() => setShowShare(false)}>取消</button>
          </div>
        </div>
      )}
    </div>
  );
}
