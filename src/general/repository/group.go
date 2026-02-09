package repository

import (
	"context"
	"fmt"

	"saas/src/general/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GroupRepository handles group data access and permission management
type GroupRepository interface {
	// Group CRUD operations
	CreateGroup(ctx context.Context, conn *pgxpool.Conn, group *domain.Group) (*domain.Group, error)
	GetGroup(ctx context.Context, conn *pgxpool.Conn, id int) (*domain.Group, error)
	UpdateGroup(ctx context.Context, conn *pgxpool.Conn, group *domain.Group) (*domain.Group, error)
	DeleteGroup(ctx context.Context, conn *pgxpool.Conn, id int) error
	ListGroups(ctx context.Context, conn *pgxpool.Conn, page, pageSize int) ([]*domain.Group, error)
	SearchGroups(ctx context.Context, conn *pgxpool.Conn, query string, page, pageSize int) ([]*domain.Group, error)

	// Permission management for groups
	AddPermissionToGroup(ctx context.Context, conn *pgxpool.Conn, groupID, permissionID int) error
	RemovePermissionFromGroup(ctx context.Context, conn *pgxpool.Conn, groupID, permissionID int) error
	GetGroupPermissions(ctx context.Context, conn *pgxpool.Conn, groupID int) ([]*domain.Permission, error)
	SetGroupPermissions(ctx context.Context, conn *pgxpool.Conn, groupID int, permissionIDs []int) error

	// Permission CRUD operations
	CreatePermission(ctx context.Context, conn *pgxpool.Conn, permission *domain.Permission) (*domain.Permission, error)
	GetPermission(ctx context.Context, conn *pgxpool.Conn, id int) (*domain.Permission, error)
	GetPermissionByName(ctx context.Context, conn *pgxpool.Conn, name string) (*domain.Permission, error)
	UpdatePermission(ctx context.Context, conn *pgxpool.Conn, permission *domain.Permission) (*domain.Permission, error)
	DeletePermission(ctx context.Context, conn *pgxpool.Conn, id int) error
	ListPermissions(ctx context.Context, conn *pgxpool.Conn, page, pageSize int) ([]*domain.Permission, error)

	// Check operations
	GroupExists(ctx context.Context, conn *pgxpool.Conn, id int) (bool, error)
	PermissionExists(ctx context.Context, conn *pgxpool.Conn, id int) (bool, error)
	HasPermission(ctx context.Context, conn *pgxpool.Conn, groupID, permissionID int) (bool, error)
}

// groupRepository implements GroupRepository
type groupRepository struct{}

// NewGroupRepository creates a new GroupRepository instance
func NewGroupRepository() GroupRepository {
	return &groupRepository{}
}

// CreateGroup creates a new group
func (gr *groupRepository) CreateGroup(ctx context.Context, conn *pgxpool.Conn, group *domain.Group) (*domain.Group, error) {
	query := `
		INSERT INTO groups (name)
		VALUES ($1)
		RETURNING id, created_at, updated_at
	`

	err := conn.QueryRow(ctx, query, group.Name).Scan(
		&group.ID,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create group: %w", err)
	}

	// Set permissions if provided
	if len(group.PermissionsID) > 0 {
		err = gr.SetGroupPermissions(ctx, conn, group.ID, group.PermissionsID)
		if err != nil {
			return nil, fmt.Errorf("failed to set group permissions: %w", err)
		}
	}

	return group, nil
}

// GetGroup retrieves a group by ID with its permissions
func (gr *groupRepository) GetGroup(ctx context.Context, conn *pgxpool.Conn, id int) (*domain.Group, error) {
	// Get group basic info
	query := `
		SELECT id, name, created_at, updated_at
		FROM groups
		WHERE id = $1
	`

	group := &domain.Group{}
	err := conn.QueryRow(ctx, query, id).Scan(
		&group.ID,
		&group.Name,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("group not found")
		}
		return nil, fmt.Errorf("failed to get group: %w", err)
	}

	// Get group permissions
	permissions, err := gr.GetGroupPermissions(ctx, conn, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get group permissions: %w", err)
	}

	// Extract permission IDs
	group.PermissionsID = make([]int, len(permissions))
	for i, permission := range permissions {
		group.PermissionsID[i] = permission.ID
	}

	return group, nil
}

