package service

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/heth/STM/internal/model"
	"github.com/heth/STM/internal/repository"
	"github.com/heth/STM/internal/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService handles authentication business logic.
type AuthService struct {
	userRepo         *repository.UserRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	jwtService       *utils.JWTService
}

// NewAuthService creates a new AuthService.
func NewAuthService(
	userRepo *repository.UserRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	jwtService *utils.JWTService,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		refreshTokenRepo: refreshTokenRepo,
		jwtService:       jwtService,
	}
}

// RegisterRequest for user registration.
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Username string `json:"username" binding:"required,min=3,max=50"`
}

// LoginRequest for user login.
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse contains tokens and user info.
type AuthResponse struct {
	AccessToken  string      `json:"access_token"`
	RefreshToken string      `json:"refresh_token"`
	ExpiresIn    int         `json:"expires_in"` // seconds
	User         *model.User `json:"user"`
}

// Register creates a new user and returns tokens.
func (s *AuthService) Register(req *RegisterRequest) (*AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	username := strings.TrimSpace(req.Username)

	exists, err := s.userRepo.ExistsByEmail(email)
	if err != nil {
		return nil, utils.NewAppError(500, "database error", err)
	}
	if exists {
		return nil, utils.NewAppError(409, "email already registered", nil)
	}

	exists, err = s.userRepo.ExistsByUsername(username)
	if err != nil {
		return nil, utils.NewAppError(500, "database error", err)
	}
	if exists {
		return nil, utils.NewAppError(409, "username already taken", nil)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, utils.NewAppError(500, "failed to hash password", err)
	}

	user := &model.User{
		Email:        email,
		Username:     username,
		PasswordHash: string(hash),
		Role:         model.RoleUser,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, utils.NewAppError(500, "failed to create user", err)
	}

	return s.issueTokens(user)
}

// Login validates credentials and returns tokens.
func (s *AuthService) Login(req *LoginRequest) (*AuthResponse, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))

	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrUnauthorized
		}
		return nil, utils.NewAppError(500, "database error", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, utils.ErrUnauthorized
	}

	return s.issueTokens(user)
}

// Refresh rotates refresh token and returns new tokens.
func (s *AuthService) Refresh(refreshToken string) (*AuthResponse, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return nil, utils.ErrUnauthorized
	}

	claims, err := s.jwtService.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, utils.ErrUnauthorized
	}

	// Verify token exists in store (rotation: invalidate old)
	stored, err := s.refreshTokenRepo.GetByToken(refreshToken)
	if err != nil || stored == nil {
		return nil, utils.ErrUnauthorized
	}

	user, err := s.userRepo.GetByID(claims.UserID)
	if err != nil || user == nil {
		return nil, utils.ErrUnauthorized
	}

	// Delete old refresh token (rotation)
	_ = s.refreshTokenRepo.DeleteByID(stored.ID)

	return s.issueTokens(user)
}

func (s *AuthService) issueTokens(user *model.User) (*AuthResponse, error) {
	tokenID := uuid.New().String()

	accessToken, err := s.jwtService.GenerateAccessToken(user.ID, user.Role, user.Email)
	if err != nil {
		return nil, utils.NewAppError(500, "failed to generate access token", err)
	}

	refreshToken, err := s.jwtService.GenerateRefreshToken(user.ID, tokenID)
	if err != nil {
		return nil, utils.NewAppError(500, "failed to generate refresh token", err)
	}

	// Store refresh token for validation (rotation)
	rt := &model.RefreshToken{
		Token:     refreshToken,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(s.jwtService.RefreshExpiry),
	}
	if err := s.refreshTokenRepo.Create(rt); err != nil {
		return nil, utils.NewAppError(500, "failed to store refresh token", err)
	}

	expiresIn := int(s.jwtService.AccessExpiry.Seconds())
	return &AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresIn,
		User:         user,
	}, nil
</think>
Adding the missing import and fixing ExpiresAt in `auth_service.go`:
<｜tool▁calls▁begin｜><｜tool▁call▁begin｜>
StrReplace