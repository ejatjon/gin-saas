package middleware

import (
	"saas/src/bootstrap"
	"saas/src/general/service"
)

type middleware struct {
	Config       *bootstrap.Config
	authService  *service.AuthService
	groupService service.GroupService
}

func NewMiddleware(config *bootstrap.Config, authService *service.AuthService, groupService service.GroupService) *middleware {
	return &middleware{
		Config:       config,
		authService:  authService,
		groupService: groupService,
	}
}
