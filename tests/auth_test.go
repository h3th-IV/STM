package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/heth/STM/internal/config"
	"github.com/heth/STM/internal/controller"
	"github.com/heth/STM/internal/model"
	"github.com/heth/STM/internal/repository"
	"github.com/heth/STM/internal/router"
	"github.com/heth/STM/internal/service"
	"github.com/heth/STM/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestApp(t *testing.T) (*gin.Engine, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&model.User{}, &model.Task{}, &model.RefreshToken{}))

	cfg := &config.Config{
		Port:          8080,
		JWTSecret:     "test-secret-key-at-least-32-characters-long",
		JWTIssuer:     "test",
		JWTExpiry:     15,
		RefreshExpiry: 7,
		DBPath:        ":memory:",
		Env:           "test",
	}

	userRepo := repository.NewUserRepository(db)
	taskRepo := repository.NewTaskRepository(db)
	refreshTokenRepo := repository.NewRefreshTokenRepository(db)
	jwtService := utils.NewJWTService(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTExpiry, cfg.RefreshExpiry)
	authService := service.NewAuthService(userRepo, refreshTokenRepo, jwtService)
	taskService := service.NewTaskService(taskRepo)

	authCtrl := controller.NewAuthController(authService)
	userCtrl := controller.NewUserController(userRepo)
	taskCtrl := controller.NewTaskController(taskService)
	adminCtrl := controller.NewAdminController(taskService)

	r := router.Setup(cfg, authCtrl, userCtrl, taskCtrl, adminCtrl, jwtService)
	return r, db
}

func TestRegisterLogin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupTestApp(t)

	// Register
	body := `{"email":"test@example.com","password":"password123","username":"testuser"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["access_token"])
	assert.NotEmpty(t, resp["refresh_token"])

	// Login
	body = `{"email":"test@example.com","password":"password123"}`
	req = httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["access_token"])
}

func TestLoginInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupTestApp(t)

	body := `{"email":"wrong@example.com","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestHealth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, _ := setupTestApp(t)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]string
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "up", resp["status"])
}