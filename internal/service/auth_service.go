package service

import (
	"context"
	"errors"
	"time"

	"task/internal/model"
	"task/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users     *repository.UserRepository
	jwtSecret []byte
	ttl       time.Duration
}

type LoginResult struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type Claims struct {
	UserID   uint64 `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func NewAuthService(users *repository.UserRepository, jwtSecret string, ttl time.Duration) *AuthService {
	return &AuthService{users: users, jwtSecret: []byte(jwtSecret), ttl: ttl}
}

func (s *AuthService) Register(ctx context.Context, username, password string, email *string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	u := &model.User{Username: username, Email: email, PasswordHash: string(hash)}
	return s.users.Create(ctx, u)
}

func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	u, err := s.users.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, errors.New("invalid credentials")
	}
	now := time.Now()
	expiresAt := now.Add(s.ttl)
	claims := Claims{
		UserID:   u.ID,
		Username: u.Username,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString(s.jwtSecret)
	if err != nil {
		return nil, err
	}
	return &LoginResult{AccessToken: signed, TokenType: "bearer", ExpiresIn: int64(s.ttl.Seconds())}, nil
}
