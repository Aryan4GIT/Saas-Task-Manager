package models

import (
	"time"

	"github.com/google/uuid"
)

type Organization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type User struct {
	ID           uuid.UUID `json:"id"`
	OrgID        uuid.UUID `json:"org_id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never expose in JSON
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type RefreshToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	OrgID     uuid.UUID  `json:"org_id"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	CreatedAt time.Time  `json:"created_at"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
}

type Task struct {
	ID             uuid.UUID  `json:"id"`
	OrgID          uuid.UUID  `json:"org_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Status         string     `json:"status"` // todo, in_progress, done, verified, approved
	Priority       string     `json:"priority"`
	AssignedTo     *uuid.UUID `json:"assigned_to,omitempty"`
	AssignedToName *string    `json:"assigned_to_name,omitempty"`
	CreatedBy      uuid.UUID  `json:"created_by"`
	CreatedByName  *string    `json:"created_by_name,omitempty"`
	VerifiedBy     *uuid.UUID `json:"verified_by,omitempty"`
	VerifiedByName *string    `json:"verified_by_name,omitempty"`
	VerifiedAt     *time.Time `json:"verified_at,omitempty"`
	ApprovedBy     *uuid.UUID `json:"approved_by,omitempty"`
	ApprovedByName *string    `json:"approved_by_name,omitempty"`
	ApprovedAt     *time.Time `json:"approved_at,omitempty"`
	DueDate        *time.Time `json:"due_date,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type Issue struct {
	ID             uuid.UUID  `json:"id"`
	OrgID          uuid.UUID  `json:"org_id"`
	Title          string     `json:"title"`
	Description    string     `json:"description"`
	Severity       string     `json:"severity"`
	Status         string     `json:"status"`
	ReportedBy     uuid.UUID  `json:"reported_by"`
	ReportedByName *string    `json:"reported_by_name,omitempty"`
	AssignedTo     *uuid.UUID `json:"assigned_to,omitempty"`
	AssignedToName *string    `json:"assigned_to_name,omitempty"`
	AISummary      *string    `json:"ai_summary,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
}

type AuditLog struct {
	ID         uuid.UUID              `json:"id"`
	OrgID      uuid.UUID              `json:"org_id"`
	UserID     *uuid.UUID             `json:"user_id,omitempty"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   *uuid.UUID             `json:"entity_id,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
	IPAddress  *string                `json:"ip_address,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}
