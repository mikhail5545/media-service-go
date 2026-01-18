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
	serviceerrors "github.com/mikhail5545/media-service-go/internal/errors"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// HandleUploadWebhook processes incoming webhook notifications from Cloudinary regarding asset uploads.
// It verifies the webhook signature and updates the asset status accordingly.
func (s *Service) HandleUploadWebhook(ctx context.Context, payload []byte, timestamp, signature string) error {
	var data *assetmodel.CloudinaryUploadWebhook
	if err := json.Unmarshal(payload, &data); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}

	timestampInt64, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		return serviceerrors.NewInvalidArgumentError(err)
	}
	// Verify notification signature
	if !s.apiClient.VerifyNotificationSignature(ctx, &apiclient.VerificationParams{
		Payload:           string(payload),
		ReceivedSignature: signature,
		Timestamp:         timestampInt64.Unix(),
		ValidFor:          7200, // validFor as two hours
	}) {
		s.logger.Warn("received webhook with invalid signature")
		return serviceerrors.NewPermissionDeniedError("invalid signature")
	}

	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		if data.AssetID == "" {
			return serviceerrors.NewInvalidArgumentError("asset ID is empty")
		}
		asset, err := s.getByAssetID(ctx, txRepo, data.AssetID)
		if err != nil {
			return err
		}

		updates := buildUpdatesFromWebhook(asset, data)
		if len(updates) == 0 {
			return nil
		}

		if _, err := txRepo.Update(ctx, updates, assetrepo.StateOperationOptions{IDs: uuid.UUIDs{asset.ID}}); err != nil {
			s.logger.Error("failed to update asset from webhook", zap.Error(err), zap.String("asset_id", asset.ID.String()))
			return fmt.Errorf("failed to update asset from webhook: %w", err)
		}
		return nil
	})
}
