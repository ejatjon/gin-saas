package middleware
import (
	"saas/src/bootstrap"
)

type middleware struct {
	Config *bootstrap.Config
}

func NewMiddleware(config *bootstrap.Config) *middleware {
	return &middleware{Config: config}
}
