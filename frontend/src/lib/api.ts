import axios, { AxiosError } from "axios";

const API_BASE = import.meta.env.VITE_API_BASE_URL ?? "";

export type ApiResponse<T> = {
  success: boolean;
  data?: T;
  error?: string;
};

export type TokenPair = {
  access_token: string;
  refresh_token: string;
  expires_in?: number;
};

export type User = {
  ID?: string;
  Email?: string;
  Name?: string;
  Role?: string;
  email?: string;
  name?: string;
  role?: string;
};

export type PortfolioSummary = {
  UserID?: string;
  NetValue?: number;
  Cash?: number;
  UsedMargin?: number;
  RealizedPnL?: number;
  UnrealizedPnL?: number;
  Exposure?: number;
  UpdatedAt?: string;
};

export type Funds = {
  Available?: number;
  UsedMargin?: number;
  Opening?: number;
  Net?: number;
  PaperBalance?: number;
  UpdatedAt?: string;
};

export type Order = {
  ID?: string;
  TradingSymbol?: string;
  Exchange?: string;
  TransactionType?: string;
  OrderType?: string;
  ProductType?: string;
  Quantity?: number;
  Price?: number;
  AveragePrice?: number;
  StopLoss?: number;
  Target?: number;
  Status?: string;
  RejectReason?: string;
  Paper?: boolean;
  CreatedAt?: string;
};

export type Trade = {
  ID?: string;
  TradingSymbol?: string;
  Exchange?: string;
  TransactionType?: string;
  Quantity?: number;
  Price?: number;
  Paper?: boolean;
  TradedAt?: string;
};

export type Position = {
  ID?: string;
  Exchange?: string;
  TradingSymbol?: string;
  ProductType?: string;
  Quantity?: number;
  AveragePrice?: number;
  LastPrice?: number;
  RealizedPnL?: number;
  UnrealizedPnL?: number;
  Paper?: boolean;
  UpdatedAt?: string;
};

export type Holding = {
  ID?: string;
  Exchange?: string;
  TradingSymbol?: string;
  ISIN?: string;
  Quantity?: number;
  AveragePrice?: number;
  LastPrice?: number;
  PnL?: number;
  UpdatedAt?: string;
};

export type StrategyDefinition = {
  ID?: string;
  UserID?: string;
  Name?: string;
  Description?: string;
  Status?: "enabled" | "disabled";
  Config?: Record<string, string>;
  CreatedAt?: string;
  UpdatedAt?: string;
};

export type NotificationItem = {
  ID?: string;
  Channel?: string;
  Subject?: string;
  Body?: string;
  Status?: string;
  CreatedAt?: string;
};

export type AIConversation = {
  ID?: string;
  Title?: string;
  Messages?: Array<{ Role?: string; Content?: string; CreatedAt?: string; role?: string; content?: string }>;
};

export type MarketSnapshot = {
  Exchange?: string;
  SymbolToken?: string;
  TradingSymbol?: string;
  LastPrice?: number;
  Open?: number;
  High?: number;
  Low?: number;
  Close?: number;
  Volume?: number;
  Timestamp?: string;
};

export function getToken() {
  return localStorage.getItem("tradepilot.access_token");
}

export function getRefreshToken() {
  return localStorage.getItem("tradepilot.refresh_token");
}

export function setTokens(tokens: TokenPair) {
  localStorage.setItem("tradepilot.access_token", tokens.access_token);
  localStorage.setItem("tradepilot.refresh_token", tokens.refresh_token);
}

export function clearTokens() {
  localStorage.removeItem("tradepilot.access_token");
  localStorage.removeItem("tradepilot.refresh_token");
}

export const http = axios.create({
  baseURL: `${API_BASE}/api/v1`,
  headers: { "Content-Type": "application/json" }
});

