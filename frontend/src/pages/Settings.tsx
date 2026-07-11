import { useQuery } from "@tanstack/react-query";
import { Card, Field, inputClass, PageHeader, StatusPill, Toggle } from "../components/ui";
import { services } from "../lib/api";
import { useTheme } from "../lib/theme";

export function Settings() {
  const { theme, toggleTheme } = useTheme();
  const metrics = useQuery({ queryKey: ["metrics"], queryFn: services.system.metrics });
  return (
    <section>
      <PageHeader title="Settings" eyebrow="Broker, risk, API status" />
      <div className="grid gap-4 xl:grid-cols-2">
        <Card>
          <h2 className="mb-4 font-semibold">Broker Configuration</h2>
          <div className="grid gap-3 sm:grid-cols-2">
            <Field label="Broker"><input className={inputClass} value="SMC" readOnly /></Field>
            <Field label="Connection"><input className={inputClass} value="Awaiting live API approval" readOnly /></Field>
          </div>
        </Card>
        <Card>
          <h2 className="mb-4 font-semibold">Trading Preferences</h2>
          <div className="space-y-3">
            <SettingRow label="Paper trading default" enabled />
            <SettingRow label="Require risk validation" enabled />
            <SettingRow label="Allow AI execution" enabled={false} />
          </div>
        </Card>
        <Card>
          <h2 className="mb-4 font-semibold">Theme</h2>
          <div className="flex items-center justify-between rounded-lg bg-slate-50 p-3 dark:bg-slate-900">
            <span className="capitalize">{theme} mode</span>
            <Toggle checked={theme === "dark"} onChange={toggleTheme} />
          </div>
        </Card>
        <Card>
          <h2 className="mb-4 font-semibold">API Status</h2>
          <div className="flex items-center justify-between">
            <span className="text-sm text-slate-500 dark:text-slate-400">System metrics</span>
            <StatusPill label={metrics.isError ? "Degraded" : "Healthy"} tone={metrics.isError ? "bad" : "good"} />
          </div>
          <pre className="mt-4 overflow-auto rounded-lg bg-slate-50 p-3 text-xs dark:bg-slate-900">{JSON.stringify(metrics.data ?? {}, null, 2)}</pre>
        </Card>
      </div>
    </section>
  );
}

function SettingRow({ label, enabled }: { label: string; enabled: boolean }) {
  return (
    <div className="flex items-center justify-between rounded-lg bg-slate-50 p-3 dark:bg-slate-900">
      <span>{label}</span>
      <StatusPill label={enabled ? "On" : "Off"} tone={enabled ? "good" : "bad"} />
    </div>
  );
}
