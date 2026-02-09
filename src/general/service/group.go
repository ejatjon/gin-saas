package service

import (
	"context"
	"fmt"
	"log"

	"saas/src/bootstrap"
	"saas/src/general/domain"
	"saas/src/general/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

// GroupService handles group and permission management operations
type GroupService interface {
	// Group management
	CreateGroup(ctx context.Context, tenantName string, name string) (*domain.Group, error)
	GetGroup(ctx context.Context, tenantName string, id int) (*domain.Group, error)
	UpdateGroup(ctx context.Context, tenantName string, id int, name string) (*domain.Group, error)
	DeleteGroup(ctx context.Context, tenantName string, id int) (*domain.Group, error)
	ListGroups(ctx context.Context, tenantName string, page int, pageSize int) ([]*domain.Group, error)
	SearchGroups(ctx context.Context, tenantName string, query string, page int, pageSize int) ([]*domain.Group, error)

	// Permission management
	AddPermissionToGroup(ctx context.Context, tenantName string, groupId int, permissionId int) error
	RemovePermissionFromGroup(ctx context.Context, tenantName string, groupId int, permissionId int) error
	GetGroupPermissions(ctx context.Context, tenantName string, groupId int) ([]*domain.Permission, error)
	SetGroupPermissions(ctx context.Context, tenantName string, groupId int, permissionIds []int) error

	// Permission CRUD
	CreatePermission(ctx context.Context, tenantName string, name string) (*domain.Permission, error)
	GetPermission(ctx context.Context, tenantName string, id int) (*domain.Permission, error)
	GetPermissionByName(ctx context.Context, tenantName string, name string) (*domain.Permission, error)
	UpdatePermission(ctx context.Context, tenantName string, id int, name string) (*domain.Permission, error)
	DeletePermission(ctx context.Context, tenantName string, id int) error
	ListPermissions(ctx context.Context, tenantName string, page int, pageSize int) ([]*domain.Permission, error)

	// Permission check for users
	UserHasPermission(ctx context.Context, tenantName string, userId int, permissionId int) (bool, error)
	UserHasPermissionByName(ctx context.Context, tenantName string, userId int, permissionName string) (bool, error)
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

// getConnForTenant 获取对应租户schema 的链接
func (s *groupService) getConnForTenant(ctx context.Context, tenantName string) (*pgxpool.Conn, error) {
	conn, err := s.dbManager.GetConnForSchema(ctx, tenantName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection for tenant %s: %w", tenantName, err)
	}
	return conn, nil
}

// CreateGroup 创建新的用户组
func (s *groupService) CreateGroup(ctx context.Context, tenantName string, name string) (*domain.Group, error) {
	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	

	createdGroup, err := s.groupRepo.CreateGroup(ctx, conn, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	return createdGroup, nil
}

// GetGroup retrieves a group by ID
func (s *groupService) GetGroup(ctx context.Context, tenantName string, id int) (*domain.Group, error) {

	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	group, err := s.groupRepo.GetGroup(ctx, conn, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	return group, nil
}

// UpdateGroup updates group information
func (s *groupService) UpdateGroup(ctx context.Context, tenantName string, id int, name string) (*domain.Group, error) {
	if name == "" {
		return nil, fmt.Errorf("group name is required")
	}

	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	group := &domain.Group{
		ID:   id,
		Name: name,
	}

	updatedGroup, err := s.groupRepo.UpdateGroup(ctx, conn, group)
	if err != nil {
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	return updatedGroup, nil
}

// DeleteGroup deletes a group
func (s *groupService) DeleteGroup(ctx context.Context, tenantName string, id int) (*domain.Group, error) {

	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// Get group first to return it
	group, err := s.groupRepo.GetGroup(ctx, conn, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Delete group
	err = s.groupRepo.DeleteGroup(ctx, conn, id)
	if err != nil {
		return nil, fmt.Errorf("failed to delete group: %w", err)
	}

	return group, nil
}

// ListGroups lists groups with pagination
func (s *groupService) ListGroups(ctx context.Context, tenantName string, page int, pageSize int) ([]*domain.Group, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	conn, err := s.getConnForTenant(ctx, tenantName)
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
func (s *groupService) SearchGroups(ctx context.Context, tenantName string, query string, page int, pageSize int) ([]*domain.Group, error) {
	if query == "" {
		return s.ListGroups(ctx, tenantName, page, pageSize)
	}
	
	conn, err := s.getConnForTenant(ctx, tenantName)
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
func (s *groupService) AddPermissionToGroup(ctx context.Context, tenantName string, groupId int, permissionId int) error {

	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return err
	}
	defer conn.Release()

	err = s.groupRepo.AddPermissionToGroup(ctx, conn, groupId, permissionId)
	if err != nil {
		return fmt.Errorf("failed to add permission to group: %w", err)
	}

	return nil
}

// RemovePermissionFromGroup removes a permission from a group
func (s *groupService) RemovePermissionFromGroup(ctx context.Context, tenantName string, groupId int, permissionId int) error {
	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return err
	}
	defer conn.Release()

	group, err := s.groupRepo.GetGroup(ctx, conn, groupId)
	if err != nil {
		return fmt.Errorf("failed to get group details: %w", err)
	}

	if group.Name == "admin" {
		permission, err := s.groupRepo.GetPermission(ctx, conn, permissionId)
		if err != nil {
			return fmt.Errorf("failed to get permission details: %w", err)
		}

		isSystemPermission := false
		for _, perm := range domain.DefaultPermissions {
			if permission.Name == perm {
				isSystemPermission = true
				break
			}
		}
		
		if isSystemPermission {
			return fmt.Errorf("cannot remove system permission '%s' from admin group", permission.Name)
		}
	}

	err = s.groupRepo.RemovePermissionFromGroup(ctx, conn, groupId, permissionId)
	if err != nil {
		return fmt.Errorf("failed to remove permission from group: %w", err)
	}

	return nil
}

// GetGroupPermissions gets all permissions for a group
func (s *groupService) GetGroupPermissions(ctx context.Context, tenantName string, groupId int) ([]*domain.Permission, error) {

	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	permissions, err := s.groupRepo.GetGroupPermissions(ctx, conn, groupId)
	if err != nil {
		return nil, fmt.Errorf("failed to get group permissions: %w", err)
	}

	return permissions, nil
}

// SetGroupPermissions sets all permissions for a group (replaces existing)
func (s *groupService) SetGroupPermissions(ctx context.Context, tenantName string, groupId int, permissionIds []int) error {
	
	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return err
	}
	defer conn.Release()

	// Check if this is the admin group
	group, err := s.groupRepo.GetGroup(ctx, conn, groupId)
	if err != nil {
		return fmt.Errorf("failed to get group details: %w", err)
	}

	// If it's the admin group, ensure all system permissions are included
	if group.Name == "admin" {
		// Get all system permission names
		systemPermissions := domain.DefaultPermissions

		// Create a map of existing permission IDs for quick lookup
		existingPermIds := make(map[int]bool)
		for _, pid := range permissionIds {
			existingPermIds[pid] = true
		}

		// For each system permission, ensure it's in the list
		for _, sysPerm := range systemPermissions {
			// Get permission by name
			perm, err := s.groupRepo.GetPermissionByName(ctx, conn, sysPerm)
			if err != nil {
				log.Printf("Warning: System permission '%s' not found for admin group in tenant %s",
					sysPerm, tenantName)
				continue
			}

			// Add to list if not already present
			if !existingPermIds[perm.ID] {
				permissionIds = append(permissionIds, perm.ID)
				existingPermIds[perm.ID] = true
				log.Printf("Added system permission '%s' (ID: %d) to admin group in tenant %s",
					perm.Name, perm.ID, tenantName)
			}
		}
	}

	err = s.groupRepo.SetGroupPermissions(ctx, conn, groupId, permissionIds)
	if err != nil {
		return fmt.Errorf("failed to set group permissions: %w", err)
	}

	return nil
}

// CreatePermission creates a new permission
func (s *groupService) CreatePermission(ctx context.Context, tenantName string, name string) (*domain.Permission, error) {
	if name == "" {
		return nil, fmt.Errorf("permission name is required")
	}

	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	createdPermission, err := s.groupRepo.CreatePermission(ctx, conn, name)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}
	
	return createdPermission, nil
}

// GetPermission retrieves a permission by ID
func (s *groupService) GetPermission(ctx context.Context, tenantName string, id int) (*domain.Permission, error) {
	
	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	permission, err := s.groupRepo.GetPermission(ctx, conn, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permission, nil
}

// GetPermissionByName retrieves a permission by name
func (s *groupService) GetPermissionByName(ctx context.Context, tenantName string, name string) (*domain.Permission, error) {
	if name == "" {
		return nil, fmt.Errorf("permission name is required")
	}

	conn, err := s.getConnForTenant(ctx, tenantName)
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

func (s *groupService) UpdatePermission(ctx context.Context, tenantName string, id int, name string) (*domain.Permission,error) {
	if name == "" {
		return nil, fmt.Errorf("permission name is required")
	}
	
	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return nil, err
	}
	defer conn.Release()
	
	permission, err := s.groupRepo.UpdatePermission(ctx, conn, &domain.Permission{
		ID: id,
		Name: name,
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}
	return permission,nil
}

// DeletePermission deletes a permission
func (s *groupService) DeletePermission(ctx context.Context, tenantName string, id int) error {
	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return err
	}
	defer conn.Release()

	// Get permission details to check if it's a system permission
	permission, err := s.groupRepo.GetPermission(ctx, conn, id)
	if err != nil {
		return fmt.Errorf("failed to get permission details: %w", err)
	}

	// Check if this is a system permission
	isSystemPermission := false
	for _, sysPerm := range domain.DefaultPermissions {
		if permission.Name == sysPerm {
			isSystemPermission = true
			break
		}
	}

	if isSystemPermission {
		return fmt.Errorf("cannot delete system permission: '%s'", permission.Name)
	}

	err = s.groupRepo.DeletePermission(ctx, conn, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	return nil
}

// ListPermissions lists permissions with pagination
func (s *groupService) ListPermissions(ctx context.Context, tenantName string, page int, pageSize int) ([]*domain.Permission, error) {
	conn, err := s.getConnForTenant(ctx, tenantName)
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

// UserHasPermission checks if a user has a specific permission through their groups
func (s *groupService) UserHasPermission(ctx context.Context, tenantName string, userId int, permissionId int) (bool, error) {

	
	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	hasPermission, err := s.groupRepo.UserHasPermission(ctx, conn, userId, permissionId)
	if err != nil {
		return false, fmt.Errorf("failed to check user permission: %w", err)
	}

	return hasPermission, nil
}

// UserHasPermissionByName checks if a user has a specific permission by permission name through their groups
func (s *groupService) UserHasPermissionByName(ctx context.Context, tenantName string, userId int, permissionName string) (bool, error) {
	if permissionName == "" {
		return false, fmt.Errorf("permission name is required")
	}

	conn, err := s.getConnForTenant(ctx, tenantName)
	if err != nil {
		return false, err
	}
	defer conn.Release()

	hasPermission, err := s.groupRepo.UserHasPermissionByName(ctx, conn, userId, permissionName)
	if err != nil {
		return false, fmt.Errorf("failed to check user permission by name: %w", err)
	}

	return hasPermission, nil
}