http.interceptors.request.use((config) => {
  const token = getToken();
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

http.interceptors.response.use(
  (response) => {
    const payload = response.data as ApiResponse<unknown>;
    if (payload && payload.success === false) {
      throw new Error(payload.error ?? "Request failed");
    }
    return response;
  },
  (error: AxiosError<ApiResponse<unknown>>) => {
    const message = error.response?.data?.error ?? error.message ?? "Request failed";
    return Promise.reject(new Error(message));
  }
);

export async function api<T>(path: string, init: { method?: string; body?: string } = {}): Promise<T> {
  const response = await http.request<ApiResponse<T>>({
    url: path,
    method: init.method ?? "GET",
    data: init.body ? JSON.parse(init.body) : undefined
  });
  return response.data.data as T;
}

export const services = {
  auth: {
    login: (email: string, password: string) =>
      api<{ user: User; tokens: TokenPair }>("/auth/login", { method: "POST", body: JSON.stringify({ email, password }) }),
    register: (email: string, password: string, name: string) =>
      api<{ user: User; tokens: TokenPair }>("/auth/register", { method: "POST", body: JSON.stringify({ email, password, name }) }),
    me: () => api<User>("/auth/me"),
    logout: (refreshToken: string) =>
      api<{ status: string }>("/auth/logout", { method: "POST", body: JSON.stringify({ refresh_token: refreshToken }) })
  },
  dashboard: {
    portfolio: (paper = true) => api<PortfolioSummary>(`/portfolio/summary?paper=${paper}`),
    funds: () => api<Funds>("/funds"),
    positions: (paper = true) => api<Position[]>(`/portfolio/positions?paper=${paper}`),
    trades: () => api<Trade[]>("/trades?limit=8")
  },
  portfolio: {
    summary: (paper = false) => api<PortfolioSummary>(`/portfolio/summary?paper=${paper}`),
    positions: (paper = false) => api<Position[]>(`/portfolio/positions?paper=${paper}`),
    holdings: () => api<Holding[]>("/portfolio/holdings"),
    trades: () => api<Trade[]>("/trades?limit=50")
  },
  orders: {
    list: () => api<Order[]>("/orders?limit=100"),
    paper: (payload: unknown) => api<Order>("/paper/orders", { method: "POST", body: JSON.stringify(payload) })
  },
  strategies: {
    list: () => api<StrategyDefinition[]>("/strategies"),
    create: (payload: unknown) => api<StrategyDefinition>("/strategies", { method: "POST", body: JSON.stringify(payload) }),
    enable: (id: string) => api<{ status: string }>(`/strategies/${id}/enable`, { method: "POST" }),
    disable: (id: string) => api<{ status: string }>(`/strategies/${id}/disable`, { method: "POST" }),
    explain: (id: string) => api<AIConversation>(`/strategies/${id}/explain`, { method: "POST" })
  },
  market: {
    quote: (exchange: string, symbolToken: string) => api<MarketSnapshot>(`/market/quote?exchange=${exchange}&symbol_token=${symbolToken}`)
  },
  ai: {
    ask: (prompt: string) => api<AIConversation>("/ai/ask", { method: "POST", body: JSON.stringify({ prompt }) }),
    portfolio: () => api<AIConversation>("/ai/portfolio", { method: "POST" }),
    market: () => api<AIConversation>("/ai/market?exchange=NSE&symbol_token=26000", { method: "POST" }),
    risk: () => api<AIConversation>("/ai/risk?paper=true", { method: "POST" })
  },
  notifications: {
    list: () => api<NotificationItem[]>("/notifications"),
    create: (channel: string, subject: string, body: string) =>
      api<NotificationItem>("/notifications", { method: "POST", body: JSON.stringify({ channel, subject, body }) })
  },
  system: {
    metrics: () => api<Record<string, unknown>>("/system/metrics")
  }
};

export function numberValue(record: Record<string, unknown> | undefined, key: string, fallback = 0) {
  const value = record?.[key] ?? record?.[key.charAt(0).toLowerCase() + key.slice(1)];
  return typeof value === "number" ? value : fallback;
}
