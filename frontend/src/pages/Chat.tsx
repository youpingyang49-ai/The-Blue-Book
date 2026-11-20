import { useCallback, useEffect, useRef, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import AppShell from "../components/AppShell";
import Avatar from "../components/Avatar";
import Icon from "../components/Icon";
import { EmptyState, PageLoading } from "../components/States";
import { useToast } from "../components/Toast";

interface Msg {
  id: number;
  from_user_id: number;
  to_user_id: number;
  content: string;
  created_at?: string;
}

interface ConvItem {
  user_id: number;
  nickname: string;
  avatar_url: string;
  last_message: string;
  unread: number;
}

export default function Chat() {
  const { id } = useParams();
  const peerId = Number(id);
  const navigate = useNavigate();
  const { user, token } = useAuth();
  const toast = useToast();

  const [msgs, setMsgs] = useState<Msg[]>([]);
  const [convs, setConvs] = useState<ConvItem[]>([]);
  const [peer, setPeer] = useState<any>(null);
  const [text, setText] = useState("");
  const [loading, setLoading] = useState(true);
  const [sending, setSending] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const bottomRef = useRef<HTMLDivElement | null>(null);

  const loadMsgs = useCallback(async () => {
    try {
      const result = await api.conversation(peerId);
      setMsgs(result.items || []);
      // 会话已读后刷新侧栏未读角标
      window.dispatchEvent(new Event("notifications-updated"));
    } catch (err: any) {
      toast(err?.message || "会话加载失败", "error");
    } finally {
      setLoading(false);
    }
  }, [peerId, toast]);

  const loadConvs = useCallback(async () => {
    try {
      const result = await api.conversations();
      setConvs(result.items || []);
    } catch {
      setConvs([]);
    }
  }, []);

  const loadPeer = useCallback(async () => {
    try {
      const result = await api.getUser(peerId);
      setPeer(result.user);
    } catch {
      setPeer(null);
    }
  }, [peerId]);

  useEffect(() => {
    setLoading(true);
    loadMsgs();
    loadConvs();
    loadPeer();
  }, [loadMsgs, loadConvs, loadPeer]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [msgs]);

  const send = async () => {
    const content = text.trim();
    if (!content || sending) return;
    setSending(true);
    try {
      await api.sendMessage(peerId, content);
      setText("");
      await loadMsgs();
      loadConvs();
    } catch (err: any) {
      toast(err?.message || "发送失败", "error");
    } finally {
      setSending(false);
    }
  };

  const clearChat = async () => {
    if (!window.confirm("确定清空与该用户的聊天记录吗？")) return;
    try {
      await api.deleteConversation(peerId);
      toast("聊天记录已清空");
      navigate("/notifications?tab=dm");
    } catch (err: any) {
      toast(err?.message || "操作失败", "error");
    }
  };

  if (!token) {
    return (
      <AppShell>
        <EmptyState
          icon="chat"
          title="登录后查看私信"
          action={
            <button className="btn btn-primary btn-sm" onClick={() => navigate("/login?redirect=/notifications?tab=dm")}>
              去登录
            </button>
          }
        />
      </AppShell>
    );
  }

  const peerName = peer?.nickname || "对方";
  const peerAvatar = peer?.avatar_url;

  return (
    <AppShell>
      <div className="chat-shell fade-up">
        <aside className="chat-side">
          <div className="chat-side-head">
            <Icon name="chat" size={18} />
            私信
          </div>
          <div className="chat-side-list">
            {convs.length === 0 ? (
              <p className="muted" style={{ padding: "10px 12px", fontSize: 12.5 }}>
                还没有会话
              </p>
            ) : (
              convs.map((conv) => (
                <div
                  key={conv.user_id}
                  className={`chat-side-item ${conv.user_id === peerId ? "active" : ""}`}
                  onClick={() => navigate(`/chat/${conv.user_id}`)}
                >
                  <Avatar src={conv.avatar_url} name={conv.nickname} size="sm" />
                  <div className="meta">
                    <div className="name">{conv.nickname}</div>
                    <div className="last">{conv.last_message}</div>
                  </div>
                  {conv.unread > 0 ? <span className="conv-unread">{conv.unread}</span> : null}
                </div>
              ))
            )}
          </div>
        </aside>

        <section className="chat-main">
          <header className="chat-head">
            <button type="button" className="icon-btn" onClick={() => navigate("/notifications?tab=dm")} aria-label="返回">
              <Icon name="back" size={20} />
            </button>
            <Avatar src={peerAvatar} name={peerName} size="sm" />
            <span className="name" onClick={() => navigate(`/user/${peerId}`)} style={{ cursor: "pointer" }}>
              {peerName}
            </span>
            <span className="spacer" />
            <button type="button" className="icon-btn" onClick={() => setMenuOpen(true)} aria-label="会话设置">
              <Icon name="more" size={20} />
            </button>
          </header>

          <div className="chat-body">
            {loading ? (
              <PageLoading text="正在加载会话…" />
            ) : msgs.length === 0 ? (
              <EmptyState icon="chat" title="还没有聊天记录" description={`给 ${peerName} 发第一条消息吧`} />
            ) : (
              msgs.map((msg) => (
                <div key={msg.id} className={msg.from_user_id === user?.id ? "chat-bubble mine" : "chat-bubble"}>
                  {msg.content}
                </div>
              ))
            )}
            <div ref={bottomRef} />
          </div>

          <div className="chat-input-bar">
            <input
              className="input"
              value={text}
              onChange={(event) => setText(event.target.value)}
              onKeyDown={(event) => {
                if (event.key === "Enter") send();
              }}
              placeholder={`发消息给 ${peerName}…`}
            />
            <button type="button" className="btn btn-primary btn-sm" onClick={send} disabled={!text.trim() || sending}>
              {sending ? "发送中" : "发送"}
            </button>
          </div>
        </section>
      </div>

      {menuOpen ? (
        <div className="sheet-overlay" onClick={() => setMenuOpen(false)}>
          <div className="sheet" onClick={(event) => event.stopPropagation()}>
            <button type="button" className="sheet-item" onClick={() => navigate(`/user/${peerId}`)}>
              查看 TA 的主页
            </button>
            <button type="button" className="sheet-item danger" onClick={clearChat}>
              清空聊天记录
            </button>
            <button type="button" className="sheet-cancel" onClick={() => setMenuOpen(false)}>
              取消
            </button>
          </div>
        </div>
      ) : null}
    </AppShell>
  );
}
