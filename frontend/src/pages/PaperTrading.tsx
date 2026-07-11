import { FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Card, DataTable, Field, inputClass, MetricCard, PageHeader, StatusPill, Toggle } from "../components/ui";
import { Order, Position, services } from "../lib/api";
import { money } from "../lib/format";

export function PaperTrading() {
  const [enabled, setEnabled] = useState(true);
  const [symbol, setSymbol] = useState("RELIANCE");
  const [token, setToken] = useState("2885");
  const [quantity, setQuantity] = useState(1);
  const [price, setPrice] = useState(2500);
  const client = useQueryClient();
  const portfolio = useQuery({ queryKey: ["paper", "summary"], queryFn: () => services.portfolio.summary(true) });
  const positions = useQuery({ queryKey: ["paper", "positions"], queryFn: () => services.portfolio.positions(true) });
  const orders = useQuery({ queryKey: ["paper", "orders"], queryFn: services.orders.list });
  const placeOrder = useMutation({
    mutationFn: () =>
      services.orders.paper({
        signal: {
          Exchange: "NSE",
          SymbolToken: token,
          TradingSymbol: symbol,
          Action: "BUY",
          ProductType: "INTRADAY",
          OrderType: "LIMIT",
          Quantity: quantity,
          Price: price,
          StopLoss: price * 0.98,
          Target: price * 1.03
        },
        last_price: price
      }),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: ["paper"] });
      client.invalidateQueries({ queryKey: ["orders"] });
    }
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    if (enabled) placeOrder.mutate();
  }

  return (
    <section>
      <PageHeader
        title="Paper Trading"
        eyebrow="Simulation engine"
        action={
          <div className="flex items-center gap-3 rounded-lg border border-slate-200 bg-white px-3 py-2 dark:border-slate-800 dark:bg-slate-950">
            <span className="text-sm">Enabled</span>
            <Toggle checked={enabled} onChange={setEnabled} />
          </div>
        }
      />
      <div className="grid gap-4 md:grid-cols-3">
        <MetricCard label="Paper Portfolio" value={money(portfolio.data?.NetValue)} />
        <MetricCard label="Paper PnL" value={money((portfolio.data?.RealizedPnL ?? 0) + (portfolio.data?.UnrealizedPnL ?? 0))} />
        <MetricCard label="Paper Exposure" value={money(portfolio.data?.Exposure)} />
      </div>
      <div className="mt-4 grid gap-4 xl:grid-cols-[0.75fr_1.25fr]">
        <Card>
          <h2 className="mb-4 font-semibold">Quick Paper Order</h2>
          <form className="space-y-3" onSubmit={submit}>
            <Field label="Stock"><input className={inputClass} value={symbol} onChange={(e) => setSymbol(e.target.value.toUpperCase())} /></Field>
            <Field label="Symbol Token"><input className={inputClass} value={token} onChange={(e) => setToken(e.target.value)} /></Field>
            <Field label="Quantity"><input className={inputClass} type="number" value={quantity} onChange={(e) => setQuantity(Number(e.target.value))} /></Field>
            <Field label="Entry Price"><input className={inputClass} type="number" value={price} onChange={(e) => setPrice(Number(e.target.value))} /></Field>
            <button disabled={!enabled || placeOrder.isPending} className="w-full rounded-md bg-cyan-500 px-4 py-2 font-semibold text-slate-950 disabled:opacity-50">
              {placeOrder.isPending ? "Placing" : "Place Paper BUY"}
            </button>
            {placeOrder.isError && <p className="text-sm text-rose-500">{placeOrder.error.message}</p>}
          </form>
        </Card>
        <Card>
          <h2 className="mb-4 font-semibold">Paper Positions</h2>
          <DataTable<Position>
            data={positions.data}
            empty="No paper positions"
            columns={[
              { key: "symbol", header: "Symbol", render: (row) => row.TradingSymbol ?? "-" },
              { key: "qty", header: "Qty", align: "right", render: (row) => row.Quantity ?? 0 },
              { key: "avg", header: "Avg", align: "right", render: (row) => money(row.AveragePrice) },
              { key: "pnl", header: "PnL", align: "right", render: (row) => money(row.UnrealizedPnL) }
            ]}
          />
        </Card>
      </div>
      <Card className="mt-4">
        <h2 className="mb-4 font-semibold">Paper Orders</h2>
        <DataTable<Order>
          data={(orders.data ?? []).filter((order) => order.Paper)}
          empty="No paper orders"
          columns={[
            { key: "symbol", header: "Symbol", render: (row) => row.TradingSymbol ?? "-" },
            { key: "side", header: "Side", render: (row) => <StatusPill label={row.TransactionType ?? "-"} tone="good" /> },
            { key: "qty", header: "Qty", align: "right", render: (row) => row.Quantity ?? 0 },
            { key: "status", header: "Status", render: (row) => <StatusPill label={row.Status ?? "-"} tone={row.Status === "rejected" ? "bad" : "good"} /> }
          ]}
        />
      </Card>
    </section>
  );
}
