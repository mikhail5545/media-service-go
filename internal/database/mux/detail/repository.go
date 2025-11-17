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

package detail

import (
	"context"

	detailmodel "github.com/mikhail5545/media-service-go/internal/models/mux/detail"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Repository defines the interface for asset detail data operations.
type Repository interface {
	// Get retrieves a single asset detail record.
	Get(ctx context.Context, assetID string) (*detailmodel.AssetDetail, error)
	// ListByAssetIDs retrieves multiple asset detail records by their asset IDs.
	ListByAssetIDs(ctx context.Context, assetIDs ...string) (map[string]*detailmodel.AssetDetail, error)
	// Upsert creates or updates an asset detail record.
	// It uses `clauses.OnConflict` to perform an "upsert" operation,
	// updating the 'tracks' column if the asset_id already exists.
	Upsert(ctx context.Context, details *detailmodel.AssetDetail) error
	// DB returns the underlying gorm.DB instance.
	DB() *gorm.DB
	// WithTx returns a new repository instance with the given transaction.
	WithTx(tx *gorm.DB) Repository
}

type gormRepository struct {
	db *gorm.DB
}

// New creates a new GORM-based asset detail repository.
func New(db *gorm.DB) Repository {
	return &gormRepository{db: db}
}

// DB returns the underlying gorm.DB instance.
func (r *gormRepository) DB() *gorm.DB {
	return r.db
}

// WithTx returns a new repository instance with the given transaction.
func (r *gormRepository) WithTx(tx *gorm.DB) Repository {
	return &gormRepository{db: tx}
}

// Get retrieves a single asset detail record.
func (r *gormRepository) Get(ctx context.Context, assetID string) (*detailmodel.AssetDetail, error) {
	var detail detailmodel.AssetDetail
	err := r.db.WithContext(ctx).First(&detail, "asset_id = ?", assetID).Error
	return &detail, err
}

// ListByAssetIDs retrieves multiple asset detail records and returns them as a map.
func (r *gormRepository) ListByAssetIDs(ctx context.Context, assetIDs ...string) (map[string]*detailmodel.AssetDetail, error) {
	if len(assetIDs) == 0 {
		return make(map[string]*detailmodel.AssetDetail), nil
	}

	var details []detailmodel.AssetDetail
	err := r.db.WithContext(ctx).Where("asset_id IN ?", assetIDs).Find(&details).Error
	if err != nil {
		return nil, err
	}

	detailMap := make(map[string]*detailmodel.AssetDetail)
	for i := range details {
		detailMap[details[i].AssetID] = &details[i]
	}

	return detailMap, nil
}

// Upsert creates or updates an asset detail record.
// It uses `clauses.OnConflict` to perform an "upsert" operation,
// updating the 'tracks' column if the asset_id already exists.
func (r *gormRepository) Upsert(ctx context.Context, detail *detailmodel.AssetDetail) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "asset_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"tracks"}),
	}).Create(detail).Error
}
