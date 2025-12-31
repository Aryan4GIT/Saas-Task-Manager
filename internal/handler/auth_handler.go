package handler

import (
	"net/http"

	"saas-backend/config"
	"saas-backend/internal/middleware"
	"saas-backend/internal/models"
	"saas-backend/internal/service"
	"saas-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *service.AuthService
	cfg         *config.Config
}

func NewAuthHandler(authService *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		cfg:         cfg,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if !utils.BindJSON(c, &req) {
		return
	}

	response, err := h.authService.Register(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "registration failed", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusCreated, response)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if !utils.BindJSON(c, &req) {
		return
	}

	response, err := h.authService.Login(&req)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "login failed", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, response)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest
	if !utils.BindJSON(c, &req) {
		return
	}

	response, err := h.authService.RefreshToken(req.RefreshToken)
	if err != nil {
		utils.RespondWithError(c, http.StatusUnauthorized, "token refresh failed", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, response)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		utils.RespondWithError(c, http.StatusUnauthorized, "user not authenticated", "")
		return
	}

	if err := h.authService.Logout(userID); err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "logout failed", err.Error())
		return
	}

	utils.RespondWithMessage(c, http.StatusOK, "logged out successfully")
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, _ := middleware.GetUserID(c)
	orgID, _ := middleware.GetOrgID(c)
	role, _ := middleware.GetRole(c)

	utils.RespondWithSuccess(c, http.StatusOK, gin.H{
		"user_id": userID,
		"org_id":  orgID,
		"role":    role,
	})
}
