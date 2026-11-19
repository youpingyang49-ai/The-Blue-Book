import { useEffect, useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Icon from "./Icon";

interface HeaderSearchProps {
  initialValue?: string;
}

export default function HeaderSearch({ initialValue = "" }: HeaderSearchProps) {
  const navigate = useNavigate();
  const location = useLocation();
  const params = new URLSearchParams(location.search);
  const queryParam = params.get("q") || "";
  const [value, setValue] = useState(initialValue || queryParam);

  useEffect(() => {
    if (location.pathname === "/search") setValue(queryParam);
  }, [location.pathname, queryParam]);

  const submit = () => {
    const keyword = value.trim();
    if (!keyword) {
      navigate("/search");
      return;
    }
    navigate(`/search?q=${encodeURIComponent(keyword)}`);
  };

  return (
    <div className="search-pill">
      <Icon name="search" size={18} />
      <input
        value={value}
        onChange={(event) => setValue(event.target.value)}
        onKeyDown={(event) => {
          if (event.key === "Enter") submit();
        }}
        placeholder="搜索你感兴趣的笔记、内容和话题…"
        aria-label="搜索笔记"
      />
      <button type="button" className="search-submit" onClick={submit}>
        搜索
      </button>
    </div>
  );
}
