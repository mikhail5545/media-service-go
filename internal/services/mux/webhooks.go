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

/*
Package mux provides service-layer business logic for for mux asset model.
*/
package mux

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	metarepo "github.com/mikhail5545/media-service-go/internal/database/arango/mux/metadata"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	detailmodel "github.com/mikhail5545/media-service-go/internal/models/mux/detail"
	metamodel "github.com/mikhail5545/media-service-go/internal/models/mux/metadata"
	"gorm.io/gorm"
)

// HandleAssetCreatedWebhook processes an incoming Mux webhook with "video.asset.created" event type, finds the corresponding asset,
// and updates it in a patch-like manner.
func (s *service) HandleAssetCreatedWebhook(ctx context.Context, payload *assetmodel.MuxWebhook) error {
	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.Repo.WithTx(tx)
		txDetailRepo := s.detailRepo.WithTx(tx)

		var asset *assetmodel.Asset
		var err error

		if payload.Data.UploadID != nil && *payload.Data.UploadID != "" {
			asset, err = txRepo.GetByUploadID(ctx, *payload.Data.UploadID)
		} else {
			asset, err = txRepo.GetByAssetID(ctx, payload.Data.ID) // data.ID is required, so no pointer
		}

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: asset not found for upload_id '%s' or asset_id '%s'", ErrNotFound, *payload.Data.UploadID, payload.Data.ID)
			}
			return fmt.Errorf("failed to retrieve asset for webhook: %w", err)
		}

		updates := buildAssetUpdates(asset, &payload.Data)

		if len(updates) == 0 {
			return nil
		}

		// Separately handle the bulky 'Tracks' data by upserting it.
		if len(payload.Data.Tracks) > 0 {
			details := detailmodel.AssetDetail{AssetID: asset.ID, Tracks: payload.Data.Tracks}
			if err := txDetailRepo.Upsert(ctx, &details); err != nil {
				return fmt.Errorf("failed to upsert asset details from webhook: %w", err)
			}
		}

		if _, err := txRepo.Update(ctx, asset, updates); err != nil {
			return fmt.Errorf("failed to update asset from webhook: %w", err)
		}

		return nil
	})
}

// HandleAssetReadyWebhook processes an incoming Mux webhook with "video.asset.ready" event type, finds the corresponding asset,
// and updates it in a patch-like manner.
func (s *service) HandleAssetReadyWebhook(ctx context.Context, payload *assetmodel.MuxWebhook) error {
	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.Repo.WithTx(tx)
		txDetailRepo := s.detailRepo.WithTx(tx)

		var asset *assetmodel.Asset
		var err error

		if payload.Data.UploadID != nil && *payload.Data.UploadID != "" {
			asset, err = txRepo.GetByUploadID(ctx, *payload.Data.UploadID)
		} else {
			asset, err = txRepo.GetByAssetID(ctx, payload.Data.ID) // data.ID is required, so no pointer
		}

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: asset not found for upload_id '%s' or asset_id '%s'", ErrNotFound, *payload.Data.UploadID, payload.Data.ID)
			}
			return fmt.Errorf("failed to retrieve asset for webhook: %w", err)
		}

		updates := buildAssetUpdates(asset, &payload.Data)

		if len(updates) == 0 {
			return nil
		}

		// Separately handle the bulky 'Tracks' data by upserting it.
		if len(payload.Data.Tracks) > 0 {
			details := detailmodel.AssetDetail{AssetID: asset.ID, Tracks: payload.Data.Tracks}
			if err := txDetailRepo.Upsert(ctx, &details); err != nil {
				return fmt.Errorf("failed to upsert asset details from webhook: %w", err)
			}
		}

		if _, err := txRepo.Update(ctx, asset, updates); err != nil {
			return fmt.Errorf("failed to update asset from webhook: %w", err)
		}

		return nil
	})
}

