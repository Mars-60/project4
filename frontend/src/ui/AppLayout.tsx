import { NavLink, Outlet, useLocation, useNavigate } from "react-router-dom";
import { clearTokens, getRefreshToken, services } from "../lib/api";
import { useTheme } from "../lib/theme";
import { StatusPill } from "../components/ui";

const links = [
  ["/", "Dashboard"],
  ["/strategies", "Strategy Builder"],
  ["/portfolio", "Portfolio"],
  ["/orders", "Orders"],
  ["/funds", "Funds"],
  ["/paper", "Paper Trading"],
  ["/ai", "AI Assistant"],
  ["/notifications", "Notifications"],
  ["/profile", "Profile"],
  ["/settings", "Settings"]
];

export function AppLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const { theme, toggleTheme } = useTheme();

  async function logout() {
    const refreshToken = getRefreshToken();
    if (refreshToken) {
      try {
        await services.auth.logout(refreshToken);
      } catch {
        // Local logout should still succeed if the server token was already revoked.
      }
    }
    clearTokens();
    navigate("/login");
  }

  return (
    <div className="min-h-screen bg-slate-100 text-slate-900 dark:bg-[#080b12] dark:text-slate-100">
      <aside className="fixed inset-y-0 left-0 z-30 hidden w-72 border-r border-slate-200 bg-white/95 px-4 py-5 backdrop-blur dark:border-slate-800 dark:bg-slate-950/95 lg:block">
        <div className="flex items-center gap-3 px-2">
          <div className="grid h-10 w-10 place-items-center rounded-lg bg-cyan-500 font-bold text-slate-950">TP</div>
          <div>
            <div className="font-semibold text-slate-950 dark:text-white">TradePilot AI</div>
            <div className="text-xs text-slate-500 dark:text-slate-400">Trading operations</div>
          </div>
        </div>
        <nav className="mt-7 space-y-1">
          {links.map(([to, label]) => (
            <NavLink
              key={to}
              to={to}
              className={({ isActive }) =>
                `block rounded-md px-3 py-2.5 text-sm font-medium transition ${
                  isActive
                    ? "bg-slate-950 text-white dark:bg-cyan-400 dark:text-slate-950"
                    : "text-slate-600 hover:bg-slate-100 hover:text-slate-950 dark:text-slate-400 dark:hover:bg-slate-900 dark:hover:text-white"
                }`
              }
            >
              {label}
            </NavLink>
          ))}
        </nav>
        <div className="absolute bottom-5 left-4 right-4 rounded-lg border border-slate-200 bg-slate-50 p-3 text-xs dark:border-slate-800 dark:bg-slate-900">
          <div className="flex items-center justify-between">
            <span className="text-slate-500 dark:text-slate-400">Bot status</span>
            <StatusPill label="Guarded" tone="good" />
          </div>
          <p className="mt-2 text-slate-500 dark:text-slate-400">AI can explain trades and risk, but cannot place orders.</p>
        </div>
      </aside>

      <main className="lg:pl-72">
        <header className="sticky top-0 z-20 border-b border-slate-200 bg-white/85 backdrop-blur dark:border-slate-800 dark:bg-slate-950/80">
          <div className="flex min-h-16 flex-col gap-3 px-4 py-3 sm:flex-row sm:items-center sm:justify-between lg:px-6">
            <div>
              <div className="text-xs uppercase tracking-[0.2em] text-slate-500 dark:text-slate-400">Workspace</div>
              <div className="font-semibold text-slate-950 dark:text-white">{pageName(location.pathname)}</div>
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <StatusPill label="Market: Watch" tone="warn" />
              <StatusPill label="Broker: Awaiting SMC" tone="neutral" />
              <button className="rounded-md border border-slate-300 px-3 py-2 text-sm dark:border-slate-700" onClick={toggleTheme}>
                {theme === "dark" ? "Light" : "Dark"}
              </button>
              <button className="rounded-md bg-slate-950 px-3 py-2 text-sm font-medium text-white dark:bg-cyan-400 dark:text-slate-950" onClick={logout}>
                Logout
              </button>
            </div>
          </div>
          <div className="flex gap-2 overflow-x-auto px-4 pb-3 lg:hidden">
            {links.map(([to, label]) => (
              <NavLink
                key={to}
                to={to}
                className={({ isActive }) =>
                  `whitespace-nowrap rounded-md px-3 py-2 text-sm ${isActive ? "bg-slate-950 text-white dark:bg-cyan-400 dark:text-slate-950" : "bg-slate-100 text-slate-700 dark:bg-slate-900 dark:text-slate-300"}`
                }
              >
                {label}
              </NavLink>
            ))}
          </div>
        </header>
        <div className="p-4 lg:p-6">
          <Outlet />
        </div>
      </main>
    </div>
  );
}

function pageName(pathname: string) {
  const found = links.find(([to]) => to === pathname);
  return found?.[1] ?? "Dashboard";
}
