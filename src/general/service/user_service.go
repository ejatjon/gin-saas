package public

    import(
        "context"
        "errors"
        "fmt"
        "time"

        "saas/src/domain/public"
        "saas/src/repository/public")

        type UserService struct {
    userRepo* repository.UserRepository
}

    func
    NewUserService(userRepo* repository.UserRepository)
    * UserService
{
    return &UserService
    {
    userRepo:
        userRepo,
    }
}

// Register creates a new user for a tenant
func(s* UserService) Register(ctx context.Context, tenantName string, user* public.User, password string)(*public.User, error)
{
    // Validate input
    if user
        .Username == ""
        {
            return nil, errors.New("username is required")
        }
    if user
        .Email == ""
        {
            return nil, errors.New("email is required")
        }
    if password
        == "" {
            return nil, errors.New("password is required")
        }

            // Hash password (simplified - in real app use proper hashing)
            // For now, just store as-is
            user.Password
            = password user.Status = "active" user.Role = "user" user.CreatedAt = time.Now() user.UpdatedAt = time.Now()

                                                                                                              // Create user
                                                                                                              if err : = s.userRepo.Create(ctx, tenantName, user);
    err != nil
    {
        return nil, fmt.Errorf("failed to create user: %w", err)
    }

    return user, nil
}

// Login authenticates a user
func(s* UserService) Login(ctx context.Context, tenantName, username, password string)(*public.User, error) {
    // First find user by username or email
    // For simplicity, we'll assume username is provided
    // In real app, you might need to check both

    // We need to get user - but repository may not have GetByUsername method
    // For now, let's assume we have a way to get user
    // We'll implement this later

    return nil, errors.New("login not implemented yet")
}

// GetUser retrieves a user by ID
func(s* UserService) GetUser(ctx context.Context, tenantName string, userID int)(*public.User, error)
{
    user, err : = s.userRepo.GetByID(ctx, tenantName, userID) if err != nil
    {
        return nil, fmt.Errorf("failed to get user: %w", err)
    }
    if user
        == nil
        {
            return nil, errors.New("user not found")
        }
    return user, nil
}

// UpdateUser updates user information
func(s* UserService) UpdateUser(ctx context.Context, tenantName string, user* public.User) error
{
    // Validate
    if user
        .ID == 0
        {
            return errors.New("user ID is required")
        }
    if user
        .Username == ""
        {
            return errors.New("username is required")
        }
    if user
        .Email == ""
        {
            return errors.New("email is required")
        }

    // Update
    if err := s.userRepo.Update(ctx, tenantName, user);
    err != nil
    {
        return fmt.Errorf("failed to update user: %w", err)
    }

    return nil
}

// DeleteUser soft deletes a user
func(s* UserService) DeleteUser(ctx context.Context, tenantName string, userID, tenantID int) error
{
    if err := s.userRepo.Delete(ctx, tenantName, userID, tenantID);
    err != nil
    {
        return fmt.Errorf("failed to delete user: %w", err)
    }
    return nil
}

// ListUsers retrieves a paginated list of users with filters
func(s* UserService) ListUsers(ctx context.Context, tenantName string, filter public.UserFilter, pagination public.Pagination)(*public.UserListResponse, error)
{
    users, total, err : = s.userRepo.List(ctx, tenantName, filter, pagination) if err != nil { return nil, fmt.Errorf("failed to list users: %w", err) }

                            totalPages : = 0 if pagination.PageSize > 0
    {
        totalPages = (total + pagination.PageSize - 1) / pagination.PageSize
    }

    return &public.UserListResponse {
        Users : users,
        Total : total,
        Page : pagination.Page,
        PageSize : pagination.PageSize,
        TotalPages : totalPages,
    },
           nil
}

// ChangePassword changes user password
func(s* UserService) ChangePassword(ctx context.Context, tenantName string, userID int, oldPassword, newPassword string) error
{
    // Get user first
    user, err : = s.userRepo.GetByID(ctx, tenantName, userID) if err != nil
    {
        return fmt.Errorf("failed to get user: %w", err)
    }
    if user
        == nil
        {
            return errors.New("user not found")
        }

    // Verify old password (simplified)
    if user
        .Password != oldPassword { return errors.New("invalid old password") }

                     // Update password
                     user.Password
            = newPassword if err : = s.userRepo.Update(ctx, tenantName, user);
    err != nil
    {
        return fmt.Errorf("failed to update password: %w", err)
    }

    return nil
}

// UpdateStatus updates user status
func(s* UserService) UpdateStatus(ctx context.Context, tenantName string, userID int, status string) error
{
    user, err : = s.userRepo.GetByID(ctx, tenantName, userID) if err != nil
    {
        return fmt.Errorf("failed to get user: %w", err)
    }
    if user
        == nil {
            return errors.New("user not found")
        }

            // Validate status
            validStatuses : = map[string] bool
        {
            "active" : true,
                       "inactive" : true,
                                    "suspended" : true,
        }
    if !validStatuses
        [status] {
            return fmt.Errorf("invalid status: %s", status)
        }

        user.Status
            = status if err : = s.userRepo.Update(ctx, tenantName, user);
    err != nil
    {
        return fmt.Errorf("failed to update status: %w", err)
    }

    return nil
}
