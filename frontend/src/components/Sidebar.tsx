import { Link, useLocation, useNavigate } from "react-router-dom";
import Icon from "./Icon";
import type { IconName } from "./Icon";
import Avatar from "./Avatar";
import { useAuth } from "../store/auth";

interface SidebarProps {
  unread: number;
  onAbout: () => void;
}

interface NavEntry {
  to: string;
  label: string;
  icon: IconName;
  exact?: boolean;
  badge?: number;
  tab?: string;
}

export default function Sidebar({ unread, onAbout }: SidebarProps) {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, token, logout } = useAuth();

  const params = new URLSearchParams(location.search);
  const dmActive = location.pathname === "/notifications" && params.get("tab") === "dm";

  const entries: NavEntry[] = [
    { to: "/", label: "首页", icon: "home", exact: true },
    { to: "/search", label: "发现", icon: "compass" },
    { to: "/notifications", label: "消息", icon: "bell", badge: unread },
    { to: "/notifications?tab=dm", label: "私信", icon: "chat", tab: "dm" },
    { to: "/profile", label: "我的", icon: "user" },
  ];

  const isActive = (entry: NavEntry) => {
    if (entry.tab === "dm") return dmActive;
    if (entry.exact) return location.pathname === entry.to;
    if (entry.to === "/notifications") return location.pathname === "/notifications" && !dmActive;
    return location.pathname.startsWith(entry.to);
  };

  return (
    <aside className="sidebar">
      <Link className="sidebar-brand" to="/">
        <span className="brand-mark">B</span>
        <span className="brand-text">
          <span className="brand-name">The Blue Book</span>
          <span className="brand-sub">notes &amp; life</span>
        </span>
      </Link>

      <nav className="nav">
        {entries.slice(0, 2).map((entry) => (
          <Link key={entry.to} to={entry.to} className={`nav-item ${isActive(entry) ? "active" : ""}`}>
            <Icon name={entry.icon} size={20} />
            {entry.label}
          </Link>
        ))}

        <button type="button" className="nav-publish" onClick={() => navigate("/create")}>
          <Icon name="plus" size={18} strokeWidth={2.2} />
          发布笔记
        </button>

        {entries.slice(2).map((entry) => (
          <Link key={entry.to} to={entry.to} className={`nav-item ${isActive(entry) ? "active" : ""}`}>
            <Icon name={entry.icon} size={20} />
            {entry.label}
            {entry.badge && entry.badge > 0 ? (
              <span className="nav-badge">{entry.badge > 99 ? "99+" : entry.badge}</span>
            ) : null}
          </Link>
        ))}
      </nav>

      <div className="sidebar-spacer" />

      {token && user ? (
        <div className="sidebar-user" onClick={() => navigate("/profile")}>
          <Avatar src={user.avatar_url} name={user.nickname} size="sm" />
          <div className="sidebar-user-info">
            <div className="sidebar-user-name">{user.nickname}</div>
            <div className="sidebar-user-sub">查看个人主页</div>
          </div>
        </div>
      ) : (
        <button className="btn btn-primary btn-block btn-sm" onClick={() => navigate("/login")}>
          登录 / 注册
        </button>
      )}

      <div className="sidebar-foot">
        <button type="button" onClick={onAbout}>
          关于
        </button>
        {token ? (
          <button
            type="button"
            onClick={() => {
              logout();
              navigate("/login");
            }}
          >
            退出登录
          </button>
        ) : (
          <button type="button" onClick={() => navigate("/login")}>
            登录
          </button>
        )}
      </div>
    </aside>
  );
}
