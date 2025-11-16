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

// Package asset provides repository-layer logic for cloudinary asset models.
package asset

import (
	"context"

	"github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	"gorm.io/gorm"
)

// Repository defines the interface for cloudinary asset data operations.
type Repository interface {
	// --- Only not soft-deleted ---

	// Get retrieves a single asset record from the database.
	Get(ctx context.Context, id string) (*asset.Asset, error)
	// List retrieves all asset records from the database.
	List(ctx context.Context, limit, offset int) ([]asset.Asset, error)
	// ListByIDs retrieves a paginated liat of asset records from the database by their IDs.
	ListByIDs(ctx context.Context, limit, offset int, ids ...string) ([]asset.Asset, error)
	// ListAllCloudinaryAssetIDs returns all asset record's cloudinary asset id field value.
	// This method efficiently fetches only the cloudinary_asset_id column from assets table
	// and returns them in a map[string]struct{} for quick, O(1) lookups.
	ListAllCloudinaryAssetIDs(ctx context.Context) (map[string]struct{}, error)
	// Count counts the total number of asset records in the database.
	Count(ctx context.Context) (int64, error)

	// --- With soft-deleted ---

	// GetWithDeleted retrieves a single asset record from the database including soft-deleted ones.
	GetWithDeleted(ctx context.Context, id string) (*asset.Asset, error)
	// GetWithDeletedByAssetID retrieves a single asset record from the database by it's external
	// `CloudinaryAssetID` including soft-deleted ones.
	GetWithDeletedByAssetID(ctx context.Context, assetID string) (*asset.Asset, error)
	// ListSelect returls a list of all assets with specified fields populated.
	ListSelect(ctx context.Context, fields ...string) ([]asset.Asset, error)
	// ListDeleted retrieves all soft-deleted asset records from the database.
	ListDeleted(ctx context.Context, limit, offset int) ([]asset.Asset, error)
	// CountDeleted counts the total number of soft-deleted asset records in the database.
	CountDeleted(ctx context.Context) (int64, error)

	// --- Common ---

	// Create creates a new asset record in the database.
	Create(ctx context.Context, Asset *asset.Asset) error
	// Update performs partial update of asset record in the database using updates.
	Update(ctx context.Context, Asset *asset.Asset, updates any) (int64, error)
	// Delete performs soft-delete of asset record.
	Delete(ctx context.Context, id string) (int64, error)
	// DeletePermanent performs permanent delete of asset record.
	DeletePermanent(ctx context.Context, id string) (int64, error)
	// Restore restores soft-deleted asset record.
	Restore(ctx context.Context, id string) (int64, error)
	// DB returns the underlying gorm.DB instance.
	DB() *gorm.DB
	// WithTx returns a new repository instance with the given transaction.
	WithTx(tx *gorm.DB) Repository
}

// gormRepository holds gorm.DB for GORM-based database operations.
type gormRepository struct {
	db *gorm.DB
}

// New creates a new GORM-based asset repository.
func New(db *gorm.DB) Repository {
	return &gormRepository{
		db: db,
	}
}

// DB returns the underlying gorm.DB instance.
func (r *gormRepository) DB() *gorm.DB {
	return r.db
}

// WithTx returns a new repository instance with the given transaction.
func (r *gormRepository) WithTx(tx *gorm.DB) Repository {
	return &gormRepository{
		db: tx,
	}
}

// --- Only not soft-deleted ---

// Get retrieves a single asset record from the database.
func (r *gormRepository) Get(ctx context.Context, id string) (*asset.Asset, error) {
	var Asset asset.Asset
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&Asset).Error
	return &Asset, err
}

// List retrieves all asset records from the database.
func (r *gormRepository) List(ctx context.Context, limit, offset int) ([]asset.Asset, error) {
	var Assets []asset.Asset
	err := r.db.WithContext(ctx).
		Model(&asset.Asset{}).Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&Assets).Error
	return Assets, err
}

// ListByIDs retrieves a paginated liat of asset records from the database by their IDs.
func (r *gormRepository) ListByIDs(ctx context.Context, limit, offset int, ids ...string) ([]asset.Asset, error) {
	var assets []asset.Asset
	err := r.db.WithContext(ctx).
		Model(&asset.Asset{}).
		Where("id IN ?", ids).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&assets).Error
	return assets, err
}

