package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/heth/STM/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTaskCRUD(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r, db := setupTestApp(t)

	// Create user and get token
	token := mustRegisterAndLogin(t, r)

	// Create task
	body := `{"title":"Test task","description":"Do something"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/tasks", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	var task model.Task
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &task))
	assert.Equal(t, "Test task", task.Title)
	taskID := task.ID

	// List tasks
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tasks", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Get task
	req = httptest.NewRequest(http.MethodGet, "/api/v1/tasks/"+fmt.Sprintf("%d", taskID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Update task
	body = `{"title":"Updated task","completed":true}`
	req = httptest.NewRequest(http.MethodPut, "/api/v1/tasks/"+fmt.Sprintf("%d", taskID), bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Delete task
	req = httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/"+fmt.Sprintf("%d", taskID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusNoContent, w.Code)

	_ = db // avoid unused variable
}

func mustRegisterAndLogin(t *testing.T, r *gin.Engine) string {
	t.Helper()
	body := `{"email":"crud@example.com","password":"password123","username":"cruduser"}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	token, _ := resp["access_token"].(string)
	require.NotEmpty(t, token)
	return token
}
