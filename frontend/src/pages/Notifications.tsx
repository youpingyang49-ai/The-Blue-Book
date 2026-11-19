import { useCallback, useEffect, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import AppShell from "../components/AppShell";
import Avatar from "../components/Avatar";
import Icon from "../components/Icon";
import { EmptyState, ErrorState, LoadingRow } from "../components/States";
import { useToast } from "../components/Toast";

interface NotifItem {
  id: number;
  type: string;
  post_id: number;
  content: string;
  read: boolean;
  created_at?: string;
  actor: { id: number; nickname: string; avatar_url: string };
}

interface ConvItem {
  user_id: number;
  nickname: string;
  avatar_url: string;
  last_message: string;
  unread: number;
  created_at?: string;
}

const TYPE_TEXT: Record<string, string> = {
  like: "赞了你的笔记",
  favorite: "收藏了你的笔记",
  comment: "评论了你的笔记",
  follow: "关注了你",
  mention: "在评论中提到了你",
};

const TABS = [
  { key: "like", label: "点赞", type: "like,favorite" },
  { key: "comment", label: "评论", type: "comment" },
  { key: "follow", label: "关注", type: "follow" },
  { key: "mention", label: "@我", type: "mention" },
  { key: "dm", label: "私信", type: "" },
];

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

export default function Notifications() {
  const navigate = useNavigate();
  const { token } = useAuth();
  const toast = useToast();
  const [params, setParams] = useSearchParams();
  const initial = params.get("tab") === "dm" ? "dm" : "like";
  const [tab, setTab] = useState(initial);
  const [notifs, setNotifs] = useState<NotifItem[]>([]);
  const [convs, setConvs] = useState<ConvItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");
  const [actionConv, setActionConv] = useState<ConvItem | null>(null);

  const isDm = tab === "dm";

  const load = useCallback(
    async (tabKey: string) => {
      if (!token) {
        setLoading(false);
        return;
      }
      setLoading(true);
      setError("");
      try {
        if (tabKey === "dm") {
          const result = await api.conversations();
          setConvs(result.items || []);
        } else {
          const type = TABS.find((item) => item.key === tabKey)?.type || "";
          const result = await api.notifications(type);
          setNotifs(result.items || []);
        }
      } catch (err: any) {
        setError(err?.message || "消息加载失败");
      } finally {
        setLoading(false);
      }
    },
    [token]
  );

  useEffect(() => {
    load(tab);
  }, [tab, load]);

  const switchTab = (next: string) => {
    setTab(next);
    if (next === "dm") setParams({ tab: "dm" });
    else setParams({});
  };

  const markAll = async () => {
    try {
      await api.markNotificationsRead();
      setNotifs((prev) => prev.map((item) => ({ ...item, read: true })));
      toast("已全部标为已读");
    } catch (err: any) {
      toast(err?.message || "操作失败", "error");
    }
  };

  const handleConvAction = async (action: "unread" | "hide" | "delete") => {
    if (!actionConv) return;
    try {
      if (action === "unread") await api.markConversationUnread(actionConv.user_id);
      if (action === "hide") await api.hideConversation(actionConv.user_id);
      if (action === "delete") await api.deleteConversation(actionConv.user_id);
      setActionConv(null);
      load("dm");
      toast("操作成功");
    } catch (err: any) {
      toast(err?.message || "操作失败", "error");
    }
  };

  if (!token) {
    return (
      <AppShell>
        <EmptyState
          icon="bell"
          title="登录后查看消息"
          description="点赞、评论、关注和私信都会出现在这里"
          action={
            <button className="btn btn-primary btn-sm" onClick={() => navigate("/login?redirect=/notifications")}>
              去登录
            </button>
          }
        />
      </AppShell>
    );
  }

  return (
    <AppShell>
      <div className="msg-page">
        <div className="msg-top">
          <h1>消息</h1>
          <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
            {!isDm ? (
              <button type="button" className="btn btn-ghost btn-sm" onClick={markAll}>
                全部已读
              </button>
            ) : null}
          </div>
        </div>

        <div className="tabs" style={{ borderBottom: "1px solid var(--border)", paddingBottom: 10 }}>
          {TABS.map((item) => (
            <button
              key={item.key}
              type="button"
              className={`tab ${tab === item.key ? "active" : ""}`}
              onClick={() => switchTab(item.key)}
            >
              {item.label}
            </button>
          ))}
        </div>

        <div className="msg-layout">
          {loading ? (
            <LoadingRow text="正在加载消息…" />
          ) : error ? (
            <ErrorState message={error} onRetry={() => load(tab)} />
          ) : isDm ? (
            convs.length === 0 ? (
              <EmptyState icon="chat" title="还没有私信" description="在笔记详情页找到感兴趣的人，给 TA 发第一条消息" />
            ) : (
              <div className="notif-list">
                {convs.map((conv) => (
                  <div className="notif-item" key={conv.user_id} onClick={() => navigate(`/chat/${conv.user_id}`)}>
                    <Avatar src={conv.avatar_url} name={conv.nickname} size="md" />
                    <div className="notif-main">
                      <p style={{ display: "flex", alignItems: "center", gap: 6 }}>
                        <b>{conv.nickname}</b>
                        {conv.unread > 0 ? <span className="conv-unread">{conv.unread}</span> : null}
                      </p>
                      <p className="notif-content">{conv.last_message}</p>
                    </div>
                    <span className="notif-time">{timeAgo(conv.created_at)}</span>
                    <button
                      type="button"
                      className="more-btn"
                      aria-label="会话操作"
                      onClick={(event) => {
                        event.stopPropagation();
                        setActionConv(conv);
                      }}
                    >
                      <Icon name="more" size={18} />
                    </button>
                  </div>
                ))}
              </div>
            )
          ) : notifs.length === 0 ? (
            <EmptyState
              icon={tab === "follow" ? "user" : tab === "comment" ? "comment" : "heart"}
              title="暂时没有新消息"
              description="当你发布更多笔记、与大家互动时，消息会出现在这里"
            />
          ) : (
            <div className="notif-list">
              {notifs.map((item) => (
                <div
                  key={item.id}
                  className={item.read ? "notif-item" : "notif-item unread"}
                  onClick={() => {
                    if (item.post_id) navigate(`/post/${item.post_id}`);
                    else if (item.actor?.id) navigate(`/user/${item.actor.id}`);
                  }}
                >
                  <Avatar src={item.actor?.avatar_url} name={item.actor?.nickname} size="md" />
                  <div className="notif-main">
                    <p>
                      <b>{item.actor?.nickname}</b> {TYPE_TEXT[item.type] || item.type}
                    </p>
                    {item.content ? <p className="notif-content">“{item.content}”</p> : null}
                  </div>
                  <span className="notif-time">{timeAgo(item.created_at)}</span>
                  {!item.read ? <span className="notif-dot" /> : null}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {actionConv ? (
        <div className="sheet-overlay" onClick={() => setActionConv(null)}>
          <div className="sheet" onClick={(event) => event.stopPropagation()}>
            <button type="button" className="sheet-item" onClick={() => handleConvAction("unread")}>
              标为未读
            </button>
            <button type="button" className="sheet-item" onClick={() => handleConvAction("hide")}>
              不显示该会话
            </button>
            <button type="button" className="sheet-item danger" onClick={() => handleConvAction("delete")}>
              删除会话
            </button>
            <button type="button" className="sheet-cancel" onClick={() => setActionConv(null)}>
              取消
            </button>
          </div>
        </div>
      ) : null}
    </AppShell>
  );
}
