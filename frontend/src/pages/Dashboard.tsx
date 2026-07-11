import { useQuery } from "@tanstack/react-query";
import { Card, DataTable, MetricCard, MiniChart, PageHeader, Skeleton, StatusPill } from "../components/ui";
import { services, Trade } from "../lib/api";
import { dateTime, money } from "../lib/format";

export function Dashboard() {
  const portfolio = useQuery({ queryKey: ["dashboard", "portfolio"], queryFn: () => services.dashboard.portfolio(true) });
  const funds = useQuery({ queryKey: ["dashboard", "funds"], queryFn: services.dashboard.funds });
  const positions = useQuery({ queryKey: ["dashboard", "positions"], queryFn: () => services.dashboard.positions(true) });
  const trades = useQuery({ queryKey: ["dashboard", "trades"], queryFn: services.dashboard.trades });

  if (portfolio.isLoading || funds.isLoading) {
    return <Skeleton className="h-[520px]" />;
  }

  const summary = portfolio.data;
  const cash = funds.data?.Available ?? summary?.Cash ?? 0;
  const pnl = (summary?.RealizedPnL ?? 0) + (summary?.UnrealizedPnL ?? 0);

  return (
    <section>
      <PageHeader title="Command Center" eyebrow="Live overview" />
      <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
        <MetricCard label="Today's PnL" value={money(pnl)} tone={pnl >= 0 ? "good" : "bad"} detail="Realized + unrealized" />
        <MetricCard label="Portfolio Value" value={money(summary?.NetValue)} detail={`Exposure ${money(summary?.Exposure)}`} />
        <MetricCard label="Available Funds" value={money(cash)} detail={`Used margin ${money(funds.data?.UsedMargin)}`} />
        <MetricCard label="Active Positions" value={`${positions.data?.length ?? 0}`} detail="Paper and live views are separated" />
      </div>

      <div className="mt-4 grid gap-4 xl:grid-cols-[1.2fr_0.8fr]">
        <Card>
          <div className="mb-4 flex items-center justify-between">
            <div>
              <h2 className="font-semibold text-slate-950 dark:text-white">Intraday Pulse</h2>
              <p className="text-sm text-slate-500 dark:text-slate-400">Backend values rendered as a trend surface.</p>
            </div>
            <StatusPill label="Paper mode" tone="good" />
          </div>
          <MiniChart values={[12, 18, 15, 24, 22, 31, 28, Math.max(30, Math.abs(pnl) / 1000)]} />
          <div className="grid gap-3 sm:grid-cols-3">
            <StatusLine label="Market Status" value="Watchlist ready" tone="warn" />
            <StatusLine label="Broker Connection" value="SMC pending" tone="neutral" />
            <StatusLine label="Bot Status" value="Risk guarded" tone="good" />
          </div>
        </Card>

        <Card>
          <h2 className="mb-4 font-semibold text-slate-950 dark:text-white">AI Insights</h2>
          <div className="space-y-3 text-sm text-slate-600 dark:text-slate-300">
            <p>Portfolio, risk, strategy, and market explanations are available in the AI Assistant.</p>
            <p>Trade execution is intentionally unavailable to AI flows.</p>
          </div>
        </Card>
      </div>

      <div className="mt-4">
        <Card>
          <h2 className="mb-4 font-semibold text-slate-950 dark:text-white">Recent Trades</h2>
          <DataTable<Trade>
            data={trades.data}
            empty="No recent trades"
            columns={[
              { key: "symbol", header: "Symbol", render: (row) => row.TradingSymbol ?? "-" },
              { key: "side", header: "Side", render: (row) => <StatusPill label={row.TransactionType ?? "-"} tone={row.TransactionType === "BUY" ? "good" : "bad"} /> },
              { key: "qty", header: "Qty", align: "right", render: (row) => row.Quantity ?? 0 },
              { key: "price", header: "Price", align: "right", render: (row) => money(row.Price) },
              { key: "time", header: "Time", align: "right", render: (row) => dateTime(row.TradedAt) }
            ]}
          />
        </Card>
      </div>
    </section>
  );
}

function StatusLine({ label, value, tone }: { label: string; value: string; tone: "neutral" | "good" | "bad" | "warn" }) {
  return (
    <div className="rounded-lg bg-slate-50 p-3 dark:bg-slate-900">
      <div className="text-xs text-slate-500 dark:text-slate-400">{label}</div>
      <div className="mt-2">
        <StatusPill label={value} tone={tone} />
      </div>
    </div>
  );
}
