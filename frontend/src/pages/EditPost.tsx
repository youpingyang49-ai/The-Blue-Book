import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";

export default function EditPost() {
  const { id } = useParams();
  const navigate = useNavigate();
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [msg, setMsg] = useState("");

  const load = async () => {
    try {
      const r = await api.postDetail(Number(id));
      setTitle(r.post.title || "");
      setContent(r.post.content || "");
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  useEffect(() => {
    load();
  }, [id]);

  const submit = async () => {
    try {
      await api.updatePost(Number(id), { title, content });
      navigate("/post/" + id);
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  return (
    <div className="create-page">
      <header className="topbar">
        <button className="back" onClick={() => navigate(-1)}>←</button>
        <span className="title">编辑笔记</span>
        <button className="primary small" onClick={submit}>保存</button>
      </header>
      <div className="form">
        <input placeholder="标题" value={title} onChange={(e) => setTitle(e.target.value)} />
        <textarea placeholder="内容..." value={content} onChange={(e) => setContent(e.target.value)} />
      </div>
      {msg && <p className="msg">{msg}</p>}
    </div>
  );
}
