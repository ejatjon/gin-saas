package service

import (
	"saas/src/bootstrap"
)


type TenantService struct {
	dbManager *bootstrap.DBManager
	tableName string
}

func NewTenantService(dbManager *bootstrap.DBManager) *TenantService {
	return &TenantService{
		dbManager: dbManager,
		tableName: "tenants",
	}
}

