import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";

export default function Login() {
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [msg, setMsg] = useState("");
  const [countdown, setCountdown] = useState(0);
  const navigate = useNavigate();
  const setAuth = useAuth((s) => s.setAuth);

  const sendCode = async () => {
    if (!/^1[3-9]\d{9}$/.test(phone)) {
      setMsg("手机号格式不正确");
      return;
    }
    try {
      const r = await api.sendCode(phone);
      setMsg(r.debug_code ? "验证码已发送（开发环境：" + r.debug_code + "）" : "验证码已发送");
      setCountdown(60);
      const timer = setInterval(() => {
        setCountdown((c) => {
          if (c <= 1) clearInterval(timer);
          return c - 1;
        });
      }, 1000);
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  const login = async () => {
    if (!phone || !code) {
      setMsg("请输入手机号和验证码");
      return;
    }
    try {
      const r = await api.login(phone, code);
      setAuth(r.access_token, r.user);
      navigate("/");
    } catch (e: any) {
      setMsg(e.message);
    }
  };

  return (
    <div className="login-page">
      <h1>The Blue Book</h1>
      <p className="subtitle">小红书类轻量社区</p>
      <input
        placeholder="手机号"
        value={phone}
        maxLength={11}
        onChange={(e) => setPhone(e.target.value.replace(/\D/g, ""))}
      />
      <div className="code-row">
        <input
          placeholder="验证码"
          value={code}
          maxLength={6}
          onChange={(e) => setCode(e.target.value.replace(/\D/g, ""))}
        />
        <button onClick={sendCode} disabled={countdown > 0}>
          {countdown > 0 ? countdown + "s" : "发送验证码"}
        </button>
      </div>
      <button className="primary" onClick={login}>登录</button>
      {msg && <p className="msg">{msg}</p>}
    </div>
  );
}
