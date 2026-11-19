import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";

interface NotifItem {
  id: number;
  type: string;
  post_id: number;
  content: string;
  read: boolean;
  actor: { id: number; nickname: string };
}

interface ConvItem {
  user_id: number;
  nickname: string;
  last_message: string;
  unread: number;
}

const typeText: Record<string, string> = {
  like: "赞了你的笔记",
  favorite: "收藏了你的笔记",
  comment: "评论了你的笔记",
  follow: "关注了你",
  mention: "在评论中提到了你",
};

const tabs = [
  { key: "like", label: "点赞", type: "like,favorite" },
  { key: "comment", label: "评论", type: "comment" },
  { key: "follow", label: "关注", type: "follow" },
  { key: "mention", label: "@我", type: "mention" },
  { key: "dm", label: "私信", type: "" },
];

export default function Notifications() {
  const navigate = useNavigate();
  const [tab, setTab] = useState("like");
  const [notifs, setNotifs] = useState<NotifItem[]>([]);
  const [convs, setConvs] = useState<ConvItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [actionConv, setActionConv] = useState<ConvItem | null>(null);

  const loadNotifs = async (type: string) => {
    setLoading(true);
    try {
      const r = await api.notifications(type);
      setNotifs(r.items || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const loadConvs = async () => {
    setLoading(true);
    try {
      const r = await api.conversations();
      setConvs(r.items || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (tab === "dm") {
      loadConvs();
    } else {
      const t = tabs.find((x) => x.key === tab);
      if (t) loadNotifs(t.type);
    }
  }, [tab]);

  const markAll = async () => {
    try {
      await api.markNotificationsRead();
      const t = tabs.find((x) => x.key === tab);
      if (t && tab !== "dm") loadNotifs(t.type);
    } catch (e) {
      console.error(e);
    }
  };

  const openNotif = (n: NotifItem) => {
    if (n.post_id) navigate("/post/" + n.post_id);
  };

  const markConvUnread = async () => {
    if (!actionConv) return;
    try {
      await api.markConversationUnread(actionConv.user_id);
      setActionConv(null);
      loadConvs();
    } catch (e) {
      console.error(e);
    }
  };

  const hideConv = async () => {
    if (!actionConv) return;
    try {
      await api.hideConversation(actionConv.user_id);
      setActionConv(null);
      loadConvs();
    } catch (e) {
      console.error(e);
    }
  };

  const deleteConv = async () => {
    if (!actionConv) return;
    if (!window.confirm("确定删除该会话吗？")) {
      setActionConv(null);
      return;
    }
    try {
      await api.deleteConversation(actionConv.user_id);
      setActionConv(null);
      loadConvs();
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="notif-page">
      <header className="topbar">
        <button className="back" onClick={() => navigate(-1)}>←</button>
        <span className="title">消息</span>
        {tab !== "dm" && <button className="share-btn" onClick={markAll}>全部已读</button>}
      </header>
      <div className="msg-tabs">
        {tabs.map((t) => (
          <button key={t.key} className={tab === t.key ? "active" : ""} onClick={() => setTab(t.key)}>
            {t.label}
          </button>
        ))}
      </div>
      {loading ? (
        <p className="empty">加载中...</p>
      ) : tab === "dm" ? (
        convs.length === 0 ? (
          <p className="empty">还没有私信</p>
        ) : (
          <div className="notif-list">
            {convs.map((c) => (
              <div className="notif-item" key={c.user_id} onClick={() => navigate("/chat/" + c.user_id)}>
                <div className="avatar">{(c.nickname || "U").charAt(0)}</div>
                <div className="notif-main">
                  <p>
                    <b>{c.nickname}</b>
                    {c.unread > 0 ? "（" + c.unread + " 条未读）" : ""}
                  </p>
                  <p className="notif-content">{c.last_message}</p>
                </div>
                <button
                  className="more-btn"
                  onClick={(e) => {
                    e.stopPropagation();
                    setActionConv(c);
                  }}
                >
                  ⋯
                </button>
              </div>
            ))}
          </div>
        )
      ) : notifs.length === 0 ? (
        <p className="empty">暂无相关通知</p>
      ) : (
        <div className="notif-list">
          {notifs.map((n) => (
            <div className={n.read ? "notif-item read" : "notif-item"} key={n.id} onClick={() => openNotif(n)}>
              <div className="avatar">{(n.actor?.nickname || "U").charAt(0)}</div>
              <div className="notif-main">
                <p>
                  <b>{n.actor?.nickname}</b> {typeText[n.type] || n.type}
                </p>
                {n.content && <p className="notif-content">"{n.content}"</p>}
              </div>
            </div>
          ))}
        </div>
      )}
      {actionConv && (
        <div className="share-overlay" onClick={() => setActionConv(null)}>
          <div className="share-sheet" onClick={(e) => e.stopPropagation()}>
            <button onClick={markConvUnread}>标为未读</button>
            <button onClick={hideConv}>不显示</button>
            <button onClick={deleteConv}>删除</button>
            <button className="cancel" onClick={() => setActionConv(null)}>取消</button>
          </div>
        </div>
      )}
    </div>
  );
}
