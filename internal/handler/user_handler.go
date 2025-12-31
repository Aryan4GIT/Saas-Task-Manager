package handler

import (
	"net/http"

	"saas-backend/internal/middleware"
	"saas-backend/internal/models"
	"saas-backend/internal/service"
	"saas-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, _ := middleware.GetUserID(c)

	var req models.CreateUserRequest
	if !utils.BindJSON(c, &req) {
		return
	}

	user, err := h.userService.CreateUser(orgID, userID, &req)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "failed to create user", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusCreated, user)
}

func (h *UserHandler) GetUser(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	userID, ok := utils.ParseUUID(c, "id", "user ID")
	if !ok {
		return
	}

	user, err := h.userService.GetUser(orgID, userID)
	if err != nil {
		utils.RespondWithError(c, http.StatusNotFound, "user not found", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, user)
}

func (h *UserHandler) ListUsers(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)

	users, err := h.userService.ListUsers(orgID)
	if err != nil {
		utils.RespondWithError(c, http.StatusInternalServerError, "failed to list users", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, users)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	currentUserID, _ := middleware.GetUserID(c)
	targetUserID, ok := utils.ParseUUID(c, "id", "user ID")
	if !ok {
		return
	}

	var updates models.User
	if !utils.BindJSON(c, &updates) {
		return
	}

	user, err := h.userService.UpdateUser(orgID, targetUserID, currentUserID, &updates)
	if err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "failed to update user", err.Error())
		return
	}

	utils.RespondWithSuccess(c, http.StatusOK, user)
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	orgID, _ := middleware.GetOrgID(c)
	currentUserID, _ := middleware.GetUserID(c)
	targetUserID, ok := utils.ParseUUID(c, "id", "user ID")
	if !ok {
		return
	}

	if err := h.userService.DeleteUser(orgID, targetUserID, currentUserID); err != nil {
		utils.RespondWithError(c, http.StatusBadRequest, "failed to delete user", err.Error())
		return
	}

	utils.RespondWithMessage(c, http.StatusOK, "user deleted successfully")
}
