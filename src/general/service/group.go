package service

import (
	"context"
	"fmt"
	"strconv"

	"saas/src/bootstrap"
	"saas/src/general/domain"
	"saas/src/general/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GroupService handles group and permission management operations
type GroupService interface {
	// Group management
	CreateGroup(ctx context.Context, tenantId string, name string) (*domain.Group, error)
	GetGroup(ctx context.Context, tenantId string, id string) (*domain.Group, error)
	UpdateGroup(ctx context.Context, tenantId string, id string, name string) (*domain.Group, error)
	DeleteGroup(ctx context.Context, tenantId string, id string) (*domain.Group, error)
	ListGroups(ctx context.Context, tenantId string, page int, pageSize int) ([]*domain.Group, error)
	SearchGroups(ctx context.Context, tenantId string, query string, page int, pageSize int) ([]*domain.Group, error)

	// Permission management
	AddPermissionToGroup(ctx context.Context, tenantId string, groupId string, permissionId string) error
	RemovePermissionFromGroup(ctx context.Context, tenantId string, groupId string, permissionId string) error
	GetGroupPermissions(ctx context.Context, tenantId string, groupId string) ([]*domain.Permission, error)
	SetGroupPermissions(ctx context.Context, tenantId string, groupId string, permissionIds []string) error

	// Permission CRUD
	CreatePermission(ctx context.Context, tenantId string, name string) (*domain.Permission, error)
	GetPermission(ctx context.Context, tenantId string, id string) (*domain.Permission, error)
	GetPermissionByName(ctx context.Context, tenantId string, name string) (*domain.Permission, error)
	UpdatePermission(ctx context.Context, tenantId string, id string, name string) (*domain.Permission, error)
	DeletePermission(ctx context.Context, tenantId string, id string) error
	ListPermissions(ctx context.Context, tenantId string, page int, pageSize int) ([]*domain.Permission, error)
}

type groupService struct {
	dbManager *bootstrap.DBManager
	groupRepo repository.GroupRepository
}

// NewGroupService creates a new GroupService instance
func NewGroupService(dbManager *bootstrap.DBManager, groupRepo repository.GroupRepository) GroupService {
	return &groupService{
		dbManager: dbManager,
		groupRepo: groupRepo,
	}
}

// getConnForTenant gets database connection for tenant schema
func (s *groupService) getConnForTenant(ctx context.Context, tenantId string) (*pgxpool.Conn, error) {
	conn, err := s.dbManager.GetConnForSchema(ctx, tenantId)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection for tenant %s: %w", tenantId, err)
	}
	return conn, nil
}

// parseStringToInt parses string to int with error handling
func parseStringToInt(id string, fieldName string) (int, error) {
	value, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", fieldName, err)
	}
	return value, nil
}

// parseStringSliceToIntSlice parses slice of strings to slice of ints
func parseStringSliceToIntSlice(stringIds []string, fieldName string) ([]int, error) {
	intIds := make([]int, len(stringIds))
	for i, str := range stringIds {
		val, err := strconv.Atoi(str)
		if err != nil {
			return nil, fmt.Errorf("invalid %s at index %d: %w", fieldName, i, err)
		}
		intIds[i] = val
	}
	return intIds, nil
}

// CreateGroup creates a new group
func (s *groupService) CreateGroup(ctx context.Context, tenantId string, name string) (*domain.Group, error) {
	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	group := &domain.Group{
		Name: name,
		// PermissionsID will be set by repository if needed
	}

	createdGroup, err := s.groupRepo.CreateGroup(ctx, conn, group)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	return createdGroup, nil
}

