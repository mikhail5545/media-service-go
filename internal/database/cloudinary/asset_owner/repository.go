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

// Package asset provides repository-layer logic for cloudinary asset owner models.
package assetowner

import (
	"context"

	assetownermodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset_owner"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository defines the interface for cloudinary asset owner data operations.
type Repository interface {
	// CreateBatch creates multiple asset owner records, ignoring any that already exist.
	CreateBatch(ctx context.Context, owners []assetownermodel.AssetOwner) error
	// DeleteByOwnerTypeAndIDs deletes owner links for a specific asset and owner type.
	DeleteByOwnerTypeAndIDs(ctx context.Context, assetID, ownerType string, ownerIDs []string) (int64, error)
	// ListByAssetID retrieves all owner records for a given asset.
	ListByAssetID(ctx context.Context, assetID string) ([]assetownermodel.AssetOwner, error)
	// WithTx returns a new repository instance with the given transaction.
	WithTx(tx *gorm.DB) Repository
}

// gormRepository holds gorm.DB for GORM-based database operations.
type gormRepository struct {
	db *gorm.DB
}

// New creates a new GORM-based asset owner repository.
func New(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// WithTx returns a new repository instance with the given transaction.
func (r *gormRepository) WithTx(tx *gorm.DB) Repository {
	return &gormRepository{db: tx}
}

// CreateBatch creates multiple asset owner records, ignoring any that already exist.
func (r *gormRepository) CreateBatch(ctx context.Context, owners []assetownermodel.AssetOwner) error {
	if len(owners) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&owners).Error
}

// DeleteByOwnerTypeAndIDs deletes owner links for a specific asset and owner type.
func (r *gormRepository) DeleteByOwnerTypeAndIDs(ctx context.Context, assetID, ownerType string, ownerIDs []string) (int64, error) {
	if len(ownerIDs) == 0 {
		return 0, nil
	}
	res := r.db.WithContext(ctx).Where("asset_id = ? AND owner_type = ? AND owner_id IN ?", assetID, ownerType, ownerIDs).Delete(&assetownermodel.AssetOwner{})
	return res.RowsAffected, res.Error
}

// ListByAssetID retrieves all owner records for a given asset.
func (r *gormRepository) ListByAssetID(ctx context.Context, assetID string) ([]assetownermodel.AssetOwner, error) {
	var owners []assetownermodel.AssetOwner
	err := r.db.WithContext(ctx).Where("asset_id = ?", assetID).Find(&owners).Error
	return owners, err
}
