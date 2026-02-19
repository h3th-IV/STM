package model

import (
	"time"

	"gorm.io/gorm"
)

// RefreshToken stores refresh tokens for JWT rotation.
type RefreshToken struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Token     string         `gorm:"uniqueIndex;not null" json:"-"`
	UserID    uint           `gorm:"index;not null" json:"user_id"`
	ExpiresAt time.Time      `gorm:"not null" json:"expires_at"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// TableName overrides the table name.
func (RefreshToken) TableName() string {
	return "refresh_tokens"
}
