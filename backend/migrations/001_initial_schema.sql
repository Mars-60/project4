create table if not exists users (
  id text primary key,
  email text not null unique,
  name text not null,
  role text not null,
  password_hash text not null,
  active boolean not null default true,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists strategies (
  id text primary key,
  user_id text not null references users(id),
  name text not null,
  description text not null,
  status text not null,
  config jsonb not null default '{}',
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists orders (
  id text primary key,
  user_id text not null references users(id),
  strategy_id text,
  broker_order_id text not null default '',
  exchange text not null,
  symbol_token text not null,
  trading_symbol text not null,
  transaction_type text not null,
  order_type text not null,
  product_type text not null,
  quantity integer not null,
  filled_quantity integer not null default 0,
  price numeric(18,4) not null,
  average_price numeric(18,4) not null,
  stop_loss numeric(18,4) not null default 0,
  target numeric(18,4) not null default 0,
  trailing_stop numeric(18,4) not null default 0,
  trail_by numeric(18,4) not null default 0,
  status text not null,
  reject_reason text not null default '',
  paper boolean not null default false,
  created_at timestamptz not null,
  updated_at timestamptz not null,
  unique(user_id, trading_symbol)
);

create table if not exists trades (
  id text primary key,
  user_id text not null references users(id),
  order_id text not null references orders(id),
  exchange text not null,
  trading_symbol text not null,
  transaction_type text not null,
  quantity integer not null,
  price numeric(18,4) not null,
  paper boolean not null default false,
  traded_at timestamptz not null
);

create table if not exists positions (
  id text primary key,
  user_id text not null references users(id),
  exchange text not null,
  trading_symbol text not null,
  product_type text not null,
  quantity integer not null,
  average_price numeric(18,4) not null,
  last_price numeric(18,4) not null,
  realized_pnl numeric(18,4) not null,
  unrealized_pnl numeric(18,4) not null,
  paper boolean not null default false,
  updated_at timestamptz not null,
  unique(user_id, trading_symbol, product_type, paper)
);

create table if not exists holdings (
  id text primary key,
  user_id text not null references users(id),
  exchange text not null,
  trading_symbol text not null,
  isin text not null,
  quantity integer not null,
  average_price numeric(18,4) not null,
  last_price numeric(18,4) not null,
  pnl numeric(18,4) not null,
  updated_at timestamptz not null
);

create table if not exists funds (
  user_id text primary key references users(id),
  available numeric(18,4) not null,
  used_margin numeric(18,4) not null,
  opening numeric(18,4) not null,
  net numeric(18,4) not null,
  paper_balance numeric(18,4) not null,
  updated_at timestamptz not null
);

create table if not exists ai_conversations (
  id text primary key,
  user_id text not null references users(id),
  title text not null,
  messages jsonb not null,
  created_at timestamptz not null,
  updated_at timestamptz not null
);

create table if not exists notifications (
  id text primary key,
  user_id text not null references users(id),
  channel text not null,
  subject text not null,
  body text not null,
  status text not null,
  created_at timestamptz not null
);

create table if not exists refresh_sessions (
  token text primary key,
  user_id text not null references users(id),
  role text not null,
  expires_at timestamptz not null,
  revoked_at timestamptz,
  created_at timestamptz not null
);

create table if not exists app_logs (
  id bigserial primary key,
  level text not null,
  message text not null,
  fields jsonb not null default '{}',
  created_at timestamptz not null default now()
);
