import { useState } from "react";
import { useNavigate } from "react-router-dom";

export default function Search() {
  const [q, setQ] = useState("");
  const navigate = useNavigate();
  return (
    <div className="search-page">
      <header className="topbar">
        <button className="back" onClick={() => navigate(-1)}>←</button>
        <input placeholder="搜索标题、内容..." value={q} onChange={(e) => setQ(e.target.value)} />
      </header>
      <p className="empty">搜索功能开发中，敬请期待</p>
    </div>
  );
}
