import { useEffect, useState } from "react";
import type { ReactNode } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import HeaderSearch from "./HeaderSearch";
import Icon from "./Icon";
import type { IconName } from "./Icon";
import Sidebar from "./Sidebar";

export default function AppShell({ children }: { children: ReactNode }) {
  const { token } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();
  const [unread, setUnread] = useState(0);
  const [aboutOpen, setAboutOpen] = useState(false);

  useEffect(() => {
    if (!token) {
      setUnread(0);
      return;
    }
    let alive = true;
    const load = async () => {
      try {
        const result: any = await api.unreadCount();
        if (!alive) return;
        const value = result?.count ?? result?.unread_count ?? result?.unread ?? 0;
        setUnread(Number(value) || 0);
      } catch {
        /* silent: badge is optional */
      }
    };
    load();
    const timer = window.setInterval(load, 30000);
    return () => {
      alive = false;
      window.clearInterval(timer);
    };
  }, [token, location.pathname]);

  const mobileNav: Array<{
    to: string;
    label: string;
    icon: IconName;
    exact?: boolean;
    publish?: boolean;
    badge?: number;
  }> = [
    { to: "/", label: "首页", icon: "home", exact: true },
    { to: "/search", label: "发现", icon: "compass" },
    { to: "/create", label: "发布", icon: "plus", publish: true },
    { to: "/notifications", label: "消息", icon: "bell", badge: unread },
    { to: token ? "/profile" : "/login", label: "我的", icon: "user" },
  ];

  return (
    <div className="app-shell">
      <Sidebar unread={unread} onAbout={() => setAboutOpen(true)} />

      <div className="app-main">
        <header className="app-header">
          <div className="header-inner">
            <span className="mobile-brand" aria-hidden="true">
              B
            </span>
            <HeaderSearch />
          </div>
        </header>

        <main className="app-content">{children}</main>
      </div>

      <nav className="bottom-nav" aria-label="底部导航">
        {mobileNav.map((item) => {
          const active = item.exact ? location.pathname === item.to : location.pathname.startsWith(item.to);
          return (
            <button
              key={item.label}
              type="button"
              className={`bottom-nav-item ${active ? "active" : ""} ${item.publish ? "publish" : ""}`}
              onClick={() => navigate(item.to)}
            >
              {item.publish ? (
                <span className="plus-circle">
                  <Icon name="plus" size={18} strokeWidth={2.4} />
                </span>
              ) : (
                <Icon name={item.icon} size={22} />
              )}
              <span>{item.label}</span>
              {item.badge && item.badge > 0 ? (
                <span className="nav-badge" style={{ position: "absolute", transform: "translate(14px,-4px)" }}>
                  {item.badge > 99 ? "99+" : item.badge}
                </span>
              ) : null}
            </button>
          );
        })}
      </nav>

      {aboutOpen ? (
        <div className="overlay" onClick={() => setAboutOpen(false)}>
          <div className="modal" onClick={(event) => event.stopPropagation()}>
            <div className="modal-title">关于 The Blue Book</div>
            <div className="modal-text">
              一个以图文笔记为核心的轻量内容社区。记录灵感、分享生活，让每一次发现都值得分享。
            </div>
            <div className="about-list">
              <div className="about-row">
                <span>前端</span>
                <b>React 19 · TypeScript · Vite</b>
              </div>
              <div className="about-row">
                <span>后端</span>
                <b>Go · Gin · GORM</b>
              </div>
              <div className="about-row">
                <span>基础设施</span>
                <b>MySQL · Redis · RabbitMQ</b>
              </div>
              <div className="about-row">
                <span>版本</span>
                <b>v1.0.0</b>
              </div>
            </div>
            <div className="modal-actions">
              <button type="button" className="btn btn-primary btn-sm" onClick={() => setAboutOpen(false)}>
                知道了
              </button>
            </div>
          </div>
        </div>
      ) : null}
    </div>
  );
}
