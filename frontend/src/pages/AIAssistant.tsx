import { FormEvent, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { Card, EmptyState, Field, inputClass, PageHeader, StatusPill } from "../components/ui";
import { AIConversation, services } from "../lib/api";

export function AIAssistant() {
  const [prompt, setPrompt] = useState("Explain my paper portfolio risk and where exposure may be concentrated.");
  const [conversation, setConversation] = useState<AIConversation | null>(null);
  const ask = useMutation({ mutationFn: () => services.ai.ask(prompt), onSuccess: setConversation });
  const portfolio = useMutation({ mutationFn: services.ai.portfolio, onSuccess: setConversation });
  const market = useMutation({ mutationFn: services.ai.market, onSuccess: setConversation });
  const risk = useMutation({ mutationFn: services.ai.risk, onSuccess: setConversation });

  function submit(event: FormEvent) {
    event.preventDefault();
    ask.mutate();
  }

  return (
    <section>
      <PageHeader title="AI Assistant" eyebrow="Advisory only" action={<StatusPill label="Cannot place trades" tone="warn" />} />
      <div className="grid gap-4 xl:grid-cols-[0.8fr_1.2fr]">
        <Card>
          <form onSubmit={submit} className="space-y-4">
            <Field label="Ask TradePilot AI">
              <textarea className={`${inputClass} min-h-36 resize-none`} value={prompt} onChange={(event) => setPrompt(event.target.value)} />
            </Field>
            <button className="w-full rounded-md bg-cyan-500 px-4 py-2 font-semibold text-slate-950" disabled={ask.isPending}>
              {ask.isPending ? "Thinking" : "Ask"}
            </button>
          </form>
          <div className="mt-4 grid gap-2 sm:grid-cols-2">
            <QuickAction label="Portfolio Explanation" pending={portfolio.isPending} onClick={() => portfolio.mutate()} />
            <QuickAction label="Risk Explanation" pending={risk.isPending} onClick={() => risk.mutate()} />
            <QuickAction label="Market Explanation" pending={market.isPending} onClick={() => market.mutate()} />
            <QuickAction label="Strategy Explanation" pending={false} onClick={() => setPrompt("Explain my selected strategy configuration and risk controls.")} />
          </div>
        </Card>
        <Card>
          <h2 className="mb-4 font-semibold">Conversation</h2>
          {!conversation ? (
            <EmptyState title="No AI response yet" body="Ask a question or use a quick explanation action." />
          ) : (
            <div className="space-y-3">
              {conversation.Messages?.map((message, index) => (
                <div key={index} className={`rounded-lg p-4 ${message.Role === "assistant" || message.role === "assistant" ? "bg-cyan-50 text-slate-900 dark:bg-cyan-950/40 dark:text-slate-100" : "bg-slate-100 dark:bg-slate-900"}`}>
                  <div className="mb-1 text-xs font-semibold uppercase text-slate-500 dark:text-slate-400">{message.Role ?? message.role}</div>
                  <p className="whitespace-pre-wrap text-sm leading-6">{message.Content ?? message.content}</p>
                </div>
              ))}
            </div>
          )}
          {(ask.isError || portfolio.isError || market.isError || risk.isError) && <p className="mt-4 text-sm text-rose-500">AI request failed. Check Groq configuration.</p>}
        </Card>
      </div>
    </section>
  );
}

function QuickAction({ label, pending, onClick }: { label: string; pending: boolean; onClick: () => void }) {
  return (
    <button className="rounded-md border border-slate-300 px-3 py-2 text-left text-sm font-medium hover:bg-slate-50 dark:border-slate-700 dark:hover:bg-slate-900" onClick={onClick} disabled={pending}>
      {pending ? "Loading" : label}
    </button>
  );
}
