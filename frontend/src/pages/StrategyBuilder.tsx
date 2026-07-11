import { FormEvent, useMemo, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Card, DataTable, Field, inputClass, PageHeader, StatusPill, Toggle } from "../components/ui";
import { AIConversation, services, StrategyDefinition } from "../lib/api";
import { dateTime } from "../lib/format";

const strategyTypes = [
  "Buy at Target Price",
  "Sell at Target Price",
  "Buying Pressure Strategy",
  "Volume Breakout",
  "Moving Average",
  "RSI",
  "MACD",
  "AI Suggested Strategy",
  "Custom Rule Builder"
];

type BuilderState = {
  name: string;
  type: string;
  broker: string;
  exchange: string;
  stock: string;
  symbolToken: string;
  quantity: number;
  orderType: string;
  entryPrice: number;
  targetPrice: number;
  stopLoss: number;
  trailingStop: number;
  maxLoss: number;
  maxProfit: number;
  startTime: string;
  endTime: string;
  squareOffTime: string;
  entryNotExecuted: string;
  enabled: boolean;
  indicatorPeriod: number;
  threshold: number;
  ruleLeft: string;
  ruleOperator: string;
  ruleRight: string;
};

const initialState: BuilderState = {
  name: "Morning breakout pilot",
  type: "Buy at Target Price",
  broker: "SMC",
  exchange: "NSE",
  stock: "RELIANCE",
  symbolToken: "2885",
  quantity: 10,
  orderType: "LIMIT",
  entryPrice: 2500,
  targetPrice: 2560,
  stopLoss: 2460,
  trailingStop: 20,
  maxLoss: 1500,
  maxProfit: 4000,
  startTime: "09:20",
  endTime: "14:45",
  squareOffTime: "15:15",
  entryNotExecuted: "Cancel",
  enabled: false,
  indicatorPeriod: 14,
  threshold: 60,
  ruleLeft: "last_price",
  ruleOperator: ">=",
  ruleRight: "entry_price"
};

