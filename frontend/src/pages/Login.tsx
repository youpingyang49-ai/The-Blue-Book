import { useEffect, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import Icon from "../components/Icon";

const PHONE_RE = /^1[3-9]\d{9}$/;

export default function Login() {
  const navigate = useNavigate();
  const [params] = useSearchParams();
  const redirect = params.get("redirect") || "/";
  const setAuth = useAuth((state) => state.setAuth);
  const token = useAuth((state) => state.token);

  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [agreed, setAgreed] = useState(true);
  const [countdown, setCountdown] = useState(0);
  const [sending, setSending] = useState(false);
  const [submitting, setSubmitting] = useState(false);
  const [message, setMessage] = useState<{ text: string; kind: "ok" | "error" } | null>(null);
  const timerRef = useRef<number | null>(null);

  useEffect(() => {
    return () => {
      if (timerRef.current) window.clearInterval(timerRef.current);
    };
  }, []);

  const startCountdown = () => {
    setCountdown(60);
    if (timerRef.current) window.clearInterval(timerRef.current);
    timerRef.current = window.setInterval(() => {
      setCountdown((value) => {
        if (value <= 1) {
          if (timerRef.current) window.clearInterval(timerRef.current);
          return 0;
        }
        return value - 1;
      });
    }, 1000);
  };

  const sendCode = async () => {
    if (countdown > 0 || sending) return;
    if (!PHONE_RE.test(phone)) {
      setMessage({ text: "请输入正确的 11 位手机号", kind: "error" });
      return;
    }
    setSending(true);
    setMessage(null);
    try {
      const result = await api.sendCode(phone);
      startCountdown();
      setMessage({
        text: result?.debug_code ? `验证码已发送，开发环境验证码：${result.debug_code}` : "验证码已发送，请注意查收",
        kind: "ok",
      });
    } catch (error: any) {
      setMessage({ text: error?.message || "发送失败，请稍后重试", kind: "error" });
    } finally {
      setSending(false);
    }
  };

  const submit = async () => {
    if (submitting) return;
    if (!PHONE_RE.test(phone)) {
      setMessage({ text: "请输入正确的 11 位手机号", kind: "error" });
      return;
    }
    if (code.length !== 6) {
      setMessage({ text: "请输入 6 位短信验证码", kind: "error" });
      return;
    }
    if (!agreed) {
      setMessage({ text: "请先阅读并同意用户协议与隐私政策", kind: "error" });
      return;
    }
    setSubmitting(true);
    setMessage(null);
    try {
      const result = await api.login(phone, code);
      setAuth(result.access_token, result.user);
      navigate(redirect, { replace: true });
    } catch (error: any) {
      setMessage({ text: error?.message || "登录失败，请检查验证码", kind: "error" });
    } finally {
      setSubmitting(false);
    }
  };

  const canSubmit = PHONE_RE.test(phone) && code.length === 6 && !submitting;

  return (
    <div className="login-page">
      <section className="login-card">
        <div className="login-brand">
          <div className="login-brand-head">
            <span className="brand-mark">B</span>
            <span className="brand-text">
              <span className="brand-name">The Blue Book</span>
              <span className="brand-sub">notes &amp; life</span>
            </span>
          </div>

          <div className="login-slogan">
            <h1>
              记录灵感
              <br />
              分享<em>生活</em>
            </h1>
            <p>让每一次发现，都成为值得分享的内容。在这里收藏你的日常，也遇见别人的世界。</p>
          </div>

          <div className="login-features">
            <div className="login-feature">
              <span className="dot">
                <Icon name="grid" size={14} />
              </span>
              瀑布流发现，一眼看到喜欢的内容
            </div>
            <div className="login-feature">
              <span className="dot">
                <Icon name="edit" size={14} />
              </span>
              图文笔记，记录真实的生活质感
            </div>
            <div className="login-feature">
              <span className="dot">
                <Icon name="chat" size={14} />
              </span>
              关注、评论、私信，和有趣的人连接
            </div>
          </div>

          <div className="login-cards" aria-hidden="true">
            <div className="login-mini-card">
              <div className="login-mini-cover" />
              <div className="login-mini-line" />
              <div className="login-mini-line short" />
            </div>
            <div className="login-mini-card">
              <div className="login-mini-cover" />
              <div className="login-mini-line" />
              <div className="login-mini-line short" />
            </div>
            <div className="login-mini-card">
              <div className="login-mini-cover" />
              <div className="login-mini-line" />
              <div className="login-mini-line short" />
            </div>
          </div>
        </div>

        <div className="login-panel">
          <h2 className="login-title">手机号登录</h2>
          <p className="login-desc">未注册的手机号验证后将自动创建账号</p>

          <div className="login-form">
            <div className="input-group">
              <span className="prefix">+86</span>
              <input
                value={phone}
                inputMode="numeric"
                autoComplete="tel"
                maxLength={11}
                placeholder="请输入手机号"
                onChange={(event) => setPhone(event.target.value.replace(/\D/g, "").slice(0, 11))}
                onKeyDown={(event) => {
                  if (event.key === "Enter") submit();
                }}
              />
            </div>

            <div className="input-group">
              <input
                value={code}
                inputMode="numeric"
                autoComplete="one-time-code"
                maxLength={6}
                placeholder="请输入 6 位验证码"
                onChange={(event) => setCode(event.target.value.replace(/\D/g, "").slice(0, 6))}
                onKeyDown={(event) => {
                  if (event.key === "Enter") submit();
                }}
              />
              <button type="button" className="suffix-btn" onClick={sendCode} disabled={countdown > 0 || sending}>
                {countdown > 0 ? `${countdown}s 后重发` : sending ? "发送中…" : "获取验证码"}
              </button>
            </div>

            {message ? (
              <div className={`login-msg ${message.kind === "error" ? "error" : ""}`}>{message.text}</div>
            ) : null}

            <label className="checkbox login-terms">
              <input type="checkbox" checked={agreed} onChange={(event) => setAgreed(event.target.checked)} />
              <span>
                我已阅读并同意 <a onClick={(event) => event.preventDefault()}>《用户协议》</a> 与{" "}
                <a onClick={(event) => event.preventDefault()}>《隐私政策》</a>
              </span>
            </label>

            <button
              type="button"
              className="btn btn-primary btn-lg btn-block"
              onClick={submit}
              disabled={!canSubmit}
            >
              {submitting ? <span className="spinner" /> : null}
              {submitting ? "登录中…" : "登录"}
            </button>
          </div>

          <p className="login-foot">
            {token ? "你已登录，可继续使用当前账号" : "登录即表示你同意遵守社区规范"}
          </p>
        </div>
      </section>
    </div>
  );
}
