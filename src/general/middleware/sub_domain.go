package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func (m *middleware)SubDomainMiddleware() gin.HandlerFunc {

	return func(c *gin.Context) {
		tenant := strings.ToLower(GetSubDomain(c.Request,m.Config.ServerConfig.Domain))
		if tenant == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}
		c.Set("TenantName", tenant)
		c.Next()
	}
}

func GetSubDomain(r *http.Request,domain string) string {
	host := r.Host
	var subdomain string = "public"
	sd := strings.TrimSuffix(strings.Split(host, ":")[0],domain);
	if len(sd) > 0 {
		subdomain = strings.TrimSuffix(sd, ".")
	}
	return subdomain
}
