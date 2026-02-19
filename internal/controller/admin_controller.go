package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/heth/STM/internal/service"
)

// AdminController handles admin-only HTTP handlers.
type AdminController struct {
	taskService *service.TaskService
}

// NewAdminController creates a new AdminController.
func NewAdminController(taskService *service.TaskService) *AdminController {
	return &AdminController{taskService: taskService}
}

// ForceDeleteTask godoc
// @Summary Admin: Force delete any task
// @Tags admin
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Success 204
// @Failure 401,403,404 {object} map[string]string
// @Router /admin/tasks/{id} [delete]
func (c *AdminController) ForceDeleteTask(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := c.taskService.AdminForceDelete(uint(id64)); err != nil {
		respondError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