// UpdateGroup updates group information
func (gr *groupRepository) UpdateGroup(ctx context.Context, conn *pgxpool.Conn, group *domain.Group) (*domain.Group, error) {
	query := `
		UPDATE groups
		SET name = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING id, name, created_at, updated_at
	`

	updatedGroup := &domain.Group{}
	err := conn.QueryRow(ctx, query, group.ID, group.Name).Scan(
		&updatedGroup.ID,
		&updatedGroup.Name,
		&updatedGroup.CreatedAt,
		&updatedGroup.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("group not found")
		}
		return nil, fmt.Errorf("failed to update group: %w", err)
	}

	// Update permissions if provided
	if group.PermissionsID != nil {
		err = gr.SetGroupPermissions(ctx, conn, group.ID, group.PermissionsID)
		if err != nil {
			return nil, fmt.Errorf("failed to update group permissions: %w", err)
		}
	}

	// Get updated permissions
	permissions, err := gr.GetGroupPermissions(ctx, conn, updatedGroup.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get updated permissions: %w", err)
	}

	updatedGroup.PermissionsID = make([]int, len(permissions))
	for i, permission := range permissions {
		updatedGroup.PermissionsID[i] = permission.ID
	}

	return updatedGroup, nil
}

// DeleteGroup deletes a group
func (gr *groupRepository) DeleteGroup(ctx context.Context, conn *pgxpool.Conn, id int) error {
	query := `DELETE FROM groups WHERE id = $1`
	result, err := conn.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("group not found")
	}

	return nil
}

// ListGroups lists groups with pagination
func (gr *groupRepository) ListGroups(ctx context.Context, conn *pgxpool.Conn, page, pageSize int) ([]*domain.Group, error) {
	query := `
		SELECT g.id, g.name, g.created_at, g.updated_at,
		       COALESCE(ARRAY_AGG(gp.permission_id) FILTER (WHERE gp.permission_id IS NOT NULL), '{}') as permissions
		FROM groups g
		LEFT JOIN group_permissions gp ON g.id = gp.group_id
		GROUP BY g.id, g.name, g.created_at, g.updated_at
		ORDER BY g.id
		LIMIT $1 OFFSET $2
	`

	rows, err := conn.Query(ctx, query, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list groups: %w", err)
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		group := &domain.Group{}
		var permissions []int32

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.CreatedAt,
			&group.UpdatedAt,
			&permissions,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group row: %w", err)
		}

		// Convert []int32 to []int
		group.PermissionsID = make([]int, len(permissions))
		for i, pid := range permissions {
			group.PermissionsID[i] = int(pid)
		}

		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating groups: %w", err)
	}

	return groups, nil
}

// SearchGroups searches groups by name
func (gr *groupRepository) SearchGroups(ctx context.Context, conn *pgxpool.Conn, query string, page, pageSize int) ([]*domain.Group, error) {
	searchQuery := `
		SELECT g.id, g.name, g.created_at, g.updated_at,
		       COALESCE(ARRAY_AGG(gp.permission_id) FILTER (WHERE gp.permission_id IS NOT NULL), '{}') as permissions
		FROM groups g
		LEFT JOIN group_permissions gp ON g.id = gp.group_id
		WHERE g.name ILIKE $1
		GROUP BY g.id, g.name, g.created_at, g.updated_at
		ORDER BY g.id
		LIMIT $2 OFFSET $3
	`

	rows, err := conn.Query(ctx, searchQuery, fmt.Sprintf("%%%s%%", query), pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to search groups: %w", err)
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		group := &domain.Group{}
		var permissions []int32

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.CreatedAt,
			&group.UpdatedAt,
			&permissions,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group row: %w", err)
		}

		// Convert []int32 to []int
		group.PermissionsID = make([]int, len(permissions))
		for i, pid := range permissions {
			group.PermissionsID[i] = int(pid)
		}

		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating groups: %w", err)
	}

	return groups, nil
}

