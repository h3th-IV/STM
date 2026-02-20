package service

import (
	"fmt"
	"strings"
	"time"

	"github.com/heth/STM/internal/model"
	"github.com/heth/STM/internal/repository"
	"github.com/heth/STM/internal/utils"
	"github.com/heth/STM/proto"
	"gorm.io/gorm"
)

// TaskService handles task business logic.
type TaskService struct {
	taskRepo   *repository.TaskRepository
	taskNotify TaskNotifier // optional: broadcasts task events for real-time subscribers
}

// NewTaskService creates a new TaskService.
func NewTaskService(taskRepo *repository.TaskRepository) *TaskService {
	return &TaskService{taskRepo: taskRepo}
}

// SetTaskNotifier sets the optional notifier for real-time task events.
func (s *TaskService) SetTaskNotifier(n TaskNotifier) {
	s.taskNotify = n
}

// CreateTaskRequest for creating a task.
type CreateTaskRequest struct {
	Title       string `json:"title" binding:"required,max=255"`
	Description string `json:"description" binding:"max=2000"`
	DueDate     *string `json:"due_date,omitempty"` // RFC3339 format
}

// UpdateTaskRequest for updating a task.
type UpdateTaskRequest struct {
	Title       *string `json:"title,omitempty" binding:"omitempty,max=255"`
	Description *string `json:"description,omitempty" binding:"omitempty,max=2000"`
	DueDate     *string `json:"due_date,omitempty"`
	Completed   *bool   `json:"completed,omitempty"`
}

// Create creates a new task for the user.
func (s *TaskService) Create(userID uint, req *CreateTaskRequest) (*model.Task, error) {
	task := &model.Task{
		Title:       strings.TrimSpace(req.Title),
		Description: strings.TrimSpace(req.Description),
		UserID:      userID,
	}
	if req.DueDate != nil {
		if t, err := parseDate(*req.DueDate); err == nil {
			task.DueDate = &t
		}
	}
	if err := s.taskRepo.Create(task); err != nil {
		return nil, utils.NewAppError(500, "failed to create task", err)
	}
	if s.taskNotify != nil {
		s.taskNotify.BroadcastTaskEvent(uintToStr(task.UserID), &proto.TaskEvent{
			Type: proto.TaskEvent_CREATE,
			Task: ModelTaskToProto(task),
		})
	}
	return task, nil
}

// GetByID fetches a task. Returns error if not owner and not admin.
func (s *TaskService) GetByID(id uint, userID uint, isAdmin bool) (*model.Task, error) {
	task, err := s.taskRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrNotFound
		}
		return nil, utils.NewAppError(500, "database error", err)
	}
	if task.UserID != userID && !isAdmin {
		return nil, utils.ErrForbidden
	}
	return task, nil
}

// ListByUser fetches all tasks for a user.
func (s *TaskService) ListByUser(userID uint) ([]model.Task, error) {
	return s.taskRepo.GetByUserID(userID)
}

// Update updates a task. Only owner can update.
func (s *TaskService) Update(id uint, userID uint, req *UpdateTaskRequest) (*model.Task, error) {
	task, err := s.taskRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, utils.ErrNotFound
		}
		return nil, utils.NewAppError(500, "database error", err)
	}
	if task.UserID != userID {
		return nil, utils.ErrForbidden
	}

	if req.Title != nil {
		task.Title = strings.TrimSpace(*req.Title)
	}
	if req.Description != nil {
		task.Description = strings.TrimSpace(*req.Description)
	}
	if req.DueDate != nil {
		if t, err := parseDate(*req.DueDate); err == nil {
			task.DueDate = &t
		} else if *req.DueDate == "" {
			task.DueDate = nil
		}
	}
	if req.Completed != nil {
		task.Completed = *req.Completed
	}

	if err := s.taskRepo.Update(task); err != nil {
		return nil, utils.NewAppError(500, "failed to update task", err)
	}
	if s.taskNotify != nil {
		s.taskNotify.BroadcastTaskEvent(uintToStr(task.UserID), &proto.TaskEvent{
			Type: proto.TaskEvent_UPDATE,
			Task: ModelTaskToProto(task),
		})
	}
	return task, nil
}

// Delete deletes a task. Owner or admin can delete.
func (s *TaskService) Delete(id uint, userID uint, isAdmin bool) error {
	task, err := s.taskRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.ErrNotFound
		}
		return utils.NewAppError(500, "database error", err)
	}
	if task.UserID != userID && !isAdmin {
		return utils.ErrForbidden
	}
	if err := s.taskRepo.Delete(id); err != nil {
		return err
	}
	if s.taskNotify != nil {
		s.taskNotify.BroadcastTaskEvent(uintToStr(task.UserID), &proto.TaskEvent{
			Type: proto.TaskEvent_DELETE,
			Task: ModelTaskToProto(task),
		})
	}
	return nil
}

// AdminForceDelete permanently deletes any task (admin only).
func (s *TaskService) AdminForceDelete(id uint) error {
	task, err := s.taskRepo.GetByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return utils.ErrNotFound
		}
		return utils.NewAppError(500, "database error", err)
	}
	if err := s.taskRepo.HardDelete(id); err != nil {
		return err
	}
	if s.taskNotify != nil {
		s.taskNotify.BroadcastTaskEvent(uintToStr(task.UserID), &proto.TaskEvent{
			Type: proto.TaskEvent_DELETE,
			Task: ModelTaskToProto(task),
		})
	}
	return nil
}

func uintToStr(u uint) string {
	return fmt.Sprintf("%d", u)
}

func parseDate(s string) (time.Time, error) {
	return time.Parse(time.RFC3339, strings.TrimSpace(s))
}