// GetGroup retrieves a group by ID
func (s *groupService) GetGroup(ctx context.Context, tenantId string, id string) (*domain.Group, error) {
	groupID, err := parseStringToInt(id, "group ID")
	if err != nil {
		return nil, err
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	group, err := s.groupRepo.GetGroup(ctx, conn, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return group, nil
}

// UpdateGroup updates group information
func (s *groupService) UpdateGroup(ctx context.Context, tenantId string, id string, name string) (*domain.Group, error) {
	groupID, err := parseStringToInt(id, "group ID")
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	group := &domain.Group{
		ID:   groupID,
		Name: name,
	}

	updatedGroup, err := s.groupRepo.UpdateGroup(ctx, conn, group)
	if err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	return updatedGroup, nil
}

// DeleteGroup deletes a group
func (s *groupService) DeleteGroup(ctx context.Context, tenantId string, id string) (*domain.Group, error) {
	groupID, err := parseStringToInt(id, "group ID")
	if err != nil {
		return nil, err
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// Get group first to return it
	group, err := s.groupRepo.GetGroup(ctx, conn, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Delete group
	err = s.groupRepo.DeleteGroup(ctx, conn, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete group: %w", err)
	}

	return group, nil
}

// ListGroups lists groups with pagination
func (s *groupService) ListGroups(ctx context.Context, tenantId string, page int, pageSize int) ([]*domain.Group, error) {
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

	groups, err := s.groupRepo.ListGroups(ctx, conn, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}

	return groups, nil
}

// SearchGroups searches groups by name
func (s *groupService) SearchGroups(ctx context.Context, tenantId string, query string, page int, pageSize int) ([]*domain.Group, error) {
	if query == "" {
		return s.ListGroups(ctx, tenantId, page, pageSize)
	}

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

	groups, err := s.groupRepo.SearchGroups(ctx, conn, query, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to search groups: %w", err)
	}

	return groups, nil
}

// AddPermissionToGroup adds a permission to a group
func (s *groupService) AddPermissionToGroup(ctx context.Context, tenantId string, groupId string, permissionId string) error {
	groupID, err := parseStringToInt(groupId, "group ID")
	if err != nil {
		return err
	}

	permissionID, err := parseStringToInt(permissionId, "permission ID")
	if err != nil {
		return err
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return err
	}
	defer conn.Release()

	err = s.groupRepo.AddPermissionToGroup(ctx, conn, groupID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to add permission to group: %w", err)
	}

	return nil
}

// RemovePermissionFromGroup removes a permission from a group
func (s *groupService) RemovePermissionFromGroup(ctx context.Context, tenantId string, groupId string, permissionId string) error {
	groupID, err := parseStringToInt(groupId, "group ID")
	if err != nil {
		return err
	}

	permissionID, err := parseStringToInt(permissionId, "permission ID")
	if err != nil {
		return err
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return err
	}
	defer conn.Release()

	err = s.groupRepo.RemovePermissionFromGroup(ctx, conn, groupID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from group: %w", err)
	}

	return nil
}

// GetGroupPermissions gets all permissions for a group
func (s *groupService) GetGroupPermissions(ctx context.Context, tenantId string, groupId string) ([]*domain.Permission, error) {
	groupID, err := parseStringToInt(groupId, "group ID")
	if err != nil {
		return nil, err
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	permissions, err := s.groupRepo.GetGroupPermissions(ctx, conn, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group permissions: %w", err)
	}

	return permissions, nil
}

// SetGroupPermissions sets all permissions for a group (replaces existing)
func (s *groupService) SetGroupPermissions(ctx context.Context, tenantId string, groupId string, permissionIds []string) error {
	groupID, err := parseStringToInt(groupId, "group ID")
	if err != nil {
		return err
	}

	permissionIDs, err := parseStringSliceToIntSlice(permissionIds, "permission ID")
	if err != nil {
		return err
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return err
	}
	defer conn.Release()

	err = s.groupRepo.SetGroupPermissions(ctx, conn, groupID, permissionIDs)
	if err != nil {
		return fmt.Errorf("failed to set group permissions: %w", err)
	}

	return nil
}

// CreatePermission creates a new permission
func (s *groupService) CreatePermission(ctx context.Context, tenantId string, name string) (*domain.Permission, error) {
	if name == "" {
		return nil, fmt.Errorf("permission name is required")
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	permission := &domain.Permission{
		Name: name,
	}

	createdPermission, err := s.groupRepo.CreatePermission(ctx, conn, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return createdPermission, nil
}

// GetPermission retrieves a permission by ID
func (s *groupService) GetPermission(ctx context.Context, tenantId string, id string) (*domain.Permission, error) {
	permissionID, err := parseStringToInt(id, "permission ID")
	if err != nil {
		return nil, err
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	permission, err := s.groupRepo.GetPermission(ctx, conn, permissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permission, nil
}

// GetPermissionByName retrieves a permission by name
func (s *groupService) GetPermissionByName(ctx context.Context, tenantId string, name string) (*domain.Permission, error) {
	if name == "" {
		return nil, fmt.Errorf("permission name is required")
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	permission, err := s.groupRepo.GetPermissionByName(ctx, conn, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission by name: %w", err)
	}

	return permission, nil
}

// UpdatePermission updates permission information
func (s *groupService) UpdatePermission(ctx context.Context, tenantId string, id string, name string) (*domain.Permission, error) {
	permissionID, err := parseStringToInt(id, "permission ID")
	if err != nil {
		return nil, err
	}

	if name == "" {
		return nil, fmt.Errorf("permission name is required")
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	permission := &domain.Permission{
		ID:   permissionID,
		Name: name,
	}

	updatedPermission, err := s.groupRepo.UpdatePermission(ctx, conn, permission)
	if err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return updatedPermission, nil
}

// DeletePermission deletes a permission
func (s *groupService) DeletePermission(ctx context.Context, tenantId string, id string) error {
	permissionID, err := parseStringToInt(id, "permission ID")
	if err != nil {
		return err
	}

	conn, err := s.getConnForTenant(ctx, tenantId)
	if err != nil {
		return err
	}
	defer conn.Release()

	err = s.groupRepo.DeletePermission(ctx, conn, permissionID)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

// ListPermissions lists permissions with pagination
func (s *groupService) ListPermissions(ctx context.Context, tenantId string, page int, pageSize int) ([]*domain.Permission, error) {
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

	permissions, err := s.groupRepo.ListPermissions(ctx, conn, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}

	return permissions, nil
}