export function StrategyBuilder() {
  const [state, setState] = useState<BuilderState>(initialState);
  const [explanation, setExplanation] = useState<AIConversation | null>(null);
  const client = useQueryClient();
  const strategies = useQuery({ queryKey: ["strategies"], queryFn: services.strategies.list });
  const create = useMutation({
    mutationFn: () =>
      services.strategies.create({
        name: state.name,
        description: `${state.type} for ${state.exchange}:${state.stock}`,
        config: buildConfig(state)
      }),
    onSuccess: async (created) => {
      if (state.enabled && created.ID) {
        await services.strategies.enable(created.ID);
      }
      client.invalidateQueries({ queryKey: ["strategies"] });
    }
  });
  const explain = useMutation({
    mutationFn: (id: string) => services.strategies.explain(id),
    onSuccess: setExplanation
  });

  const dynamicFields = useMemo(() => dynamicPanel(state.type), [state.type]);

  function patch<K extends keyof BuilderState>(key: K, value: BuilderState[K]) {
    setState((current) => ({ ...current, [key]: value }));
  }

  function submit(event: FormEvent) {
    event.preventDefault();
    create.mutate();
  }

  return (
    <section>
      <PageHeader title="Strategy Builder" eyebrow="No-code automation" />
      <div className="grid gap-4 xl:grid-cols-[0.95fr_1.05fr]">
        <Card>
          <form onSubmit={submit} className="space-y-5">
            <div className="grid gap-3 md:grid-cols-2">
              <Field label="Strategy Name"><input className={inputClass} value={state.name} onChange={(e) => patch("name", e.target.value)} /></Field>
              <Field label="Strategy Type">
                <select className={inputClass} value={state.type} onChange={(e) => patch("type", e.target.value)}>
                  {strategyTypes.map((type) => <option key={type}>{type}</option>)}
                </select>
              </Field>
              <Field label="Broker"><select className={inputClass} value={state.broker} onChange={(e) => patch("broker", e.target.value)}><option>SMC</option></select></Field>
              <Field label="Exchange"><select className={inputClass} value={state.exchange} onChange={(e) => patch("exchange", e.target.value)}><option>NSE</option><option>BSE</option><option>NFO</option></select></Field>
              <Field label="Stock"><input className={inputClass} value={state.stock} onChange={(e) => patch("stock", e.target.value.toUpperCase())} /></Field>
              <Field label="Symbol Token"><input className={inputClass} value={state.symbolToken} onChange={(e) => patch("symbolToken", e.target.value)} /></Field>
              <Field label="Quantity"><input className={inputClass} type="number" value={state.quantity} onChange={(e) => patch("quantity", Number(e.target.value))} /></Field>
              <Field label="Order Type"><select className={inputClass} value={state.orderType} onChange={(e) => patch("orderType", e.target.value)}><option>LIMIT</option><option>MARKET</option></select></Field>
              <Field label="Entry Price"><input className={inputClass} type="number" value={state.entryPrice} onChange={(e) => patch("entryPrice", Number(e.target.value))} /></Field>
              <Field label="Target Price"><input className={inputClass} type="number" value={state.targetPrice} onChange={(e) => patch("targetPrice", Number(e.target.value))} /></Field>
              <Field label="Stop Loss"><input className={inputClass} type="number" value={state.stopLoss} onChange={(e) => patch("stopLoss", Number(e.target.value))} /></Field>
              <Field label="Trailing Stop"><input className={inputClass} type="number" value={state.trailingStop} onChange={(e) => patch("trailingStop", Number(e.target.value))} /></Field>
              <Field label="Maximum Loss"><input className={inputClass} type="number" value={state.maxLoss} onChange={(e) => patch("maxLoss", Number(e.target.value))} /></Field>
              <Field label="Maximum Profit"><input className={inputClass} type="number" value={state.maxProfit} onChange={(e) => patch("maxProfit", Number(e.target.value))} /></Field>
              <Field label="Start Time"><input className={inputClass} type="time" value={state.startTime} onChange={(e) => patch("startTime", e.target.value)} /></Field>
              <Field label="End Time"><input className={inputClass} type="time" value={state.endTime} onChange={(e) => patch("endTime", e.target.value)} /></Field>
              <Field label="Square Off Time"><input className={inputClass} type="time" value={state.squareOffTime} onChange={(e) => patch("squareOffTime", e.target.value)} /></Field>
              <Field label="If Entry Not Executed">
                <select className={inputClass} value={state.entryNotExecuted} onChange={(e) => patch("entryNotExecuted", e.target.value)}>
                  <option>Cancel</option><option>Market Order</option><option>Retry</option><option>Notify User</option>
                </select>
              </Field>
            </div>

            <div className="rounded-lg border border-slate-200 p-4 dark:border-slate-800">
              <h2 className="mb-3 font-semibold">Dynamic Configuration</h2>
              <div className="grid gap-3 md:grid-cols-2">
                {dynamicFields.includes("period") && <Field label="Indicator Period"><input className={inputClass} type="number" value={state.indicatorPeriod} onChange={(e) => patch("indicatorPeriod", Number(e.target.value))} /></Field>}
                {dynamicFields.includes("threshold") && <Field label="Threshold"><input className={inputClass} type="number" value={state.threshold} onChange={(e) => patch("threshold", Number(e.target.value))} /></Field>}
                {dynamicFields.includes("rules") && (
                  <>
                    <Field label="Rule Left"><input className={inputClass} value={state.ruleLeft} onChange={(e) => patch("ruleLeft", e.target.value)} /></Field>
                    <Field label="Operator"><select className={inputClass} value={state.ruleOperator} onChange={(e) => patch("ruleOperator", e.target.value)}><option>&gt;=</option><option>&lt;=</option><option>==</option><option>crosses_above</option></select></Field>
                    <Field label="Rule Right"><input className={inputClass} value={state.ruleRight} onChange={(e) => patch("ruleRight", e.target.value)} /></Field>
                  </>
                )}
              </div>
            </div>

            <div className="flex flex-col gap-3 border-t border-slate-200 pt-4 dark:border-slate-800 sm:flex-row sm:items-center sm:justify-between">
              <div className="flex items-center gap-3">
                <Toggle checked={state.enabled} onChange={(checked) => patch("enabled", checked)} />
                <span className="text-sm font-medium">Enable Strategy after save</span>
              </div>
              <button className="rounded-md bg-cyan-500 px-5 py-2 font-semibold text-slate-950" disabled={create.isPending}>
                {create.isPending ? "Saving" : "Save Strategy"}
              </button>
            </div>
            {create.isError && <p className="text-sm text-rose-500">{create.error.message}</p>}
            {create.isSuccess && <p className="text-sm text-emerald-500">Strategy saved successfully.</p>}
          </form>
        </Card>

        <div className="space-y-4">
          <Card>
            <h2 className="mb-4 font-semibold">Strategy Preview</h2>
            <div className="space-y-3 text-sm">
              <Preview label="Rule" value={`${state.type}: ${state.stock} ${state.orderType} at ${state.entryPrice || "market"}`} />
              <Preview label="Risk" value={`SL ${state.stopLoss}, target ${state.targetPrice}, trail ${state.trailingStop}`} />
              <Preview label="Window" value={`${state.startTime} to ${state.endTime}, square-off ${state.squareOffTime}`} />
              <Preview label="Fallback" value={state.entryNotExecuted} />
            </div>
          </Card>
          <Card>
            <h2 className="mb-4 font-semibold">Saved Strategies</h2>
            <DataTable<StrategyDefinition>
              data={strategies.data}
              empty="No strategies saved"
              columns={[
                { key: "name", header: "Name", render: (row) => row.Name ?? "-" },
                { key: "status", header: "Status", render: (row) => <StatusPill label={row.Status ?? "disabled"} tone={row.Status === "enabled" ? "good" : "neutral"} /> },
                { key: "created", header: "Created", align: "right", render: (row) => dateTime(row.CreatedAt) },
                {
                  key: "ai",
                  header: "AI",
                  align: "right",
                  render: (row) => (
                    <button className="rounded-md border border-slate-300 px-2 py-1 text-xs dark:border-slate-700" onClick={() => row.ID && explain.mutate(row.ID)}>
                      Explain
                    </button>
                  )
                }
              ]}
            />
          </Card>
          {explanation && (
            <Card>
              <h2 className="mb-3 font-semibold">AI Strategy Explanation</h2>
              <p className="whitespace-pre-wrap text-sm text-slate-600 dark:text-slate-300">
                {lastMessage(explanation)?.Content ?? lastMessage(explanation)?.content}
              </p>
            </Card>
          )}
        </div>
      </div>
    </section>
  );
}

