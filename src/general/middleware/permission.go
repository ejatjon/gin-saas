package middleware

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)


// PermissionMiddleware 权限中间件
func (m *middleware) PermissionMiddleware(permissionNames []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var userID int
		var tenantName string

		if value, exists := c.Get("TenantName"); !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Tenant not found in context",
			})
			return
		}else{
			tenantName = value.(string)
		}

		if value, exists := c.Get("UserID"); !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "User not authenticated",
			})
			return
		}else{
			userID = value.(int)
		}
		hasPermission := true
		for _, permissionName := range permissionNames {
			if has, err := m.groupService.UserHasPermissionByName(c.Request.Context(), tenantName, userID, permissionName); err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": fmt.Sprintf("Failed to check permission '%s': %v", permissionName, err),
				})
				return
			}else if !has{
				hasPermission = false
				break
			}
		}

		if !hasPermission {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "User does not have required permissions"})
			return
		}

		c.Next()
	}
}
