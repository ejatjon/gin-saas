package middleware
import (
	"saas/src/bootstrap"
	"saas/src/general/service"
)

type middleware struct {
	Config *bootstrap.Config
	authService service.AuthService
}

func NewMiddleware(config *bootstrap.Config) *middleware {
	return &middleware{Config: config}
}