function buildConfig(state: BuilderState) {
  return Object.fromEntries(
    Object.entries({
      type: state.type,
      broker: state.broker,
      exchange: state.exchange,
      stock: state.stock,
      symbol_token: state.symbolToken,
      quantity: String(state.quantity),
      order_type: state.orderType,
      entry_price: String(state.entryPrice),
      target_price: String(state.targetPrice),
      stop_loss: String(state.stopLoss),
      trailing_stop: String(state.trailingStop),
      max_loss: String(state.maxLoss),
      max_profit: String(state.maxProfit),
      start_time: state.startTime,
      end_time: state.endTime,
      square_off_time: state.squareOffTime,
      entry_not_executed: state.entryNotExecuted,
      indicator_period: String(state.indicatorPeriod),
      threshold: String(state.threshold),
      rule_left: state.ruleLeft,
      rule_operator: state.ruleOperator,
      rule_right: state.ruleRight
    }).filter(([, value]) => value !== "")
  );
}

function dynamicPanel(type: string) {
  if (["Moving Average", "RSI", "MACD", "Volume Breakout", "Buying Pressure Strategy"].includes(type)) return ["period", "threshold"];
  if (type === "Custom Rule Builder" || type === "AI Suggested Strategy") return ["rules", "period", "threshold"];
  return ["threshold"];
}

function Preview({ label, value }: { label: string; value: string }) {
  return (
    <div className="flex items-start justify-between gap-4 rounded-lg bg-slate-50 p-3 dark:bg-slate-900">
      <span className="text-slate-500 dark:text-slate-400">{label}</span>
      <span className="text-right font-medium text-slate-900 dark:text-white">{value}</span>
    </div>
  );
}

function lastMessage(conversation: AIConversation) {
  const messages = conversation.Messages ?? [];
  return messages[messages.length - 1];
}