// ListSelect returls a list of all assets with specified fields populated.
func (r *gormRepository) ListSelect(ctx context.Context, fields ...string) ([]asset.Asset, error) {
	var assets []asset.Asset
	err := r.db.WithContext(ctx).
		Model(&asset.Asset{}).
		Select(fields).Find(&assets).Error
	return assets, err
}

// ListAllCloudinaryAssetIDs returns all asset record's cloudinary asset id field value.
// This method efficiently fetches only the cloudinary_asset_id column from assets table
// and returns them in a map[string]struct{} for quick, O(1) lookups.
func (r *gormRepository) ListAllCloudinaryAssetIDs(ctx context.Context) (map[string]struct{}, error) {
	var ids []string
	err := r.db.WithContext(ctx).
		Model(&asset.Asset{}).
		Select("cloudinary_asset_id").
		Find(&ids).Error

	res := make(map[string]struct{})
	for _, id := range ids {
		res[id] = struct{}{}
	}

	return res, err
}

// Count counts the total number of asset records in the database.
func (r *gormRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&asset.Asset{}).Count(&count).Error
	return count, err
}

// --- With soft-deleted ---

// GetWithDeleted retrieves a single asset record from the database including soft-deleted ones.
func (r *gormRepository) GetWithDeleted(ctx context.Context, id string) (*asset.Asset, error) {
	var Asset asset.Asset
	err := r.db.WithContext(ctx).Unscoped().First(&Asset, "id = ?", id).Error
	return &Asset, err
}

// GetWithDeletedByAssetID retrieves a single asset record from the database by it's external
// `CloudinaryAssetID` including soft-deleted ones.
func (r *gormRepository) GetWithDeletedByAssetID(ctx context.Context, assetID string) (*asset.Asset, error) {
	var asset asset.Asset
	err := r.db.WithContext(ctx).Unscoped().First(&asset, "cloudinary_asset_id = ?", assetID).Error
	return &asset, err
}

// ListDeleted retrieves all soft-deleted asset records from the database.
func (r *gormRepository) ListDeleted(ctx context.Context, limit, offset int) ([]asset.Asset, error) {
	var Assets []asset.Asset
	err := r.db.WithContext(ctx).Unscoped().
		Where("deleted_at IS NOT NULL").
		Limit(limit).Offset(offset).
		Order("created_at DESC").
		Find(&Assets).Error
	return Assets, err
}

// CountDeleted counts the total number of soft-deleted asset records in the database.
func (r *gormRepository) CountDeleted(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&asset.Asset{}).Where("deleted_at IS NOT NULL").Count(&count).Error
	return count, err
}

// --- Common ---

// Create creates a new asset record in the database.
func (r *gormRepository) Create(ctx context.Context, Asset *asset.Asset) error {
	return r.db.WithContext(ctx).Create(Asset).Error
}

// Update performs partial update of asset record in the database using updates.
func (r *gormRepository) Update(ctx context.Context, Asset *asset.Asset, updates any) (int64, error) {
	res := r.db.WithContext(ctx).Model(Asset).Updates(updates)
	return res.RowsAffected, res.Error
}

// Delete performs soft-delete of asset record.
func (r *gormRepository) Delete(ctx context.Context, id string) (int64, error) {
	res := r.db.WithContext(ctx).Delete(&asset.Asset{}, id)
	return res.RowsAffected, res.Error
}

// DeletePermanent performs permanent delete of asset record.
func (r *gormRepository) DeletePermanent(ctx context.Context, id string) (int64, error) {
	res := r.db.WithContext(ctx).Unscoped().Delete(&asset.Asset{}, id)
	return res.RowsAffected, res.Error
}

// Restore restores soft-deleted asset record.
func (r *gormRepository) Restore(ctx context.Context, id string) (int64, error) {
	res := r.db.WithContext(ctx).Unscoped().
		Model(&asset.Asset{}).
		Where("id = ?", id).
		Update("deleted_at", nil)
	return res.RowsAffected, res.Error
}
