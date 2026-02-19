package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/heth/STM/internal/service"
	"github.com/heth/STM/internal/utils"
)

// AuthController handles auth HTTP handlers.
type AuthController struct {
	authService *service.AuthService
}

// NewAuthController creates a new AuthController.
func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

// Register godoc
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body service.RegisterRequest true "Registration data"
// @Success 201 {object} service.AuthResponse
// @Failure 400,409,500 {object} map[string]string
// @Router /auth/register [post]
func (c *AuthController) Register(ctx *gin.Context) {
	var req service.RegisterRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.authService.Register(&req)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body service.LoginRequest true "Login credentials"
// @Success 200 {object} service.AuthResponse
// @Failure 401,500 {object} map[string]string
// @Router /auth/login [post]
func (c *AuthController) Login(ctx *gin.Context) {
	var req service.LoginRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.authService.Login(&req)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

// Refresh godoc
// @Summary Refresh tokens
// @Tags auth
// @Accept json
// @Produce json
// @Param body body object{refresh_token=string} true "Refresh token"
// @Success 200 {object} service.AuthResponse
// @Failure 401 {object} map[string]string
// @Router /auth/refresh [post]
func (c *AuthController) Refresh(ctx *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := c.authService.Refresh(req.RefreshToken)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, resp)
}

func respondError(ctx *gin.Context, err error) {
	if appErr, ok := err.(*utils.AppError); ok {
		ctx.JSON(appErr.Code, gin.H{"error": appErr.Message})
		return
	}
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": utils.ErrInternalServerError.Message})
}
