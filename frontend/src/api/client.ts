import { useAuth } from "../store/auth";

const BASE = "/api/v1";

function getToken(): string | null {
  return localStorage.getItem("token");
}

function getRefreshToken(): string | null {
  return localStorage.getItem("refresh_token");
}

// 同一时刻只发起一次刷新，避免并发 401 触发多次刷新
let refreshInFlight: Promise<boolean> | null = null;

async function refreshAccessToken(): Promise<boolean> {
  const refreshToken = getRefreshToken();
  if (!refreshToken) return false;
  if (!refreshInFlight) {
    refreshInFlight = (async () => {
      try {
        const res = await fetch(BASE + "/auth/refresh", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ refresh_token: refreshToken }),
        });
        if (!res.ok) return false;
        const data = await res.json();
        if (!data?.access_token) return false;
        localStorage.setItem("token", data.access_token);
        if (data.refresh_token) localStorage.setItem("refresh_token", data.refresh_token);
        useAuth.setState({ token: data.access_token, user: data.user ?? useAuth.getState().user });
        return true;
      } catch {
        return false;
      } finally {
        refreshInFlight = null;
      }
    })();
  }
  return refreshInFlight;
}

function forceReLogin() {
  useAuth.getState().logout();
  const path = window.location.pathname + window.location.search;
  if (!path.startsWith("/login")) {
    window.location.href = "/login?redirect=" + encodeURIComponent(path);
  }
}

async function request(path: string, options: RequestInit = {}, allowRetry = true): Promise<any> {
  const token = getToken();
  const headers: Record<string, string> = {};
  if (options.headers) {
    Object.assign(headers, options.headers);
  }
  if (token) {
    headers["Authorization"] = "Bearer " + token;
  }
  if (options.body && !(options.body instanceof FormData)) {
    headers["Content-Type"] = "application/json";
  }
  const res = await fetch(BASE + path, { ...options, headers });

  // access token 过期：静默刷新后重试一次
  if (res.status === 401 && allowRetry && !path.startsWith("/auth/")) {
    const refreshed = await refreshAccessToken();
    if (refreshed) return request(path, options, false);
    forceReLogin();
    throw new Error("登录已过期，请重新登录");
  }

  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: "请求失败" }));
    throw new Error(err.error || "请求失败");
  }
  return res.json();
}

export const api = {
  sendCode: (phone: string) =>
    request("/auth/send-code", { method: "POST", body: JSON.stringify({ phone }) }),
  login: (phone: string, code: string) =>
    request("/auth/login", { method: "POST", body: JSON.stringify({ phone, code }) }),
  me: () => request("/users/me"),
  feed: (sort = "latest", page = 1) =>
    request("/posts?sort=" + sort + "&page=" + page + "&page_size=20"),
  postDetail: (id: number) => request("/posts/" + id),
  createPost: (data: { title: string; content: string; image_keys: string[]; topics?: string[] }) =>
    request("/posts", { method: "POST", body: JSON.stringify(data) }),
  updatePost: (id: number, data: { title: string; content: string }) =>
    request("/posts/" + id, { method: "PUT", body: JSON.stringify(data) }),
  deletePost: (id: number) => request("/posts/" + id, { method: "DELETE" }),
  like: (id: number) => request("/posts/" + id + "/like", { method: "POST" }),
  unlike: (id: number) => request("/posts/" + id + "/like", { method: "DELETE" }),
  upload: (file: File) => {
    const fd = new FormData();
    fd.append("file", file);
    return request("/upload", { method: "POST", body: fd });
  },
  search: (q: string, page = 1) =>
    request("/search?q=" + encodeURIComponent(q) + "&page=" + page + "&page_size=20"),
  myPosts: (page = 1) => request("/users/me/posts?page=" + page + "&page_size=20"),
  myLikes: (page = 1) => request("/users/me/likes?page=" + page + "&page_size=20"),
  myFavorites: (page = 1) => request("/users/me/favorites?page=" + page + "&page_size=20"),
  favorite: (id: number) => request("/posts/" + id + "/favorite", { method: "POST" }),
  unfavorite: (id: number) => request("/posts/" + id + "/favorite", { method: "DELETE" }),
  listComments: (id: number, page = 1) => request("/posts/" + id + "/comments?page=" + page + "&page_size=20"),
  createComment: (id: number, content: string, parentId = 0) =>
    request("/posts/" + id + "/comments", { method: "POST", body: JSON.stringify({ content, parent_id: parentId }) }),
  follow: (id: number) => request("/users/" + id + "/follow", { method: "POST" }),
  unfollow: (id: number) => request("/users/" + id + "/follow", { method: "DELETE" }),
  getUser: (id: number) => request("/users/" + id),
  userPosts: (id: number, page = 1) => request("/users/" + id + "/posts?page=" + page + "&page_size=20"),
  following: (id: number, page = 1) => request("/users/" + id + "/following?page=" + page + "&page_size=20"),
  followers: (id: number, page = 1) => request("/users/" + id + "/followers?page=" + page + "&page_size=20"),
  notifications: (type = "", page = 1) => request("/notifications?type=" + type + "&page=" + page + "&page_size=20"),
  conversations: () => request("/conversations"),
  conversation: (id: number) => request("/messages/" + id),
  sendMessage: (id: number, content: string) => request("/messages/" + id, { method: "POST", body: JSON.stringify({ content }) }),
  markConversationUnread: (id: number) => request("/messages/" + id + "/unread", { method: "POST" }),
  deleteConversation: (id: number) => request("/messages/" + id, { method: "DELETE" }),
  hideConversation: (id: number) => request("/conversations/" + id + "/hide", { method: "POST" }),
  unreadCount: () => request("/notifications/unread-count"),
  markNotificationsRead: () => request("/notifications/read", { method: "POST" }),
};
