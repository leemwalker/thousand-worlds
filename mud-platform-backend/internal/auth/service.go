package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

// Config holds JWT configuration
type Config struct {
	SecretKey       []byte
	TokenExpiration time.Duration
}

// Service handles authentication logic
type Service struct {
	config *Config
	repo   Repository
}

// NewService creates a new auth service
func NewService(config *Config, repo Repository) *Service {
	return &Service{
		config: config,
		repo:   repo,
	}
}

// Register creates a new user account
func (s *Service) Register(ctx context.Context, email, password string) (*User, error) {
	// Check if user exists
	existing, err := s.repo.GetUserByEmail(ctx, email)
	if err == nil && existing != nil {
		return nil, ErrUserExists
	}
	// If error is not "not found", it's a real error
	if err != nil && err != ErrUserNotFound {
		return nil, err
	}

	// Hash password using Argon2id
	hasher := NewPasswordHasher()
	hashedPassword, err := hasher.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &User{
		UserID:       uuid.New(),
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    time.Now().UTC(),
	}

	if err := s.repo.CreateUser(ctx, user); err != nil {
		if err == ErrDuplicateEmail {
			return nil, ErrUserExists
		}
		return nil, err
	}

	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *Service) Login(ctx context.Context, email, password string) (string, *User, error) {
	// Get user
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Verify password using Argon2id
	hasher := NewPasswordHasher()
	match, err := hasher.ComparePassword(password, user.PasswordHash)
	if err != nil || !match {
		return "", nil, ErrInvalidCredentials
	}

	// Generate token
	token, err := s.GenerateToken(user.UserID, uuid.Nil)
	if err != nil {
		return "", nil, err
	}

	// Update last login
	user.LastLogin = timePtr(time.Now().UTC())
	s.repo.UpdateUser(ctx, user)

	return token, user, nil
}

// GenerateToken creates a new JWT token
func (s *Service) GenerateToken(userID, characterID uuid.UUID) (string, error) {
	claims := &Claims{
		UserID: userID.String(),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.config.TokenExpiration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.config.SecretKey)
}

// ValidateToken validates a JWT token and returns the claims
func (s *Service) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return s.config.SecretKey, nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, ErrInvalidToken
}

// GenerateSecretKey generates a random secret key
func GenerateSecretKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return []byte(base64.StdEncoding.EncodeToString(key)), nil
}

func timePtr(t time.Time) *time.Time {
	return &t
}
