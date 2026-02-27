package service

import (
	"authentication_service/internal/config"
	"authentication_service/internal/database/model"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Config *config.AuthService
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type UserClaims struct {
	Permissions []string `json:"perms"`
	jwt.RegisteredClaims
}

func (as AuthService) createAccessToken(claims *UserClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(as.Config.SecretKey)
}

func (as AuthService) createRefreshToken() string {
	return uuid.New().String()
}

func (as AuthService) createClaims(user *model.User) *UserClaims {
	currentTime := time.Now()
	claims := &UserClaims{
		Permissions: user.Permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "auth-service",
			Subject:   fmt.Sprint(user.ID),
			Audience:  jwt.ClaimStrings{"ai-assistant-app"},
			IssuedAt:  jwt.NewNumericDate(currentTime),
			ExpiresAt: jwt.NewNumericDate(currentTime.Add(as.Config.TokenTTL)),
		},
	}
	return claims
}

func (as AuthService) CreateTokenPair(user *model.User) (*TokenPair, error) {
	return as.CreateTokenPairWithClaims(as.createClaims(user))
}

func (as AuthService) CreateTokenPairWithClaims(claims *UserClaims) (*TokenPair, error) {
	accessToken, err := as.createAccessToken(claims)
	if err != nil {
		return nil, err
	}
	refreshToken := as.createRefreshToken()
	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

const defaultHashCost = 12

func (as AuthService) CreatePasswordHash(password string) ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(password), defaultHashCost)
}

func (as AuthService) ValidatePassword(password string, passwordHash []byte) error {
	return bcrypt.CompareHashAndPassword(passwordHash, []byte(password))
}

func (as AuthService) ParseToken(tokenStr string) (*UserClaims, error) {
	claims := &UserClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != "HS256" {
			return nil, fmt.Errorf("unexpected signing method: %s", token.Method.Alg())
		}
		return as.Config.SecretKey, nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid jwt token")
	}
	return claims, nil
}
