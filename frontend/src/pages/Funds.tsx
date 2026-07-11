import { useQuery } from "@tanstack/react-query";
import { Card, MetricCard, PageHeader } from "../components/ui";
import { services } from "../lib/api";
import { money } from "../lib/format";

export function FundsPage() {
  const funds = useQuery({ queryKey: ["funds"], queryFn: services.dashboard.funds });
  return (
    <section>
      <PageHeader title="Funds" eyebrow="Capital and margin" />
      <div className="grid gap-4 md:grid-cols-4">
        <MetricCard label="Available" value={money(funds.data?.Available)} />
        <MetricCard label="Used Margin" value={money(funds.data?.UsedMargin)} />
        <MetricCard label="Opening" value={money(funds.data?.Opening)} />
        <MetricCard label="Paper Balance" value={money(funds.data?.PaperBalance)} />
      </div>
      <Card className="mt-4">
        <h2 className="font-semibold">Fund Controls</h2>
        <div className="mt-4 grid gap-3 sm:grid-cols-3">
          {["Add funds", "Withdraw", "View ledger"].map((item) => (
            <button key={item} className="rounded-md border border-slate-300 px-4 py-3 text-left text-sm font-medium dark:border-slate-700">
              {item}
            </button>
          ))}
        </div>
      </Card>
    </section>
  );
}
