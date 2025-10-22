// github.com/mikhail5545/media-service-go
// microservice for vitianmove project family
// Copyright (C) 2025  Mikhail Kulik

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package database

import (
	"context"

	"github.com/mikhail5545/media-service-go/internal/models"
	"gorm.io/gorm"
)

type MUXRepository interface {
	// Read operations
	Find(ctx context.Context, id string) (*models.MUXUpload, error)
	FindAll(ctx context.Context) ([]*models.MUXUpload, error)

	// Write operations
	Create(ctx context.Context, muxUpload *models.MUXUpload) error
	Delete(ctx context.Context, id string) error
	DB() *gorm.DB
	WithTx(tx *gorm.DB) MUXRepository
}

type gormMUXRepository struct {
	db *gorm.DB
}

func NewMUXRepository(db *gorm.DB) MUXRepository {
	return &gormMUXRepository{
		db: db,
	}
}

func (r *gormMUXRepository) DB() *gorm.DB {
	return r.db
}

func (r *gormMUXRepository) WithTx(tx *gorm.DB) MUXRepository {
	return &gormMUXRepository{
		db: tx,
	}
}

func (r *gormMUXRepository) Find(ctx context.Context, id string) (*models.MUXUpload, error) {
	var muxUpload models.MUXUpload
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&muxUpload).Error
	if err != nil {
		return nil, err
	}

	return &muxUpload, nil
}

func (r *gormMUXRepository) FindAll(ctx context.Context) ([]*models.MUXUpload, error) {
	var muxUploads []*models.MUXUpload
	err := r.db.WithContext(ctx).Find(&muxUploads).Error
	if err != nil {
		return nil, err
	}

	return muxUploads, nil
}

func (r *gormMUXRepository) Create(ctx context.Context, muxUpload *models.MUXUpload) error {
	return r.db.WithContext(ctx).Create(muxUpload).Error
}

func (r *gormMUXRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.MUXUpload{}).Error
}
