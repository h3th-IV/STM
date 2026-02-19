package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/heth/STM/internal/middleware"
	"github.com/heth/STM/internal/repository"
)

// UserController handles user HTTP handlers.
type UserController struct {
	userRepo *repository.UserRepository
}

// NewUserController creates a new UserController.
func NewUserController(userRepo *repository.UserRepository) *UserController {
	return &UserController{userRepo: userRepo}
}

// Me godoc
// @Summary Get current user profile
// @Tags users
// @Security BearerAuth
// @Produce json
// @Success 200 {object} model.User
// @Failure 401,404 {object} map[string]string
// @Router /users/me [get]
func (c *UserController) Me(ctx *gin.Context) {
	userID, _ := ctx.Get(middleware.UserIDKey)
	uid := userID.(uint)

	user, err := c.userRepo.GetByID(uid)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	ctx.JSON(http.StatusOK, user)
}
