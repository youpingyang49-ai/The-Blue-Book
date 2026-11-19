import { useEffect, useRef, useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { api } from "../api/client";
import { useAuth } from "../store/auth";
import Avatar from "../components/Avatar";
import type { PostItem } from "../components/PostCard";

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
  const [previews, setPreviews] = useState<PostItem[]>([]);
  const timerRef = useRef<number | null>(null);

  useEffect(() => {
    return () => {
      if (timerRef.current) window.clearInterval(timerRef.current);
    };
  }, []);

  // 左侧展示社区真实热门笔记
  useEffect(() => {
    let alive = true;
    (async () => {
      try {
        const result = await api.feed("hot", 1);
        if (!alive) return;
        setPreviews((result.items || []).slice(0, 3));
      } catch {
        if (alive) setPreviews([]);
      }
    })();
    return () => {
      alive = false;
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
              把生活里
              <br />
              <em>值得记住的瞬间</em>
              <br />
              都留在这里
            </h1>
            <p>记录一顿饭、一次远行、一段心情。在 The Blue Book，遇见和你一样认真生活的人。</p>
          </div>

          <div className="login-showcase">
            {previews.length > 0
              ? previews.map((post) => (
                  <figure key={post.id}>
                    <img src={post.cover_image_url} alt={post.title} loading="lazy" />
                    <figcaption>
                      <div className="showcase-title">{post.title}</div>
                      <div className="showcase-author">@{post.author?.nickname}</div>
                    </figcaption>
                  </figure>
                ))
              : Array.from({ length: 3 }).map((_, index) => (
                  <figure key={index} aria-hidden="true">
                    <div style={{ width: "100%", height: "100%", background: "rgba(255,255,255,.14)" }} />
                  </figure>
                ))}
          </div>

          <div className="login-creators">
            <div className="avatar-stack">
              {previews.slice(0, 3).map((post) => (
                <Avatar
                  key={post.id}
                  src={post.author?.avatar_url}
                  name={post.author?.nickname}
                  size="xs"
                />
              ))}
            </div>
            <span>今日社区精选 · 来自真实创作者</span>
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
