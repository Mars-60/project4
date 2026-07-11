import { useQuery } from "@tanstack/react-query";
import { Card, DataTable, PageHeader, StatusPill } from "../components/ui";
import { Order, services } from "../lib/api";
import { dateTime, money } from "../lib/format";

export function Orders() {
  const orders = useQuery({ queryKey: ["orders"], queryFn: services.orders.list });
  const rows = orders.data ?? [];

  return (
    <section>
      <PageHeader title="Orders" eyebrow="Pending, executed, rejected, cancelled" />
      <div className="grid gap-4 sm:grid-cols-4">
        <OrderBucket label="Pending" value={count(rows, "pending")} tone="warn" />
        <OrderBucket label="Executed" value={count(rows, "filled") + count(rows, "placed")} tone="good" />
        <OrderBucket label="Rejected" value={count(rows, "rejected")} tone="bad" />
        <OrderBucket label="Cancelled" value={count(rows, "cancelled")} tone="neutral" />
      </div>
      <Card className="mt-4">
        <DataTable<Order>
          data={rows}
          empty="No orders"
          columns={[
            { key: "symbol", header: "Symbol", render: (row) => row.TradingSymbol ?? "-" },
            { key: "side", header: "Side", render: (row) => <StatusPill label={row.TransactionType ?? "-"} tone={row.TransactionType === "BUY" ? "good" : "bad"} /> },
            { key: "qty", header: "Qty", align: "right", render: (row) => row.Quantity ?? 0 },
            { key: "price", header: "Price", align: "right", render: (row) => money(row.Price) },
            { key: "status", header: "Status", render: (row) => <StatusPill label={row.Status ?? "-"} tone={statusTone(row.Status)} /> },
            { key: "time", header: "Created", align: "right", render: (row) => dateTime(row.CreatedAt) },
            { key: "actions", header: "Actions", align: "right", render: () => <button className="rounded-md border border-slate-300 px-2 py-1 text-xs opacity-60 dark:border-slate-700">Broker gated</button> }
          ]}
        />
      </Card>
    </section>
  );
}

function count(orders: Order[], status: string) {
  return orders.filter((order) => order.Status === status).length;
}

function statusTone(status?: string) {
  if (status === "filled" || status === "placed") return "good" as const;
  if (status === "rejected") return "bad" as const;
  if (status === "pending" || status === "validated") return "warn" as const;
  return "neutral" as const;
}

function OrderBucket({ label, value, tone }: { label: string; value: number; tone: "neutral" | "good" | "bad" | "warn" }) {
  return (
    <Card>
      <div className="flex items-center justify-between">
        <span className="text-sm text-slate-500 dark:text-slate-400">{label}</span>
        <StatusPill label={`${value}`} tone={tone} />
      </div>
    </Card>
  );
}
