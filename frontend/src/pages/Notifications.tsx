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

const typeText: Record<string, string> = {
  like: "赞了你的笔记",
  favorite: "收藏了你的笔记",
  comment: "评论了你的笔记",
  follow: "关注了你",
};

export default function Notifications() {
  const navigate = useNavigate();
  const [items, setItems] = useState<NotifItem[]>([]);
  const [loading, setLoading] = useState(true);

  const load = async () => {
    try {
      const r = await api.notifications();
      setItems(r.items || []);
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    load();
  }, []);

  const markAll = async () => {
    try {
      await api.markNotificationsRead();
      load();
    } catch (e) {
      console.error(e);
    }
  };

  const openItem = (n: NotifItem) => {
    if (n.post_id) navigate("/post/" + n.post_id);
  };

  return (
    <div className="notif-page">
      <header className="topbar">
        <button className="back" onClick={() => navigate(-1)}>←</button>
        <span className="title">消息通知</span>
        <button className="share-btn" onClick={markAll}>全部已读</button>
      </header>
      {loading ? (
        <p className="empty">加载中...</p>
      ) : items.length === 0 ? (
        <p className="empty">还没有通知</p>
      ) : (
        <div className="notif-list">
          {items.map((n) => (
            <div className={n.read ? "notif-item read" : "notif-item"} key={n.id} onClick={() => openItem(n)}>
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
    </div>
  );
}