// AddPermissionToGroup adds a permission to a group
func (gr *groupRepository) AddPermissionToGroup(ctx context.Context, conn *pgxpool.Conn, groupID, permissionID int) error {
	// Check if group exists
	groupExists, err := gr.GroupExists(ctx, conn, groupID)
	if err != nil {
		return fmt.Errorf("failed to check group existence: %w", err)
	}
	if !groupExists {
		return fmt.Errorf("group not found")
	}

	// Check if permission exists
	permissionExists, err := gr.PermissionExists(ctx, conn, permissionID)
	if err != nil {
		return fmt.Errorf("failed to check permission existence: %w", err)
	}
	if !permissionExists {
		return fmt.Errorf("permission not found")
	}

	// Check if already has permission
	hasPermission, err := gr.HasPermission(ctx, conn, groupID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to check permission: %w", err)
	}
	if hasPermission {
		return fmt.Errorf("group already has this permission")
	}

	query := `
		INSERT INTO group_permissions (group_id, permission_id)
		VALUES ($1, $2)
		ON CONFLICT (group_id, permission_id) DO NOTHING
	`

	_, err = conn.Exec(ctx, query, groupID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to add permission to group: %w", err)
	}

	return nil
}

// RemovePermissionFromGroup removes a permission from a group
func (gr *groupRepository) RemovePermissionFromGroup(ctx context.Context, conn *pgxpool.Conn, groupID, permissionID int) error {
	query := `
		DELETE FROM group_permissions
		WHERE group_id = $1 AND permission_id = $2
	`

	result, err := conn.Exec(ctx, query, groupID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from group: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("permission not found in group")
	}

	return nil
}

