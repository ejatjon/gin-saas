package domain

import "github.com/golang-jwt/jwt/v5"

// AccessJwtClaims 访问token
type AccessJwtClaims struct {
	UserID    int   `json:"userID"`
	TenantID  int   `json:"tenantID"`
	jwt.RegisteredClaims
}

// RefreshJwtClaims 刷新token
type RefreshJwtClaims struct {
	UserID    int   `json:"userID"`
	TenantID  int   `json:"tenantID"`
    jwt.RegisteredClaims
}
