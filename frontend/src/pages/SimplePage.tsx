import { useQuery } from "@tanstack/react-query";
import { api } from "../lib/api";

export function SimplePage({ title, endpoint }: { title: string; endpoint?: string }) {
  const query = useQuery({
    queryKey: [title, endpoint],
    queryFn: () => endpoint ? api<unknown>(endpoint) : Promise.resolve({ status: "ready" }),
    retry: false
  });

  return (
    <section>
      <h1 className="mb-4 text-2xl font-semibold">{title}</h1>
      <div className="rounded-lg border border-slate-200 bg-white p-4">
        {query.isLoading && <p className="text-slate-600">Loading</p>}
        {query.isError && <p className="text-red-700">{query.error.message}</p>}
        {query.isSuccess && <pre className="overflow-auto text-sm text-slate-700">{JSON.stringify(query.data, null, 2)}</pre>}
      </div>
    </section>
  );
}
