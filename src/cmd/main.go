package main

import (
	"context"
	"fmt"
	"log"

	"saas/src/bootstrap"
	general_domain "saas/src/general/domain"
	general_middleware "saas/src/general/middleware"
	general_repository "saas/src/general/repository"
	general_service "saas/src/general/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// Initialize configuration
	config := bootstrap.InitConfig()

	// Setup Gin router
	if config.ServerConfig.Debug == "true" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add CORS middleware (optional, for development)
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Initialize database connection
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		config.Database.User, config.Database.Password,
		config.Database.Host, config.Database.Port,
		config.Database.Name)

	log.Printf("Connecting to database: %s@%s:%s",
		config.Database.User, config.Database.Host, config.Database.Port)

	pool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	// Test database connection
	conn, err := pool.Acquire(context.Background())
	if err != nil {
		log.Fatalf("Failed to acquire database connection: %v", err)
	}
	conn.Release()
	log.Println("Database connection established successfully")

	// Initialize DBManager
	dbManager := bootstrap.NewDBManager(pool)

	// Register table creation functions with DBManager
	dbManager.RegisterSchemaCreator("users", general_domain.CreateUserTable)
	dbManager.RegisterSchemaCreator("groups_permissions", general_domain.CreateGroupPermissionTable)

	// Initialize middleware
	generalMiddleware := general_middleware.NewMiddleware(config)

	// Initialize repository
	userRepo := general_repository.NewUserRepository()

	// Initialize user service
	userService := general_service.NewUserService(dbManager, userRepo)
	_ = userService // Mark as used for now

	// Initialize controllers
	// Note: authController currently not implemented, skipping for now

	// Initialize auth middleware
	// Note: authMiddleware currently not implemented, skipping for now

	// Apply subdomain middleware to all routes
	router.Use(generalMiddleware.SubDomainMiddleware())

	// Public routes (no authentication required)
	// These are accessible from any subdomain
	publicRoutes := router.Group("/api")
	{
		// Health check endpoint
		publicRoutes.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"status":  "healthy",
				"service": config.ServerConfig.Name,
			})
		})

		// Database health check
		publicRoutes.GET("/health/db", func(c *gin.Context) {
			conn, err := pool.Acquire(context.Background())
			if err != nil {
				c.JSON(500, gin.H{
					"status": "unhealthy",
					"error":  err.Error(),
				})
				return
			}
			defer conn.Release()

			var result int
			err = conn.QueryRow(context.Background(), "SELECT 1").Scan(&result)
			if err != nil {
				c.JSON(500, gin.H{
					"status": "unhealthy",
					"error":  err.Error(),
				})
				return
			}

			c.JSON(200, gin.H{
				"status":   "healthy",
				"database": "connected",
			})
		})

		// Register authentication routes
		// authController.RegisterRoutes(publicRoutes)
	}

	// Tenant-specific routes
	// These routes require a valid tenant subdomain
	tenantRoutes := router.Group("/api/tenant")
	{
		// Test endpoint to verify tenant isolation
		tenantRoutes.GET("/info", func(c *gin.Context) {
			tenant := c.MustGet("tenant").(string)

			// Get connection for tenant schema
			conn, err := dbManager.GetConnForSchema(context.Background(), tenant)
			if err != nil {
				c.JSON(500, gin.H{
					"error":   "Failed to get tenant database connection",
					"details": err.Error(),
				})
				return
			}
			defer conn.Release()

			c.JSON(200, gin.H{
				"tenant":   tenant,
				"message":  fmt.Sprintf("Welcome to tenant: %s", tenant),
				"database": "connected",
			})
		})

		// Create tenant schema endpoint (for initial setup)
		tenantRoutes.POST("/setup", func(c *gin.Context) {
			tenant := c.MustGet("tenant").(string)

			// Create schema and get connection
			conn, err := dbManager.CreateSchema(context.Background(), tenant)
			if err != nil {
				c.JSON(500, gin.H{
					"error":   "Failed to create tenant schema",
					"details": err.Error(),
				})
				return
			}
			defer conn.Release()

			// Create user table in tenant schema
			err = general_domain.CreateUserTable(context.Background(), conn)
			if err != nil {
				c.JSON(500, gin.H{
					"error":   "Failed to create user table",
					"details": err.Error(),
				})
				return
			}

			// Create group and permission tables
			err = general_domain.CreateGroupPermissionTable(context.Background(), conn)
			if err != nil {
				c.JSON(500, gin.H{
					"error":   "Failed to create group and permission tables",
					"details": err.Error(),
				})
				return
			}

			c.JSON(200, gin.H{
				"message":        fmt.Sprintf("Tenant schema '%s' created successfully", tenant),
				"tables_created": []string{"users", "groups", "permissions", "group_permissions"},
			})
		})

		// Protected tenant routes (require authentication)
		protectedTenantRoutes := tenantRoutes.Group("/")
		// TODO: Add authentication middleware when implemented
		// protectedTenantRoutes.Use(authMiddleware.RequireAuth())
		{
			// User profile endpoint
			protectedTenantRoutes.GET("/profile", func(c *gin.Context) {
				// TODO: Get user ID from JWT token after authentication is implemented
				// For now, use a test user ID
				userID := 1

				tenant := c.MustGet("tenant").(string)

				// Get connection for tenant schema
				conn, err := dbManager.GetConnForSchema(context.Background(), tenant)
				if err != nil {
					c.JSON(500, gin.H{
						"error":   "Failed to get tenant database connection",
						"details": err.Error(),
					})
					return
				}
				defer conn.Release()

				// Get user from repository
				user, err := userRepo.GetUser(context.Background(), conn, userID)
				if err != nil {
					c.JSON(404, gin.H{
						"error":   "User not found",
						"details": err.Error(),
					})
					return
				}

				// Return user info without password
				c.JSON(200, gin.H{
					"id":         user.ID,
					"username":   user.Username,
					"email":      user.Email,
					"first_name": user.FirstName,
					"last_name":  user.LastName,
					"status":     user.Status,
					// "last_login": user.LastLogin, // LastLogin field not available in UserModel
					"created_at": user.CreatedAt,
					"updated_at": user.UpdatedAt,
				})
			})

			// Update user profile endpoint
			protectedTenantRoutes.PUT("/profile", func(c *gin.Context) {
				// TODO: Get user ID from JWT token after authentication is implemented
				// For now, use a test user ID
				userID := 1

				tenant := c.MustGet("tenant").(string)

				// Get connection for tenant schema
				conn, err := dbManager.GetConnForSchema(context.Background(), tenant)
				if err != nil {
					c.JSON(500, gin.H{
						"error":   "Failed to get tenant database connection",
						"details": err.Error(),
					})
					return
				}
				defer conn.Release()

				// Get existing user
				user, err := userRepo.GetUser(context.Background(), conn, userID)
				if err != nil {
					c.JSON(404, gin.H{
						"error":   "User not found",
						"details": err.Error(),
					})
					return
				}

				// Parse update request
				type UpdateProfileRequest struct {
					FirstName string `json:"first_name"`
					LastName  string `json:"last_name"`
					Email     string `json:"email"`
				}

				var updateReq UpdateProfileRequest
				if err := c.ShouldBindJSON(&updateReq); err != nil {
					c.JSON(400, gin.H{
						"error":   "Invalid request",
						"details": err.Error(),
					})
					return
				}

				// Update user fields
				if updateReq.FirstName != "" {
					user.FirstName = updateReq.FirstName
				}
				if updateReq.LastName != "" {
					user.LastName = updateReq.LastName
				}
				if updateReq.Email != "" {
					// Validate email format
					// TODO: Add email validation when service is implemented
					// if err := passwordService.ValidateEmail(updateReq.Email); err != nil {
					// 	c.JSON(400, gin.H{
					// 		"error":   "Invalid email",
					// 		"details": err.Error(),
					// 	})
					// 	return
					// }
					user.Email = updateReq.Email
				}

				// Save updated user
				updatedUser, err := userRepo.UpdateUser(context.Background(), conn, user)
				if err != nil {
					c.JSON(500, gin.H{
						"error":   "Failed to update profile",
						"details": err.Error(),
					})
					return
				}
				user = updatedUser

				c.JSON(200, gin.H{
					"message": "Profile updated successfully",
					"user": gin.H{
						"id":         user.ID,
						"username":   user.Username,
						"email":      user.Email,
						"first_name": user.FirstName,
						"last_name":  user.LastName,
						"status":     user.Status,
					},
				})
			})
		}
	}

	// Default route
	router.GET("/", func(c *gin.Context) {
		tenant, exists := c.Get("tenant")
		if !exists {
			tenant = "public"
		}

		c.JSON(200, gin.H{
			"service": config.ServerConfig.Name,
			"version": "1.0.0",
			"tenant":  tenant,
			"domain":  config.ServerConfig.Domain,
			"message": "SaaS API Service",
			"endpoints": gin.H{
				"health":       "/api/health",
				"auth":         "/api/auth",
				"tenant_info":  "/api/tenant/info",
				"tenant_setup": "/api/tenant/setup",
			},
		})
	})

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"error": "Endpoint not found",
			"path":  c.Request.URL.Path,
		})
	})

	// Start server
	serverAddr := fmt.Sprintf("%s:%s", config.ServerConfig.Host, config.ServerConfig.Port)
	log.Printf("Starting server on %s", serverAddr)
	log.Printf("Server domain: %s", config.ServerConfig.Domain)
	log.Printf("Debug mode: %s", config.ServerConfig.Debug)

	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
