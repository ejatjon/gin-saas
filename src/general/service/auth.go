package service

import (
	"errors"
	"fmt"
	"saas/src/bootstrap"
	"saas/src/general/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// AuthService 用户验证服务，负责生成token 和 验证token的有效性
type AuthService struct {
	config *bootstrap.Config
}

func NewAuthService(config *bootstrap.Config) *AuthService {
	return &AuthService{
		config: config,
	}
}

// GenerateTokens 生成访问token和刷新token
func (s *AuthService) GenerateTokens(userID, tenantID int) (accessToken, refreshToken string, err error) {
	accessToken, err = s.generateAccessToken(userID, tenantID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err = s.generateRefreshToken(userID, tenantID)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return accessToken, refreshToken, nil
}

// generateAccessToken 生成访问token
func (s *AuthService) generateAccessToken(userID, tenantID int) (string, error) {
	now := time.Now()
	claims := &domain.AccessJwtClaims{
		UserID:   userID,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.PublicJWTConfig.AccessIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.config.PublicJWTConfig.AccessExpire))),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.config.PublicJWTConfig.AccessSecretKey)
}

// generateRefreshToken 生成刷新token
func (s *AuthService) generateRefreshToken(userID, tenantID int) (string, error) {
	now := time.Now()
	claims := &domain.RefreshJwtClaims{
		UserID:   userID,
		TenantID: tenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.config.PublicJWTConfig.RefreshIssuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.config.PublicJWTConfig.RefreshExpire))),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.config.PublicJWTConfig.RefreshSecretKey)
}

// ValidateAccessToken 验证访问token的有效性
func (s *AuthService) ValidateAccessToken(tokenString string) (*domain.AccessJwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.AccessJwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.config.PublicJWTConfig.AccessSecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*domain.AccessJwtClaims); ok && token.Valid {
		if claims.Issuer != s.config.PublicJWTConfig.AccessIssuer {
			return nil, errors.New("invalid issuer")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ValidateRefreshToken 验证刷新token的有效性
func (s *AuthService) ValidateRefreshToken(tokenString string) (*domain.RefreshJwtClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.RefreshJwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.config.PublicJWTConfig.RefreshSecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*domain.RefreshJwtClaims); ok && token.Valid {
		if claims.Issuer != s.config.PublicJWTConfig.RefreshIssuer {
			return nil, errors.New("invalid issuer")
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshTokens 通过刷新token 生成访问token
func (s *AuthService) RefreshTokens(refreshToken string) (string, error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("invalid refresh token: %w", err)
	}

	return s.generateAccessToken(claims.UserID, claims.TenantID)
}

// ExtractUserInfoFromAccessToken 返回访问token中的租户id和用户id
func (s *AuthService) ExtractUserInfoFromAccessToken(accessToken string) (userID, tenantID int, err error) {
	claims, err := s.ValidateAccessToken(accessToken)
	if err != nil {
		return 0, 0, err
	}
	return claims.UserID, claims.TenantID, nil
}

// ExtractUserInfoFromRefreshToken 返回刷新token中的租户id和用户id
func (s *AuthService) ExtractUserInfoFromRefreshToken(refreshToken string) (userID, tenantID int, err error) {
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return 0, 0, err
	}
	return claims.UserID, claims.TenantID, nil
}