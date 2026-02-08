package public

    import(
        "net/http"
        "strconv"

        "saas/src/domain/public"
        "saas/src/service/public"

        "github.com/gin-gonic/gin")

        type UserController struct {
    userService* public.UserService
}

    func NewUserController(userService* public.UserService) *UserController
{
    return &UserController
    {
    userService:
        userService,
    }
}

// ListUsers handles GET /users
func(c* UserController) ListUsers(ctx* gin.Context)
{
    // Get tenant from context (set by middleware)
    tenant, exists : = ctx.Get("tenant") if !exists { ctx.JSON(http.StatusBadRequest, gin.H { "error" : "tenant not found in context" }) return } tenantName : = tenant.(string)

                                                                                                                                                                 // Parse pagination
                                                                                                                                                                 var pagination public.Pagination if err : = ctx.ShouldBindQuery(&pagination);
    err != nil { pagination.Page = 1 pagination.PageSize = 20 }

           // Parse filters
           var filter public.UserFilter if err : = ctx.ShouldBindQuery(&filter);
    err != nil
    {
        // Continue with empty filter if binding fails
    }

    // Get users
    response, err : = c.userService.ListUsers(ctx.Request.Context(), tenantName, filter, pagination) if err != nil { ctx.JSON(http.StatusInternalServerError, gin.H { "error" : err.Error() }) return }

                                                                                                               ctx.JSON(http.StatusOK, response)
}

// GetUser handles GET /users/:id
func(c* UserController) GetUser(ctx* gin.Context)
{
    // Get tenant from context
    tenant, exists : = ctx.Get("tenant") if !exists { ctx.JSON(http.StatusBadRequest, gin.H { "error" : "tenant not found in context" }) return } tenantName : = tenant.(string)

                                                                                                                                                                 // Get user ID from path
                                                                                                                                                                 idStr : = ctx.Param("id") userID,
                   err : = strconv.Atoi(idStr) if err != nil { ctx.JSON(http.StatusBadRequest, gin.H { "error" : "invalid user ID" }) return }

                             // Get user
                             user,
                   err : = c.userService.GetUser(ctx.Request.Context(), tenantName, userID) if err != nil
    {
        ctx.JSON(http.StatusInternalServerError, gin.H { "error" : err.Error() }) return
    }

    if user
        == nil {
            ctx.JSON(http.StatusNotFound, gin.H { "error" : "user not found" }) return
        }

            ctx.JSON(http.StatusOK, user)
}

// CreateUser handles POST /users
func(c* UserController) CreateUser(ctx* gin.Context)
{
    // Get tenant from context
    tenant, exists : = ctx.Get("tenant") if !exists { ctx.JSON(http.StatusBadRequest, gin.H { "error" : "tenant not found in context" }) return } tenantName : = tenant.(string)

                                                                                                                                                                 // Parse request body
                                                                                                                                                                 var req struct {
        Username string `json : "username" binding : "required"` Email string `json : "email" binding : "required,email"` Password string `json : "password" binding : "required,min=6"` FirstName string `json : "first_name"` LastName string `json : "last_name"` Role string `json : "role"` TenantID int    `json : "tenant_id" binding : "required"`
    }

    if err : = ctx.ShouldBindJSON(&req);
    err != nil { ctx.JSON(http.StatusBadRequest, gin.H { "error" : err.Error() }) return }

        // Create user object
        user : = &public.User {
        TenantID : req.TenantID,
        Username : req.Username,
        Email : req.Email,
        Password : req.Password,
        FirstName : req.FirstName,
        LastName : req.LastName,
        Role : req.Role,
    }

                  // Register user
                  createdUser,
             err : = c.userService.Register(ctx.Request.Context(), tenantName, user, req.Password) if err != nil { ctx.JSON(http.StatusInternalServerError, gin.H { "error" : err.Error() }) return }

                                                                                                             ctx.JSON(http.StatusCreated, createdUser)
}

