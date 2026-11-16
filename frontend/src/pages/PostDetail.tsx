import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";

export default function PostDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const token = useAuth((s) => s.token);
  const [post, setPost] = useState<any>(null);
  const [msg, setMsg] = useState("");

  const load = async () => {
    try {
      const r = await api.postDetail(Number(id));
      setPost(r.post);
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  useEffect(() => {
    load();
  }, [id]);

  const toggleLike = async () => {
    if (!token) {
      navigate("/login");
      return;
    }
    try {
      if (post.is_liked) {
        await api.unlike(post.id);
      } else {
        await api.like(post.id);
      }
      load();
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  if (!post) {
    return <div className="empty">{msg || "加载中..."}</div>;
  }

  return (
    <div className="detail-page">
      <header className="topbar">
        <button className="back" onClick={() => navigate(-1)}>←</button>
        <span className="title">笔记详情</span>
      </header>
      <div className="detail-images">
        {(post.image_urls || []).map((u: string) => (
          <img key={u} src={u} alt="" />
        ))}
      </div>
      <div className="detail-body">
        <h2>{post.title}</h2>
        <p className="content">{post.content}</p>
        <div className="author-row">
          <div className="avatar">{(post.author?.nickname || "U").charAt(0)}</div>
          <span>{post.author?.nickname}</span>
        </div>
        <button className={post.is_liked ? "like-btn liked" : "like-btn"} onClick={toggleLike}>
          {post.is_liked ? "❤ 已赞" : "🤍 点赞"} {post.like_count}
        </button>
        {msg && <p className="msg">{msg}</p>}
      </div>
    </div>
  );
}
