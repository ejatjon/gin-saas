package bootstrap

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBManager 数据库管理器，准确的说数据库链接管理器，用于管理数据库连接池。
type DBManager struct {
	pool              *pgxpool.Pool                                                  // 数据库连接池
	schemaCreatorsMap map[string]func(ctx context.Context, conn *pgxpool.Conn) error // schema创建器映射 (跟pgsql的schema不是一个东西，这个是表示一堆表结构。具体服务注册这个方法，当租户启用某个服务时执行这个方法创建对应数据表)
}

func NewDBManager(pool *pgxpool.Pool) *DBManager {
	return &DBManager{
		pool:              pool,
		schemaCreatorsMap: make(map[string]func(ctx context.Context, conn *pgxpool.Conn) error),
	}
}

// 获取指定schema的连接 ，需要手动 conn.Release()
func (m *DBManager) GetConnForSchema(ctx context.Context, schema string) (*pgxpool.Conn, error) {
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	// INFO:需要提前创建好schema
	_, err = conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", pgx.Identifier{schema}.Sanitize()))
	if err != nil {
		conn.Release()
		return nil, err
	}

	return conn, nil
}

// 需要手动 conn.Release()
func (m *DBManager) CreateSchema(ctx context.Context, schema string) (*pgxpool.Conn, error) {
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(ctx, fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, pgx.Identifier{schema}.Sanitize()))
	if err != nil {
		conn.Release()
		return nil, fmt.Errorf("failed to create schema %s: %w", schema, err)
	}
	_, err = conn.Exec(ctx, fmt.Sprintf("SET search_path TO %s", pgx.Identifier{schema}.Sanitize()))
	if err != nil {
		conn.Release()
		return nil, err
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		conn.Release()
		return nil, err
	}
	defer tx.Rollback(ctx)
	// 给每个schema 添加必要的通用函数
	_, err = tx.Exec(ctx, `
    	CREATE OR REPLACE FUNCTION update_column()
		RETURNS TRIGGER AS $$
		BEGIN
      		NEW.updated_at = CURRENT_TIMESTAMP;
       		RETURN NEW;
        END;
        $$ language 'plpgsql';
    `)
	if err != nil {
		return nil, fmt.Errorf("failed to create table %s: %w", schema, err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return conn, nil
}

// ListSchemas 列出所有已存在的schema
func (m *DBManager) ListSchemas(ctx context.Context) ([]string, error) {
	conn, err := m.pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// Query all schemas, excluding system schemas
	query := `
		SELECT schema_name
		FROM information_schema.schemata
		WHERE schema_name NOT IN ('information_schema', 'pg_catalog', 'pg_toast')
		AND schema_name NOT LIKE 'pg_%'
		ORDER BY schema_name
	`

	rows, err := conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list schemas: %w", err)
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schemaName string
		if err := rows.Scan(&schemaName); err != nil {
			return nil, fmt.Errorf("failed to scan schema name: %w", err)
		}
		schemas = append(schemas, schemaName)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating schema rows: %w", err)
	}

	return schemas, nil
}

// 注册schema创建器
func (m *DBManager) RegisterSchemaCreator(serviceName string, creator func(ctx context.Context, conn *pgxpool.Conn) error) {
	m.schemaCreatorsMap[serviceName] = creator
}

func (m *DBManager) GetSchemaCreator(serviceName string) func(ctx context.Context, conn *pgxpool.Conn) error {
	return m.schemaCreatorsMap[serviceName]
}
