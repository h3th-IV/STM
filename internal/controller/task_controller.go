package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/heth/STM/internal/middleware"
	"github.com/heth/STM/internal/service"
)

// TaskController handles task HTTP handlers.
type TaskController struct {
	taskService *service.TaskService
}

// NewTaskController creates a new TaskController.
func NewTaskController(taskService *service.TaskService) *TaskController {
	return &TaskController{taskService: taskService}
}

// List godoc
// @Summary List my tasks
// @Tags tasks
// @Security BearerAuth
// @Produce json
// @Success 200 {array} model.Task
// @Failure 401 {object} map[string]string
// @Router /tasks [get]
func (c *TaskController) List(ctx *gin.Context) {
	userID := getUserID(ctx)

	tasks, err := c.taskService.ListByUser(userID)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, tasks)
}

// Create godoc
// @Summary Create a task
// @Tags tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body service.CreateTaskRequest true "Task data"
// @Success 201 {object} model.Task
// @Failure 400,401,500 {object} map[string]string
// @Router /tasks [post]
func (c *TaskController) Create(ctx *gin.Context) {
	userID := getUserID(ctx)

	var req service.CreateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := c.taskService.Create(userID, &req)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, task)
}

// Get godoc
// @Summary Get a task
// @Tags tasks
// @Security BearerAuth
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} model.Task
// @Failure 401,403,404 {object} map[string]string
// @Router /tasks/{id} [get]
func (c *TaskController) Get(ctx *gin.Context) {
	id, userID, isAdmin := getTaskParams(ctx)
	if id == 0 {
		return
	}

	task, err := c.taskService.GetByID(id, userID, isAdmin)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, task)
}

// Update godoc
// @Summary Update a task
// @Tags tasks
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param body body service.UpdateTaskRequest true "Task updates"
// @Success 200 {object} model.Task
// @Failure 400,401,403,404 {object} map[string]string
// @Router /tasks/{id} [put]
func (c *TaskController) Update(ctx *gin.Context) {
	id, userID, _ := getTaskParams(ctx)
	if id == 0 {
		return
	}

	var req service.UpdateTaskRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := c.taskService.Update(id, userID, &req)
	if err != nil {
		respondError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, task)
}

// Delete godoc
// @Summary Delete a task
// @Tags tasks
// @Security BearerAuth
// @Param id path int true "Task ID"
// @Success 204
// @Failure 401,403,404 {object} map[string]string
// @Router /tasks/{id} [delete]
func (c *TaskController) Delete(ctx *gin.Context) {
	id, userID, isAdmin := getTaskParams(ctx)
	if id == 0 {
		return
	}

	if err := c.taskService.Delete(id, userID, isAdmin); err != nil {
		respondError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func getUserID(ctx *gin.Context) uint {
	userID, _ := ctx.Get(middleware.UserIDKey)
	return userID.(uint)
}

func getIsAdmin(ctx *gin.Context) bool {
	role, _ := ctx.Get(middleware.RoleKey)
	return role == "admin"
}

func getTaskParams(ctx *gin.Context) (id uint, userID uint, isAdmin bool) {
	idStr := ctx.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return 0, 0, false
	}
	return uint(id64), getUserID(ctx), getIsAdmin(ctx)
}
