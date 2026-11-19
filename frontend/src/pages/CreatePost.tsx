import { useRef, useState } from "react";
import type { ChangeEvent } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import AppShell from "../components/AppShell";
import Icon from "../components/Icon";
import { useToast } from "../components/Toast";

interface UploadedImage {
  key: string;
  url: string;
}

export default function CreatePost() {
  const navigate = useNavigate();
  const { token } = useAuth();
  const toast = useToast();
  const fileRef = useRef<HTMLInputElement | null>(null);

  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [topicsText, setTopicsText] = useState("");
  const [topics, setTopics] = useState<string[]>([]);
  const [images, setImages] = useState<UploadedImage[]>([]);
  const [uploading, setUploading] = useState(false);
  const [publishing, setPublishing] = useState(false);

  const addTopic = (raw: string) => {
    const name = raw.replace(/^#/, "").trim();
    if (!name) return;
    setTopics((prev) => (prev.includes(name) || prev.length >= 5 ? prev : [...prev, name]));
  };

  const commitTopics = () => {
    topicsText
      .split(/[,，\s]+/)
      .filter(Boolean)
      .forEach(addTopic);
    setTopicsText("");
  };

  const onPickFiles = async (event: ChangeEvent<HTMLInputElement>) => {
    const files = Array.from(event.target.files || []);
    event.target.value = "";
    if (files.length === 0) return;
    if (images.length + files.length > 9) {
      toast("最多上传 9 张图片", "error");
      return;
    }
    setUploading(true);
    try {
      for (const file of files) {
        if (!file.type.startsWith("image/")) {
          toast(`${file.name} 不是图片文件`, "error");
          continue;
        }
        const result = await api.upload(file);
        setImages((prev) => [...prev, { key: result.key, url: result.url }]);
      }
    } catch (error: any) {
      toast(error?.message || "图片上传失败", "error");
    } finally {
      setUploading(false);
    }
  };

  const publish = async () => {
    if (publishing) return;
    if (!token) {
      navigate("/login?redirect=/create");
      return;
    }
    if (!title.trim()) {
      toast("给笔记起一个标题吧", "error");
      return;
    }
    if (!content.trim()) {
      toast("写点正文内容再发布", "error");
      return;
    }
    if (images.length === 0) {
      toast("至少上传一张图片", "error");
      return;
    }
    if (topicsText.trim()) commitTopics();
    setPublishing(true);
    try {
      const result = await api.createPost({
        title: title.trim(),
        content: content.trim(),
        image_keys: images.map((image) => image.key),
        topics,
      });
      toast("发布成功");
      navigate(result?.post_id ? `/post/${result.post_id}` : "/");
    } catch (error: any) {
      toast(error?.message || "发布失败", "error");
    } finally {
      setPublishing(false);
    }
  };

  return (
    <AppShell>
      <div className="editor-wrap">
        <div className="page-head" style={{ paddingTop: 4 }}>
          <h1>发布笔记</h1>
          <p>分享你的生活片段，让更多人看到</p>
        </div>

        <div className="editor-grid">
          <div>
            {images.length === 0 ? (
              <div className="uploader" onClick={() => fileRef.current?.click()}>
                <span className="upload-icon">
                  <Icon name="image" size={22} />
                </span>
                <strong>{uploading ? "正在上传…" : "添加图片"}</strong>
                <span>支持多选，最多 9 张，建议竖版构图</span>
              </div>
            ) : (
              <div
                className="uploader"
                style={{ padding: 14, cursor: "default", minHeight: 0, alignItems: "stretch" }}
                onClick={(event) => event.stopPropagation()}
              >
                <div className="preview-grid" style={{ marginTop: 0 }}>
                  {images.map((image) => (
                    <div className="preview-item" key={image.key}>
                      <img src={image.url} alt="" />
                      <button
                        type="button"
                        className="remove"
                        aria-label="移除图片"
                        onClick={() => setImages((prev) => prev.filter((item) => item.key !== image.key))}
                      >
                        <Icon name="close" size={13} />
                      </button>
                    </div>
                  ))}
                  {uploading ? (
                    <div className="preview-item uploading">
                      <span className="spinner" />
                    </div>
                  ) : null}
                  {images.length < 9 ? (
                    <button
                      type="button"
                      className="preview-item"
                      style={{ display: "grid", placeItems: "center", color: "var(--text-3)", background: "#fff", border: "1px dashed var(--border-strong)" }}
                      onClick={() => fileRef.current?.click()}
                      aria-label="继续添加图片"
                    >
                      <Icon name="plus" size={22} />
                    </button>
                  ) : null}
                </div>
              </div>
            )}
            <input ref={fileRef} type="file" accept="image/*" multiple hidden onChange={onPickFiles} />
            <p className="field-hint" style={{ marginTop: 8 }}>
              图片将按上传顺序展示，第一张作为封面
            </p>
          </div>

          <div className="editor-form">
            <div className="field">
              <label className="field-label" htmlFor="post-title">
                标题
              </label>
              <input
                id="post-title"
                className="input"
                value={title}
                maxLength={60}
                onChange={(event) => setTitle(event.target.value)}
                placeholder="给这篇笔记起个标题"
              />
              <span className="field-hint">{title.length}/60</span>
            </div>

            <div className="field">
              <label className="field-label" htmlFor="post-content">
                正文
              </label>
              <textarea
                id="post-content"
                className="textarea"
                value={content}
                onChange={(event) => setContent(event.target.value)}
                placeholder="记录此刻的想法、经验或者心情…"
              />
            </div>

            <div className="field">
              <label className="field-label" htmlFor="post-topics">
                话题
              </label>
              <input
                id="post-topics"
                className="input"
                value={topicsText}
                onChange={(event) => setTopicsText(event.target.value)}
                onBlur={commitTopics}
                onKeyDown={(event) => {
                  if (event.key === "Enter" || event.key === " ") {
                    event.preventDefault();
                    commitTopics();
                  }
                }}
                placeholder="输入话题后回车添加，最多 5 个，例如：美食 探店"
              />
              {topics.length > 0 ? (
                <div style={{ display: "flex", flexWrap: "wrap", gap: 6, marginTop: 4 }}>
                  {topics.map((topic) => (
                    <span className="tag" key={topic}>
                      #{topic}
                      <button
                        type="button"
                        aria-label={`移除话题 ${topic}`}
                        style={{ marginLeft: 6, display: "flex" }}
                        onClick={() => setTopics((prev) => prev.filter((item) => item !== topic))}
                      >
                        <Icon name="close" size={12} />
                      </button>
                    </span>
                  ))}
                </div>
              ) : null}
            </div>

            <div className="editor-actions">
              <button type="button" className="btn btn-ghost" onClick={() => navigate(-1)}>
                取消
              </button>
              <button type="button" className="btn btn-primary" onClick={publish} disabled={publishing || uploading}>
                {publishing ? <span className="spinner" /> : null}
                {publishing ? "发布中…" : "发布笔记"}
              </button>
            </div>
          </div>
        </div>
      </div>
    </AppShell>
  );
}
