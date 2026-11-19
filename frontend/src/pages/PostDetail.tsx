import { useCallback, useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import AppShell from "../components/AppShell";
import Avatar from "../components/Avatar";
import Icon from "../components/Icon";
import { EmptyState, ErrorState, PageLoading } from "../components/States";
import { useToast } from "../components/Toast";

interface CommentItem {
  id: number;
  parent_id: number;
  content: string;
  reply_to_name?: string;
  author: { id: number; nickname: string; avatar_url: string };
  created_at?: string;
}

function timeAgo(iso?: string) {
  if (!iso) return "";
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return "";
  const diff = (Date.now() - date.getTime()) / 1000;
  if (diff < 60) return "刚刚";
  if (diff < 3600) return `${Math.floor(diff / 60)} 分钟前`;
  if (diff < 86400) return `${Math.floor(diff / 3600)} 小时前`;
  if (diff < 86400 * 30) return `${Math.floor(diff / 86400)} 天前`;
  return date.toLocaleDateString("zh-CN");
}

export default function PostDetail() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { token, user } = useAuth();
  const toast = useToast();

  const [post, setPost] = useState<any>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [comments, setComments] = useState<CommentItem[]>([]);
  const [commentText, setCommentText] = useState("");
  const [submitting, setSubmitting] = useState(false);
  const [replyingTo, setReplyingTo] = useState<{ id: number; nickname: string } | null>(null);
  const [activeImage, setActiveImage] = useState(0);
  const [shareOpen, setShareOpen] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);

  const postId = Number(id);

  const loadPost = useCallback(async () => {
    setLoading(true);
    setError("");
    try {
      const result = await api.postDetail(postId);
      setPost(result.post);
    } catch (err: any) {
      setError(err?.message || "笔记加载失败");
    } finally {
      setLoading(false);
    }
  }, [postId]);

  const loadComments = useCallback(async () => {
    try {
      const result = await api.listComments(postId);
      setComments(result.items || []);
    } catch {
      setComments([]);
    }
  }, [postId]);

  useEffect(() => {
    loadPost();
    loadComments();
    setActiveImage(0);
  }, [loadPost, loadComments]);

  const requireLogin = () => {
    if (!token) {
      toast("登录后即可操作", "error");
      navigate("/login?redirect=/post/" + postId);
      return true;
    }
    return false;
  };

  const patchPost = (patch: Record<string, any>) => setPost((prev: any) => (prev ? { ...prev, ...patch } : prev));

  const toggleLike = async () => {
    if (requireLogin()) return;
    const next = !post.is_liked;
    patchPost({ is_liked: next, like_count: Math.max(0, post.like_count + (next ? 1 : -1)) });
    try {
      if (next) await api.like(post.id);
      else await api.unlike(post.id);
    } catch (err: any) {
      patchPost({ is_liked: !next, like_count: post.like_count });
      toast(err?.message || "操作失败", "error");
    }
  };

  const toggleFavorite = async () => {
    if (requireLogin()) return;
    const next = !post.is_favorited;
    patchPost({ is_favorited: next, favorite_count: Math.max(0, (post.favorite_count || 0) + (next ? 1 : -1)) });
    try {
      if (next) await api.favorite(post.id);
      else await api.unfavorite(post.id);
      toast(next ? "已收藏" : "已取消收藏");
    } catch (err: any) {
      patchPost({ is_favorited: !next, favorite_count: post.favorite_count });
      toast(err?.message || "操作失败", "error");
    }
  };

  const toggleFollow = async () => {
    if (requireLogin()) return;
    const next = !post.is_following;
    patchPost({ is_following: next });
    try {
      if (next) await api.follow(post.author.id);
      else await api.unfollow(post.author.id);
    } catch (err: any) {
      patchPost({ is_following: !next });
      toast(err?.message || "操作失败", "error");
    }
  };

  const submitComment = async () => {
    const content = commentText.trim();
    if (!content) return;
    if (requireLogin()) return;
    setSubmitting(true);
    try {
      await api.createComment(postId, content, replyingTo?.id || 0);
      setCommentText("");
      setReplyingTo(null);
      await loadComments();
      patchPost({ comment_count: (post.comment_count || 0) + 1 });
      toast("评论已发布");
    } catch (err: any) {
      toast(err?.message || "评论失败", "error");
    } finally {
      setSubmitting(false);
    }
  };

  const removePost = async () => {
    if (!window.confirm("确定删除这篇笔记吗？删除后不可恢复")) return;
    try {
      await api.deletePost(postId);
      toast("笔记已删除");
      navigate("/");
    } catch (err: any) {
      toast(err?.message || "删除失败", "error");
    }
  };

  const share = async () => {
    try {
      await navigator.clipboard.writeText(window.location.href);
      toast("链接已复制");
    } catch {
      toast("复制失败，请手动复制地址栏链接", "error");
    }
    setShareOpen(false);
  };

  const imageList: string[] = post?.image_urls?.length ? post.image_urls : post?.cover_image_url ? [post.cover_image_url] : [];
  const isSelf = Boolean(user && post && user.id === post.author?.id);

  const tree = useMemo(() => {
    const top = comments.filter((comment) => !comment.parent_id);
    const replies = (parentId: number) => comments.filter((comment) => comment.parent_id === parentId);
    return { top, replies };
  }, [comments]);

  const renderComment = (comment: CommentItem, depth: number) => (
    <div key={comment.id}>
      <div className={depth > 0 ? "comment-item reply" : "comment-item"}>
        <Avatar src={comment.author?.avatar_url} name={comment.author?.nickname} size="sm" />
        <div className="comment-main">
          <div className="comment-top">
            <b>{comment.author?.nickname}</b>
            {comment.reply_to_name ? <span>回复 @{comment.reply_to_name}</span> : null}
            <span>·</span>
            <span>{timeAgo(comment.created_at)}</span>
          </div>
          <p className="comment-text">{comment.content}</p>
          <div className="comment-actions">
            <button
              type="button"
              onClick={() => setReplyingTo({ id: comment.id, nickname: comment.author?.nickname || "" })}
            >
              回复
            </button>
          </div>
        </div>
      </div>
      {tree.replies(comment.id).map((reply) => renderComment(reply, depth + 1))}
    </div>
  );

  return (
    <AppShell>
      {loading ? (
        <PageLoading text="正在打开笔记…" />
      ) : error || !post ? (
        <ErrorState message={error || "笔记不存在"} onRetry={loadPost} />
      ) : (
        <div className="detail-wrap fade-up">
          <div className="detail-media">
            {imageList.length > 0 ? (
              <img src={imageList[activeImage]} alt={post.title} />
            ) : (
              <div className="post-cover-skeleton" style={{ borderRadius: 16 }} />
            )}
            {imageList.length > 1 ? (
              <div className="detail-media-dots">
                {imageList.map((_, index) => (
                  <button
                    key={index}
                    type="button"
                    aria-label={`第 ${index + 1} 张`}
                    onClick={() => setActiveImage(index)}
                    style={{
                      width: index === activeImage ? 18 : 6,
                      height: 6,
                      borderRadius: 999,
                      background: index === activeImage ? "var(--brand-500)" : "var(--border-strong)",
                      transition: "all .2s",
                    }}
                  />
                ))}
              </div>
            ) : null}
          </div>

          <div className="detail-panel">
            <div className="detail-author">
              <Avatar
                src={post.author?.avatar_url}
                name={post.author?.nickname}
                size="md"
                className="avatar-clickable"
              />
              <div style={{ minWidth: 0 }}>
                <div className="detail-author-name" onClick={() => navigate(`/user/${post.author.id}`)}>
                  {post.author?.nickname}
                </div>
                <div className="detail-meta-line">{timeAgo(post.created_at)}</div>
              </div>
              <span className="spacer" />
              {!isSelf ? (
                <button
                  type="button"
                  className={post.is_following ? "btn btn-ghost btn-sm" : "btn btn-soft btn-sm"}
                  onClick={toggleFollow}
                >
                  {post.is_following ? "已关注" : "+ 关注"}
                </button>
              ) : null}
              <button type="button" className="icon-btn" onClick={() => setMenuOpen(true)} aria-label="更多操作">
                <Icon name="more" size={20} />
              </button>
            </div>

            <h1 className="detail-title">{post.title}</h1>

            {post.topics?.length ? (
              <div className="detail-topics">
                {post.topics.map((topic: string) => (
                  <button
                    key={topic}
                    type="button"
                    className="tag"
                    onClick={() => navigate(`/search?q=${encodeURIComponent(topic)}`)}
                  >
                    #{topic}
                  </button>
                ))}
              </div>
            ) : null}

            {post.content ? <p className="detail-content">{post.content}</p> : null}

            <div className="detail-actions">
              <button type="button" className={post.is_liked ? "action-pill liked" : "action-pill"} onClick={toggleLike}>
                <Icon name="heart" size={17} />
                {post.like_count || 0}
              </button>
              <button
                type="button"
                className={post.is_favorited ? "action-pill faved" : "action-pill"}
                onClick={toggleFavorite}
              >
                <Icon name="star" size={17} />
                {post.favorite_count || 0}
              </button>
              <button
                type="button"
                className="action-pill"
                onClick={() => document.getElementById("comments")?.scrollIntoView({ behavior: "smooth" })}
              >
                <Icon name="comment" size={17} />
                {post.comment_count || 0}
              </button>
              <span className="spacer" />
              <button type="button" className="action-pill icon-only" onClick={() => setShareOpen(true)} aria-label="分享">
                <Icon name="share" size={17} />
              </button>
            </div>

            <div className="comments-block" id="comments">
              <div className="comments-head">
                评论 <span>{comments.length}</span>
              </div>

              <div className="comment-compose">
                <input
                  className="input"
                  value={commentText}
                  onChange={(event) => setCommentText(event.target.value)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" && !event.shiftKey) submitComment();
                  }}
                  placeholder={replyingTo ? `回复 @${replyingTo.nickname}…` : token ? "说点什么…" : "登录后参与评论"}
                  onFocus={() => {
                    if (!token) requireLogin();
                  }}
                />
                <button
                  type="button"
                  className="btn btn-primary btn-sm"
                  onClick={submitComment}
                  disabled={!commentText.trim() || submitting}
                >
                  {submitting ? "发送中" : "发送"}
                </button>
              </div>
              {replyingTo ? (
                <button type="button" className="cancel-reply" onClick={() => setReplyingTo(null)}>
                  取消回复
                </button>
              ) : null}

              {comments.length === 0 ? (
                <EmptyState icon="comment" title="还没有评论" description="来说点什么都好" />
              ) : (
                <div style={{ display: "flex", flexDirection: "column", gap: 6 }}>
                  {tree.top.map((comment) => renderComment(comment, 0))}
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {menuOpen ? (
        <div className="sheet-overlay" onClick={() => setMenuOpen(false)}>
          <div className="sheet" onClick={(event) => event.stopPropagation()}>
            <button type="button" className="sheet-item" onClick={() => { setMenuOpen(false); setShareOpen(true); }}>
              分享笔记
            </button>
            {isSelf ? (
              <>
                <button type="button" className="sheet-item" onClick={() => navigate(`/edit/${post.id}`)}>
                  编辑笔记
                </button>
                <button type="button" className="sheet-item danger" onClick={removePost}>
                  删除笔记
                </button>
              </>
            ) : (
              <button
                type="button"
                className="sheet-item"
                onClick={() => {
                  setMenuOpen(false);
                  navigate(`/user/${post.author.id}`);
                }}
              >
                查看作者主页
              </button>
            )}
            <button type="button" className="sheet-cancel" onClick={() => setMenuOpen(false)}>
              取消
            </button>
          </div>
        </div>
      ) : null}

      {shareOpen ? (
        <div className="sheet-overlay" onClick={() => setShareOpen(false)}>
          <div className="sheet" onClick={(event) => event.stopPropagation()}>
            <button type="button" className="sheet-item" onClick={share}>
              复制链接
            </button>
            <button type="button" className="sheet-cancel" onClick={() => setShareOpen(false)}>
              取消
            </button>
          </div>
        </div>
      ) : null}
    </AppShell>
  );
}
