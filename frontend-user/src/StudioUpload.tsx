import { useState } from "react";
import { api, uploadChunk } from "./api";

export function StudioUpload({ push }: { push: (s: string, d?: boolean) => void }) {
  const [title, setTitle] = useState("");
  const [file, setFile] = useState<File | null>(null);
  const [prog, setProg] = useState(0);
  const run = async () => {
    if (!file || !title.trim()) {
      push("标题与文件都必填", true);
      return;
    }
    const buf = await file.arrayBuffer();
    const hash = Array.from(new Uint8Array(await crypto.subtle.digest("SHA-256", buf)))
      .map((b) => b.toString(16).padStart(2, "0"))
      .join("");
    const init = await api.initUpload({ filename: file.name, sha256: hash, sizeBytes: file.size, title });
    if (init.instant) {
      push("秒传命中，已复用转码缓存");
      return;
    }
    const size = 5 * 1024 * 1024;
    const chunks = Math.ceil(file.size / size);
    for (let i = 0; i < chunks; i++) {
      await uploadChunk(init.uploadId, i, file.slice(i * size, (i + 1) * size));
      setProg(Math.round(((i + 1) / chunks) * 100));
    }
    await api.completeUpload(init.uploadId, title);
    push("上传完成，转码排队中");
  };
  return (
    <div className="rounded-3xl border border-white/5 p-6">
      <h3 className="font-display text-3xl">把一首未发行的作品放上货架</h3>
      <input className="mt-4 w-full rounded-xl bg-ink p-3" placeholder="曲名" value={title} onChange={(e) => setTitle(e.target.value)} />
      <input className="mt-3 w-full" type="file" accept=".flac,.wav,.mp3,.m4a,audio/*" onChange={(e) => setFile(e.target.files?.[0] || null)} />
      <button onClick={run} className="mt-4 rounded-full bg-amber px-5 py-2 text-ink">开始上传</button>
      {prog > 0 && <p className="mt-2 text-sm">进度 {prog}%</p>}
    </div>
  );
}
