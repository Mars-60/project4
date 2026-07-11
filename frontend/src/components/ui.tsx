import { ReactNode } from "react";

export function PageHeader({
  title,
  eyebrow,
  action
}: {
  title: string;
  eyebrow?: string;
  action?: ReactNode;
}) {
  return (
    <div className="mb-5 flex flex-col gap-3 sm:flex-row sm:items-end sm:justify-between">
      <div>
        {eyebrow && <p className="text-xs font-semibold uppercase tracking-[0.22em] text-cyan-600 dark:text-cyan-300">{eyebrow}</p>}
        <h1 className="mt-1 text-2xl font-semibold text-slate-950 dark:text-white">{title}</h1>
      </div>
      {action}
    </div>
  );
}

export function Card({ children, className = "" }: { children: ReactNode; className?: string }) {
  return (
    <section className={`rounded-lg border border-slate-200 bg-white p-4 shadow-sm dark:border-slate-800 dark:bg-slate-950 ${className}`}>
      {children}
    </section>
  );
}

export function MetricCard({
  label,
  value,
  detail,
  tone = "neutral"
}: {
  label: string;
  value: string;
  detail?: string;
  tone?: "neutral" | "good" | "bad" | "warn";
}) {
  const tones = {
    neutral: "text-slate-950 dark:text-white",
    good: "text-emerald-600 dark:text-emerald-300",
    bad: "text-rose-600 dark:text-rose-300",
    warn: "text-amber-600 dark:text-amber-300"
  };
  return (
    <Card>
      <div className="text-xs font-medium uppercase tracking-wide text-slate-500 dark:text-slate-400">{label}</div>
      <div className={`mt-3 text-2xl font-semibold ${tones[tone]}`}>{value}</div>
      {detail && <div className="mt-2 text-sm text-slate-500 dark:text-slate-400">{detail}</div>}
    </Card>
  );
}

export function StatusPill({ label, tone = "neutral" }: { label: string; tone?: "neutral" | "good" | "bad" | "warn" }) {
  const tones = {
    neutral: "bg-slate-100 text-slate-700 dark:bg-slate-800 dark:text-slate-300",
    good: "bg-emerald-100 text-emerald-700 dark:bg-emerald-950 dark:text-emerald-300",
    bad: "bg-rose-100 text-rose-700 dark:bg-rose-950 dark:text-rose-300",
    warn: "bg-amber-100 text-amber-700 dark:bg-amber-950 dark:text-amber-300"
  };
  return <span className={`inline-flex rounded-full px-2.5 py-1 text-xs font-semibold ${tones[tone]}`}>{label}</span>;
}

export function Skeleton({ className = "" }: { className?: string }) {
  return <div className={`animate-pulse rounded-md bg-slate-200 dark:bg-slate-800 ${className}`} />;
}

export function EmptyState({ title, body }: { title: string; body: string }) {
  return (
    <div className="rounded-lg border border-dashed border-slate-300 p-6 text-center dark:border-slate-700">
      <div className="font-semibold text-slate-900 dark:text-white">{title}</div>
      <p className="mt-2 text-sm text-slate-500 dark:text-slate-400">{body}</p>
    </div>
  );
}

export function DataTable<T>({
  data,
  columns,
  empty = "No rows yet"
}: {
  data?: T[];
  empty?: string;
  columns: Array<{ key: string; header: string; render: (row: T) => ReactNode; align?: "left" | "right" }>;
}) {
  if (!data?.length) {
    return <EmptyState title={empty} body="Once the backend has records, they will appear here automatically." />;
  }
  return (
    <div className="overflow-hidden rounded-lg border border-slate-200 dark:border-slate-800">
      <div className="overflow-x-auto">
        <table className="min-w-full divide-y divide-slate-200 text-sm dark:divide-slate-800">
          <thead className="bg-slate-50 dark:bg-slate-900">
            <tr>
              {columns.map((column) => (
                <th key={column.key} className={`px-4 py-3 font-semibold text-slate-500 dark:text-slate-400 ${column.align === "right" ? "text-right" : "text-left"}`}>
                  {column.header}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-100 bg-white dark:divide-slate-900 dark:bg-slate-950">
            {data.map((row, index) => (
              <tr key={index} className="hover:bg-slate-50 dark:hover:bg-slate-900/70">
                {columns.map((column) => (
                  <td key={column.key} className={`px-4 py-3 text-slate-700 dark:text-slate-200 ${column.align === "right" ? "text-right" : "text-left"}`}>
                    {column.render(row)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}

export function MiniChart({ values }: { values: number[] }) {
  const max = Math.max(...values, 1);
  const points = values.map((value, index) => `${(index / Math.max(values.length - 1, 1)) * 100},${48 - (value / max) * 40}`).join(" ");
  return (
    <svg viewBox="0 0 100 52" className="h-28 w-full text-cyan-500" role="img" aria-label="Performance trend">
      <polyline fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" points={points} />
    </svg>
  );
}

export function Field({ label, children }: { label: string; children: ReactNode }) {
  return (
    <label className="block text-sm">
      <span className="mb-1.5 block font-medium text-slate-700 dark:text-slate-300">{label}</span>
      {children}
    </label>
  );
}

export const inputClass =
  "w-full rounded-md border border-slate-300 bg-white px-3 py-2 text-sm text-slate-900 outline-none transition focus:border-cyan-500 focus:ring-2 focus:ring-cyan-500/20 dark:border-slate-700 dark:bg-slate-900 dark:text-white";

export function Toggle({ checked, onChange }: { checked: boolean; onChange: (checked: boolean) => void }) {
  return (
    <button
      type="button"
      onClick={() => onChange(!checked)}
      className={`relative h-6 w-11 rounded-full transition ${checked ? "bg-cyan-500" : "bg-slate-300 dark:bg-slate-700"}`}
      aria-pressed={checked}
    >
      <span className={`absolute top-1 h-4 w-4 rounded-full bg-white transition ${checked ? "left-6" : "left-1"}`} />
    </button>
  );
}
