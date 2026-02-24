package auth

import (
	"context"
	"github.com/Guram-Gurych/gophermart.git/internal/domain"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type AuthRepository interface {
	CreateUser(ctx context.Context, userID uuid.UUID, login, hashPassword string) error
	GetUserByLogin(ctx context.Context, login string) (domain.Users, error)
}

type service struct {
	repository AuthRepository
	secretKey  string
}

func NewService(repo AuthRepository, key string) *service {
	return &service{
		repository: repo,
		secretKey:  key,
	}
}

func (s *service) HashPassword(password string) (string, error) {
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hashPassword), err
}

func (s *service) CheckHashPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *service) GenerateToken(userID uuid.UUID) (string, error) {
	claims := domain.TokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "service",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &claims)

	tokenString, err := token.SignedString([]byte(s.secretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *service) Register(ctx context.Context, login, password string) (string, error) {
	hashPassword, err := s.HashPassword(password)
	if err != nil {
		return "", err
	}

	userID, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	if err = s.repository.CreateUser(ctx, userID, login, hashPassword); err != nil {
		return "", err
	}

	token, err := s.GenerateToken(userID)
	return token, err
}

func (s *service) Login(ctx context.Context, login, password string) (string, error) {
	user, err := s.repository.GetUserByLogin(ctx, login)
	if err != nil {
		return "", err
	}
	if !s.CheckHashPassword(password, user.HashPassword) {
		return "", domain.ErrorInvalidCredentials
	}

	token, err := s.GenerateToken(user.ID)
	return token, err
}
