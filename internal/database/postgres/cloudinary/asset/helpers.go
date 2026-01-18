package asset

import (
	"context"
	"fmt"

	"github.com/mikhail5545/media-service-go/internal/database/postgres/pagination"
	"github.com/mikhail5545/media-service-go/internal/database/types"
	cldassetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	"gorm.io/gorm"
)

func (r *Repository) get(ctx context.Context, filter *Filter) (*cldassetmodel.Asset, error) {
	cleanFilter(filter)
	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}
	if !hasIdentifyingFilters(filter) {
		return nil, fmt.Errorf("at least one identifying filter must be provided")
	}

	db := r.db.WithContext(ctx)
	db = applyStatusFilter(db, filter.Statuses)

	if len(filter.Fields) > 0 {
		db = db.Select(filter.Fields)
	}

	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)

	var asset cldassetmodel.Asset
	err := db.First(&asset).Error
	return &asset, err
}

func (r *Repository) list(ctx context.Context, filter *Filter) ([]*cldassetmodel.Asset, string, error) {
	cleanFilter(filter)
	if err := filter.Validate(); err != nil {
		return nil, "", fmt.Errorf("invalid filter: %w", err)
	}

	db := r.db.WithContext(ctx)
	db = applyStatusFilter(db, filter.Statuses)

	if len(filter.Fields) > 0 {
		db = db.Select(filter.Fields)
	}

	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)

	db, err := pagination.ApplyCursor(db, pagination.ApplyCursorParams{
		PageSize:   filter.PageSize,
		PageToken:  filter.PageToken,
		OrderField: string(filter.OrderField),
		OrderDir:   string(filter.OrderDir),
	})
	if err != nil {
		return nil, "", err
	}

	var assets []*cldassetmodel.Asset
	if err := db.Find(&assets).Error; err != nil {
		return nil, "", err
	}

	var nextToken string
	if len(assets) == filter.PageSize+1 {
		last := assets[filter.PageSize-1]
		cursorVal := getCursorValue(last, filter.OrderField)
		nextToken = pagination.EncodePageToken(cursorVal, last.ID)
		assets = assets[:filter.PageSize]
	}
	return assets, nextToken, nil
}

func (r *Repository) listAll(ctx context.Context, filter *Filter) ([]*cldassetmodel.Asset, error) {
	cleanFilter(filter)
	if err := filter.Validate(); err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	db := r.db.WithContext(ctx)
	db = applyStatusFilter(db, filter.Statuses)

	if len(filter.Fields) > 0 {
		db = db.Select(filter.Fields)
	}

	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)
	db = applyOrdering(db, filter)

	var assets []*cldassetmodel.Asset
	err := db.Find(&assets).Error
	return assets, err
}

func (r *Repository) restore(ctx context.Context, filter *Filter, opts *types.AuditTrailOptions) (int64, error) {
	cleanFilter(filter)
	if err := filter.Validate(); err != nil {
		return 0, fmt.Errorf("invalid filter: %w", err)
	}
	if err := opts.Validate(); err != nil {
		return 0, fmt.Errorf("invalid audit trail options: %w", err)
	}
	if !hasIdentifyingFilters(filter) {
		return 0, fmt.Errorf("at least one identifying filter must be provided")
	}

	db := r.db.WithContext(ctx).Model(&cldassetmodel.Asset{})
	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)

	db = db.Where("deleted_at IS NOT NULL AND status = ?", cldassetmodel.StatusArchived)

	updates := restoreUpdates(opts)
	res := db.Updates(updates)
	return res.RowsAffected, res.Error
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

	db := r.db.WithContext(ctx).Model(&cldassetmodel.Asset{})
	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)

	// check if updates contains "status" field
	if _, ok := updates["status"]; ok {
		return 0, fmt.Errorf("cannot update status field using update method")
	}

	res := db.Updates(updates)
	return res.RowsAffected, res.Error
}

func (r *Repository) archive(ctx context.Context, filter *Filter, opts *types.AuditTrailOptions) (int64, error) {
	cleanFilter(filter)
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
		db := tx.Model(&cldassetmodel.Asset{})
		db = applyIdentifyingFilters(db, filter)
		db = applySpecificFilters(db, filter)

		db = db.Where("deleted_at IS NULL AND status <> ?", cldassetmodel.StatusArchived) // only archive non-archived records
		updates := archiveUpdates(opts)
		res := db.Updates(updates)
		if res.Error != nil {
			return res.Error
		}

		db = tx.Model(&cldassetmodel.Asset{})
		db = applyIdentifyingFilters(db, filter)
		db = applySpecificFilters(db, filter)

		res = db.Delete(&cldassetmodel.Asset{})
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
	if !hasIdentifyingFilters(filter) {
		return 0, fmt.Errorf("filter does not contain identifying fields")
	}
	if err := filter.Validate(); err != nil {
		return 0, fmt.Errorf("invalid filter: %w", err)
	}

	db := r.db.WithContext(ctx).Unscoped().Model(&cldassetmodel.Asset{})
	db = applyIdentifyingFilters(db, filter)
	db = applySpecificFilters(db, filter)

	db = db.Where("deleted_at IS NOT NULL AND status = ?", cldassetmodel.StatusArchived) // only delete archived records

	res := db.Delete(&cldassetmodel.Asset{})
	return res.RowsAffected, res.Error
}
