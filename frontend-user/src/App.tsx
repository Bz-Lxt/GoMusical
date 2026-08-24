import { useEffect, useState } from "react";
import { Link, Navigate, Route, Routes, useNavigate, useParams, useSearchParams } from "react-router-dom";
import { api, setCsrf, yuan } from "./api";
import { FieldError, Modal, Skeleton, ToastHost, useToasts } from "./ui";
import { WavePlayer } from "./WavePlayer";
import { CreatorHome } from "@creator/CreatorHome";
import { AdminHome } from "@admin/AdminHome";

export default function App() {
  const [user, setUser] = useState<any>(null);
  const { toasts, push, close } = useToasts();
  useEffect(() => {
    api.me().then((d) => {
      setUser(d.user);
      setCsrf(d.csrf);
    }).catch(() => {});
  }, []);
  return (
    <div className="min-h-screen">
      <ToastHost toasts={toasts} onClose={close} />
      <Header user={user} onLogout={async () => { await api.logout(); setUser(null); }} />
      <main className="w-full px-4 py-8 md:px-8">
        <Routes>
          <Route path="/" element={<Discover push={push} />} />
          <Route path="/login" element={<Auth mode="login" onOk={(u, c) => { setUser(u); setCsrf(c); }} push={push} />} />
          <Route path="/register" element={<Auth mode="register" onOk={(u, c) => { setUser(u); setCsrf(c); }} push={push} />} />
          <Route path="/track/:id" element={<TrackPage user={user} push={push} />} />
          <Route path="/creator/:id" element={<PublicCreator user={user} push={push} />} />
          <Route path="/pay/mock" element={<MockPay push={push} />} />
          <Route path="/studio/*" element={user ? <CreatorHome user={user} push={push} /> : <Navigate to="/login" />} />
          <Route path="/admin" element={user?.role === "ADMIN" ? <AdminHome /> : <Navigate to="/" />} />
        </Routes>
      </main>
    </div>
  );
}

function Header({ user, onLogout }: any) {
  return (
    <header className="flex w-full flex-wrap items-center justify-between gap-4 border-b border-white/5 px-4 py-4 md:px-8">
      <Link to="/" className="font-display text-3xl tracking-tight">GoMusical</Link>
      <nav className="flex flex-wrap items-center gap-4 text-sm">
        <Link to="/" className="text-paper/70 hover:text-amber">货架</Link>
        {user?.role === "CREATOR" || user?.role === "ADMIN" ? <Link to="/studio" className="text-paper/70 hover:text-amber">工作室</Link> : null}
        {user?.role === "ADMIN" ? <Link to="/admin" className="text-paper/70 hover:text-amber">管理</Link> : null}
        {user ? (
          <>
            <span className="text-paper/50">{user.displayName}</span>
            <button onClick={onLogout} className="text-amber">退出</button>
          </>
        ) : (
          <>
            <Link to="/login">登录</Link>
            <Link to="/register" className="rounded-full bg-amber px-3 py-1 text-ink">加入货架</Link>
          </>
        )}
      </nav>
    </header>
  );
}

function Discover({ push }: any) {
  const [tracks, setTracks] = useState<any[] | null>(null);
  useEffect(() => {
    api.tracks().then((d) => setTracks(d.items || [])).catch((e) => push(e.message, true));
  }, []);
  if (!tracks) return <Skeleton />;
  if (!tracks.length) {
    return (
      <div className="rounded-3xl border border-dashed border-amber/30 p-16 text-center">
        <p className="font-display text-4xl">货架还是空的</p>
        <p className="mt-2 text-paper/50">创作者上传第一首作品后，这里会亮起来。</p>
      </div>
    );
  }
  return (
    <section>
      <p className="text-xs uppercase tracking-[0.3em] text-amber/70">独立声音的数字货架</p>
      <h2 className="mb-8 font-display text-5xl">今晚想听哪一首</h2>
      <div className="grid gap-5 md:grid-cols-2 xl:grid-cols-3">
        {tracks.map((t) => (
          <Link key={t.id} to={"/track/" + t.id} className="group rounded-3xl border border-white/5 bg-[#1a1410] p-5 transition hover:border-amber/40">
            <img src={"/api/tracks/" + t.id + "/cover"} alt="" className="mb-4 aspect-square w-full rounded-2xl object-cover" />
            <h3 className="font-display text-2xl group-hover:text-amber">{t.title}</h3>
            <p className="text-sm text-paper/50">{t.creatorName} · {t.accessTier === "PREVIEW" ? `试听 ${t.previewSeconds}s` : "已解锁"}</p>
          </Link>
        ))}
      </div>
    </section>
  );
}

