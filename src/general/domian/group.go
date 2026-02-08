package general

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Permission 权限。不可修改，系统自己管理
type Permission struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}


type group struct {
	ID        int   `json:"id" `
	Name      string `json:"name"`
	PermissionsID []int `json:"permissions_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}


func CreateGroupPermissionTable(ctx context.Context, conn *pgxpool.Conn) error {
	query := `
		CREATE TABLE IF NOT EXISTS groups (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS permissions (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) UNIQUE NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		
		CREATE TABLE IF NOT EXISTS group_permissions (
			group_id INT NOT NULL,
			permission_id INT NOT NULL,
			PRIMARY KEY (group_id, permission_id),
			FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
			FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
		);
		
		CREATE INDEX IF NOT EXISTS idx_group_permissions_group_id ON group_permissions(group_id);
		CREATE INDEX IF NOT EXISTS idx_group_permissions_permission_id ON group_permissions(permission_id);
		
		DROP TRIGGER IF EXISTS update_group ON groups;
		
		CREATE TRIGGER update_group
		AFTER UPDATE ON groups
		FOR EACH ROW
		EXECUTE FUNCTION update_column();
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

	return tx.Commit(ctx)
}
