package service

import (
	"context"
	"saas/src/general/domain"
)

type UserService interface {
	Login(ctx context.Context,tenantId string, email string, password string) (*domain.UserModel, error)
	Register(ctx context.Context,tenantId string, username string, email string, password string, firstName string, lastName string) (*domain.UserModel, error)
	Profile(ctx context.Context,tenantId string, id string) (*domain.UserModel, error)
	UpdateProfile(ctx context.Context,tenantId string, id string, firstName string, lastName string) (*domain.UserModel, error)
	ChangeUsername(ctx context.Context,tenantId string, id string, firstName string, lastName string) (*domain.UserModel, error)
	ChangeEmail(ctx context.Context,tenantId string, id string, oldEmail string, newEmail string) (*domain.UserModel, error)
	ChangePassword(ctx context.Context,tenantId string, id string, oldPassword string, newPassword string) (*domain.UserModel, error)
	
	// Manager 
	GetUser(ctx context.Context,tenantId string, id string) (*domain.UserModel, error)
	CreateUser(ctx context.Context,tenantId string,username string, email string, password string, firstName string, lastName string) (*domain.UserModel, error)
	UpdateUser(ctx context.Context,tenantId string, id string, username string, email string, firstName string, lastName string) (*domain.UserModel, error)
	DeleteUser(ctx context.Context,tenantId string, id string) (*domain.UserModel, error)
	
	GetUsersList(ctx context.Context,tenantId string,page int,pageSize int) ([]*domain.UserModel, error)
	SearchUsers(ctx context.Context,tenantId string, query string,page int,pageSize int) ([]*domain.UserModel, error)
}

type userService struct{}

func NewUserService() UserService {
	return &userService{}
}

func (s *userService) Login(ctx context.Context, tenantId string, email string, password string) (*domain.UserModel, error) {
	
	return nil, nil
}

func (s *userService) Register(ctx context.Context, tenantId string, username string, email string, password string, firstName string, lastName string) (*domain.UserModel, error) {
	return nil, nil
}

func (s *userService) Profile(ctx context.Context, tenantId string, id string) (*domain.UserModel, error) {
	return nil, nil
}

func (s *userService) UpdateProfile(ctx context.Context,tenantId string, id string, firstName string, lastName string) (*domain.UserModel, error) {
	return nil, nil
}

func (s *userService) ChangeUsername(ctx context.Context,tenantId string, id string, firstName string, lastName string) (*domain.UserModel, error) {
	return nil, nil
}

func (s *userService) ChangeEmail(ctx context.Context,tenantId string, id string, oldEmail string, newEmail string) (*domain.UserModel, error) {
	return nil, nil
}

func (s *userService) ChangePassword(ctx context.Context,tenantId string, id string, oldPassword string, newPassword string) (*domain.UserModel, error) {
	return nil, nil
}

func (s *userService) GetUser(ctx context.Context,tenantId string, id string) (*domain.UserModel, error) {
	return nil, nil
}

func (s *userService) CreateUser(ctx context.Context,tenantId string,username string, email string, password string, firstName string, lastName string) (*domain.UserModel, error) {
	return nil, nil
}

func (s *userService) UpdateUser(ctx context.Context,tenantId string, id string, username string, email string, firstName string, lastName string) (*domain.UserModel, error) {
	return nil, nil
}

func (s *userService) DeleteUser(ctx context.Context,tenantId string, id string) (*domain.UserModel, error) {
	return nil, nil
}

func (s *userService) GetUsersList(ctx context.Context,tenantId string, page int, pageSize int) ([]*domain.UserModel, error) {
	return nil, nil
}

func (s *userService) SearchUsers(ctx context.Context,tenantId string, query string,page int,pageSize int) ([]*domain.UserModel, error) {
	return nil, nil
}
