import { useEffect, useState } from "react";
import { api } from "../../frontend-user/src/api";

export function AdminHome() {
  const [data, setData] = useState<any>(null);
  useEffect(() => {
    api.admin().then(setData);
  }, []);
  if (!data) return <p>加载看板…</p>;
  return (
    <div>
      <h2 className="font-display text-5xl">风控与履约看板</h2>
      <div className="mt-6 grid gap-4 md:grid-cols-3">
        <Card label="用户" value={data.stats.users} />
        <Card label="作品" value={data.stats.tracks} />
        <Card label="已支付订单" value={data.stats.paidOrders} />
      </div>
      <h3 className="mt-10 font-display text-3xl">审计流</h3>
      <ul className="mt-4 space-y-2 text-sm">
        {(data.audit || []).map((a: any) => (
          <li key={a.id} className="rounded-xl bg-white/5 px-4 py-2">
            <span className="text-amber">{a.action}</span> · {a.reason}
          </li>
        ))}
      </ul>
    </div>
  );
}

function Card({ label, value }: any) {
  return (
    <div className="rounded-2xl border border-amber/20 p-6">
      <p className="text-xs uppercase tracking-widest text-paper/40">{label}</p>
      <p className="font-display text-4xl">{value}</p>
    </div>
  );
}
