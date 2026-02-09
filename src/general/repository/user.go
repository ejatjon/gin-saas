package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"saas/src/general/domain"
)

// UserRepository handles user data access
type UserRepository interface {
	// Create creates a new user
	CreateUser(ctx context.Context, conn *pgxpool.Conn, user *domain.UserModel)(*domain.UserModel, error)
	// GetUserByID returns a user by ID
	GetUser(ctx context.Context, conn *pgxpool.Conn, id int)(*domain.UserModel, error)
	// Update updates user information
	UpdateUser(ctx context.Context, conn *pgxpool.Conn, user *domain.UserModel)(*domain.UserModel, error)
	// Delete deletes a user (should not be used, use status change instead)
	DeleteUser(ctx context.Context, conn *pgxpool.Conn, id int) error
	// List returns a list of users matching filters
	ListUsers(ctx context.Context, conn *pgxpool.Conn, page, pageSize int) ([]*domain.UserModel, error)
	// Search searches users by search term
	SearchUsers(ctx context.Context, conn *pgxpool.Conn, query string, page, pageSize int) ([]*domain.UserModel, error)

}

// userRepository implements UserRepository
type userRepository struct{}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository() UserRepository {
	return &userRepository{}
}

// CreateUser 创建用户
func (ur *userRepository) CreateUser(ctx context.Context, conn *pgxpool.Conn, user *domain.UserModel)(*domain.UserModel, error) {

	query := `
		INSERT INTO users (
			username, email, password, first_name, last_name, status, groups
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	
	err := conn.QueryRow(ctx, query,
		user.Username,
		user.Email,
		user.Password,
		user.FirstName,
		user.LastName,
		user.Status,
		user.Groups,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// GetUser 获取用户
func (ur *userRepository) GetUser(ctx context.Context, conn *pgxpool.Conn, id int)(*domain.UserModel, error) {

	query := `
		SELECT u.id, u.created_at, u.updated_at, u.username, u.email, u.password, u.first_name, u.last_name, u.status,ARRAY_AGG(g.name) AS groups
		FROM users u
		JOIN user_groups ug ON u.id = ug.user_id
		JOIN groups g ON ug.group_id = g.id
		GROUP BY u.id, u.created_at, u.updated_at, u.username, u.email, u.password, u.first_name, u.last_name, u.status
	`
	
	user := &domain.UserModel{}
	err := conn.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.FirstName,
		&user.LastName,
		&user.Status,
		&user.Groups,
	)
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

// UpdateUser 更新用户信息 不更新用户组信息（用户组信息不属于用户表）
func (ur *userRepository) UpdateUser(ctx context.Context, conn *pgxpool.Conn, user *domain.UserModel)(*domain.UserModel, error) {
	query := `
		UPDATE users
		SET username = $2, email = $3, password = $4, first_name = $5, last_name = $6, status = $7,
		WHERE id = $1
		RETURNING id, created_at, updated_at
	`
	
	err := conn.QueryRow(ctx, query, user.ID, user.Username, user.Email, user.Password, user.FirstName, user.LastName, user.Status).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (ur *userRepository) DeleteUser(ctx context.Context, conn *pgxpool.Conn, id int) error {
	query := `
		DELETE FROM users WHERE id = $1
	`
	
	_, err := conn.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	
	return nil
}

func (ur *userRepository) SearchUsers(ctx context.Context, conn *pgxpool.Conn, query string, page, pageSize int) ([]*domain.UserModel, error) {
	query = `
		SELECT u.id, u.created_at, u.updated_at, u.username, u.email, u.password, u.first_name, u.last_name, u.status,ARRAY_AGG(g.name) AS groups
		FROM users u
		JOIN user_groups ug ON u.id = ug.user_id
		JOIN groups g ON ug.group_id = g.id
		WHERE u.username LIKE $1 OR u.email LIKE $1 OR g.name LIKE $1 OR u.first_name LIKE $1 OR u.last_name LIKE $1
		GROUP BY u.id, u.created_at, u.updated_at, u.username, u.email, u.password, u.first_name, u.last_name, u.status
		LIMIT $2 OFFSET $3
	`
	
	rows, err := conn.Query(ctx, query, fmt.Sprintf("%%%s%%", query), pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	users, err := pgx.CollectRows(rows, pgx.RowTo[*domain.UserModel])
	if err != nil {
		return nil, err
	}
	
	return users, nil
}

func (ur *userRepository) ListUsers(ctx context.Context, conn *pgxpool.Conn, page, pageSize int) ([]*domain.UserModel, error) {
	query := `
		SELECT u.id, u.created_at, u.updated_at, u.username, u.email, u.password, u.first_name, u.last_name, u.status, ARRAY_AGG(g.name) AS groups
		FROM users as u
		JOIN user_groups ug ON u.id = ug.user_id
		JOIN groups g ON ug.group_id = g.id
		LIMIT $1 OFFSET $2
	`
	
	rows, err := conn.Query(ctx, query, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	users, err := pgx.CollectRows(rows, pgx.RowTo[*domain.UserModel])
	if err != nil {
		return nil, err
	}
	
	return users, nil
}
