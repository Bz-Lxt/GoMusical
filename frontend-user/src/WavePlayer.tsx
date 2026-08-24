import { useEffect, useRef, useState } from "react";
import WaveSurfer from "wavesurfer.js";
import Hls from "hls.js";
import { api, fmtTime } from "./api";

const SPEEDS = [1, 1.25, 1.5, 2];

export function WavePlayer({ track, user, onToast }: { track: any; user: any; onToast: (s: string, d?: boolean) => void }) {
  const wrap = useRef<HTMLDivElement>(null);
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const wsRef = useRef<WaveSurfer | null>(null);
  const hlsRef = useRef<Hls | null>(null);
  const [ready, setReady] = useState(false);
  const [rate, setRate] = useState(1);
  const [playing, setPlaying] = useState(false);
  const [t, setT] = useState(0);
  const [comments, setComments] = useState<any[]>([]);
  const [draft, setDraft] = useState("");
  const [at, setAt] = useState<number | null>(null);
  const until = track.accessUntilMs || track.previewSeconds * 1000 || 30000;
  const locked = (track.durationMs || 0) > until + 200;

  useEffect(() => {
    let dead = false;
    (async () => {
      const [peaksBody, stream] = await Promise.all([api.peaks(track.id), api.openStream(track.id)]);
      const peaks = peaksBody.channels?.[0] || peaksBody.data?.channels?.[0];
      const audio = document.createElement("audio");
      audio.preload = "auto";
      (audio as any).preservesPitch = true;
      audio.preservesPitch = true;
      audioRef.current = audio;
      const playlist = "/api" + String(stream.playlist).replace(/^\/api/, "");
      if (Hls.isSupported()) {
        const hls = new Hls({ enableWorker: true });
        hls.loadSource(playlist);
        hls.attachMedia(audio);
        hlsRef.current = hls;
      } else {
        audio.src = playlist;
      }
      if (!wrap.current || dead) return;
      const ws = WaveSurfer.create({
        container: wrap.current,
        media: audio,
        peaks: peaks ? [peaks] : undefined,
        duration: (track.durationMs || 36000) / 1000,
        height: 128,
        waveColor: "#5c4630",
        progressColor: "#c9a36a",
        cursorColor: "#f3e6d2",
        barWidth: 2,
        barGap: 1,
        barRadius: 2,
        normalize: true,
      });
      ws.on("ready", () => setReady(true));
      ws.on("timeupdate", (sec) => setT(sec * 1000));
      ws.on("play", () => setPlaying(true));
      ws.on("pause", () => setPlaying(false));
      ws.on("click", (rel) => {
        const ms = rel * (track.durationMs || 0);
        if (ms > until) {
          onToast("试听区间外已锁定，赞助后可写全曲乐评", true);
          return;
        }
        setAt(ms);
      });
      wsRef.current = ws;
    })().catch((e) => onToast(e.message, true));
    api.comments(track.id).then((d) => setComments(d.items || [])).catch(() => {});
    return () => {
      dead = true;
      wsRef.current?.destroy();
      hlsRef.current?.destroy();
      audioRef.current = null;
    };
  }, [track.id]);

  useEffect(() => {
    if (audioRef.current) audioRef.current.playbackRate = rate;
  }, [rate]);

  const submit = async () => {
    if (at == null || !draft.trim()) return;
    try {
      const c = await api.addComment(track.id, Math.round(at), draft.trim());
      setComments((s) => [...s, c]);
      setDraft("");
      setAt(null);
      onToast("乐评已钉在波形上");
    } catch (e: any) {
      onToast(e.message, true);
    }
  };

  return (
    <div className="rounded-3xl border border-amber/20 bg-[#1a1410]/80 p-5 md:p-8">
      <div className="mb-4 flex flex-wrap items-end justify-between gap-4">
        <div>
          <p className="text-xs uppercase tracking-[0.25em] text-amber/70">高保真货架 · AAC 流播</p>
          <h1 className="font-display text-4xl md:text-5xl">{track.title}</h1>
          <p className="text-paper/60">{track.creatorName} · {fmtTime(track.durationMs)} · {track.format?.toUpperCase()}</p>
        </div>
        <div className="flex gap-2">
          {SPEEDS.map((s) => (
            <button key={s} onClick={() => setRate(s)} className={`rounded-full px-3 py-1 text-sm ${rate === s ? "bg-amber text-ink" : "bg-white/5"}`}>
              {s.toFixed(2).replace(/\.00$/, ".0")}x
            </button>
          ))}
        </div>
      </div>
      <div className="relative">
        <div ref={wrap} className="min-h-[128px]" />
        {locked && (
          <div className="pointer-events-none absolute inset-y-0 right-0 overflow-hidden rounded-xl" style={{ left: `${(until / (track.durationMs || 1)) * 100}%` }}>
            <div className="h-full w-full bg-[repeating-linear-gradient(135deg,rgba(0,0,0,.45)_0_8px,rgba(201,163,106,.12)_8px_16px)]">
              <div className="absolute right-3 top-3 rounded-full bg-ink/70 px-3 py-1 text-xs">锁定 · 赞助解锁全曲</div>
            </div>
          </div>
        )}
        {comments.filter((c) => !c.hidden).map((c) => (
          <div key={c.id} className="absolute -bottom-1 h-3 w-3 -translate-x-1/2 rounded-full bg-rust" style={{ left: `${(c.timestampMs / (track.durationMs || 1)) * 100}%` }} title={`${c.authorName}: ${c.body}`} />
        ))}
      </div>
      <div className="mt-5 flex flex-wrap items-center gap-3">
        <button onClick={() => wsRef.current?.playPause()} className="rounded-full bg-amber px-6 py-2 font-semibold text-ink">
          {playing ? "暂停" : "播放"}
        </button>
        <span className="text-sm text-paper/70">{fmtTime(t)} / {fmtTime(until)}{locked ? ` · 试听 ${track.previewSeconds}s` : ""}</span>
        {!ready && <span className="text-sm text-paper/40">波形加载中…</span>}
      </div>
      {at != null && (
        <div className="mt-5 rounded-2xl border border-amber/20 bg-black/20 p-4">
          <p className="mb-2 text-sm text-amber">在 {fmtTime(at)} 写下乐评</p>
          <textarea value={draft} onChange={(e) => setDraft(e.target.value)} className="w-full rounded-xl bg-ink/60 p-3 outline-none" rows={3} placeholder={user ? "这一秒让你想到什么？" : "登录后即可留言"} />
          <div className="mt-2 flex gap-2">
            <button disabled={!user} onClick={submit} className="rounded-full bg-rust px-4 py-1.5 text-sm">钉在波形上</button>
            <button onClick={() => setAt(null)} className="text-sm text-paper/50">取消</button>
          </div>
        </div>
      )}
      <ul className="mt-6 space-y-3">
        {comments.filter((c) => !c.hidden).map((c) => (
          <li key={c.id} className="flex gap-3 border-b border-white/5 pb-3">
            <button className="text-amber" onClick={() => wsRef.current?.setTime(c.timestampMs / 1000)}>{fmtTime(c.timestampMs)}</button>
            <div>
              <p className="text-sm"><span className="text-paper/50">{c.authorName}</span> {c.body}</p>
              {c.reply && <p className="text-xs text-moss">创作者：{c.reply}</p>}
            </div>
          </li>
        ))}
      </ul>
    </div>
  );
}
