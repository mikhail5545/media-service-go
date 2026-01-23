package asset

import (
	"slices"

	"github.com/mikhail5545/media-service-go/internal/database/types"
	cldassetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	"github.com/mikhail5545/media-service-go/internal/util/parsing"
	"gorm.io/gorm"
)

func cleanFilter(filter *Filter) {
	if filter == nil {
		return
	}

	filter.IDs = parsing.CleanIDs(filter.IDs)
	filter.Formats = parsing.CleanStrings(filter.Formats)
	filter.ResourceTypes = parsing.CleanStrings(filter.ResourceTypes)
	filter.CloudinaryPublicIDs = parsing.CleanStrings(filter.CloudinaryPublicIDs)
	filter.CloudinaryAssetIDs = parsing.CleanStrings(filter.CloudinaryAssetIDs)
}

func extractScopes(scopes []Scope) []cldassetmodel.Status {
	var statuses []cldassetmodel.Status
	if len(scopes) > 0 {
		if slices.Contains(scopes, ScopeAll) {
			statuses = []cldassetmodel.Status{
				cldassetmodel.StatusUploadURLGenerated,
				cldassetmodel.StatusActive,
				cldassetmodel.StatusArchived,
				cldassetmodel.StatusBroken,
			}
		}
		statuses = make([]cldassetmodel.Status, 0, len(scopes))
		if slices.Contains(scopes, ScopeUploadURLGenerated) {
			statuses = append(statuses, cldassetmodel.StatusUploadURLGenerated)
		}
		if slices.Contains(scopes, ScopeActive) {
			statuses = append(statuses, cldassetmodel.StatusActive)
		}
		if slices.Contains(scopes, ScopeArchived) {
			statuses = append(statuses, cldassetmodel.StatusArchived)
		}
		if slices.Contains(scopes, ScopeBroken) {
			statuses = append(statuses, cldassetmodel.StatusBroken)
		}
	} else {
		statuses = []cldassetmodel.Status{cldassetmodel.StatusActive} // Only active by default
	}
	return statuses
}

func getCursorValue(asset *cldassetmodel.Asset, orderBy cldassetmodel.OrderField) any {
	switch orderBy {
	case cldassetmodel.OrderResourceType:
		return asset.ResourceType
	case cldassetmodel.OrderFormat:
		return asset.Format
	case cldassetmodel.OrderUpdatedAt:
		return asset.UpdatedAt
	case cldassetmodel.OrderCreatedAt:
		fallthrough
	default:
		return asset.CreatedAt
	}
}

func applyStatusFilter(db *gorm.DB, statuses []cldassetmodel.Status) *gorm.DB {
	if len(statuses) == 0 {
		return db
	}
	if slices.Contains(statuses, cldassetmodel.StatusArchived) {
		db = db.Unscoped()
	}
	return db.Model(&cldassetmodel.Asset{}).Where("status IN ?", statuses)
}

func applyIdentifyingFilters(db *gorm.DB, filter *Filter) *gorm.DB {
	if len(filter.IDs) > 0 {
		db = db.Where("id IN ?", filter.IDs)
	}
	if len(filter.CloudinaryAssetIDs) > 0 {
		db = db.Where("cloudinary_asset_id IN ?", filter.CloudinaryAssetIDs)
	}
	if len(filter.CloudinaryPublicIDs) > 0 {
		db = db.Where("cloudinary_public_id IN ?", filter.CloudinaryPublicIDs)
	}
	return db
}

func applySpecificFilters(db *gorm.DB, filter *Filter) *gorm.DB {
	if len(filter.ResourceTypes) > 0 {
		db = db.Where("resource_type IN ?", filter.ResourceTypes)
	}
	if len(filter.Formats) > 0 {
		db = db.Where("format IN ?", filter.Formats)
	}
	return db
}

func applyOrdering(db *gorm.DB, filter *Filter) *gorm.DB {
	orderField := string(cldassetmodel.OrderCreatedAt)
	if filter.OrderField != "" {
		orderField = string(filter.OrderField)
	}

	orderDirection := "DESC"
	if filter.OrderDir == cldassetmodel.OrderAscending {
		orderDirection = "ASC"
	}

	db = db.Order(orderField + " " + orderDirection)
	return db
}

func hasIdentifyingFilters(filter *Filter) bool {
	if len(filter.IDs) > 0 {
		return true
	}
	if len(filter.CloudinaryAssetIDs) > 0 {
		return true
	}
	if len(filter.CloudinaryPublicIDs) > 0 {
		return true
	}
	return false
}

func restoreUpdates(opts *types.AuditTrailOptions) map[string]any {
	return map[string]any{
		"status":           cldassetmodel.StatusActive,
		"deleted_at":       nil,
		"restored_by":      opts.AdminID,
		"restored_by_name": opts.AdminName,
		"note":             opts.Note,
	}
}

func archiveUpdates(opts *types.AuditTrailOptions) map[string]any {
	return map[string]any{
		"status":           cldassetmodel.StatusArchived,
		"archived_by":      opts.AdminID,
		"archived_by_name": opts.AdminName,
		"archive_reason":   opts.Note,
	}
}

func populateFromStateOperationOptions(opts *StateOperationOptions) *Filter {
	return &Filter{
		IDs:                 opts.IDs,
		CloudinaryAssetIDs:  opts.CloudinaryAssetIDs,
		CloudinaryPublicIDs: opts.CloudinaryPublicIDs,
		ResourceTypes:       opts.ResourceTypes,
		Formats:             opts.Formats,
	}
}

func populateFromListOptions(opts *ListOptions, scopes []Scope) *Filter {
	return &Filter{
		IDs:                 opts.IDs,
		CloudinaryAssetIDs:  opts.CloudinaryAssetIDs,
		CloudinaryPublicIDs: opts.CloudinaryPublicIDs,
		ResourceTypes:       opts.ResourceTypes,
		Formats:             opts.Formats,
		Fields:              opts.Fields,
		Statuses:            extractScopes(scopes),
		OrderDir:            opts.OrderDir,
		OrderField:          opts.OrderField,
		PageSize:            opts.PageSize,
		PageToken:           opts.PageToken,
	}
}
