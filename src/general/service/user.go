package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"saas/src/bootstrap"
	"saas/src/general/domain"
	"saas/src/general/repository"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	Login(ctx context.Context, tenantId string, email string, password string) (*domain.UserModel, error)
	Register(ctx context.Context, tenantId string, username string, email string, password string, firstName string, lastName string) (*domain.UserModel, error)
	Profile(ctx context.Context, tenantId string, id string) (*domain.UserModel, error)
	UpdateProfile(ctx context.Context, tenantId string, id string, firstName string, lastName string) (*domain.UserModel, error)
	ChangeUsername(ctx context.Context, tenantId string, id string, firstName string, lastName string) (*domain.UserModel, error)
	ChangeEmail(ctx context.Context, tenantId string, id string, oldEmail string, newEmail string) (*domain.UserModel, error)
	ChangePassword(ctx context.Context, tenantId string, id string, oldPassword string, newPassword string) (*domain.UserModel, error)

	// Manager
	GetUser(ctx context.Context, tenantId string, id string) (*domain.UserModel, error)
	CreateUser(ctx context.Context, tenantId string, username string, email string, password string, firstName string, lastName string) (*domain.UserModel, error)
	UpdateUser(ctx context.Context, tenantId string, id string, username string, email string, firstName string, lastName string) (*domain.UserModel, error)
	DeleteUser(ctx context.Context, tenantId string, id string) (*domain.UserModel, error)

	GetUsersList(ctx context.Context, tenantId string, page int, pageSize int) ([]*domain.UserModel, error)
	SearchUsers(ctx context.Context, tenantId string, query string, page int, pageSize int) ([]*domain.UserModel, error)
}

type userService struct {
	dbManager *bootstrap.DBManager
	userRepo  repository.UserRepository
}

func NewUserService(dbManager *bootstrap.DBManager, userRepo repository.UserRepository) UserService {
	return &userService{
		dbManager: dbManager,
		userRepo:  userRepo,
	}
}

// validatePassword validates password length and allows only alphanumeric characters
func validatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("password must be at least 6 characters long")
	}
	// Only allow letters (a-z, A-Z) and numbers (0-9)
	matched, err := regexp.MatchString("^[a-zA-Z0-9]+$", password)
	if err != nil {
		return fmt.Errorf("password validation error: %w", err)
	}
	if !matched {
		return fmt.Errorf("password can only contain letters and numbers")
	}
	return nil
}

// hashPassword hashes password using bcrypt
func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}
	return string(hashed), nil
}

// verifyPassword compares plain password with hashed password
func verifyPassword(plainPassword, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}

// getConnForTenant gets database connection for tenant schema
func (s *userService) getConnForTenant(ctx context.Context, tenantId string) (*pgxpool.Conn, error) {
	conn, err := s.dbManager.GetConnForSchema(ctx, tenantId)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection for tenant %s: %w", tenantId, err)
	}
	return conn, nil
}

// getUserByEmail gets user by email with exact match
func (s *userService) getUserByEmail(ctx context.Context, conn *pgxpool.Conn, email string) (*domain.UserModel, error) {
	// Search for user by email using SearchUsers (exact match)
	users, err := s.userRepo.SearchUsers(ctx, conn, email, 1, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to search for user: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("user not found")
	}

	// Should only get one result since email should be unique
	user := users[0]

	// Additional check to ensure exact email match
	if user.Email != email {
		return nil, fmt.Errorf("email mismatch")
	}

	return user, nil
}

