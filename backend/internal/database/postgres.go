package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/Mars-60/project4/backend/internal/core"
)

var ErrNotFound = errors.New("record not found")

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *DB) *PostgresRepository {
	return &PostgresRepository{db: db.SQL()}
}

func (r *PostgresRepository) WithinTx(ctx context.Context, fn func(context.Context) error) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	txCtx := context.WithValue(ctx, txKey{}, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *PostgresRepository) CreateUser(ctx context.Context, user core.User) (core.User, error) {
	_, err := executor(ctx, r.db).ExecContext(ctx, `
		insert into users (id,email,name,role,password_hash,active,created_at,updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8)`,
		user.ID, user.Email, user.Name, user.Role, user.PasswordHash, user.Active, user.CreatedAt, user.UpdatedAt)
	return user, err
}

func (r *PostgresRepository) FindUserByEmail(ctx context.Context, email string) (core.User, error) {
	return r.scanUser(executor(ctx, r.db).QueryRowContext(ctx, `
		select id,email,name,role,password_hash,active,created_at,updated_at from users where email=$1`, email))
}

func (r *PostgresRepository) FindUserByID(ctx context.Context, id core.ID) (core.User, error) {
	return r.scanUser(executor(ctx, r.db).QueryRowContext(ctx, `
		select id,email,name,role,password_hash,active,created_at,updated_at from users where id=$1`, id))
}

func (r *PostgresRepository) scanUser(row *sql.Row) (core.User, error) {
	var user core.User
	err := row.Scan(&user.ID, &user.Email, &user.Name, &user.Role, &user.PasswordHash, &user.Active, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return core.User{}, ErrNotFound
	}
	return user, err
}

func (r *PostgresRepository) CreateStrategy(ctx context.Context, strategy core.StrategyDefinition) (core.StrategyDefinition, error) {
	cfg, err := json.Marshal(strategy.Config)
	if err != nil {
		return core.StrategyDefinition{}, err
	}
	_, err = executor(ctx, r.db).ExecContext(ctx, `
		insert into strategies (id,user_id,name,description,status,config,created_at,updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8)`,
		strategy.ID, strategy.UserID, strategy.Name, strategy.Description, strategy.Status, cfg, strategy.CreatedAt, strategy.UpdatedAt)
	return strategy, err
}

