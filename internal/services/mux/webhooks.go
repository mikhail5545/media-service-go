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

package mux

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	assetrepo "github.com/mikhail5545/media-service-go/internal/database/postgres/mux/asset"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	muxtypes "github.com/mikhail5545/media-service-go/internal/models/mux/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HandleAssetWebhook processes incoming MUX asset webhooks based on their type.
// It routes the webhook to the appropriate handler function.
// It does not return any error, as we want to avoid retrying the webhook processing in case of failure.
func (s *Service) HandleAssetWebhook(ctx context.Context, payload *muxtypes.MuxWebhook) error {
	switch payload.Type {
	case "video.asset.created", "video.asset.ready", "video.asset.updated":
		return s.handleDataRichWebhook(ctx, payload)
	case "video.asset.errored":
		return s.handleAssetErroredWebhook(ctx, payload)
	case "video.asset.deleted":
		return s.handleAssetDeletedWebhook(ctx, payload)
	default:
		s.logger.Warn(
			"received unsupported webhook type",
			zap.String("type", payload.Type),
			zap.String("event_id", payload.ID),
		)
		return nil
	}
	// TODO: notify other services about the asset update if needed to sync state/cache
}

// handleDataRichWebhook processes webhooks that contain rich data about the asset.
// It updates the local asset record with the information provided in the webhook.
// It does not return any error, as we want to avoid retrying the webhook processing in case of failure.
// This includes 'video.asset.created', 'video.asset.ready', and 'video.asset.updated' types.
func (s *Service) handleDataRichWebhook(ctx context.Context, payload *muxtypes.MuxWebhook) error {
	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		asset := s.getAssetFromWebhook(ctx, txRepo, payload)
		if asset == nil {
			return nil
		}

		updates := buildAssetUpdatesFromWebhook(asset, &payload.Data)
		// explicitly extract playback IDs hot paths
		if err := extractPlaybackIDs(updates, &payload.Data); err != nil {
			s.logger.Warn(
				"failed to extract playback IDs from webhook",
				zap.Error(err),
				zap.String("asset_id", asset.ID.String()),
				zap.String("event_id", payload.ID),
			)
			return nil
		}
		if len(updates) > 0 {
			if _, err := txRepo.Update(ctx, updates, assetrepo.StateOperationOptions{IDs: uuid.UUIDs{asset.ID}}); err != nil {
				s.logger.Warn(
					"failed to update asset from webhook",
					zap.Error(err),
					zap.String("asset_id", asset.ID.String()),
					zap.String("event_id", payload.ID),
				)
				return nil
			}
		}

		if err := s.updateMetadataFromWebhook(ctx, asset.ID, &payload.Data); err != nil {
			s.logger.Warn(
				"failed to update asset metadata from webhook",
				zap.Error(err),
				zap.String("asset_id", asset.ID.String()),
				zap.String("event_id", payload.ID),
			)
			return nil
		}
		return nil
	})
}

// handleAssetErroredWebhook processes 'video.asset.errored' type webhooks specifically.
// It does not return any error, as we want to avoid retrying the webhook processing in case of failure.
func (s *Service) handleAssetErroredWebhook(ctx context.Context, payload *muxtypes.MuxWebhook) error {
	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		asset := s.getAssetFromWebhook(ctx, txRepo, payload)
		if asset == nil {
			// getAssetFromWebhook already logs the missing asset case
			return nil
		}
		// In case of errored webhook, we only update the status to 'errored'.
		updates := map[string]any{
			"upload_status": "errored",
			"status":        assetmodel.StatusBroken,
		}
		if payload.Data.Errors != nil {
			updates["mux_error"] = payload.Data.Errors
		}
		if _, err := txRepo.Update(ctx, updates, assetrepo.StateOperationOptions{IDs: uuid.UUIDs{asset.ID}}); err != nil {
			s.logger.Warn(
				"failed to update asset to errored from webhook",
				zap.Error(err),
				zap.String("asset_id", asset.ID.String()),
				zap.String("event_id", payload.ID),
			)
			return nil
		}
		return nil
	})
}

// handleAssetDeletedWebhook processes 'video.asset.deleted' type webhooks specifically.
// It archives the asset locally if it was not already archived.
// It does not return any error, as we want to avoid retrying the webhook processing in case of failure.
func (s *Service) handleAssetDeletedWebhook(ctx context.Context, payload *muxtypes.MuxWebhook) error {
	var assetIDtoDelete *uuid.UUID
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		asset := s.getAssetFromWebhook(ctx, txRepo, payload)
		if asset == nil {
			// getAssetFromWebhook already logs the missing asset case
			return nil
		}

		// There is two cases here:
		// 1. Asset was already archived in our system (deleted locally and from MUX, we received webhook about this) - in this case we do nothing.
		// 2. Asset was not archived in our system (deleted from MUX directly) - in this case we archive it locally.
		if asset.Status == assetmodel.StatusArchived {
			return nil
		}
		if err := s.archiveAssetOnWebhook(ctx, txRepo, asset, payload.ID); err != nil {
			return nil
		}
		assetIDtoDelete = &asset.ID
		return nil
	})
	if err == nil && assetIDtoDelete != nil {
		if err := s.deleteMetadataOnWebhook(ctx, *assetIDtoDelete, payload); err != nil {
			s.logger.Warn(
				"failed to delete asset metadata on deleted webhook",
				zap.Error(err),
				zap.String("asset_id", assetIDtoDelete.String()),
				zap.String("event_id", payload.ID),
			)
		}
	}
	return err
}

func extractPlaybackIDs(updates map[string]any, data *muxtypes.MuxWebhookData) error {
	if len(data.PlaybackIDs) == 0 {
		return nil
	}
	for _, id := range data.PlaybackIDs {
		switch id.Policy {
		case "public":
			if _, exists := updates["primary_public_playback_id"]; !exists {
				updates["primary_public_playback_id"] = id.ID
			}
		case "signed":
			if _, exists := updates["primary_signed_playback_id"]; !exists {
				updates["primary_signed_playback_id"] = id.ID
			}
		default:
			return fmt.Errorf("encountered unknown playback ID policy: %s", id.Policy)
		}
	}
	return nil
}
