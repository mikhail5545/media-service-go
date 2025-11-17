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
Package cloudinary provides service-layer logic for Cloudinary asset management and asset models.
*/
package cloudinary

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	"gorm.io/gorm"
)

// HandleUploadWebhook processes an incoming Cloudinary upload webhook, finds the corresponding asset,
// and updates it in a patch-like manner.
func (s *service) HandleUploadWebhook(ctx context.Context, payload []byte, recievedTimestamp, recievedSignature string) error {
	var data *assetmodel.CloudinaryUploadWebhook
	if err := json.Unmarshal(payload, &data); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	timestamp, err := time.Parse(time.RFC3339, recievedTimestamp)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	// Verify notification signature
	if !s.Client.VerifyNotificationSignature(ctx, string(payload), recievedSignature, timestamp.Unix(), 7200) { // validFor as two hours
		return ErrInvalidSignature
	}

	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.Repo.WithTx(tx)

		if data.AssetID == "" {
			return fmt.Errorf("%w: AssetID is empty", ErrInvalidArgument)
		}

		asset, err := txRepo.GetWithDeletedByAssetID(ctx, data.AssetID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: %w", ErrNotFound, err)
			}
			return fmt.Errorf("failed to retrieve asset: %w", err)
		}

		updates := buildAssetUpdates(asset, data)

		if len(updates) > 0 {
			if _, err := txRepo.Update(ctx, asset, updates); err != nil {
				return fmt.Errorf("failed to update asset: %w", err)
			}
		}

		return nil
	})
}

// buildAssetUpdates compares the existing asset with the webhook data and constructs a
// map of fields that need to be updated. This implements the "patch-like" update.
func buildAssetUpdates(asset *assetmodel.Asset, data *assetmodel.CloudinaryUploadWebhook) map[string]any {
	updates := make(map[string]any)

	if data.DisplayName != "" && data.DisplayName != asset.DisplayName {
		updates["display_name"] = data.DisplayName
	}
	if data.AssetFolder != "" && data.AssetFolder != asset.AssetFolder {
		updates["asset_folder"] = data.AssetFolder
	}
	if data.Url != "" && data.Url != asset.URL {
		updates["url"] = data.Url
	}
	if data.SecureUrl != "" && data.SecureUrl != asset.SecureURL {
		updates["secure_url"] = data.SecureUrl
	}
	if data.PublicID != "" && data.PublicID != asset.CloudinaryPublicID { // Should not change, but good to have
		updates["cloudinary_public_id"] = data.PublicID
	}
	if data.Height != 0 && (asset.Height == nil || data.Height != *asset.Height) {
		updates["height"] = data.Height
	}
	if data.Width != 0 && (asset.Width == nil || data.Width != *asset.Width) {
		updates["width"] = data.Width
	}
	if data.Format != "" && data.Format != asset.Format {
		updates["format"] = data.Format
	}
	if data.ResourceType != "" && data.ResourceType != asset.ResourceType {
		updates["resource_type"] = data.ResourceType
	}
	if len(data.Tags) > 0 && !reflect.DeepEqual(data.Tags, asset.Tags) {
		updates["tags"] = data.Tags
	}

	return updates
}
