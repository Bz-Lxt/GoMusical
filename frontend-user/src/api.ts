const API = import.meta.env.VITE_API_BASE || "/api";

let csrf = "";

export function setCsrf(t: string) {
  csrf = t || "";
}

export async function req<T = any>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  if (init.body && !(init.body instanceof FormData) && !headers.has("Content-Type")) {
    headers.set("Content-Type", "application/json");
  }
  if (csrf && init.method && init.method !== "GET") {
    headers.set("X-CSRF-Token", csrf);
  }
  const res = await fetch(API + path.replace(/^\/api/, ""), {
    ...init,
    headers,
    credentials: "include",
  });
  const text = await res.text();
  let body: any = {};
  try {
    body = text ? JSON.parse(text) : {};
  } catch {
    body = { raw: text };
  }
  if (!res.ok) {
    const msg = body?.error?.message || body?.message || res.statusText;
    throw new Error(msg);
  }
  return (body.data ?? body) as T;
}

export const api = {
  me: () => req("/auth/me"),
  login: (email: string, password: string) =>
    req("/auth/login", { method: "POST", body: JSON.stringify({ email, password }) }),
  register: (payload: any) => req("/auth/register", { method: "POST", body: JSON.stringify(payload) }),
  logout: () => req("/auth/logout", { method: "POST" }),
  tracks: (q = "") => req("/tracks" + q),
  track: (id: string) => req("/tracks/" + id),
  peaks: (id: string) => fetch(API + "/tracks/" + id + "/peaks", { credentials: "include" }).then((r) => r.json()),
  comments: (id: string) => req("/tracks/" + id + "/comments"),
  addComment: (id: string, timestampMs: number, body: string) =>
    req("/tracks/" + id + "/comments", { method: "POST", body: JSON.stringify({ timestampMs, body }) }),
  openStream: (id: string) => req("/stream/" + id + "/open"),
  sponsor: (id: string, amountCents: number) =>
    req("/tracks/" + id + "/sponsor", { method: "POST", body: JSON.stringify({ amountCents }) }),
  subscribe: (id: string, amountCents: number) =>
    req("/creators/" + id + "/subscribe", { method: "POST", body: JSON.stringify({ amountCents }) }),
  ticket: (id: string) => req("/tracks/" + id + "/ticket", { method: "POST" }),
  ticketStatus: (nonce: string) => req("/tickets/" + nonce),
  creators: () => req("/creators"),
  creator: (id: string) => req("/creators/" + id),
  albums: (q = "") => req("/albums" + q),
  createAlbum: (payload: any) => req("/albums", { method: "POST", body: JSON.stringify(payload) }),
  patchTrack: (id: string, payload: any) => req("/tracks/" + id, { method: "PATCH", body: JSON.stringify(payload) }),
  initUpload: (payload: any) => req("/uploads", { method: "POST", body: JSON.stringify(payload) }),
  completeUpload: (id: string, title: string) =>
    req("/uploads/" + id + "/complete", { method: "POST", body: JSON.stringify({ title }) }),
  jobs: () => req("/jobs"),
  retry: (id: string) => req("/tracks/" + id + "/transcode", { method: "POST" }),
  myOrders: () => req("/me/orders"),
  creatorOrders: () => req("/creator/orders"),
  admin: () => req("/admin/stats"),
  payCallback: (payload: any) => req("/pay/callback", { method: "POST", body: JSON.stringify(payload) }),
  modComment: (id: string, payload: any) => req("/comments/" + id, { method: "PATCH", body: JSON.stringify(payload) }),
};

export async function uploadChunk(uploadId: string, index: number, blob: Blob) {
  const headers = new Headers();
  if (csrf) headers.set("X-CSRF-Token", csrf);
  const res = await fetch(API + "/uploads/" + uploadId + "/chunk?index=" + index, {
    method: "PUT",
    credentials: "include",
    headers,
    body: blob,
  });
  if (!res.ok) throw new Error("分片上传失败");
}

export function yuan(cents: number) {
  return "¥" + (cents / 100).toFixed(2);
}

export function fmtTime(ms: number) {
  const s = Math.max(0, Math.floor(ms / 1000));
  const m = Math.floor(s / 60);
  const r = s % 60;
  return `${m}:${r.toString().padStart(2, "0")}`;
}
