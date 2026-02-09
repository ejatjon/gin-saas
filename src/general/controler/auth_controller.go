package controler

import (
	"net/http"

	"saas/src/general/middleware"
	general_repository "saas/src/general/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// AuthController handles authentication-related HTTP requests
type AuthController struct {
	userAuthService *general_repository.UserAuthService
	passwordService *general_repository.PasswordService
	authService     *general_repository.AuthService
}

// NewAuthController creates a new AuthController instance
func NewAuthController(
	userAuthService *general_repository.UserAuthService,
	passwordService *general_repository.PasswordService,
	authService *general_repository.AuthService,
) *AuthController {
	return &AuthController{
		userAuthService: userAuthService,
		passwordService: passwordService,
		authService:     authService,
	}
}

// RegisterRoutes registers authentication routes
func (c *AuthController) RegisterRoutes(router *gin.RouterGroup) {
	// Public routes (no authentication required)
	public := router.Group("/auth")
	{
		public.POST("/login", c.Login)
		public.POST("/register", c.Register)
		public.POST("/refresh", c.RefreshToken)
		public.GET("/verify", c.VerifyToken)
	}

	// Protected routes (authentication required)
	protected := router.Group("/auth")
	protected.Use(middleware.NewAuthMiddleware(c.authService).RequireAuth())
	{
		protected.POST("/logout", c.Logout)
		protected.POST("/change-password", c.ChangePassword)
	}
}

// Login handles user login
// @Summary User login
// @Description Authenticate user with username/email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body general_repository.LoginRequest true "Login credentials"
// @Success 200 {object} general_repository.LoginResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	// Get tenant from context (set by SubDomainMiddleware)
	tenant, exists := ctx.Get("tenant")
	if !exists {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Tenant information not found",
			Message: "Cannot determine tenant from request",
		})
		return
	}

	tenantSchema := tenant.(string)
	_ = tenantSchema // Currently not used in placeholder implementation

	// Parse request body
	var req general_repository.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Validate required fields
	if req.Username == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Username is required",
			Message: "Please provide a username or email",
		})
		return
	}

	if req.Password == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Password is required",
			Message: "Please provide a password",
		})
		return
	}

	// Note: The actual login logic is implemented in UserAuthService
	// Since it requires database integration, we'll return a placeholder response
	// In a real implementation, we would call c.userAuthService.Login()

	ctx.JSON(http.StatusNotImplemented, ErrorResponse{
		Error:   "Login not implemented",
		Message: "Login endpoint requires database integration",
	})
}

// Register handles user registration
// @Summary User registration
// @Description Register a new user account
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body general_repository.RegisterRequest true "Registration data"
// @Success 201 {object} general_repository.RegisterResponse
// @Failure 400 {object} ErrorResponse
// @Failure 409 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	// Get tenant from context
	tenant, exists := ctx.Get("tenant")
	if !exists {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Tenant information not found",
			Message: "Cannot determine tenant from request",
		})
		return
	}

	tenantSchema := tenant.(string)
	_ = tenantSchema // Currently not used in placeholder implementation

	// Parse request body
	var req general_repository.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Validate request data
	if req.Username == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Username is required",
			Message: "Please provide a username",
		})
		return
	}

	if req.Email == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Email is required",
			Message: "Please provide an email address",
		})
		return
	}

	// Validate email format
	if err := c.passwordService.ValidateEmail(req.Email); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid email",
			Message: err.Error(),
		})
		return
	}

	if req.Password == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Password is required",
			Message: "Please provide a password",
		})
		return
	}

	// Validate password complexity
	if err := c.passwordService.ValidatePassword(req.Password); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Weak password",
			Message: err.Error(),
		})
		return
	}

	if req.FirstName == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "First name is required",
			Message: "Please provide your first name",
		})
		return
	}

	if req.LastName == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Last name is required",
			Message: "Please provide your last name",
		})
		return
	}

	// Note: The actual registration logic is implemented in UserAuthService
	// Since it requires database integration, we'll return a placeholder response
	// In a real implementation, we would call c.userAuthService.Register()

	ctx.JSON(http.StatusNotImplemented, ErrorResponse{
		Error:   "Registration not implemented",
		Message: "Registration endpoint requires database integration",
	})
}

