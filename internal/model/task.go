package model

import (
	"time"

	"gorm.io/gorm"
)

// Task represents a user's task.
type Task struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"not null" json:"title"`
	Description string         `json:"description"`
	DueDate     *time.Time     `json:"due_date,omitempty"`
	Completed   bool           `gorm:"default:false" json:"completed"`
	UserID      uint           `gorm:"index;not null" json:"user_id"`
	User        User           `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides the table name.
func (Task) TableName() string {
	return "tasks"
}
