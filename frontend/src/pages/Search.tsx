import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import PostCard from "../components/PostCard";
import type { PostItem } from "../components/PostCard";

export default function Search() {
  const [q, setQ] = useState("");
  const [items, setItems] = useState<PostItem[]>([]);
  const [searched, setSearched] = useState(false);
  const [loading, setLoading] = useState(false);
  const navigate = useNavigate();

  const doSearch = async () => {
    const kw = q.trim();
    if (!kw) return;
    setLoading(true);
    setSearched(true);
    try {
      const r = await api.search(kw);
      setItems(r.items || []);
    } catch (e: any) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="search-page">
      <header className="topbar">
        <button className="back" onClick={() => navigate(-1)}>←</button>
        <input
          placeholder="搜索标题、内容..."
          value={q}
          onChange={(e) => setQ(e.target.value)}
          onKeyDown={(e) => {
            if (e.key === "Enter") doSearch();
          }}
        />
        <button className="primary small" onClick={doSearch}>搜索</button>
      </header>
      {loading ? (
        <p className="empty">搜索中...</p>
      ) : searched && items.length === 0 ? (
        <p className="empty">没有找到相关笔记</p>
      ) : items.length > 0 ? (
        <div className="waterfall">
          {items.map((p) => (
            <PostCard key={p.id} post={p} />
          ))}
        </div>
      ) : (
        <p className="empty">输入关键词搜索笔记</p>
      )}
    </div>
  );
}
