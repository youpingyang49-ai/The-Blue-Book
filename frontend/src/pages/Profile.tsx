import { useNavigate } from "react-router-dom";
import { useAuth } from "../store/auth";
import TabBar from "../components/TabBar";

export default function Profile() {
  const { token, user, logout } = useAuth();
  const navigate = useNavigate();

  if (!token) {
    return (
      <div className="profile-page">
        <p className="empty">还没有登录</p>
        <button className="primary" onClick={() => navigate("/login")}>去登录</button>
        <TabBar />
      </div>
    );
  }

  return (
    <div className="profile-page">
      <div className="profile-head">
        <div className="avatar big">{(user?.nickname || "U").charAt(0)}</div>
        <div>
          <h2>{user?.nickname}</h2>
          <p style={{ color: "#999" }}>{user?.phone}</p>
        </div>
      </div>
      <button className="primary" onClick={() => { logout(); navigate("/login"); }}>退出登录</button>
      <TabBar />
    </div>
  );
}