// Logout handles user logout
// @Summary User logout
// @Description Logout user by invalidating refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param refresh_token body string false "Refresh token to invalidate"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/logout [post]
func (c *AuthController) Logout(ctx *gin.Context) {
	// Parse request body for optional refresh token
	type LogoutRequest struct {
		RefreshToken string `json:"refresh_token"`
	}

	var req LogoutRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		// EOF means no body, which is acceptable for logout
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// In a stateless JWT system, logout is typically handled client-side
	// If a refresh token is provided, we can validate it before telling client to discard
	if req.RefreshToken != "" {
		_, err := c.authService.ValidateRefreshToken(req.RefreshToken)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Invalid refresh token",
				Message: "The provided refresh token is invalid",
			})
			return
		}
	}

	ctx.JSON(http.StatusOK, SuccessResponse{
		Message: "Logged out successfully",
		Success: true,
	})
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Generate new access and refresh tokens using a valid refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param refresh_token body string true "Refresh token"
// @Success 200 {object} general_repository.TokenPair
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/refresh [post]
func (c *AuthController) RefreshToken(ctx *gin.Context) {
	// Parse request body
	type RefreshRequest struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	var req RefreshRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Use the auth service to refresh tokens
	accessToken, refreshToken, err := c.authService.RefreshTokens(req.RefreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Token refresh failed",
			Message: err.Error(),
		})
		return
	}

	// TODO: Get expires_in from auth service configuration
	// For now, use default of 3600 seconds (1 hour)
	expiresIn := int64(3600)

	ctx.JSON(http.StatusOK, general_repository.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
	})
}

// VerifyToken verifies an access token
// @Summary Verify access token
// @Description Verify if an access token is valid and get user information
// @Tags Authentication
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {object} TokenVerificationResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /auth/verify [get]
func (c *AuthController) VerifyToken(ctx *gin.Context) {
	// Extract token from Authorization header
	authHeader := ctx.GetHeader("Authorization")
	if authHeader == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Authorization header required",
			Message: "Please provide an Authorization header with Bearer token",
		})
		return
	}

	// Check if the header has the Bearer prefix
	var tokenString string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	} else {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid Authorization header",
			Message: "Authorization header must be in format: Bearer <token>",
		})
		return
	}

	// Validate the access token
	claims, err := c.authService.ValidateAccessToken(tokenString)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Invalid token",
			Message: err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, TokenVerificationResponse{
		Valid:     true,
		UserID:    claims.UserID,
		TenantID:  claims.TenantID,
		Issuer:    claims.Issuer,
		IssuedAt:  claims.IssuedAt,
		ExpiresAt: claims.ExpiresAt,
	})
}

// ChangePassword changes user password
// @Summary Change user password
// @Description Change the password for the authenticated user
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "Password change data"
// @Success 200 {object} SuccessResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /auth/change-password [post]
func (c *AuthController) ChangePassword(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := middleware.GetUserIDFromContext(ctx)
	if !exists {
		ctx.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Authentication required",
			Message: "User must be authenticated to change password",
		})
		return
	}
	_ = userID // Currently not used in placeholder implementation

	// Get tenant from context
	tenant, exists := ctx.Get("tenant")
	if !exists {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Tenant information not found",
			Message: "Cannot determine tenant from request",
		})
		return
	}

	tenantSchema := tenant.(string)
	_ = tenantSchema // Currently not used in placeholder implementation

	// Parse request body
	var req ChangePasswordRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Invalid request",
			Message: err.Error(),
		})
		return
	}

	// Validate request data
	if req.OldPassword == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Old password is required",
			Message: "Please provide your current password",
		})
		return
	}

	if req.NewPassword == "" {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "New password is required",
			Message: "Please provide a new password",
		})
		return
	}

	// Validate new password complexity
	if err := c.passwordService.ValidatePassword(req.NewPassword); err != nil {
		ctx.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Weak password",
			Message: err.Error(),
		})
		return
	}

	// Note: The actual password change logic is implemented in UserAuthService
	// Since it requires database integration, we'll return a placeholder response
	// In a real implementation, we would call c.userAuthService.ChangePassword()

	ctx.JSON(http.StatusNotImplemented, ErrorResponse{
		Error:   "Change password not implemented",
		Message: "Change password endpoint requires database integration",
	})
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// SuccessResponse represents a success response
type SuccessResponse struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// TokenVerificationResponse represents a token verification response
type TokenVerificationResponse struct {
	Valid     bool             `json:"valid"`
	UserID    int              `json:"user_id"`
	TenantID  int              `json:"tenant_id"`
	Issuer    string           `json:"issuer,omitempty"`
	IssuedAt  *jwt.NumericDate `json:"issued_at,omitempty"`
	ExpiresAt *jwt.NumericDate `json:"expires_at,omitempty"`
}

// ChangePasswordRequest represents a password change request
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}
