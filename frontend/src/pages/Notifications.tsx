import { FormEvent, useState } from "react";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { Card, DataTable, Field, inputClass, PageHeader, Toggle } from "../components/ui";
import { NotificationItem, services } from "../lib/api";
import { dateTime } from "../lib/format";

export function Notifications() {
  const [channels, setChannels] = useState({ email: true, telegram: true, push: false, whatsapp: false });
  const [subject, setSubject] = useState("Risk alert");
  const [body, setBody] = useState("Paper portfolio exposure needs review.");
  const [channel, setChannel] = useState("email");
  const client = useQueryClient();
  const notifications = useQuery({ queryKey: ["notifications"], queryFn: services.notifications.list });
  const send = useMutation({
    mutationFn: () => services.notifications.create(channel, subject, body),
    onSuccess: () => client.invalidateQueries({ queryKey: ["notifications"] })
  });

  function submit(event: FormEvent) {
    event.preventDefault();
    send.mutate();
  }

  return (
    <section>
      <PageHeader title="Notifications" eyebrow="History and channels" />
      <div className="grid gap-4 xl:grid-cols-[0.8fr_1.2fr]">
        <Card>
          <h2 className="mb-4 font-semibold">Channel Settings</h2>
          <div className="space-y-3">
            {Object.entries(channels).map(([name, enabled]) => (
              <div key={name} className="flex items-center justify-between rounded-lg bg-slate-50 p-3 capitalize dark:bg-slate-900">
                <span>{name}</span>
                <Toggle checked={enabled} onChange={(checked) => setChannels((current) => ({ ...current, [name]: checked }))} />
              </div>
            ))}
          </div>
          <form onSubmit={submit} className="mt-5 space-y-3">
            <Field label="Channel"><select className={inputClass} value={channel} onChange={(e) => setChannel(e.target.value)}><option>email</option><option>telegram</option><option>push</option><option>whatsapp</option></select></Field>
            <Field label="Subject"><input className={inputClass} value={subject} onChange={(e) => setSubject(e.target.value)} /></Field>
            <Field label="Body"><textarea className={`${inputClass} min-h-24`} value={body} onChange={(e) => setBody(e.target.value)} /></Field>
            <button className="w-full rounded-md bg-cyan-500 px-4 py-2 font-semibold text-slate-950">Send Test Notification</button>
          </form>
        </Card>
        <Card>
          <h2 className="mb-4 font-semibold">Notification History</h2>
          <DataTable<NotificationItem>
            data={notifications.data}
            empty="No notifications"
            columns={[
              { key: "channel", header: "Channel", render: (row) => row.Channel ?? "-" },
              { key: "subject", header: "Subject", render: (row) => row.Subject ?? "-" },
              { key: "status", header: "Status", render: (row) => row.Status ?? "-" },
              { key: "created", header: "Created", align: "right", render: (row) => dateTime(row.CreatedAt) }
            ]}
          />
        </Card>
      </div>
    </section>
  );
}
