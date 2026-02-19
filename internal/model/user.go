package model

import (
	"time"

	"gorm.io/gorm"
)

// Role constants for RBAC.
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User represents a registered user.
type User struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	Username     string         `gorm:"uniqueIndex;not null" json:"username"`
	PasswordHash string         `gorm:"column:password_hash;not null" json:"-"`
	Role         string         `gorm:"default:user" json:"role"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides the table name.
func (User) TableName() string {
	return "users"
}