// GetGroupPermissions gets all permissions for a group
func (gr *groupRepository) GetGroupPermissions(ctx context.Context, conn *pgxpool.Conn, groupID int) ([]*domain.Permission, error) {
	query := `
		SELECT p.id, p.name, p.created_at
		FROM permissions p
		INNER JOIN group_permissions gp ON p.id = gp.permission_id
		WHERE gp.group_id = $1
		ORDER BY p.id
	`

	rows, err := conn.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*domain.Permission
	for rows.Next() {
		permission := &domain.Permission{}
		err := rows.Scan(
			&permission.ID,
			&permission.Name,
			&permission.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission row: %w", err)
		}
		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}

// SetGroupPermissions sets all permissions for a group (replaces existing)
func (gr *groupRepository) SetGroupPermissions(ctx context.Context, conn *pgxpool.Conn, groupID int, permissionIDs []int) error {
	// Begin transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Clear existing permissions
	clearQuery := `DELETE FROM group_permissions WHERE group_id = $1`
	_, err = tx.Exec(ctx, clearQuery, groupID)
	if err != nil {
		return fmt.Errorf("failed to clear group permissions: %w", err)
	}

	// Add new permissions
	if len(permissionIDs) > 0 {
		// Build bulk insert query
		insertQuery := `INSERT INTO group_permissions (group_id, permission_id) VALUES `
		params := []interface{}{}
		paramCounter := 1

		for i, permissionID := range permissionIDs {
			if i > 0 {
				insertQuery += ", "
			}
			insertQuery += fmt.Sprintf("($%d, $%d)", paramCounter, paramCounter+1)
			params = append(params, groupID, permissionID)
			paramCounter += 2
		}

		_, err = tx.Exec(ctx, insertQuery, params...)
		if err != nil {
			return fmt.Errorf("failed to set group permissions: %w", err)
		}
	}

	return tx.Commit(ctx)
}

// CreatePermission creates a new permission
func (gr *groupRepository) CreatePermission(ctx context.Context, conn *pgxpool.Conn, permission *domain.Permission) (*domain.Permission, error) {
	query := `
		INSERT INTO permissions (name)
		VALUES ($1)
		RETURNING id, created_at
	`

	err := conn.QueryRow(ctx, query, permission.Name).Scan(
		&permission.ID,
		&permission.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	return permission, nil
}

// GetPermission retrieves a permission by ID
func (gr *groupRepository) GetPermission(ctx context.Context, conn *pgxpool.Conn, id int) (*domain.Permission, error) {
	query := `
		SELECT id, name, created_at
		FROM permissions
		WHERE id = $1
	`

	permission := &domain.Permission{}
	err := conn.QueryRow(ctx, query, id).Scan(
		&permission.ID,
		&permission.Name,
		&permission.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("permission not found")
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permission, nil
}

// GetPermissionByName retrieves a permission by name
func (gr *groupRepository) GetPermissionByName(ctx context.Context, conn *pgxpool.Conn, name string) (*domain.Permission, error) {
	query := `
		SELECT id, name, created_at
		FROM permissions
		WHERE name = $1
	`

	permission := &domain.Permission{}
	err := conn.QueryRow(ctx, query, name).Scan(
		&permission.ID,
		&permission.Name,
		&permission.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("permission not found")
		}
		return nil, fmt.Errorf("failed to get permission: %w", err)
	}

	return permission, nil
}

// UpdatePermission updates permission information
func (gr *groupRepository) UpdatePermission(ctx context.Context, conn *pgxpool.Conn, permission *domain.Permission) (*domain.Permission, error) {
	query := `
		UPDATE permissions
		SET name = $2
		WHERE id = $1
		RETURNING id, name, created_at
	`

	updatedPermission := &domain.Permission{}
	err := conn.QueryRow(ctx, query, permission.ID, permission.Name).Scan(
		&updatedPermission.ID,
		&updatedPermission.Name,
		&updatedPermission.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("permission not found")
		}
		return nil, fmt.Errorf("failed to update permission: %w", err)
	}

	return updatedPermission, nil
}

// DeletePermission deletes a permission
func (gr *groupRepository) DeletePermission(ctx context.Context, conn *pgxpool.Conn, id int) error {
	query := `DELETE FROM permissions WHERE id = $1`
	result, err := conn.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("permission not found")
	}

	return nil
}

// ListPermissions lists permissions with pagination
func (gr *groupRepository) ListPermissions(ctx context.Context, conn *pgxpool.Conn, page, pageSize int) ([]*domain.Permission, error) {
	query := `
		SELECT id, name, created_at
		FROM permissions
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := conn.Query(ctx, query, pageSize, (page-1)*pageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*domain.Permission
	for rows.Next() {
		permission := &domain.Permission{}
		err := rows.Scan(
			&permission.ID,
			&permission.Name,
			&permission.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission row: %w", err)
		}
		permissions = append(permissions, permission)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating permissions: %w", err)
	}

	return permissions, nil
}

// GroupExists checks if a group exists
func (gr *groupRepository) GroupExists(ctx context.Context, conn *pgxpool.Conn, id int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM groups WHERE id = $1)`
	var exists bool
	err := conn.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check group existence: %w", err)
	}
	return exists, nil
}

// PermissionExists checks if a permission exists
func (gr *groupRepository) PermissionExists(ctx context.Context, conn *pgxpool.Conn, id int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM permissions WHERE id = $1)`
	var exists bool
	err := conn.QueryRow(ctx, query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check permission existence: %w", err)
	}
	return exists, nil
}

// HasPermission checks if a group has a specific permission
func (gr *groupRepository) HasPermission(ctx context.Context, conn *pgxpool.Conn, groupID, permissionID int) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM group_permissions WHERE group_id = $1 AND permission_id = $2)`
	var exists bool
	err := conn.QueryRow(ctx, query, groupID, permissionID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check group permission: %w", err)
	}
	return exists, nil
}
