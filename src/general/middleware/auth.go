package middleware

import "github.com/gin-gonic/gin"


func (m *middleware) AuthMiddleware()  gin.HandlerFunc {
	
	return func(c *gin.Context) {
		accessToken, err := c.Cookie("AccessToken")
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		cliams, err := m.authService.ValidateAccessToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "Unauthorized"})
			return
		}
		
		c.Set("UserID", cliams.UserID)
		c.Set("TenantID", cliams.TenantID)
		c.Next()
	}
	
}