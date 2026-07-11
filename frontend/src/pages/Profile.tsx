import { useQuery } from "@tanstack/react-query";
import { Card, Field, inputClass, PageHeader, StatusPill } from "../components/ui";
import { services } from "../lib/api";

export function Profile() {
  const me = useQuery({ queryKey: ["me"], queryFn: services.auth.me });
  const email = me.data?.Email ?? me.data?.email ?? "";
  const name = me.data?.Name ?? me.data?.name ?? "TradePilot User";
  const role = me.data?.Role ?? me.data?.role ?? "user";

  return (
    <section>
      <PageHeader title="Profile" eyebrow="Account" action={<StatusPill label={role} tone="good" />} />
      <Card className="max-w-2xl">
        <div className="grid gap-4 sm:grid-cols-2">
          <Field label="Name"><input className={inputClass} value={name} readOnly /></Field>
          <Field label="Email"><input className={inputClass} value={email} readOnly /></Field>
          <Field label="Role"><input className={inputClass} value={role} readOnly /></Field>
          <Field label="Session"><input className={inputClass} value="JWT authenticated" readOnly /></Field>
        </div>
      </Card>
    </section>
  );
}
