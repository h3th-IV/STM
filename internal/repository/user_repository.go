package repository

import (
	"github.com/heth/STM/internal/model"
	"gorm.io/gorm"
)

// UserRepository handles user database operations.
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new UserRepository.
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user.
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByID fetches a user by ID.
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail fetches a user by email.
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByEmail checks if a user exists with the given email.
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// ExistsByUsername checks if a user exists with the given username.
func (r *UserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}
