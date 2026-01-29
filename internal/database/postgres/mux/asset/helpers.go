package asset

import (
	"context"
	"fmt"

	"github.com/mikhail5545/media-service-go/internal/database/postgres/pagination"
	"github.com/mikhail5545/media-service-go/internal/database/types"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	muxassetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	"gorm.io/gorm"
)

func (r *Repository) get(ctx context.Context, filter *Filter) (*muxassetmodel.Asset, error) {
	cleanFilter(filter)
	if filter == nil {
		return nil, nil
	}
	if !hasIdentifyingFilters(filter) {
		return nil, fmt.Errorf("filter does not contain identifying fields")
	}
	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	var asset muxassetmodel.Asset

	db := r.db.WithContext(ctx)
	db = applyStatusFilters(db, filter.Statuses)

	if len(filter.Fields) > 0 {
		db = db.Select(filter.Fields)
	}

	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)

	err := db.First(&asset).Error
	return &asset, err
}

func (r *Repository) list(ctx context.Context, filter *Filter) ([]*muxassetmodel.Asset, string, error) {
	cleanFilter(filter)
	if filter == nil {
		return nil, "", nil
	}
	if err := filter.Validate(); err != nil {
		return nil, "", fmt.Errorf("invalid filter: %w", err)
	}

	var assets []*muxassetmodel.Asset
	db := r.db.WithContext(ctx)
	db = applyStatusFilters(db, filter.Statuses)

	if len(filter.Fields) > 0 {
		db = db.Select(filter.Fields)
	}

	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)

	db, err := pagination.ApplyCursor(db, pagination.ApplyCursorParams{
		PageSize:   filter.PageSize,
		PageToken:  filter.PageToken,
		OrderField: string(filter.OrderBy),
		OrderDir:   string(filter.OrderDir),
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to apply pagination: %w", err)
	}

	if err := db.Find(&assets).Error; err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(assets) == filter.PageSize+1 {
		last := assets[filter.PageSize-1]
		cursorVal := getCursorValue(last, filter.OrderBy)
		nextToken = pagination.EncodePageToken(cursorVal, last.ID)
		assets = assets[:filter.PageSize]
	}
	return assets, nextToken, nil
}

func (r *Repository) listAll(ctx context.Context, filter *Filter) ([]*muxassetmodel.Asset, error) {
	cleanFilter(filter)
	if filter == nil {
		return nil, nil
	}

	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	var assets []*muxassetmodel.Asset
	db := r.db.WithContext(ctx)
	db = applyStatusFilters(db, filter.Statuses)

	if len(filter.Fields) > 0 {
		db = db.Select(filter.Fields)
	}

	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)
	db = applyOrdering(db, filter)

	err := db.Find(&assets).Error
	return assets, err
}

func (r *Repository) update(ctx context.Context, filter *Filter, updates map[string]any) (int64, error) {
	cleanFilter(filter)
	if filter == nil {
		return 0, nil
	}
	if !hasIdentifyingFilters(filter) {
		return 0, fmt.Errorf("filter does not contain identifying fields")
	}
	if err := filter.Validate(); err != nil {
		return 0, fmt.Errorf("invalid filter: %w", err)
	}

	db := r.db.WithContext(ctx).Model(&muxassetmodel.Asset{})
	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)

	// check if updates contains "status" field
	if _, ok := updates["status"]; ok {
		return 0, fmt.Errorf("cannot update status field using update method")
	}

	res := db.Updates(updates)
	return res.RowsAffected, res.Error
}

func (r *Repository) restore(ctx context.Context, filter *Filter, opts *types.AuditTrailOptions) (int64, error) {
	cleanFilter(filter)
	if filter == nil {
		return 0, nil
	}
	if !hasIdentifyingFilters(filter) {
		return 0, fmt.Errorf("filter does not contain identifying fields")
	}
	if err := filter.Validate(); err != nil {
		return 0, fmt.Errorf("invalid filter: %w", err)
	}
	if err := opts.Validate(); err != nil {
		return 0, fmt.Errorf("invalid audit trail options: %w", err)
	}

	db := r.db.WithContext(ctx).Unscoped().Model(&muxassetmodel.Asset{})
	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)

	db = db.Where("deleted_at IS NOT NULL AND status = ?", muxassetmodel.StatusArchived) // only restore archived records
	updates := restoreUpdates(opts)
	res := db.Updates(updates)

	return res.RowsAffected, res.Error
}

func (r *Repository) archive(ctx context.Context, filter *Filter, opts *types.AuditTrailOptions) (int64, error) {
	cleanFilter(filter)
	if filter == nil {
		return 0, nil
	}
	if !hasIdentifyingFilters(filter) {
		return 0, fmt.Errorf("filter does not contain identifying fields")
	}
	if err := filter.Validate(); err != nil {
		return 0, fmt.Errorf("invalid filter: %w", err)
	}
	if err := opts.Validate(); err != nil {
		return 0, fmt.Errorf("invalid audit trail options: %w", err)
	}

	var rowsAffected int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		db := tx.Model(&muxassetmodel.Asset{})
		db = applyIdentifyingFilters(db, filter)
		db = applySpecificFilters(db, filter)

		db = db.Where("status <> ?", muxassetmodel.StatusArchived) // only archive non-archived records

		updates := archiveUpdates(opts)
		res := db.Updates(updates)
		if res.Error != nil {
			return res.Error
		}

		// Re-apply filters for the delete operation
		// Note: We must re-apply because GORM methods mutate the builder or return a new one depending on usage.
		// It is safer to rebuild the query chain on the 'tx' instance.
		db = tx.Model(&muxassetmodel.Asset{})
		db = applyIdentifyingFilters(db, filter)
		db = applySpecificFilters(db, filter)

		res = db.Delete(&muxassetmodel.Asset{})
		if res.Error != nil {
			return res.Error
		}
		rowsAffected = res.RowsAffected
		return nil
	})
	return rowsAffected, err
}

func (r *Repository) delete(ctx context.Context, filter *Filter) (int64, error) {
	cleanFilter(filter)
	if filter == nil {
		return 0, nil
	}

	if !hasIdentifyingFilters(filter) {
		return 0, fmt.Errorf("filter does not contain identifying fields")
	}
	if err := filter.Validate(); err != nil {
		return 0, fmt.Errorf("invalid filter: %w", err)
	}

	db := r.db.WithContext(ctx).Unscoped().Model(&muxassetmodel.Asset{})
	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)

	db = db.Where("deleted_at IS NOT NULL AND status = ?", assetmodel.StatusArchived) // only delete soft-deleted records

	res := db.Delete(&muxassetmodel.Asset{})
	return res.RowsAffected, res.Error
}

func (r *Repository) markAsBroken(ctx context.Context, filter *Filter, auditOpts *types.AuditTrailOptions) (int64, error) {
	cleanFilter(filter)
	if !hasIdentifyingFilters(filter) {
		return 0, fmt.Errorf("filter does not contain identifying fields")
	}
	if err := filter.Validate(); err != nil {
		return 0, fmt.Errorf("invalid filter: %w", err)
	}
	if err := auditOpts.Validate(); err != nil {
		return 0, fmt.Errorf("invalid audit trail options: %w", err)
	}

	db := r.db.WithContext(ctx).Unscoped().Model(&muxassetmodel.Asset{})
	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)

	db = db.Where("status <> ?", muxassetmodel.StatusBroken) // only mark non-broken records

	updates := markAsBrokenUpdates(auditOpts)
	res := db.Updates(updates)
	return res.RowsAffected, res.Error
}
