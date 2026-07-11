import { FormEvent, useState } from "react";
import { useNavigate } from "react-router-dom";
import { services, setTokens } from "../lib/api";
import { inputClass } from "../components/ui";

export function Login() {
  const navigate = useNavigate();
  const [email, setEmail] = useState("admin@tradepilot.local");
  const [password, setPassword] = useState("");
  const [name, setName] = useState("TradePilot User");
  const [mode, setMode] = useState<"login" | "register">("login");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function submit(event: FormEvent) {
    event.preventDefault();
    setLoading(true);
    setError("");
    try {
      const payload = mode === "login" ? await services.auth.login(email, password) : await services.auth.register(email, password, name);
      setTokens(payload.tokens);
      navigate("/");
    } catch (err) {
      setError(err instanceof Error ? err.message : "Authentication failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <main className="min-h-screen bg-slate-950 text-white">
      <div className="grid min-h-screen lg:grid-cols-[1.1fr_0.9fr]">
        <section className="flex flex-col justify-between p-6 lg:p-10">
          <div className="flex items-center gap-3">
            <div className="grid h-11 w-11 place-items-center rounded-lg bg-cyan-400 font-bold text-slate-950">TP</div>
            <div>
              <div className="font-semibold">TradePilot AI</div>
              <div className="text-xs text-slate-400">Professional trading workspace</div>
            </div>
          </div>
          <div className="max-w-2xl py-16">
            <p className="text-sm font-semibold uppercase tracking-[0.28em] text-cyan-300">AI-assisted execution control</p>
            <h1 className="mt-4 text-4xl font-semibold leading-tight lg:text-6xl">Build, monitor, and explain trading strategies without writing code.</h1>
            <p className="mt-5 max-w-xl text-slate-300">A premium operations surface for portfolio visibility, paper trading, risk review, notifications, and visual strategy automation.</p>
          </div>
          <div className="grid gap-3 text-sm text-slate-300 sm:grid-cols-3">
            <span>SMC isolated</span>
            <span>AI cannot trade</span>
            <span>Risk first</span>
          </div>
        </section>
        <section className="grid place-items-center bg-white p-4 text-slate-950 dark:bg-slate-900 dark:text-white">
          <form onSubmit={submit} className="w-full max-w-md rounded-lg border border-slate-200 bg-white p-6 shadow-xl dark:border-slate-800 dark:bg-slate-950">
            <h2 className="text-xl font-semibold">{mode === "login" ? "Welcome back" : "Create workspace"}</h2>
            <p className="mt-1 text-sm text-slate-500 dark:text-slate-400">Use your TradePilot credentials.</p>
            <div className="mt-6 space-y-4">
              {mode === "register" && <label className="block text-sm">Name<input className={`${inputClass} mt-1`} value={name} onChange={(e) => setName(e.target.value)} /></label>}
              <label className="block text-sm">Email<input className={`${inputClass} mt-1`} value={email} onChange={(e) => setEmail(e.target.value)} /></label>
              <label className="block text-sm">Password<input className={`${inputClass} mt-1`} type="password" value={password} onChange={(e) => setPassword(e.target.value)} /></label>
            </div>
            {error && <div className="mt-4 rounded-md bg-rose-50 px-3 py-2 text-sm text-rose-700 dark:bg-rose-950 dark:text-rose-200">{error}</div>}
            <button disabled={loading} className="mt-5 w-full rounded-md bg-cyan-500 px-3 py-2 font-semibold text-slate-950 disabled:opacity-60">
              {loading ? "Please wait" : mode === "login" ? "Login" : "Create account"}
            </button>
            <button type="button" className="mt-4 w-full text-sm text-cyan-600 dark:text-cyan-300" onClick={() => setMode(mode === "login" ? "register" : "login")}>
              {mode === "login" ? "Create a new account" : "Use existing account"}
            </button>
          </form>
        </section>
      </div>
    </main>
  );
}
