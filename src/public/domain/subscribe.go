package domain

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// 订阅计划
type Plan struct {
	ID          int
	Name        string
	Description string
	Price       float64
	Period      time.Duration // 订阅周期，如1个月、3个月等
	Features    []string      // 订阅包含的功能列表
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func CreatePlanTable(ctx context.Context, conn *pgxpool.Conn) error {
	query := `
		CREATE TABLE IF NOT EXISTS plans (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			description TEXT NOT NULL,
			price NUMERIC(10, 2) NOT NULL,
			period INTERVAL NOT NULL,
			features TEXT[] NOT NULL,
			status VARCHAR(255) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_plans_name ON plans(name);
		CREATE INDEX IF NOT EXISTS idx_plans_status ON plans(status);
		
		DROP TRIGGER IF EXISTS update_plans ON plans;
		
		CREATE TRIGGER update_plans
		    BEFORE UPDATE ON plans
		    FOR EACH ROW
		    EXECUTE FUNCTION update_column();
	` // update_column 在创建schema 时定义的在DBManager中
	
	_, err := conn.Exec(ctx, query)
	return err
}

type Subscribe struct {
	ID        int
	TenantID  int       //外键
	PlanID    int       //外键
	Start     time.Time // 订阅开始时间
	End       time.Time // 订阅结束时间
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func CreateSubscribeTable(ctx context.Context, conn *pgxpool.Conn) error {
	query := `
		CREATE TABLE IF NOT EXISTS subscribes (
			id SERIAL PRIMARY KEY,
			tenant_id INTEGER REFERENCES tenants(id) NOT NULL,
			plan_id INTEGER REFERENCES plans(id) NOT NULL,
			start TIMESTAMP NOT NULL,
			end TIMESTAMP NOT NULL,
			status VARCHAR(50) NOT NULL CHECK (status IN ('active', 'expired', 'canceled', 'pending')),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			
			CONSTRAINT end_after_start CHECK (end > start),
		);
		
		
		CREATE INDEX IF NOT EXISTS idx_subscribes_tenant_id ON subscribes(tenant_id);
		CREATE INDEX IF NOT EXISTS idx_subscribes_plan_id ON subscribes(plan_id);
		CREATE INDEX IF NOT EXISTS idx_subscribes_status ON subscribes(status);
		CREATE INDEX IF NOT EXISTS idx_subscribes_dates ON subscribes(start, end);
        
        DROP TRIGGER IF EXISTS update_subscribes ON subscribes;
        
		CREATE TRIGGER update_subscribes
		    BEFORE UPDATE ON subscribes
		    FOR EACH ROW
		    EXECUTE FUNCTION update_column();
	` // update_column 在创建schema 时定义的在DBManager中
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return err
}

type Order struct {
	ID          int
	SubscribeID int
	Amount      float64 // 订单金额
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func CreateOrderTable(ctx context.Context, conn *pgxpool.Conn) error {
	query := `
		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			subscribe_id INTEGER REFERENCES subscribes(id) NOT NULL,
			amount NUMERIC(10, 2) NOT NULL,
			status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'paid', 'failed', 'cancelled')),
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			
			CONSTRAINT order_amount_positive CHECK (amount > 0)
		);
		
		
		CREATE INDEX IF NOT EXISTS idx_orders_subscribe_id ON orders(subscribe_id);
		CREATE INDEX IF NOT EXISTS idx_orders_payment_id ON orders(payment_id);
		CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
		CREATE INDEX IF NOT EXISTS idx_orders_dates ON orders(created_at, updated_at);
        
        DROP TRIGGER IF EXISTS update_orders ON orders;
        
		CREATE TRIGGER update_orders
		    BEFORE UPDATE ON orders
		    FOR EACH ROW
		    EXECUTE FUNCTION update_column();
	` // update_column 在创建schema 时定义的在DBManager中
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return err
}


type Payment struct {
	ID          int
	OrderID     int
	Amount      float64
	PaymentType string
	Status      string
	CreatedAt   time.Time
}

func CreatePaymentTable(ctx context.Context, conn *pgxpool.Conn) error {
	query := `
		CREATE TABLE IF NOT EXISTS payments (
			id SERIAL PRIMARY KEY,
			order_id INTEGER REFERENCES orders(id) NOT NULL,
			payment_type VARCHAR(50) NOT NULL CHECK (payment_type IN ('wechat', 'alipay', 'stripe')),
			amount NUMERIC(10, 2) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			CONSTRAINT payment_amount_positive CHECK (amount > 0)
		);
		
		
		CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
		CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
		CREATE INDEX IF NOT EXISTS idx_payments_dates ON payments(created_at, updated_at);
	`
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return err
	}
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return err
}