function Auth({ mode, onOk, push }: any) {
  const nav = useNavigate();
  const [email, setEmail] = useState(mode === "login" ? "listener@gomusical.local" : "");
  const [password, setPassword] = useState(mode === "login" ? "Listener123!" : "");
  const [displayName, setDisplayName] = useState("");
  const [role, setRole] = useState("LISTENER");
  const [errs, setErrs] = useState<any>({});
  const submit = async () => {
    const e: any = {};
    if (!email.includes("@")) e.email = "请填写有效邮箱";
    if (password.length < 8) e.password = "密码至少 8 位";
    if (mode === "register" && !displayName.trim()) e.displayName = "需要称呼";
    setErrs(e);
    if (Object.keys(e).length) {
      push("请修正表单后再提交", true);
      return;
    }
    try {
      const d = mode === "login" ? await api.login(email, password) : await api.register({ email, password, displayName, role });
      onOk(d.user, d.csrf);
      nav("/");
    } catch (err: any) {
      push(err.message, true);
    }
  };
  return (
    <div className="mx-auto w-full max-w-md rounded-3xl border border-amber/20 bg-[#1a1410] p-8">
      <h2 className="font-display text-4xl">{mode === "login" ? "回到货架" : "登记一张席位"}</h2>
      <label className="mt-6 block text-sm">邮箱<input className="mt-1 w-full rounded-xl bg-ink p-3" value={email} onChange={(e) => setEmail(e.target.value)} /></label>
      <FieldError msg={errs.email} />
      <label className="mt-4 block text-sm">密码<input type="password" className="mt-1 w-full rounded-xl bg-ink p-3" value={password} onChange={(e) => setPassword(e.target.value)} /></label>
      <FieldError msg={errs.password} />
      {mode === "register" && (
        <>
          <label className="mt-4 block text-sm">称呼<input className="mt-1 w-full rounded-xl bg-ink p-3" value={displayName} onChange={(e) => setDisplayName(e.target.value)} /></label>
          <FieldError msg={errs.displayName} />
          <label className="mt-4 block text-sm">身份
            <select className="mt-1 w-full rounded-xl bg-ink p-3" value={role} onChange={(e) => setRole(e.target.value)}>
              <option value="LISTENER">听众</option>
              <option value="CREATOR">创作者</option>
            </select>
          </label>
        </>
      )}
      <button onClick={submit} className="mt-6 w-full rounded-full bg-amber py-3 font-semibold text-ink">{mode === "login" ? "登录" : "注册"}</button>
      <p className="mt-3 text-center text-xs text-paper/40">测试账号 listener@gomusical.local / Listener123!</p>
    </div>
  );
}

