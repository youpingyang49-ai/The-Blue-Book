import { useState } from "react";
import type { ChangeEvent } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import TabBar from "../components/TabBar";

interface UploadedImage {
  key: string;
  url: string;
}

export default function CreatePost() {
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [images, setImages] = useState<UploadedImage[]>([]);
  const [uploading, setUploading] = useState(false);
  const [msg, setMsg] = useState("");
  const [topicsText, setTopicsText] = useState("");
  const navigate = useNavigate();

  const onPickFiles = async (e: ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(e.target.files || []);
    if (files.length === 0) return;
    setUploading(true);
    setMsg("");
    try {
      for (const f of files) {
        const r = await api.upload(f);
        setImages((prev) => [...prev, { key: r.key, url: r.url }]);
      }
    } catch (err: any) {
      setMsg(err.message);
    } finally {
      setUploading(false);
    }
  };

  const submit = async () => {
    if (images.length === 0) {
      setMsg("至少上传一张图片");
      return;
    }
    try {
      const topics = topicsText.split(/[，, ]+/).filter((t) => t.length > 0);
      await api.createPost({ title, content, image_keys: images.map((i) => i.key), topics });
      navigate("/");
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  return (
    <div className="create-page">
      <header className="topbar">
        <span className="title">发布笔记</span>
        <button className="primary small" onClick={submit} disabled={uploading}>发布</button>
      </header>
      <div className="form">
        <input placeholder="标题" value={title} onChange={(e) => setTitle(e.target.value)} />
        <textarea placeholder="分享你的内容..." value={content} onChange={(e) => setContent(e.target.value)} />
        <input placeholder="话题（逗号或空格分隔，如：美食 探店）" value={topicsText} onChange={(e) => setTopicsText(e.target.value)} />
      </div>
      <label className="upload-btn">
        {uploading ? "上传中..." : "选择图片（可多选）"}
        <input type="file" accept="image/*" multiple hidden onChange={onPickFiles} />
      </label>
      <div className="preview-grid">
        {images.map((img) => (
          <img key={img.key} src={img.url} alt="" />
        ))}
      </div>
      {msg && <p className="msg">{msg}</p>}
      <TabBar />
    </div>
  );
}