// UpdateUser handles PUT /users/:id
func(c* UserController) UpdateUser(ctx* gin.Context)
{
    // Get tenant from context
    tenant, exists : = ctx.Get("tenant") if !exists { ctx.JSON(http.StatusBadRequest, gin.H { "error" : "tenant not found in context" }) return } tenantName : = tenant.(string)

                                                                                                                                                                 // Get user ID from path
                                                                                                                                                                 idStr : = ctx.Param("id") userID,
                   err : = strconv.Atoi(idStr) if err != nil { ctx.JSON(http.StatusBadRequest, gin.H { "error" : "invalid user ID" }) return }

                             // Parse request body
                             var req struct {
        Username string `json : "username" binding : "required"` Email string `json : "email" binding : "required,email"` FirstName string `json : "first_name"` LastName string `json : "last_name"` Role string `json : "role"` Status string `json : "status"` TenantID int    `json : "tenant_id" binding : "required"`
    }

    if err : = ctx.ShouldBindJSON(&req);
    err != nil { ctx.JSON(http.StatusBadRequest, gin.H { "error" : err.Error() }) return }

        // Create user object
        user : = &public.User
    {
    ID:
        userID,
            TenantID : req.TenantID,
            Username : req.Username,
            Email : req.Email,
            FirstName : req.FirstName,
            LastName : req.LastName,
            Role : req.Role,
            Status : req.Status,
    }

    // Update user
    if err := c.userService.UpdateUser(ctx.Request.Context(), tenantName, user);
    err != nil { ctx.JSON(http.StatusInternalServerError, gin.H { "error" : err.Error() }) return }

           ctx.JSON(http.StatusOK, gin.H { "message" : "user updated successfully" })
}

// DeleteUser handles DELETE /users/:id
func(c* UserController) DeleteUser(ctx* gin.Context)
{
    // Get tenant from context
    tenant, exists : = ctx.Get("tenant") if !exists { ctx.JSON(http.StatusBadRequest, gin.H { "error" : "tenant not found in context" }) return } tenantName : = tenant.(string)

                                                                                                                                                                 // Get user ID from path
                                                                                                                                                                 idStr : = ctx.Param("id") userID,
                   err : = strconv.Atoi(idStr) if err != nil { ctx.JSON(http.StatusBadRequest, gin.H { "error" : "invalid user ID" }) return }

                             // Parse tenant ID from query or body
                             tenantID,
                   err : = strconv.Atoi(ctx.DefaultQuery("tenant_id", "0")) if err != nil || tenantID == 0
    {
        ctx.JSON(http.StatusBadRequest, gin.H { "error" : "tenant_id is required" }) return
    }

    // Delete user
    if err := c.userService.DeleteUser(ctx.Request.Context(), tenantName, userID, tenantID);
    err != nil { ctx.JSON(http.StatusInternalServerError, gin.H { "error" : err.Error() }) return }

           ctx.JSON(http.StatusOK, gin.H { "message" : "user deleted successfully" })
}

// ChangePassword handles POST /users/:id/change-password
func(c* UserController) ChangePassword(ctx* gin.Context)
{
    // Get tenant from context
    tenant, exists : = ctx.Get("tenant") if !exists { ctx.JSON(http.StatusBadRequest, gin.H { "error" : "tenant not found in context" }) return } tenantName : = tenant.(string)

                                                                                                                                                                 // Get user ID from path
                                                                                                                                                                 idStr : = ctx.Param("id") userID,
                   err : = strconv.Atoi(idStr) if err != nil { ctx.JSON(http.StatusBadRequest, gin.H { "error" : "invalid user ID" }) return }

                             // Parse request body
                             var req struct {
        OldPassword string `json : "old_password" binding : "required"` NewPassword string `json : "new_password" binding : "required,min=6"`
    }

    if err : = ctx.ShouldBindJSON(&req);
    err != nil
    {
        ctx.JSON(http.StatusBadRequest, gin.H { "error" : err.Error() }) return
    }

    // Change password
    if err := c.userService.ChangePassword(ctx.Request.Context(), tenantName, userID, req.OldPassword, req.NewPassword);
    err != nil { ctx.JSON(http.StatusInternalServerError, gin.H { "error" : err.Error() }) return }

           ctx.JSON(http.StatusOK, gin.H { "message" : "password changed successfully" })
}

// UpdateStatus handles PUT /users/:id/status
func(c* UserController) UpdateStatus(ctx* gin.Context)
{
    // Get tenant from context
    tenant, exists : = ctx.Get("tenant") if !exists { ctx.JSON(http.StatusBadRequest, gin.H { "error" : "tenant not found in context" }) return } tenantName : = tenant.(string)

                                                                                                                                                                 // Get user ID from path
                                                                                                                                                                 idStr : = ctx.Param("id") userID,
                   err : = strconv.Atoi(idStr) if err != nil { ctx.JSON(http.StatusBadRequest, gin.H { "error" : "invalid user ID" }) return }

                             // Parse request body
                             var req struct {
        Status string `json : "status" binding : "required,oneof=active inactive suspended"`
    }

    if err : = ctx.ShouldBindJSON(&req);
    err != nil
    {
        ctx.JSON(http.StatusBadRequest, gin.H { "error" : err.Error() }) return
    }

    // Update status
    if err := c.userService.UpdateStatus(ctx.Request.Context(), tenantName, userID, req.Status);
    err != nil { ctx.JSON(http.StatusInternalServerError, gin.H { "error" : err.Error() }) return }

           ctx.JSON(http.StatusOK, gin.H { "message" : "status updated successfully" })
}
