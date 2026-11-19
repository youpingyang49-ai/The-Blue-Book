import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";

interface Msg {
  id: number;
  from_user_id: number;
  to_user_id: number;
  content: string;
}

export default function Chat() {
  const { id } = useParams();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [msgs, setMsgs] = useState<Msg[]>([]);
  const [text, setText] = useState("");
  const [showAction, setShowAction] = useState(false);

  const load = async () => {
    try {
      const r = await api.conversation(Number(id));
      setMsgs(r.items || []);
    } catch (e) {
      console.error(e);
    }
  };

  useEffect(() => {
    load();
  }, [id]);

  const send = async () => {
    const c = text.trim();
    if (!c) return;
    try {
      await api.sendMessage(Number(id), c);
      setText("");
      load();
    } catch (e) {
      console.error(e);
    }
  };

  const clearChat = async () => {
    if (!window.confirm("确定清空聊天记录吗？")) {
      setShowAction(false);
      return;
    }
    try {
      await api.deleteConversation(Number(id));
      setShowAction(false);
      navigate("/notifications");
    } catch (e) {
      console.error(e);
    }
  };

  return (
    <div className="chat-page">
      <header className="topbar">
        <button className="back" onClick={() => navigate(-1)}>←</button>
        <span className="title">私信</span>
        <button className="share-btn" onClick={() => setShowAction(true)}>⋯</button>
      </header>
      <div className="chat-list">
        {msgs.map((m) => (
          <div key={m.id} className={m.from_user_id === user?.id ? "chat-bubble mine" : "chat-bubble"}>
            {m.content}
          </div>
        ))}
      </div>
      <div className="chat-input">
        <input
          placeholder="输入消息..."
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") send();
          }}
        />
        <button className="primary small" onClick={send}>发送</button>
      </div>
      {showAction && (
        <div className="share-overlay" onClick={() => setShowAction(false)}>
          <div className="share-sheet" onClick={(e) => e.stopPropagation()}>
            <button onClick={clearChat}>清空聊天记录</button>
            <button className="cancel" onClick={() => setShowAction(false)}>取消</button>
          </div>
        </div>
      )}
    </div>
  );
}
