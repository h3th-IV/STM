package repository

import (
	"github.com/heth/STM/internal/model"
	"gorm.io/gorm"
)

// TaskRepository handles task database operations.
type TaskRepository struct {
	db *gorm.DB
}

// NewTaskRepository creates a new TaskRepository.
func NewTaskRepository(db *gorm.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

// Create creates a new task.
func (r *TaskRepository) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

// GetByID fetches a task by ID.
func (r *TaskRepository) GetByID(id uint) (*model.Task, error) {
	var task model.Task
	err := r.db.First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// GetByUserID fetches all tasks for a user.
func (r *TaskRepository) GetByUserID(userID uint) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

// Update updates a task.
func (r *TaskRepository) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

// Delete soft-deletes a task.
func (r *TaskRepository) Delete(id uint) error {
	return r.db.Delete(&model.Task{}, id).Error
}

// HardDelete permanently deletes a task (for admin).
func (r *TaskRepository) HardDelete(id uint) error {
	return r.db.Unscoped().Delete(&model.Task{}, id).Error
}
