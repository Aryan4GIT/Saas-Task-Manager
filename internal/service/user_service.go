package service

import (
	"fmt"

	"saas-backend/internal/models"
	"saas-backend/internal/repository"
	"saas-backend/internal/utils"

	"github.com/google/uuid"
)

type UserService struct {
	userRepo     *repository.UserRepository
	auditLogRepo *repository.AuditLogRepository
}

func NewUserService(userRepo *repository.UserRepository, auditLogRepo *repository.AuditLogRepository) *UserService {
	return &UserService{
		userRepo:     userRepo,
		auditLogRepo: auditLogRepo,
	}
}

func (s *UserService) CreateUser(orgID, createdBy uuid.UUID, req *models.CreateUserRequest) (*models.User, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(orgID, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with this email already exists")
	}

	// Hash password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &models.User{
		ID:           uuid.New(),
		OrgID:        orgID,
		Email:        req.Email,
		PasswordHash: passwordHash,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Role:         req.Role,
		IsActive:     true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &createdBy,
		Action:     "create",
		EntityType: "user",
		EntityID:   &user.ID,
		Details: map[string]interface{}{
			"email": user.Email,
			"role":  user.Role,
		},
	}
	_ = s.auditLogRepo.Create(auditLog)

	return user, nil
}

func (s *UserService) GetUser(orgID, userID uuid.UUID) (*models.User, error) {
	user, err := s.userRepo.GetByID(orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return user, nil
}

func (s *UserService) ListUsers(orgID uuid.UUID) ([]models.User, error) {
	users, err := s.userRepo.List(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (s *UserService) UpdateUser(orgID, userID, updatedBy uuid.UUID, updates *models.User) (*models.User, error) {
	user, err := s.userRepo.GetByID(orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// Update fields
	user.Email = updates.Email
	user.FirstName = updates.FirstName
	user.LastName = updates.LastName
	user.Role = updates.Role
	user.IsActive = updates.IsActive

	if err := s.userRepo.Update(user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &updatedBy,
		Action:     "update",
		EntityType: "user",
		EntityID:   &user.ID,
	}
	_ = s.auditLogRepo.Create(auditLog)

	return user, nil
}

func (s *UserService) DeleteUser(orgID, userID, deletedBy uuid.UUID) error {
	if err := s.userRepo.Delete(orgID, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Create audit log
	auditLog := &models.AuditLog{
		ID:         uuid.New(),
		OrgID:      orgID,
		UserID:     &deletedBy,
		Action:     "delete",
		EntityType: "user",
		EntityID:   &userID,
	}
	_ = s.auditLogRepo.Create(auditLog)

	return nil
}
