/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package asset

import (
	"slices"

	"github.com/mikhail5545/media-service-go/internal/database/types"
	muxassetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	"github.com/mikhail5545/media-service-go/internal/util/parsing"
	"gorm.io/gorm"
)

func extractScopes(scopes []Scope) []muxassetmodel.Status {
	var statuses []muxassetmodel.Status
	if len(scopes) > 0 {
		if slices.Contains(scopes, ScopeAll) {
			statuses = []muxassetmodel.Status{
				muxassetmodel.StatusActive,
				muxassetmodel.StatusUploadURLGenerated,
				muxassetmodel.StatusArchived,
				muxassetmodel.StatusBroken,
			}
		}
		statuses = make([]muxassetmodel.Status, 0, len(scopes))
		if slices.Contains(scopes, ScopeActive) {
			statuses = append(statuses, muxassetmodel.StatusActive)
		}
		if slices.Contains(scopes, ScopeUploadURLGenerated) {
			statuses = append(statuses, muxassetmodel.StatusUploadURLGenerated)
		}
		if slices.Contains(scopes, ScopeArchived) {
			statuses = append(statuses, muxassetmodel.StatusArchived)
		}
		if slices.Contains(scopes, ScopeBroken) {
			statuses = append(statuses, muxassetmodel.StatusBroken)
		}
	} else {
		statuses = []muxassetmodel.Status{muxassetmodel.StatusActive} // Only active by default
	}
	return statuses
}

func getCursorValue(asset *muxassetmodel.Asset, orderBy muxassetmodel.OrderField) any {
	switch orderBy {
	case muxassetmodel.OrderIngestType:
		return asset.IngestType
	case muxassetmodel.OrderUpdatedAt:
		return asset.UpdatedAt
	case muxassetmodel.OrderCreatedAt:
		fallthrough
	default:
		return asset.CreatedAt
	}
}

func cleanFilter(filter *Filter) {
	if filter == nil {
		return
	}
	filter.IDs = parsing.CleanIDs(filter.IDs)
	filter.MuxAssetIDs = parsing.CleanStrings(filter.MuxAssetIDs)
	filter.MuxUploadIDs = parsing.CleanStrings(filter.MuxUploadIDs)
	filter.AspectRatios = parsing.CleanStrings(filter.AspectRatios)
	filter.ResolutionTiers = parsing.CleanStrings(filter.ResolutionTiers)
}

func applyStatusFilters(db *gorm.DB, statuses []muxassetmodel.Status) *gorm.DB {
	if len(statuses) == 0 {
		return db
	}
	if slices.Contains(statuses, muxassetmodel.StatusArchived) {
		db = db.Unscoped()
	}
	return db.Model(muxassetmodel.Asset{}).Where("status IN ?", statuses)
}

func applyIdentifyingFilters(db *gorm.DB, filter *Filter) *gorm.DB {
	if len(filter.IDs) > 0 {
		db = db.Where("id IN ?", filter.IDs)
	}
	if len(filter.MuxUploadIDs) > 0 {
		db = db.Where("mux_upload_id IN ?", filter.MuxUploadIDs)
	}
	if len(filter.MuxAssetIDs) > 0 {
		db = db.Where("mux_asset_id IN ?", filter.MuxAssetIDs)
	}
	return db
}

func applySpecificFilters(db *gorm.DB, filter *Filter) *gorm.DB {
	if len(filter.AspectRatios) > 0 {
		db = db.Where("aspect_ratio IN ?", filter.AspectRatios)
	}
	if len(filter.ResolutionTiers) > 0 {
		db = db.Where("resolution_tier IN ?", filter.ResolutionTiers)
	}
	if len(filter.IngestTypes) > 0 {
		db = db.Where("ingest_type IN ?", filter.IngestTypes)
	}
	return db
}

func applyOrdering(db *gorm.DB, filter *Filter) *gorm.DB {
	sortField := "created_at" // default
	if filter.OrderBy != "" {
		sortField = string(filter.OrderBy)
	}
	sortDir := "DESC"
	if filter.OrderDir == muxassetmodel.OrderAscending {
		sortDir = "ASC"
	}
	return db.Order(sortField + " " + sortDir)
}

func hasIdentifyingFilters(filter *Filter) bool {
	if len(filter.IDs) > 0 {
		return true
	}
	if len(filter.MuxUploadIDs) > 0 {
		return true
	}
	if len(filter.MuxAssetIDs) > 0 {
		return true
	}
	return false
}

func restoreUpdates(opts *types.AuditTrailOptions) map[string]any {
	return map[string]any{
		"restored_by":      opts.AdminID,
		"restored_by_name": opts.AdminName,
		"status":           muxassetmodel.StatusActive,
		"deleted_at":       nil,
		"note":             opts.Note,
	}
}

func archiveUpdates(opts *types.AuditTrailOptions) map[string]any {
	return map[string]any{
		"archived_by":      opts.AdminID,
		"archived_by_name": opts.AdminName,
		"status":           muxassetmodel.StatusArchived,
		"note":             opts.Note,
		"archive_event_id": opts.EventID,
	}
}

func markAsBrokenUpdates(opts *types.AuditTrailOptions) map[string]any {
	return map[string]any{
		"marked_as_broken_by":      opts.AdminID,
		"marked_as_broken_by_name": opts.AdminName,
		"status":                   muxassetmodel.StatusBroken,
		"note":                     opts.Note,
	}
}

func populateFromStateOperationOptions(opts StateOperationOptions) *Filter {
	return &Filter{
		IDs:             opts.IDs,
		MuxUploadIDs:    opts.MuxUploadIDs,
		MuxAssetIDs:     opts.MuxAssetIDs,
		States:          opts.States,
		UploadStatuses:  opts.UploadStatuses,
		AspectRatios:    opts.AspectRatios,
		ResolutionTiers: opts.ResolutionTiers,
		IngestTypes:     opts.IngestTypes,
	}
}

func populateFromListOptions(opts ListOptions, scopes []Scope) *Filter {
	statuses := extractScopes(scopes)
	return &Filter{
		IDs:             opts.IDs,
		MuxUploadIDs:    opts.MuxUploadIDs,
		MuxAssetIDs:     opts.MuxAssetIDs,
		Statuses:        statuses,
		UploadStatuses:  opts.UploadStatuses,
		AspectRatios:    opts.AspectRatios,
		ResolutionTiers: opts.ResolutionTiers,
		IngestTypes:     opts.IngestTypes,
		Fields:          opts.Fields,
		OrderBy:         opts.OrderBy,
		OrderDir:        opts.OrderDir,
		PageSize:        opts.PageSize,
		PageToken:       opts.PageToken,
	}
}