// HandleAssetErroredWebhook processes an incoming Mux webhook with "video.asset.errored" event type, finds the corresponding asset,
// and updates it in a patch-like manner. After update, it soft-deleted mux asset. If asset has owners, they will be deassociated and
// all asset metadata about it's owners will be cleared.
func (s *service) HandleAssetErroredWebhook(ctx context.Context, payload *assetmodel.MuxWebhook) error {
	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.Repo.WithTx(tx)
		txDetailRepo := s.detailRepo.WithTx(tx)

		var asset *assetmodel.Asset
		var err error

		if payload.Data.UploadID != nil && *payload.Data.UploadID != "" {
			asset, err = txRepo.GetByUploadID(ctx, *payload.Data.UploadID)
		} else {
			asset, err = txRepo.GetByAssetID(ctx, payload.Data.ID) // data.ID is required, so no pointer
		}

		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: asset not found for upload_id '%s' or asset_id '%s'", ErrNotFound, *payload.Data.UploadID, payload.Data.ID)
			}
			return fmt.Errorf("failed to retrieve asset for webhook: %w", err)
		}

		updates := buildAssetUpdates(asset, &payload.Data)

		if len(updates) == 0 {
			return nil
		}

		// Separately handle the bulky 'Tracks' data by upserting it.
		if len(payload.Data.Tracks) > 0 {
			details := detailmodel.AssetDetail{AssetID: asset.ID, Tracks: payload.Data.Tracks}
			if err := txDetailRepo.Upsert(ctx, &details); err != nil {
				return fmt.Errorf("failed to upsert asset details from webhook: %w", err)
			}
		}

		if _, err := txRepo.Update(ctx, asset, updates); err != nil {
			return fmt.Errorf("failed to update asset from webhook: %w", err)
		}

		meta, err := s.metaRepo.Get(ctx, asset.ID)
		if err != nil && !errors.Is(err, metarepo.ErrNotFound) {
			return fmt.Errorf("failed to retrieve asset metadata: %w", err)
		}

		// If asset has owners, de-associate them
		if meta != nil && len(meta.Owners) > 0 {
			toRemove := make(map[string][]string)
			for _, owner := range meta.Owners {
				toRemove[owner.OwnerType] = append(toRemove[owner.OwnerType], owner.OwnerID)
			}

			// Notify other services via gRPC about ownership changes
			if err := s.processChanges(ctx, asset, nil, toRemove); err != nil {
				return fmt.Errorf("failed to notify external services about changes: %w", err)
			}

			// Delete all information about owners from asset metadata in the ArangoDB.
			// This will keep asset metadata about Title and CreatorID untouched.
			if err := s.metaRepo.Update(ctx, asset.ID, &metamodel.AssetMetadata{Owners: []metamodel.Owner{}}); err != nil {
				return fmt.Errorf("failed to delete asset owners metadata: %w", err)
			}
		}

		// Soft-delete asset
		if _, err := s.Repo.WithTx(tx).Delete(ctx, asset.ID); err != nil {
			return fmt.Errorf("failed to delete mux upload: %w", err)
		}

		return nil
	})
}

// buildAssetUpdates compares the existing asset with the webhook data and constructs a
// map of fields that need to be updated. This implements the "patch-like" update.
func buildAssetUpdates(asset *assetmodel.Asset, data *assetmodel.MuxWebhookData) map[string]any {
	updates := make(map[string]any)

	if data.Status != nil && (asset.Status == nil || *asset.Status != *data.Status) {
		updates["status"] = *data.Status
	}
	if data.Progress.State != "" && asset.State != data.Progress.State {
		updates["state"] = data.Progress.State
	}

	if data.ID != "" && (asset.MuxAssetID == nil || *asset.MuxAssetID != data.ID) {
		updates["mux_asset_id"] = data.ID
	}
	if len(data.PlaybackIDs) > 0 && !reflect.DeepEqual(asset.MuxPlaybackIDs, data.PlaybackIDs) {
		updates["playback_ids"] = data.PlaybackIDs
	}

	if data.Duration != nil && (asset.Duration == nil || *asset.Duration != *data.Duration) {
		updates["duration"] = *data.Duration
	}
	if data.AspectRatio != nil && (asset.AspectRatio == nil || *asset.AspectRatio != *data.AspectRatio) {
		updates["aspect_ratio"] = *data.AspectRatio
	}
	if data.IngestType != nil && (asset.IngestType == nil || *asset.IngestType != *data.IngestType) {
		updates["ingest_type"] = *data.IngestType
	}
	if !data.CreatedAt.IsZero() && (asset.AssetCreatedAt == nil || !asset.AssetCreatedAt.Equal(data.CreatedAt)) {
		updates["asset_created_at"] = data.CreatedAt
	}

	if data.ResolutionTier != nil && (asset.ResolutionTier == nil || *asset.ResolutionTier != *data.ResolutionTier) {
		updates["resolution_tier"] = *data.ResolutionTier
	}

	return updates
}
