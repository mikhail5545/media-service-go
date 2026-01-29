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

package cloudinary

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	apiclient "github.com/mikhail5545/media-service-go/internal/apiclients/cloudinary"
	assetrepo "github.com/mikhail5545/media-service-go/internal/database/postgres/cloudinary/asset"
	"github.com/mikhail5545/media-service-go/internal/database/types"
	serviceerrors "github.com/mikhail5545/media-service-go/internal/errors"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	cldtypes "github.com/mikhail5545/media-service-go/internal/models/cloudinary/types"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type genericData struct {
	NotificationType string `json:"notification_type"`
}

// HandleWebhook processes incoming webhook notifications from Cloudinary.
// It validates the signature and routes the webhook to the appropriate handler based on its type.
func (s *Service) HandleWebhook(ctx context.Context, payload []byte, timestamp, signature string) error {
	// First, validate signature
	timestampInt64, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return serviceerrors.NewInvalidArgumentError(err)
	}
	if !s.apiClient.VerifyNotificationSignature(ctx, &apiclient.VerificationParams{
		Payload:           string(payload),
		ReceivedSignature: signature,
		Timestamp:         timestampInt64.Unix(),
		ValidFor:          7200, // validFor as two hours
	}) {
		s.logger.Warn("received webhook with invalid signature")
		return serviceerrors.NewPermissionDeniedError("invalid signature")
	}

	// Determine webhook type
	var generic genericData
	if err := json.Unmarshal(payload, &generic); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}

	switch generic.NotificationType {
	case "upload":
		return s.handleUploadWebhook(ctx, payload)
	case "rename":
		return s.handleRenameWebhook(ctx, payload)
	case "delete":
		return s.handleDeleteWebhook(ctx, payload)
	default:
		return nil
	}
}

// HandleUploadWebhook processes incoming webhook notifications from Cloudinary regarding asset uploads.
// It updates the local asset records with the information provided in the webhook.
func (s *Service) handleUploadWebhook(ctx context.Context, payload []byte) error {
	var data assetmodel.CloudinaryUploadWebhook
	if err := json.Unmarshal(payload, &data); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}

	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if data.PublicID == "" {
			return serviceerrors.NewInvalidArgumentError("public ID is empty")
		}
		asset, err := s.getByPublicID(ctx, txRepo, data.PublicID)
		if err != nil {
			return err
		}

		updates := buildUpdatesFromWebhook(asset, &data)
		if len(updates) == 0 {
			return nil
		}

		if _, err := txRepo.Update(ctx, updates, assetrepo.StateOperationOptions{IDs: uuid.UUIDs{asset.ID}}); err != nil {
			s.logger.Error("failed to update asset from webhook", zap.Error(err), zap.String("asset_id", asset.ID.String()), zap.String("public_id", data.PublicID))
			return fmt.Errorf("failed to update asset from webhook: %w", err)
		}
		return nil
	})
}

// handleRenameWebhook processes incoming webhook notifications from Cloudinary regarding asset renames.
// It updates the local asset records to reflect the new public ID.
func (s *Service) handleRenameWebhook(ctx context.Context, payload []byte) error {
	var data cldtypes.CloudinaryRenameWebhook
	if err := json.Unmarshal(payload, &data); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}

	logger := s.logger.With(
		zap.String("webhook_notification_type", data.NotificationType),
		zap.String("triggered_by.source", data.NotificationContext.TriggeredBy.Source),
		zap.String("triggered_by.id", data.NotificationContext.TriggeredBy.ID),
	)
	logger.Info("received Cloudinary delete webhook")

	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		asset, err := s.getByPublicID(ctx, txRepo, data.FromPublicID)
		if err != nil {
			logger.Error("failed to retrieve asset by old Cloudinary Public ID", zap.Error(err), zap.String("from_public_id", data.FromPublicID))
			return nil
		}

		updates := map[string]any{
			"cloudinary_public_id": data.ToPublicID,
		}

		if _, err := txRepo.Update(ctx, updates, assetrepo.StateOperationOptions{IDs: uuid.UUIDs{asset.ID}}); err != nil {
			logger.Error("failed to update asset Cloudinary Public ID from webhook", zap.Error(err), zap.String("asset_id", asset.ID.String()), zap.String("from_public_id", data.FromPublicID), zap.String("to_public_id", data.ToPublicID))
			return nil
		}
		return nil
	})
}

// handleDeleteWebhook processes incoming webhook notifications from Cloudinary regarding asset deletions.
// It synchronizes the local asset records by archiving them and removing associated metadata as necessary.
func (s *Service) handleDeleteWebhook(ctx context.Context, payload []byte) error {
	var data cldtypes.CloudinaryDeleteWebhook
	if err := json.Unmarshal(payload, &data); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}

	// Use dedicated logger for webhook processing
	logger := s.logger.With(
		zap.String("webhook_notification_type", data.NotificationType),
		zap.String("triggered_by.source", data.NotificationContext.TriggeredBy.Source),
		zap.String("triggered_by.id", data.NotificationContext.TriggeredBy.ID),
	)
	logger.Info("received Cloudinary delete webhook")

	var toDelete []*assetmodel.Asset
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		// This webhook may contain multiple public IDs
		pubIDs := make([]string, 0, len(data.Resources))
		for i := range data.Resources {
			pubIDs = append(pubIDs, data.Resources[i].PublicID)
		}
		// There is two cases here:
		// 1. Asset was already archived in our system (deleted locally and from Cloudinary, we received webhook about this) - in this case we do nothing.
		// 2. Asset was not archived in our system (deleted from Cloudinary directly) - in this case we archive it locally.
		affected, assets, err := s.archiveOnDeleteWebhook(ctx, txRepo, pubIDs, data.NotificationContext)
		if err != nil {
			logger.Error("failed to archive assets on Cloudinary delete webhook", zap.Error(err))
			return nil
		}
		logger.Info("archived assets on Cloudinary delete webhook", zap.Int64("affected_assets", affected))
		toDelete = assets
		return nil
	})
	if err == nil && len(toDelete) > 0 {
		// Delete asset metadata from MongoDB after successful transaction commit.
		deleted, err := s.deleteMetadataOnDeleteWebhook(ctx, toDelete)
		if err != nil {
			logger.Warn(
				"failed to delete asset metadata after Cloudinary delete webhook",
				zap.Error(err),
			)
		} else {
			logger.Info("deleted asset metadata after Cloudinary delete webhook", zap.Int64("deleted_metadata_records", deleted))
		}
	}
	return nil
}

func (s *Service) archiveOnDeleteWebhook(ctx context.Context, txRepo *assetrepo.Repository, pubIDs []string, notificationContext cldtypes.NotificationContext) (int64, []*assetmodel.Asset, error) {
	assets, err := s.listByPublicIDs(ctx, txRepo, pubIDs, assetrepo.ScopeUploadURLGenerated, assetrepo.ScopeActive, assetrepo.ScopeBroken) // only non archived assets
	if err != nil {
		return 0, nil, nil
	}
	if len(assets) == 0 {
		return 0, nil, nil
	}
	toArchive := make([]string, 0, len(assets))
	for _, asset := range assets {
		toArchive = append(toArchive, asset.CloudinaryPublicID)
	}
	affected, err := txRepo.Archive(ctx, assetrepo.StateOperationOptions{CloudinaryPublicIDs: toArchive}, &types.AuditTrailOptions{
		Note:      "Received Cloudinary delete webhook",
		AdminName: "system",
		EventID:   notificationContext.TriggeredBy.ID,
	})
	if err != nil {
		return 0, nil, err
	}
	return affected, assets, nil
}