func (s *userService) Login(ctx context.Context, tenantId string, email string, password string) (*domain.UserModel, error) {
	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// Get user by email
	user, err := s.getUserByEmail(ctx, conn, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Verify password
	if !verifyPassword(password, user.Password) {
		return nil, fmt.Errorf("invalid password")
	}

	return user, nil
}

func (s *userService) Register(ctx context.Context, tenantId string, username string, email string, password string, firstName string, lastName string) (*domain.UserModel, error) {
	// Validate password
	if err := validatePassword(password); err != nil {
		return nil, fmt.Errorf("invalid password: %w", err)
	}

	// Hash password
	hashedPassword, err := hashPassword(password)
	if err != nil {
		return nil, err
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// Create user model
	now := time.Now()
	user := &domain.UserModel{
		Username:  username,
		Email:     email,
		Password:  hashedPassword,
		FirstName: firstName,
		LastName:  lastName,
		Status:    "active",
		Groups:    []string{}, // default empty groups
		CreatedAt: now,
		UpdatedAt: now,
	}

	createdUser, err := s.userRepo.CreateUser(ctx, conn, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

func (s *userService) Profile(ctx context.Context, tenantId string, id string) (*domain.UserModel, error) {
	userID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	user, err := s.userRepo.GetUser(ctx, conn, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (s *userService) UpdateProfile(ctx context.Context, tenantId string, id string, firstName string, lastName string) (*domain.UserModel, error) {
	userID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// Get existing user
	user, err := s.userRepo.GetUser(ctx, conn, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields
	user.FirstName = firstName
	user.LastName = lastName
	user.UpdatedAt = time.Now()

	updatedUser, err := s.userRepo.UpdateUser(ctx, conn, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUser, nil
}

func (s *userService) ChangeUsername(ctx context.Context, tenantId string, id string, firstName string, lastName string) (*domain.UserModel, error) {
	// This method seems misnamed - parameters are firstName and lastName, not username
	// Assuming it's meant to update username, but parameters are wrong
	// For now, treat as UpdateProfile
	return s.UpdateProfile(ctx, tenantId, id, firstName, lastName)
}

func (s *userService) ChangeEmail(ctx context.Context, tenantId string, id string, oldEmail string, newEmail string) (*domain.UserModel, error) {
	userID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// Get existing user
	user, err := s.userRepo.GetUser(ctx, conn, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify old email matches
	if user.Email != oldEmail {
		return nil, fmt.Errorf("old email does not match")
	}

	// Update email
	user.Email = newEmail
	user.UpdatedAt = time.Now()

	updatedUser, err := s.userRepo.UpdateUser(ctx, conn, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update email: %w", err)
	}

	return updatedUser, nil
}

func (s *userService) ChangePassword(ctx context.Context, tenantId string, id string, oldPassword string, newPassword string) (*domain.UserModel, error) {
	// Validate new password
	if err := validatePassword(newPassword); err != nil {
		return nil, fmt.Errorf("invalid new password: %w", err)
	}

	userID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// Get existing user
	user, err := s.userRepo.GetUser(ctx, conn, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Verify old password
	if !verifyPassword(oldPassword, user.Password) {
		return nil, fmt.Errorf("old password is incorrect")
	}

	// Hash new password
	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return nil, err
	}

	// Update password
	user.Password = hashedPassword
	user.UpdatedAt = time.Now()

	updatedUser, err := s.userRepo.UpdateUser(ctx, conn, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update password: %w", err)
	}

	return updatedUser, nil
}

func (s *userService) GetUser(ctx context.Context, tenantId string, id string) (*domain.UserModel, error) {
	return s.Profile(ctx, tenantId, id)
}

func (s *userService) CreateUser(ctx context.Context, tenantId string, username string, email string, password string, firstName string, lastName string) (*domain.UserModel, error) {
	// This is essentially the same as Register but might be for admin use
	return s.Register(ctx, tenantId, username, email, password, firstName, lastName)
}

func (s *userService) UpdateUser(ctx context.Context, tenantId string, id string, username string, email string, firstName string, lastName string) (*domain.UserModel, error) {
	userID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// Get existing user
	user, err := s.userRepo.GetUser(ctx, conn, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields
	user.Username = username
	user.Email = email
	user.FirstName = firstName
	user.LastName = lastName
	user.UpdatedAt = time.Now()

	updatedUser, err := s.userRepo.UpdateUser(ctx, conn, user)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return updatedUser, nil
}

func (s *userService) DeleteUser(ctx context.Context, tenantId string, id string) (*domain.UserModel, error) {
	userID, err := strconv.Atoi(id)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// Get user first to return it
	user, err := s.userRepo.GetUser(ctx, conn, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Delete user
	err = s.userRepo.DeleteUser(ctx, conn, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete user: %w", err)
	}

	return user, nil
}

func (s *userService) GetUsersList(ctx context.Context, tenantId string, page int, pageSize int) ([]*domain.UserModel, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	users, err := s.userRepo.ListUsers(ctx, conn, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	return users, nil
}

func (s *userService) SearchUsers(ctx context.Context, tenantId string, query string, page int, pageSize int) ([]*domain.UserModel, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	users, err := s.userRepo.SearchUsers(ctx, conn, query, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to search users: %w", err)
	}

	return users, nil
}
