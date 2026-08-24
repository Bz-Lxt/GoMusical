import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { api, yuan } from "../../frontend-user/src/api";
import { StudioUpload } from "../../frontend-user/src/StudioUpload";

export function CreatorHome({ user, push }: any) {
  const [tab, setTab] = useState<"shelf" | "upload" | "inbox" | "orders" | "jobs">("shelf");
  const [tracks, setTracks] = useState<any[]>([]);
  const [orders, setOrders] = useState<any[]>([]);
  const [jobs, setJobs] = useState<any[]>([]);
  const [comments, setComments] = useState<any[]>([]);
  const reload = () => {
    api.tracks("?creatorId=" + user.id).then((d) => setTracks(d.items || []));
    api.creatorOrders().then((d) => setOrders(d.items || [])).catch(() => {});
    api.jobs().then((d) => setJobs(d.items || [])).catch(() => {});
  };
  useEffect(() => { reload(); }, [user.id]);
  useEffect(() => {
    (async () => {
      const all: any[] = [];
      for (const t of tracks) {
        const d = await api.comments(t.id);
        for (const c of d.items || []) all.push({ ...c, trackTitle: t.title, trackId: t.id });
      }
      setComments(all.sort((a, b) => a.timestampMs - b.timestampMs));
    })();
  }, [tracks]);

  return (
    <div>
      <h2 className="font-display text-5xl">创作者工作室</h2>
      <p className="mb-6 text-paper/50">{user.displayName} 的数字货架</p>
      <div className="mb-6 flex flex-wrap gap-2">
        {[["shelf", "货架"], ["upload", "上传"], ["inbox", "乐评"], ["orders", "赞助"], ["jobs", "转码"]].map(([k, lab]) => (
          <button key={k} onClick={() => setTab(k as any)} className={`rounded-full px-4 py-1.5 ${tab === k ? "bg-amber text-ink" : "bg-white/5"}`}>{lab}</button>
        ))}
      </div>
      {tab === "upload" && <StudioUpload push={push} />}
      {tab === "shelf" && (
        <div className="grid gap-4 md:grid-cols-2">
          {tracks.map((t) => (
            <div key={t.id} className="rounded-2xl border border-white/5 p-5">
              <div className="flex justify-between">
                <Link to={"/track/" + t.id} className="font-display text-2xl">{t.title}</Link>
                <span className="text-xs text-amber">{t.transcodeStatus}</span>
              </div>
              <p className="text-sm text-paper/50">播放 {t.playCount} · 赞助 {yuan(t.sponsorCents)} · 定价 {yuan(t.paidPriceCents)}</p>
              <div className="mt-3 flex flex-wrap gap-2 text-sm">
                <label>试听
                  <select className="ml-2 rounded bg-ink p-1" defaultValue={t.previewSeconds} onChange={(e) => api.patchTrack(t.id, { previewSeconds: Number(e.target.value) }).then(() => push("已更新试听"))}>
                    <option value={15}>15s</option>
                    <option value={30}>30s</option>
                    <option value={60}>60s</option>
                  </select>
                </label>
                <label className="flex items-center gap-1"><input type="checkbox" defaultChecked={t.paidDownload} onChange={(e) => api.patchTrack(t.id, { paidDownload: e.target.checked })} />付费下载</label>
                <label className="flex items-center gap-1"><input type="checkbox" defaultChecked={t.fanOnly} onChange={(e) => api.patchTrack(t.id, { fanOnly: e.target.checked })} />粉丝专属</label>
              </div>
            </div>
          ))}
        </div>
      )}
      {tab === "inbox" && (
        <ul className="space-y-3">
          {comments.map((c) => (
            <li key={c.id} className="rounded-2xl border border-white/5 p-4">
              <Link to={"/track/" + c.trackId} className="text-amber">{c.trackTitle} @ {(c.timestampMs / 1000).toFixed(0)}s</Link>
              <p>{c.authorName}：{c.body}</p>
              <div className="mt-2 flex gap-2 text-sm">
                <button onClick={() => api.modComment(c.id, { pinned: !c.pinned }).then(reload)}>置顶</button>
                <button onClick={() => api.modComment(c.id, { hidden: !c.hidden }).then(reload)}>隐藏</button>
              </div>
            </li>
          ))}
          {!comments.length && <p className="text-paper/40">还没有听众把句子钉在你的波形上。</p>}
        </ul>
      )}
      {tab === "orders" && (
        <table className="w-full text-left text-sm">
          <thead><tr className="text-paper/40"><th>单号</th><th>金额</th><th>状态</th><th>时间</th></tr></thead>
          <tbody>
            {orders.map((o) => (
              <tr key={o.id} className="border-t border-white/5">
                <td className="py-2">{o.orderNo}</td>
                <td>{yuan(o.amountCents)}</td>
                <td>{o.status}</td>
                <td>{o.createdAt}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
      {tab === "jobs" && (
        <ul className="space-y-2">
          {jobs.map((j) => (
            <li key={j.id} className="flex items-center justify-between rounded-xl bg-white/5 px-4 py-3">
              <span>{j.trackId.slice(0, 8)} · {j.status} · {j.progress}%</span>
              {j.status === "failed" && <button onClick={() => api.retry(j.trackId).then(() => push("已重试"))} className="text-amber">重试</button>}
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
