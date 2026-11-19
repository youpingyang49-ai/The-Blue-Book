const BASE = "/api/v1";

function getToken(): string | null {
  return localStorage.getItem("token");
}

async function request(path: string, options: RequestInit = {}): Promise<any> {
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
  notifications: (page = 1) => request("/notifications?page=" + page + "&page_size=20"),
  unreadCount: () => request("/notifications/unread-count"),
  markNotificationsRead: () => request("/notifications/read", { method: "POST" }),
};