function TrackPage({ user, push }: any) {
  const { id } = useParams();
  const [track, setTrack] = useState<any>(null);
  const [open, setOpen] = useState(false);
  const [amount, setAmount] = useState(900);
  const [ticket, setTicket] = useState<any>(null);
  useEffect(() => {
    api.track(id!).then(setTrack).catch((e) => push(e.message, true));
  }, [id]);
  if (!track) return <Skeleton />;
  const buy = async () => {
    try {
      const d = await api.sponsor(track.id, amount);
      if (d.paid) {
        push("赞助成功，全曲与下载已解锁");
        const t = await api.track(track.id);
        setTrack(t);
        setOpen(false);
      } else if (d.checkoutUrl) {
        window.location.href = d.checkoutUrl;
      }
    } catch (e: any) {
      push(e.message, true);
    }
  };
  const dl = async () => {
    try {
      const t = await api.ticket(track.id);
      setTicket(t);
      window.location.href = t.url;
    } catch (e: any) {
      push(e.message, true);
    }
  };
  return (
    <div className="grid gap-6 xl:grid-cols-[1.4fr_.6fr]">
      <WavePlayer track={track} user={user} onToast={push} />
      <aside className="rounded-3xl border border-white/5 bg-[#1a1410] p-6">
        <p className="text-sm text-paper/50">当前档位 {track.accessTier}</p>
        <p className="mt-2 font-display text-3xl">赞助 {yuan(track.paidPriceCents)}</p>
        <p className="mt-1 text-sm text-paper/50">按钮触发前会再次确认金额。支付走 Mock，不产生真实扣款。</p>
        <button onClick={() => { setAmount(track.paidPriceCents); setOpen(true); }} className="mt-4 w-full rounded-full bg-rust py-3 font-semibold">
          赞助解锁 · {yuan(track.paidPriceCents)}
        </button>
        <button onClick={dl} className="mt-3 w-full rounded-full border border-amber/40 py-3">签发无损下载凭证</button>
        {ticket && (
          <p className="mt-3 text-xs text-paper/50">凭证 {ticket.ttlSec}s 内有效，最多 {ticket.maxUses} 次。断点续传不重复扣次。</p>
        )}
        <Link to={"/creator/" + track.creatorId} className="mt-6 block text-amber">走进 {track.creatorName} 的货架 →</Link>
      </aside>
      <Modal open={open} title="确认赞助金额" onClose={() => setOpen(false)}>
        <p className="text-paper/70">本次将支付 <strong className="text-amber">{yuan(amount)}</strong>（{amount} 分）。Mock 模式立即到账。</p>
        <div className="mt-4 flex gap-2">
          {[track.paidPriceCents, 1800, 3900].map((n) => (
            <button key={n} onClick={() => setAmount(n)} className={`rounded-full px-3 py-1 ${amount === n ? "bg-amber text-ink" : "bg-white/5"}`}>{yuan(n)}</button>
          ))}
        </div>
        <label className="mt-4 block text-sm">自定义（分）
          <input type="number" min={track.paidPriceCents} className="mt-1 w-full rounded-xl bg-ink p-3" value={amount} onChange={(e) => setAmount(Number(e.target.value))} />
        </label>
        <button onClick={buy} className="mt-5 w-full rounded-full bg-amber py-3 font-semibold text-ink">确认支付 {yuan(amount)}</button>
      </Modal>
    </div>
  );
}

function PublicCreator({ user, push }: any) {
  const { id } = useParams();
  const [data, setData] = useState<any>(null);
  useEffect(() => {
    api.creator(id!).then(setData).catch((e) => push(e.message, true));
  }, [id]);
  if (!data) return <Skeleton />;
  return (
    <div>
      <h2 className="font-display text-5xl">{data.creator.displayName}</h2>
      <p className="mb-6 text-paper/60">{data.creator.bio}</p>
      <button onClick={async () => {
        try {
          const d = await api.subscribe(id!, 1800);
          push(d.paid ? "已订阅 30 天 · ¥18.00" : "请完成支付");
        } catch (e: any) { push(e.message, true); }
      }} className="mb-8 rounded-full bg-amber px-5 py-2 text-ink">成为粉丝 · ¥18.00 / 月</button>
      <div className="grid gap-4 md:grid-cols-2">
        {(data.tracks || []).map((t: any) => (
          <Link key={t.id} to={"/track/" + t.id} className="rounded-2xl border border-white/5 p-4">{t.title}</Link>
        ))}
      </div>
    </div>
  );
}

function MockPay({ push }: any) {
  const [sp] = useSearchParams();
  const nav = useNavigate();
  const orderNo = sp.get("orderNo") || "";
  return (
    <div className="mx-auto max-w-md rounded-3xl border border-amber/20 p-8">
      <h2 className="font-display text-3xl">Mock 收银台</h2>
      <p className="mt-2 text-paper/60">订单 {orderNo}。此页不会产生真实扣款。</p>
      <button onClick={async () => {
        await api.payCallback({ orderNo, status: "paid" });
        push("支付回调已核销");
        nav("/");
      }} className="mt-6 w-full rounded-full bg-amber py-3 text-ink">模拟支付成功</button>
    </div>
  );
}

