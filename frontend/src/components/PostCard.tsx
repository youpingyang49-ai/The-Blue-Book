import { useNavigate } from "react-router-dom";

export interface PostItem {
  id: number;
  title: string;
  cover_image_url: string;
  like_count: number;
  is_liked: boolean;
  author: { id: number; nickname: string; avatar_url: string };
}

export default function PostCard({ post }: { post: PostItem }) {
  const navigate = useNavigate();
  return (
    <div className="post-card" onClick={() => navigate("/post/" + post.id)}>
      <img className="post-cover" src={post.cover_image_url} alt={post.title} />
      <div className="post-meta">
        <div className="avatar">{post.author.nickname.charAt(0)}</div>
        <span className="nickname">{post.author.nickname}</span>
        <span className="like-count">❤ {post.like_count}</span>
      </div>
    </div>
  );
}
