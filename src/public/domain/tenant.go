package domain

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantModel 租户模型，用于存储租户信息
type TenantModel struct {
	ID        int `json:"id"`
	Name      string `json:"name"`  // 租户名称，同时表示租户的子域名
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func CreateTenantTable(ctx context.Context, conn *pgxpool.Conn) error {
	// 创建租户表
	query := `
		CREATE TABLE IF NOT EXISTS tenants (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE INDEX IF NOT EXISTS idx_tenants_name ON tenants(name);
	
		DROP TRIGGER IF EXISTS update_tenants ON tenants;
		
		CREATE TRIGGER update_tenants
		    BEFORE UPDATE ON tenants
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

	return tx.Commit(ctx)
}
