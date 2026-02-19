package repository

import (
	"github.com/heth/STM/internal/model"
	"gorm.io/gorm"
)

// RefreshTokenRepository handles refresh token database operations.
type RefreshTokenRepository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository creates a new RefreshTokenRepository.
func NewRefreshTokenRepository(db *gorm.DB) *RefreshTokenRepository {
	return &RefreshTokenRepository{db: db}
}

// Create stores a new refresh token.
func (r *RefreshTokenRepository) Create(token *model.RefreshToken) error {
	return r.db.Create(token).Error
}

// GetByToken fetches a refresh token by its token string.
func (r *RefreshTokenRepository) GetByToken(token string) (*model.RefreshToken, error) {
	var rt model.RefreshToken
	err := r.db.Where("token = ?", token).First(&rt).Error
	if err != nil {
		return nil, err
	}
	return &rt, nil
}

// DeleteByID removes a refresh token (for rotation).
func (r *RefreshTokenRepository) DeleteByID(id uint) error {
	return r.db.Delete(&model.RefreshToken{}, id).Error
}

// DeleteByToken removes a refresh token by its token string.
func (r *RefreshTokenRepository) DeleteByToken(token string) error {
	return r.db.Where("token = ?", token).Delete(&model.RefreshToken{}).Error
}
