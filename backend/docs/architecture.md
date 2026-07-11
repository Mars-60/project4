# Architecture

TradePilot AI follows a clean dependency direction:

`HTTP -> application container -> services -> core ports -> infrastructure`

The core package owns trading concepts such as strategies, orders, positions, funds, PnL, risk decisions, paper trading, and scheduling. It does not import HTTP, SQL, Groq, or SMC packages.

Broker integrations are isolated behind `internal/broker.Broker`. New brokers should implement this interface in their own package.

Database repositories implement core repository interfaces. PostgreSQL SQL statements stay inside `internal/database`.

Order execution runs through a transaction-aware service. Paper orders are validated by the risk engine, persisted as orders and trades, folded into aggregate positions, and reflected in paper funds. The same core pipeline can be extended to live execution by adding a broker-backed execution adapter without changing API handlers.

Authentication uses signed JWT access tokens and repository-backed refresh sessions. Refresh tokens are rotated on use and can be revoked through logout.

AI services are advisory only. They can analyze portfolios, explain strategies, and answer questions, but they cannot execute orders.

Notification delivery is provider-based. Email, Telegram, WhatsApp, and push channels are registered through the same interface, so production providers can replace log/webhook implementations independently.
