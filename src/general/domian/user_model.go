package general

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// User 通用的用户模型，不管是public 还是 tenant 的用户都使用这个用户模型
type UserModel struct {
	ID        int        `json:"id"`
	Username  string     `json:"username"`
	Email     string     `json:"email"`
	Password  string     `json:"-"` // 密码哈希，用于存储用户密码的哈希值，不直接存储明文密码。json 序列化时需要被忽略
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	Status    string     `json:"status"` // 用户状态，我们将永远不会删除用户数据，但可以更改状态。
	Groups    []int      `json:"groups"`   // 用户组，我们将通过用户组进行访问控制
	LastLogin *time.Time `json:"last_login,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func CreateUserTable(ctx context.Context, conn *pgxpool.Conn) error {
	
	tx, err := conn.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			first_name VARCHAR(255) NOT NULL,
			last_name VARCHAR(255) NOT NULL,
			status VARCHAR(255) NOT NULL DEFAULT 'active',
			groups INTEGER[] NOT NULL DEFAULT '{}',
			last_login TIMESTAMP WITH TIME ZONE,
			created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
		
		CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
		CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
		CREATE INDEX IF NOT EXISTS idx_users_groups ON users(groups);
		
		DROP TRIGGER IF EXISTS update_users ON users;
		
		CREATE TRIGGER update_users
		AFTER UPDATE ON users
		FOR EACH ROW
		EXECUTE FUNCTION update_column();
		`
	
	_, err = tx.Exec(ctx, query)
	if err != nil {
		return err
	}
	
	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	
	return nil
}


// Pagination 分页数据结构，表示当前页面的数据
type Pagination struct {
	Page     int `json:"page" form:"page" query:"page"`
	PageSize int `json:"page_size" form:"page_size" query:"page_size"`
}

// Offset 数据偏移，表示当前页面要跳过多少个数据
func (p *Pagination) Offset() int {
	if p.Page <= 0 {
		p.Page = 1
	}
	if p.PageSize <= 0 {
		p.PageSize = 20 // 默认页面大小
	}
	return (p.Page - 1) * p.PageSize
}

// Limit 返回当前页面的大小
func (p *Pagination) Limit() int {
	if p.PageSize <= 0 {
		p.PageSize = 20
	}
	return p.PageSize
}

// UserFilter 用户过滤数据结构，用于过滤用户列表
type UserFilter struct {
	Username *string `json:"username,omitempty" form:"username" query:"username"`
	Email    *string `json:"email,omitempty" form:"email" query:"email"`
	Status   *string `json:"status,omitempty" form:"status" query:"status"`
	Role     *string `json:"role,omitempty" form:"role" query:"role"`
}

// UserListResponse 用户列表响应数据结构，用于返回用户列表
type UserListResponse struct {
	Users    []UserModel `json:"users"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}
