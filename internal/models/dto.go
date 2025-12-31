package models

// Request/Response DTOs

type RegisterRequest struct {
	OrgName   string `json:"org_name" binding:"required"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type CreateTaskRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description"`
	Priority    string  `json:"priority"`
	AssignedTo  *string `json:"assigned_to"`
	DueDate     *string `json:"due_date"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Status      *string `json:"status"`
	Priority    *string `json:"priority"`
	AssignedTo  *string `json:"assigned_to"`
	DueDate     *string `json:"due_date"`
}

type CreateIssueRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description string  `json:"description" binding:"required"`
	Severity    string  `json:"severity" binding:"required"`
	AssignedTo  *string `json:"assigned_to"`
}

type UpdateIssueRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Severity    *string `json:"severity"`
	Status      *string `json:"status"`
	AssignedTo  *string `json:"assigned_to"`
}

type CreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Role      string `json:"role" binding:"required"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type SuccessResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type UploadDocumentRequest struct {
	TaskID *string `form:"task_id"`
	Title  *string `form:"title"`
}

type VerifyDocumentRequest struct {
	Question string `json:"question" binding:"required"`
}

type UpdateDocumentStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Notes  string `json:"notes"`
}

type DocumentSummaryRequest struct {
	IncludeTaskContext bool `json:"include_task_context"`
}

type VerifyDocumentCitation struct {
	ChunkIndex int    `json:"chunk_index"`
	Snippet    string `json:"snippet"`
}

type VerifyDocumentResponse struct {
	Verdict     string                   `json:"verdict"`
	Confidence  float64                  `json:"confidence"`
	Answer      string                   `json:"answer"`
	Citations   []VerifyDocumentCitation `json:"citations"`
	RawModelOut string                   `json:"raw_model_output,omitempty"`
}

type DocumentSummaryResponse struct {
	Summary                    string   `json:"summary"`
	KeyPoints                  []string `json:"key_points,omitempty"`
	DocumentType               string   `json:"document_type"`
	QualityAssessment          string   `json:"quality_assessment"`
	VerificationRecommendation string   `json:"verification_recommendation"`
	Notes                      string   `json:"notes,omitempty"`
}
