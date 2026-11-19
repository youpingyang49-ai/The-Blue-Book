import { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { api } from "../api/client";
import AppShell from "../components/AppShell";
import { PageLoading } from "../components/States";
import { useToast } from "../components/Toast";

export default function EditPost() {
  const { id } = useParams();
  const navigate = useNavigate();
  const toast = useToast();
  const [title, setTitle] = useState("");
  const [content, setContent] = useState("");
  const [images, setImages] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    (async () => {
      try {
        const result = await api.postDetail(Number(id));
        setTitle(result.post?.title || "");
        setContent(result.post?.content || "");
        setImages(result.post?.image_urls || []);
      } catch (error: any) {
        toast(error?.message || "笔记加载失败", "error");
      } finally {
        setLoading(false);
      }
    })();
  }, [id, toast]);

  const save = async () => {
    setSaving(true);
    try {
      await api.updatePost(Number(id), { title: title.trim(), content: content.trim() });
      toast("已保存");
      navigate(`/post/${id}`);
    } catch (error: any) {
      toast(error?.message || "保存失败", "error");
    } finally {
      setSaving(false);
    }
  };

  return (
    <AppShell>
      <div className="editor-wrap">
        <div className="page-head" style={{ paddingTop: 4 }}>
          <h1>编辑笔记</h1>
          <p>图片暂不支持修改，可调整标题与正文</p>
        </div>

        {loading ? (
          <PageLoading text="正在加载笔记…" />
        ) : (
          <div className="editor-grid">
            <div>
              <div className="preview-grid" style={{ marginTop: 0, gridTemplateColumns: "repeat(auto-fill, minmax(120px, 1fr))" }}>
                {images.map((url) => (
                  <div className="preview-item" key={url}>
                    <img src={url} alt="" />
                  </div>
                ))}
              </div>
            </div>

            <div className="editor-form">
              <div className="field">
                <label className="field-label" htmlFor="edit-title">
                  标题
                </label>
                <input
                  id="edit-title"
                  className="input"
                  value={title}
                  maxLength={60}
                  onChange={(event) => setTitle(event.target.value)}
                />
              </div>
              <div className="field">
                <label className="field-label" htmlFor="edit-content">
                  正文
                </label>
                <textarea
                  id="edit-content"
                  className="textarea"
                  value={content}
                  onChange={(event) => setContent(event.target.value)}
                />
              </div>
              <div className="editor-actions">
                <button type="button" className="btn btn-ghost" onClick={() => navigate(-1)}>
                  取消
                </button>
                <button type="button" className="btn btn-primary" onClick={save} disabled={saving}>
                  {saving ? "保存中…" : "保存修改"}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </AppShell>
  );
}
