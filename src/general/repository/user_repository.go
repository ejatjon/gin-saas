package public

import (
	"context"
	"fmt"
	"strings"
	"time"

	"saas/src/domain/public"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)


// UserRepository 负责用户数据的访问
type UserRepository interface {
	// Create 创建新的用户
	Create(ctx context.Context, db *pgxpool.Conn, user *public.User) error

	// GetByID 通过用户id返回用户
	GetByID(ctx context.Context, db *pgxpool.Conn, id int) (*public.User, error)

	// GetByUsername 通过用户名返回用户
	GetByUsername(ctx context.Context, db *pgxpool.Conn, username string) (*public.User, error)

	// GetByEmail 通过用户邮箱返回用户
	GetByEmail(ctx context.Context, db *pgxpool.Conn, email string) (*public.User, error)

	// Update 更新用户信息
	Update(ctx context.Context, db *pgxpool.Conn, user *public.User) error

	// Delete 删除用户（不应使用这个方法，应该使用修改状态的方式现象删除用户）
	Delete(ctx context.Context, db *pgxpool.Conn, id int) error

	// List 返回符合条件的用户列表
	List(ctx context.Context, db *pgxpool.Conn, filter *public.UserFilter, pagination *public.Pagination) ([]*public.User, int, error)

	// Count 返回符合条件的用户总数
	Count(ctx context.Context, db *pgxpool.Conn, filter *public.UserFilter) (int, error)

	// Search 通过搜索词搜索用户
	Search(ctx context.Context, db *pgxpool.Conn, searchTerm string, pagination *public.Pagination) ([]*public.User, int, error)

	// UpdateLastLogin 更新用户最后登录时间
	UpdateLastLogin(ctx context.Context, db *pgxpool.Conn, userID int) error

	// UpdateStatus 更新用户状态
	UpdateStatus(ctx context.Context, db *pgxpool.Conn, userID int, status string) error
}


// userRepository implements UserRepository
type userRepository struct{}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository() UserRepository {
	return &userRepository{}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, db *pgxpool.Conn, user *public.User) error {
	query := `
		INSERT INTO users (
			tenant_id, username, email, password, first_name, last_name,
			status, role, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	err := db.QueryRow(ctx, query,
		user.TenantID,
		user.Username,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.Status,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	return err
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, db *pgxpool.Conn, tenantName string, id int) (*public.User, error) {
	query := `
		SELECT id, tenant_id, username, email, password, first_name, last_name,
		       status, role, last_login, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	user := &public.User{}
	err := db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.TenantID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Status,
		&user.Role,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return user, nil
}

// GetByUsername retrieves a user by username
func (r *userRepository) GetByUsername(ctx context.Context, db *pgxpool.Conn, tenantName string, username string) (*public.User, error) {
	query := `
		SELECT id, tenant_id, username, email, password, first_name, last_name,
		       status, role, last_login, created_at, updated_at, deleted_at
		FROM users
		WHERE username = $1 AND deleted_at IS NULL
	`

	user := &public.User{}
	err := db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.TenantID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Status,
		&user.Role,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, db *pgxpool.Conn, tenantName string, email string) (*public.User, error) {
	query := `
		SELECT id, tenant_id, username, email, password, first_name, last_name,
		       status, role, last_login, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := &public.User{}
	err := db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.TenantID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Status,
		&user.Role,
		&user.LastLogin,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	return user, nil
}

// Update updates an existing user
func (r *userRepository) Update(ctx context.Context, db *pgxpool.Conn, tenantName string, user *public.User) error {
	query := `
		UPDATE users
		SET username = $1, email = $2, password = $3, first_name = $4, last_name = $5,
		    status =
