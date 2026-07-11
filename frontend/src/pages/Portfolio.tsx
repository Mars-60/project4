import { useQuery } from "@tanstack/react-query";
import { Card, DataTable, MetricCard, MiniChart, PageHeader, StatusPill } from "../components/ui";
import { Holding, Position, services, Trade } from "../lib/api";
import { dateTime, decimal, money } from "../lib/format";

export function Portfolio() {
  const summary = useQuery({ queryKey: ["portfolio", "summary"], queryFn: () => services.portfolio.summary(false) });
  const positions = useQuery({ queryKey: ["portfolio", "positions"], queryFn: () => services.portfolio.positions(false) });
  const holdings = useQuery({ queryKey: ["portfolio", "holdings"], queryFn: services.portfolio.holdings });
  const trades = useQuery({ queryKey: ["portfolio", "trades"], queryFn: services.portfolio.trades });

  const pnl = (summary.data?.RealizedPnL ?? 0) + (summary.data?.UnrealizedPnL ?? 0);

  return (
    <section>
      <PageHeader title="Portfolio" eyebrow="Holdings, positions, analytics" />
      <div className="grid gap-4 md:grid-cols-4">
        <MetricCard label="Net Value" value={money(summary.data?.NetValue)} />
        <MetricCard label="Cash" value={money(summary.data?.Cash)} />
        <MetricCard label="PnL" value={money(pnl)} tone={pnl >= 0 ? "good" : "bad"} />
        <MetricCard label="Exposure" value={money(summary.data?.Exposure)} />
      </div>
      <div className="mt-4 grid gap-4 xl:grid-cols-[0.9fr_1.1fr]">
        <Card>
          <h2 className="mb-4 font-semibold">Analytics</h2>
          <MiniChart values={[8, 11, 9, 17, 20, 16, 24, Math.max(25, Math.abs(pnl) / 1000)]} />
        </Card>
        <Card>
          <h2 className="mb-4 font-semibold">Current Holdings</h2>
          <DataTable<Holding>
            data={holdings.data}
            empty="No holdings"
            columns={[
              { key: "symbol", header: "Symbol", render: (row) => row.TradingSymbol ?? "-" },
              { key: "qty", header: "Qty", align: "right", render: (row) => row.Quantity ?? 0 },
              { key: "avg", header: "Avg", align: "right", render: (row) => money(row.AveragePrice) },
              { key: "ltp", header: "LTP", align: "right", render: (row) => money(row.LastPrice) },
              { key: "pnl", header: "PnL", align: "right", render: (row) => <span className={(row.PnL ?? 0) >= 0 ? "text-emerald-500" : "text-rose-500"}>{money(row.PnL)}</span> }
            ]}
          />
        </Card>
      </div>
      <div className="mt-4 grid gap-4 xl:grid-cols-2">
        <Card>
          <h2 className="mb-4 font-semibold">Positions</h2>
          <DataTable<Position>
            data={positions.data}
            empty="No open positions"
            columns={[
              { key: "symbol", header: "Symbol", render: (row) => row.TradingSymbol ?? "-" },
              { key: "product", header: "Product", render: (row) => row.ProductType ?? "-" },
              { key: "qty", header: "Qty", align: "right", render: (row) => decimal(row.Quantity, 0) },
              { key: "pnl", header: "Unrealized", align: "right", render: (row) => money(row.UnrealizedPnL) }
            ]}
          />
        </Card>
        <Card>
          <h2 className="mb-4 font-semibold">Trade History</h2>
          <DataTable<Trade>
            data={trades.data}
            empty="No trades"
            columns={[
              { key: "symbol", header: "Symbol", render: (row) => row.TradingSymbol ?? "-" },
              { key: "side", header: "Side", render: (row) => <StatusPill label={row.TransactionType ?? "-"} tone={row.TransactionType === "BUY" ? "good" : "bad"} /> },
              { key: "price", header: "Price", align: "right", render: (row) => money(row.Price) },
              { key: "time", header: "Time", align: "right", render: (row) => dateTime(row.TradedAt) }
            ]}
          />
        </Card>
      </div>
    </section>
  );
}
