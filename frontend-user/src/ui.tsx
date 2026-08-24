import React, { useEffect, useState } from "react";

export function ToastHost({ toasts, onClose }: { toasts: { id: number; text: string; danger?: boolean }[]; onClose: (id: number) => void }) {
  useEffect(() => {
    const t = setInterval(() => {
      toasts.forEach((x) => onClose(x.id));
    }, 5000);
    return () => clearInterval(t);
  }, [toasts, onClose]);
  return (
    <div className="toast-stack">
      {toasts.map((t) => (
        <div key={t.id} className={`min-w-[240px] rounded-xl border px-4 py-3 text-sm shadow-lg ${t.danger ? "border-rust/50 bg-[#3a1c14]" : "border-amber/30 bg-[#2a2118]"}`}>
          <div className="flex justify-between gap-3">
            <span>{t.text}</span>
            <button onClick={() => onClose(t.id)} aria-label="关闭">×</button>
          </div>
        </div>
      ))}
    </div>
  );
}

export function Modal({ open, title, children, onClose }: { open: boolean; title: string; children: React.ReactNode; onClose: () => void }) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-40 grid place-items-center bg-black/60 p-4" onClick={onClose}>
      <div className="w-full max-w-lg rounded-2xl border border-amber/20 bg-[#1c1612] p-6 shadow-2xl" onClick={(e) => e.stopPropagation()}>
        <div className="mb-4 flex items-center justify-between">
          <h3 className="font-display text-2xl">{title}</h3>
          <button onClick={onClose} className="text-paper/60">×</button>
        </div>
        {children}
      </div>
    </div>
  );
}

export function FieldError({ msg }: { msg?: string }) {
  if (!msg) return null;
  return <p className="mt-1 text-xs text-rust">{msg}</p>;
}

export function useToasts() {
  const [toasts, setToasts] = useState<{ id: number; text: string; danger?: boolean }[]>([]);
  const push = (text: string, danger = false) => setToasts((s) => [...s, { id: Date.now() + Math.random(), text, danger }]);
  const close = (id: number) => setToasts((s) => s.filter((x) => x.id !== id));
  return { toasts, push, close };
}

export function Skeleton() {
  return <div className="h-40 animate-pulse rounded-2xl bg-white/5" />;
}
