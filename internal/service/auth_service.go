package service

import (
	"fmt"
	"strings"
	"time"

	"saas-backend/config"
	"saas-backend/internal/models"
	"saas-backend/internal/repository"
	"saas-backend/internal/utils"

	"github.com/google/uuid"
)

type AuthService struct {
	userRepo         *repository.UserRepository
	orgRepo          *repository.OrganizationRepository
	refreshTokenRepo *repository.RefreshTokenRepository
	cfg              *config.Config
}

func NewAuthService(
	userRepo *repository.UserRepository,
	orgRepo *repository.OrganizationRepository,
	refreshTokenRepo *repository.RefreshTokenRepository,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo:         userRepo,
		orgRepo:          orgRepo,
		refreshTokenRepo: refreshTokenRepo,
		cfg:              cfg,
	}
}

func (s *AuthService) Register(req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Create organization slug from name
	slug := strings.ToLower(strings.ReplaceAll(req.OrgName, " ", "-"))

	// Check if organization slug already exists
	existingOrg, err := s.orgRepo.GetBySlug(slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check organization: %w", err)
	}
	if existingOrg != nil {
		return nil, fmt.Errorf("organization with this name already exists")
	}

	// Create organization
	org := &models.Organization{
		ID:   uuid.New(),
		Name: req.OrgName,
		Slug: slug,
	}
	if err := s.orgRepo.Create(org); err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user as admin
	user := &models.User{
		ID:           uuid.New(),
		OrgID:        org.ID,
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         "admin",
		IsActive:     true,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.OrgID, user.Role, s.cfg.JWT.AccessSecret, s.cfg.JWT.AccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.OrgID, s.cfg.JWT.RefreshSecret, s.cfg.JWT.RefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token hash
	tokenHash := utils.HashToken(refreshToken)
	refreshTokenModel := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		OrgID:     user.OrgID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshExpiry),
	}
	if err := s.refreshTokenRepo.Create(refreshTokenModel); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) Login(req *models.LoginRequest) (*models.AuthResponse, error) {
	// Find user by email - note: in production with multiple orgs, you'd want to include org context
	// For now, we'll search using a workaround by querying the database directly
	user, err := s.findUserByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to find user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check password
	if !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is inactive")
	}

	// Generate tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.OrgID, user.Role, s.cfg.JWT.AccessSecret, s.cfg.JWT.AccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := utils.GenerateRefreshToken(user.ID, user.OrgID, s.cfg.JWT.RefreshSecret, s.cfg.JWT.RefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token hash
	tokenHash := utils.HashToken(refreshToken)
	refreshTokenModel := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		OrgID:     user.OrgID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshExpiry),
	}
	if err := s.refreshTokenRepo.Create(refreshTokenModel); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) RefreshToken(tokenString string) (*models.AuthResponse, error) {
	// Validate refresh token
	claims, err := utils.ValidateToken(tokenString, s.cfg.JWT.RefreshSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token")
	}

	// Check if token exists and is not revoked
	tokenHash := utils.HashToken(tokenString)
	storedToken, err := s.refreshTokenRepo.GetByTokenHash(tokenHash)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}
	if storedToken == nil {
		return nil, fmt.Errorf("invalid or expired refresh token")
	}

	// Get user
	user, err := s.userRepo.GetByID(claims.OrgID, claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Generate new tokens
	accessToken, err := utils.GenerateAccessToken(user.ID, user.OrgID, user.Role, s.cfg.JWT.AccessSecret, s.cfg.JWT.AccessExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, err := utils.GenerateRefreshToken(user.ID, user.OrgID, s.cfg.JWT.RefreshSecret, s.cfg.JWT.RefreshExpiry)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Revoke old refresh token
	if err := s.refreshTokenRepo.RevokeToken(tokenHash); err != nil {
		return nil, fmt.Errorf("failed to revoke old token: %w", err)
	}

	// Store new refresh token
	newTokenHash := utils.HashToken(newRefreshToken)
	newRefreshTokenModel := &models.RefreshToken{
		ID:        uuid.New(),
		UserID:    user.ID,
		OrgID:     user.OrgID,
		TokenHash: newTokenHash,
		ExpiresAt: time.Now().Add(s.cfg.JWT.RefreshExpiry),
	}
	if err := s.refreshTokenRepo.Create(newRefreshTokenModel); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return &models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) Logout(userID uuid.UUID) error {
	return s.refreshTokenRepo.RevokeAllUserTokens(userID)
}

// Helper to find user by email across organizations
func (s *AuthService) findUserByEmail(email string) (*models.User, error) {
	// Find user by email globally (works across all organizations)
	// In production with many organizations, consider:
	// 1. Adding org_slug to login request for better scoping
	// 2. Creating indexed lookup tables
	// 3. Implementing caching
	return s.userRepo.GetByEmailGlobal(email)
}
