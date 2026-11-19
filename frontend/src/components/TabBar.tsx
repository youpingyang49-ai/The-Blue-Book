import { Link, useLocation } from "react-router-dom";

export default function TabBar() {
  const loc = useLocation();
  return (
    <nav className="tabbar">
      <Link className={loc.pathname === "/" ? "active" : ""} to="/">首页</Link>
      <Link className={loc.pathname === "/notifications" ? "active" : ""} to="/notifications">消息</Link>
      <Link className={loc.pathname === "/create" ? "active" : ""} to="/create">＋ 发布</Link>
      <Link className={loc.pathname === "/profile" ? "active" : ""} to="/profile">我的</Link>
    </nav>
  );
}