func (r *PostgresRepository) ListStrategies(ctx context.Context, userID core.ID) ([]core.StrategyDefinition, error) {
	rows, err := executor(ctx, r.db).QueryContext(ctx, `
		select id,user_id,name,description,status,config,created_at,updated_at from strategies where user_id=$1 order by created_at desc`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var strategies []core.StrategyDefinition
	for rows.Next() {
		var strategy core.StrategyDefinition
		var cfg []byte
		if err := rows.Scan(&strategy.ID, &strategy.UserID, &strategy.Name, &strategy.Description, &strategy.Status, &cfg, &strategy.CreatedAt, &strategy.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(cfg, &strategy.Config)
		strategies = append(strategies, strategy)
	}
	return strategies, rows.Err()
}

func (r *PostgresRepository) UpdateStrategyStatus(ctx context.Context, userID core.ID, strategyID core.ID, status core.StrategyStatus) error {
	result, err := executor(ctx, r.db).ExecContext(ctx, `update strategies set status=$1, updated_at=now() where id=$2 and user_id=$3`, status, strategyID, userID)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) CreateOrder(ctx context.Context, order core.Order) (core.Order, error) {
	_, err := executor(ctx, r.db).ExecContext(ctx, `
		insert into orders (id,user_id,strategy_id,broker_order_id,exchange,symbol_token,trading_symbol,transaction_type,order_type,product_type,quantity,filled_quantity,price,average_price,stop_loss,target,trailing_stop,trail_by,status,reject_reason,paper,created_at,updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23)`,
		order.ID, order.UserID, order.StrategyID, order.BrokerOrderID, order.Exchange, order.SymbolToken, order.TradingSymbol, order.TransactionType, order.OrderType, order.ProductType, order.Quantity, order.FilledQuantity, order.Price, order.AveragePrice, order.StopLoss, order.Target, order.TrailingStop, order.TrailBy, order.Status, order.RejectReason, order.Paper, order.CreatedAt, order.UpdatedAt)
	return order, err
}

func (r *PostgresRepository) UpdateOrder(ctx context.Context, order core.Order) error {
	_, err := executor(ctx, r.db).ExecContext(ctx, `
		update orders set broker_order_id=$1, filled_quantity=$2, average_price=$3, status=$4, reject_reason=$5, updated_at=$6 where id=$7`,
		order.BrokerOrderID, order.FilledQuantity, order.AveragePrice, order.Status, order.RejectReason, order.UpdatedAt, order.ID)
	return err
}

func (r *PostgresRepository) ListOrders(ctx context.Context, userID core.ID, filter core.PageFilter) ([]core.Order, error) {
	filter = core.NormalizePageFilter(filter)
	rows, err := executor(ctx, r.db).QueryContext(ctx, `
		select id,user_id,strategy_id,broker_order_id,exchange,symbol_token,trading_symbol,transaction_type,order_type,product_type,quantity,filled_quantity,price,average_price,stop_loss,target,trailing_stop,trail_by,status,reject_reason,paper,created_at,updated_at
		from orders where user_id=$1 order by created_at desc limit $2 offset $3`, userID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var orders []core.Order
	for rows.Next() {
		var order core.Order
		if err := rows.Scan(&order.ID, &order.UserID, &order.StrategyID, &order.BrokerOrderID, &order.Exchange, &order.SymbolToken, &order.TradingSymbol, &order.TransactionType, &order.OrderType, &order.ProductType, &order.Quantity, &order.FilledQuantity, &order.Price, &order.AveragePrice, &order.StopLoss, &order.Target, &order.TrailingStop, &order.TrailBy, &order.Status, &order.RejectReason, &order.Paper, &order.CreatedAt, &order.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, rows.Err()
}

func (r *PostgresRepository) CreateTrade(ctx context.Context, trade core.Trade) (core.Trade, error) {
	_, err := executor(ctx, r.db).ExecContext(ctx, `
		insert into trades (id,user_id,order_id,exchange,trading_symbol,transaction_type,quantity,price,paper,traded_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		trade.ID, trade.UserID, trade.OrderID, trade.Exchange, trade.TradingSymbol, trade.TransactionType, trade.Quantity, trade.Price, trade.Paper, trade.TradedAt)
	return trade, err
}

func (r *PostgresRepository) ListTrades(ctx context.Context, userID core.ID, filter core.PageFilter) ([]core.Trade, error) {
	filter = core.NormalizePageFilter(filter)
	rows, err := executor(ctx, r.db).QueryContext(ctx, `
		select id,user_id,order_id,exchange,trading_symbol,transaction_type,quantity,price,paper,traded_at
		from trades where user_id=$1 order by traded_at desc limit $2 offset $3`, userID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var trades []core.Trade
	for rows.Next() {
		var trade core.Trade
		if err := rows.Scan(&trade.ID, &trade.UserID, &trade.OrderID, &trade.Exchange, &trade.TradingSymbol, &trade.TransactionType, &trade.Quantity, &trade.Price, &trade.Paper, &trade.TradedAt); err != nil {
			return nil, err
		}
		trades = append(trades, trade)
	}
	return trades, rows.Err()
}

func (r *PostgresRepository) UpsertPosition(ctx context.Context, position core.Position) error {
	_, err := executor(ctx, r.db).ExecContext(ctx, `
		insert into positions (id,user_id,exchange,trading_symbol,product_type,quantity,average_price,last_price,realized_pnl,unrealized_pnl,paper,updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)
		on conflict (user_id, trading_symbol, product_type, paper)
		do update set quantity=$6, average_price=$7, last_price=$8, realized_pnl=$9, unrealized_pnl=$10, updated_at=$12`,
		position.ID, position.UserID, position.Exchange, position.TradingSymbol, position.ProductType, position.Quantity, position.AveragePrice, position.LastPrice, position.RealizedPnL, position.UnrealizedPnL, position.Paper, position.UpdatedAt)
	return err
}

func (r *PostgresRepository) ListPositions(ctx context.Context, userID core.ID, paper bool) ([]core.Position, error) {
	rows, err := executor(ctx, r.db).QueryContext(ctx, `
		select id,user_id,exchange,trading_symbol,product_type,quantity,average_price,last_price,realized_pnl,unrealized_pnl,paper,updated_at
		from positions where user_id=$1 and paper=$2 order by updated_at desc`, userID, paper)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var positions []core.Position
	for rows.Next() {
		var position core.Position
		if err := rows.Scan(&position.ID, &position.UserID, &position.Exchange, &position.TradingSymbol, &position.ProductType, &position.Quantity, &position.AveragePrice, &position.LastPrice, &position.RealizedPnL, &position.UnrealizedPnL, &position.Paper, &position.UpdatedAt); err != nil {
			return nil, err
		}
		positions = append(positions, position)
	}
	return positions, rows.Err()
}

func (r *PostgresRepository) UpsertHolding(ctx context.Context, holding core.Holding) error {
	_, err := executor(ctx, r.db).ExecContext(ctx, `
		insert into holdings (id,user_id,exchange,trading_symbol,isin,quantity,average_price,last_price,pnl,updated_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		on conflict (user_id, trading_symbol)
		do update set quantity=$6, average_price=$7, last_price=$8, pnl=$9, updated_at=$10`,
		holding.ID, holding.UserID, holding.Exchange, holding.TradingSymbol, holding.ISIN, holding.Quantity, holding.AveragePrice, holding.LastPrice, holding.PnL, holding.UpdatedAt)
	return err
}

func (r *PostgresRepository) ListHoldings(ctx context.Context, userID core.ID) ([]core.Holding, error) {
	rows, err := executor(ctx, r.db).QueryContext(ctx, `
		select id,user_id,exchange,trading_symbol,isin,quantity,average_price,last_price,pnl,updated_at
		from holdings where user_id=$1 order by trading_symbol`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var holdings []core.Holding
	for rows.Next() {
		var holding core.Holding
		if err := rows.Scan(&holding.ID, &holding.UserID, &holding.Exchange, &holding.TradingSymbol, &holding.ISIN, &holding.Quantity, &holding.AveragePrice, &holding.LastPrice, &holding.PnL, &holding.UpdatedAt); err != nil {
			return nil, err
		}
		holdings = append(holdings, holding)
	}
	return holdings, rows.Err()
}

func (r *PostgresRepository) GetFunds(ctx context.Context, userID core.ID) (core.Funds, error) {
	var funds core.Funds
	err := executor(ctx, r.db).QueryRowContext(ctx, `
		select user_id,available,used_margin,opening,net,paper_balance,updated_at from funds where user_id=$1`, userID).
		Scan(&funds.UserID, &funds.Available, &funds.UsedMargin, &funds.Opening, &funds.Net, &funds.PaperBalance, &funds.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return core.Funds{UserID: userID, Available: 1000000, Net: 1000000, PaperBalance: 1000000, UpdatedAt: time.Now().UTC()}, nil
	}
	return funds, err
}

func (r *PostgresRepository) SaveFunds(ctx context.Context, funds core.Funds) error {
	_, err := executor(ctx, r.db).ExecContext(ctx, `
		insert into funds (user_id,available,used_margin,opening,net,paper_balance,updated_at)
		values ($1,$2,$3,$4,$5,$6,$7)
		on conflict (user_id) do update set available=$2, used_margin=$3, opening=$4, net=$5, paper_balance=$6, updated_at=$7`,
		funds.UserID, funds.Available, funds.UsedMargin, funds.Opening, funds.Net, funds.PaperBalance, funds.UpdatedAt)
	return err
}

func (r *PostgresRepository) SaveConversation(ctx context.Context, conversation core.AIConversation) error {
	payload, err := json.Marshal(conversation.Messages)
	if err != nil {
		return err
	}
	_, err = executor(ctx, r.db).ExecContext(ctx, `
		insert into ai_conversations (id,user_id,title,messages,created_at,updated_at)
		values ($1,$2,$3,$4,$5,$6)
		on conflict (id) do update set title=$3,messages=$4,updated_at=$6`,
		conversation.ID, conversation.UserID, conversation.Title, payload, conversation.CreatedAt, conversation.UpdatedAt)
	return err
}

func (r *PostgresRepository) ListConversations(ctx context.Context, userID core.ID, filter core.PageFilter) ([]core.AIConversation, error) {
	filter = core.NormalizePageFilter(filter)
	rows, err := executor(ctx, r.db).QueryContext(ctx, `select id,user_id,title,messages,created_at,updated_at from ai_conversations where user_id=$1 order by updated_at desc limit $2 offset $3`, userID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var conversations []core.AIConversation
	for rows.Next() {
		var conversation core.AIConversation
		var payload []byte
		if err := rows.Scan(&conversation.ID, &conversation.UserID, &conversation.Title, &payload, &conversation.CreatedAt, &conversation.UpdatedAt); err != nil {
			return nil, err
		}
		_ = json.Unmarshal(payload, &conversation.Messages)
		conversations = append(conversations, conversation)
	}
	return conversations, rows.Err()
}

func (r *PostgresRepository) CreateNotification(ctx context.Context, notification core.Notification) (core.Notification, error) {
	_, err := executor(ctx, r.db).ExecContext(ctx, `insert into notifications (id,user_id,channel,subject,body,status,created_at) values ($1,$2,$3,$4,$5,$6,$7)`,
		notification.ID, notification.UserID, notification.Channel, notification.Subject, notification.Body, notification.Status, notification.CreatedAt)
	return notification, err
}

func (r *PostgresRepository) ListNotifications(ctx context.Context, userID core.ID, filter core.PageFilter) ([]core.Notification, error) {
	filter = core.NormalizePageFilter(filter)
	rows, err := executor(ctx, r.db).QueryContext(ctx, `select id,user_id,channel,subject,body,status,created_at from notifications where user_id=$1 order by created_at desc limit $2 offset $3`, userID, filter.Limit, filter.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var notifications []core.Notification
	for rows.Next() {
		var notification core.Notification
		if err := rows.Scan(&notification.ID, &notification.UserID, &notification.Channel, &notification.Subject, &notification.Body, &notification.Status, &notification.CreatedAt); err != nil {
			return nil, err
		}
		notifications = append(notifications, notification)
	}
	return notifications, rows.Err()
}

func (r *PostgresRepository) SaveRefreshSession(ctx context.Context, session core.RefreshSession) error {
	_, err := executor(ctx, r.db).ExecContext(ctx, `
		insert into refresh_sessions (token,user_id,role,expires_at,revoked_at,created_at)
		values ($1,$2,$3,$4,$5,$6)
		on conflict (token) do update set revoked_at=$5`,
		session.Token, session.UserID, session.Role, session.ExpiresAt, session.RevokedAt, session.CreatedAt)
	return err
}

func (r *PostgresRepository) FindRefreshSession(ctx context.Context, token string) (core.RefreshSession, error) {
	var session core.RefreshSession
	var revoked sql.NullTime
	err := executor(ctx, r.db).QueryRowContext(ctx, `
		select token,user_id,role,expires_at,revoked_at,created_at from refresh_sessions where token=$1`, token).
		Scan(&session.Token, &session.UserID, &session.Role, &session.ExpiresAt, &revoked, &session.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return core.RefreshSession{}, ErrNotFound
	}
	if err != nil {
		return core.RefreshSession{}, err
	}
	if revoked.Valid {
		session.RevokedAt = &revoked.Time
	}
	return session, nil
}

func (r *PostgresRepository) RevokeRefreshSession(ctx context.Context, token string, revokedAt time.Time) error {
	result, err := executor(ctx, r.db).ExecContext(ctx, `update refresh_sessions set revoked_at=$1 where token=$2`, revokedAt, token)
	if err != nil {
		return err
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}
