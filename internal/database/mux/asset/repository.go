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

// Package asset provides repository-level operations for mux asset models.
package asset

import (
	"context"

	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	"gorm.io/gorm"
)

// Repository defines the interface for mux asset data operations.
type Repository interface {
	// --- Only not soft-deleted ---

	// Get retrieves a single asset record from the database.
	Get(ctx context.Context, id string) (*assetmodel.Asset, error)
	// GetByUploadID retrieves a single asset record from the database by it's MuxUploadID.
	GetByUploadID(ctx context.Context, uploadID string) (*assetmodel.Asset, error)
	// GetByAssetID retrieves a single asset record from the database by it's MuxAssetID.
	GetByAssetID(ctx context.Context, assetID string) (*assetmodel.Asset, error)
	// List retrieves all asset records from the database.
	List(ctx context.Context, limit, offset int) ([]assetmodel.Asset, error)
	// ListByIDs retrieves a paginated liat of asset records from the database by their IDs.
	ListByIDs(ctx context.Context, limit, offset int, ids ...string) ([]assetmodel.Asset, error)
	// Count counts the total number of asset records in the database.
	Count(ctx context.Context) (int64, error)

	// --- With soft-deleted ---

	// GetWithDeleted retrieves a single asset record from the database including soft-deleted ones.
	GetWithDeleted(ctx context.Context, id string) (*assetmodel.Asset, error)
	// ListDeleted retrieves all soft-deleted asset records from the database.
	ListDeleted(ctx context.Context, limit, offset int) ([]assetmodel.Asset, error)
	// CountDeleted counts the total number of soft-deleted asset records in the database.
	CountDeleted(ctx context.Context) (int64, error)

	// --- Common ---

	// Create creates a new asset record in the database.
	Create(ctx context.Context, asset *assetmodel.Asset) error
	// RemoveOwner removes local asset association with the owner by setting it's
	// `owner_id` and `owner_type` to nil.
	RemoveOwner(ctx context.Context, id string) (int64, error)
	// SetOwner sets local asset association with the owner by setting it's
	// `owner_id` and `owner_type`.
	SetOwner(ctx context.Context, id string, ownerID, ownerType string) (int64, error)
	// Update performs partial update of asset record in the database using updates.
	Update(ctx context.Context, asset *assetmodel.Asset, updates any) (int64, error)
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

// --- Only not soft-deleted ---

// Get retrieves a single asset record from the database.
func (r *gormRepository) Get(ctx context.Context, id string) (*assetmodel.Asset, error) {
	var asset assetmodel.Asset
	err := r.db.WithContext(ctx).First(&asset, "id = ?", id).Error
	return &asset, err
}

// GetByUploadID retrieves a single asset record from the database by it's MuxUploadID.
func (r *gormRepository) GetByUploadID(ctx context.Context, uploadID string) (*assetmodel.Asset, error) {
	var asset assetmodel.Asset
	err := r.db.WithContext(ctx).First(&asset, "mux_upload_id = ?", uploadID).Error
	return &asset, err
}

// GetByAssetID retrieves a single asset record from the database by it's MuxAssetID.
func (r *gormRepository) GetByAssetID(ctx context.Context, assetID string) (*assetmodel.Asset, error) {
	var asset assetmodel.Asset
	err := r.db.WithContext(ctx).First(&asset, "mux_asset_id = ?", assetID).Error
	return &asset, err
}

// List retrieves all asset records from the database.
func (r *gormRepository) List(ctx context.Context, limit, offset int) ([]assetmodel.Asset, error) {
	var assets []assetmodel.Asset
	err := r.db.WithContext(ctx).Model(&assetmodel.Asset{}).Order("created_at DESC").Limit(limit).Offset(offset).Find(&assets).Error
	return assets, err
}

// ListByIDs retrieves a paginated liat of asset records from the database by their IDs.
func (r *gormRepository) ListByIDs(ctx context.Context, limit, offset int, ids ...string) ([]assetmodel.Asset, error) {
	var assets []assetmodel.Asset
	err := r.db.WithContext(ctx).
		Model(&assetmodel.Asset{}).
		Where("id IN ?", ids).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&assets).Error
	return assets, err
}

// Count counts the total number of asset records in the database.
func (r *gormRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&assetmodel.Asset{}).Count(&count).Error
	return count, err
}

// --- With soft-deleted ---

// GetWithDeleted retrieves a single asset record from the database including soft-deleted ones.
func (r *gormRepository) GetWithDeleted(ctx context.Context, id string) (*assetmodel.Asset, error) {
	var asset assetmodel.Asset
	err := r.db.WithContext(ctx).Unscoped().First(&asset, "id = ?", id).Error
	return &asset, err
}

// ListDeleted retrieves all soft-deleted asset records from the database.
func (r *gormRepository) ListDeleted(ctx context.Context, limit, offset int) ([]assetmodel.Asset, error) {
	var assets []assetmodel.Asset
	err := r.db.WithContext(ctx).
		Unscoped().
		Model(&assetmodel.Asset{}).
		Order("created_at DESC").
		Limit(limit).Offset(offset).
		Find(&assets).Error
	return assets, err
}

// CountDeleted counts the total number of soft-deleted asset records in the database.
func (r *gormRepository) CountDeleted(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Unscoped().
		Model(&assetmodel.Asset{}).
		Where("deleted_at IS NOT NULL").
		Count(&count).Error
	return count, err
}

// --- Common ---

// Create creates a new asset record in the database.
func (r *gormRepository) Create(ctx context.Context, asset *assetmodel.Asset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

// RemoveOwner removes local asset association with the owner by setting it's
// `owner_id` and `owner_type` to nil.
func (r *gormRepository) RemoveOwner(ctx context.Context, id string) (int64, error) {
	res := r.db.WithContext(ctx).Model(&assetmodel.Asset{}).Where("id = ?", id).Update("owner_id", nil).Update("owner_type", nil)
	return res.RowsAffected, res.Error
}

// SetOwner sets local asset association with the owner by setting it's
// `owner_id` and `owner_type`.
func (r *gormRepository) SetOwner(ctx context.Context, id string, ownerID, ownerType string) (int64, error) {
	res := r.db.WithContext(ctx).Model(&assetmodel.Asset{}).Where("id = ?", id).Update("owner_id", ownerID).Update("owner_type", ownerType)
	return res.RowsAffected, res.Error
}

// Update performs partial update of asset record in the database using updates.
func (r *gormRepository) Update(ctx context.Context, asset *assetmodel.Asset, updates any) (int64, error) {
	res := r.db.WithContext(ctx).Model(asset).Updates(updates)
	return res.RowsAffected, res.Error
}

// Delete performs soft-delete of asset record.
func (r *gormRepository) Delete(ctx context.Context, id string) (int64, error) {
	res := r.db.WithContext(ctx).Delete(&assetmodel.Asset{}, id)
	return res.RowsAffected, res.Error
}

// DeletePermanent performs permanent delete of asset record.
func (r *gormRepository) DeletePermanent(ctx context.Context, id string) (int64, error) {
	res := r.db.WithContext(ctx).Unscoped().Delete(&assetmodel.Asset{}, id)
	return res.RowsAffected, res.Error
}

// Restore restores soft-deleted asset record.
func (r *gormRepository) Restore(ctx context.Context, id string) (int64, error) {
	res := r.db.WithContext(ctx).Unscoped().
		Model(&assetmodel.Asset{}).
		Where("id = ?", id).
		Update("deleted_at", nil)
	return res.RowsAffected, res.Error
}
