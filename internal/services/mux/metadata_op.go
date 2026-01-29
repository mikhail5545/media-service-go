/*
 * Copyright (c) 2026. Mikhail Kulik
 *
 * This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published
 *  by the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 *  along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package mux

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	metadatamodel "github.com/mikhail5545/media-service-go/internal/models/mux/metadata"
	muxtypes "github.com/mikhail5545/media-service-go/internal/models/mux/types"
	"github.com/mikhail5545/media-service-go/internal/util/memory"
	"go.uber.org/zap"
)

func (s *Service) deleteAssetMetadata(ctx context.Context, assetID uuid.UUID) error {
	if err := s.metadataRepo.Delete(ctx, assetID.String()); err != nil {
		s.logger.Error("failed to delete asset metadata", zap.Error(err), zap.String("asset_id", assetID.String()))
		return fmt.Errorf("failed to delete asset metadata: %w", err)
	}
	return nil
}

func (s *Service) updateMetadataFromWebhook(ctx context.Context, assetID uuid.UUID, data *muxtypes.MuxWebhookData) error {
	if len(data.Tracks) == 0 {
		return nil
	}
	metadata, err := s.getAssetMetadata(ctx, assetID)
	if err != nil {
		return err
	}

	metadata.Tracks = memory.SlicePtr(data.Tracks...)
	if err := s.metadataRepo.Update(ctx, assetID.String(), metadata); err != nil {
		s.logger.Error("failed to update asset metadata from webhook", zap.Error(err), zap.String("asset_id", assetID.String()))
		return fmt.Errorf("failed to update asset metadata from webhook: %w", err)
	}
	return nil
}

func (s *Service) addOwner(ctx context.Context, assetID uuid.UUID, req *assetmodel.ManageOwnerRequest) error {
	metadata, err := s.getAssetMetadata(ctx, assetID)
	if err != nil {
		return err
	}

	newOwner := metadatamodel.Owner{
		OwnerID:   req.OwnerID,
		OwnerType: req.OwnerType,
	}
	if err := s.checkOwnership(ctx, &newOwner, assetID); err != nil {
		return err
	}
	metadata.Owners = append(metadata.Owners, &newOwner)

	if err := s.metadataRepo.Update(ctx, assetID.String(), metadata); err != nil {
		s.logger.Error("failed to add owner to asset metadata", zap.Error(err), zap.String("asset_id", assetID.String()))
		return fmt.Errorf("failed to add owner to asset metadata: %w", err)
	}
	return nil
}

func (s *Service) removeOwner(ctx context.Context, metadata *metadatamodel.AssetMetadata, req *assetmodel.ManageOwnerRequest) error {
	currentOwners := metadata.Owners
	for i, owner := range currentOwners {
		if owner.OwnerID == req.OwnerID && owner.OwnerType == req.OwnerType {
			// Remove owner from slice
			metadata.Owners = append(currentOwners[:i], currentOwners[i+1:]...)
			break
		}
	}
	if err := s.metadataRepo.Update(ctx, metadata.Key, metadata); err != nil {
		s.logger.Error("failed to remove owner from asset metadata",
			zap.Error(err), zap.String("owner_id", req.OwnerID), zap.String("owner_type", req.OwnerType),
		)
		return fmt.Errorf("failed to remove owner from asset metadata: %w", err)
	}
	return nil
}

func (s *Service) deleteMetadataOnWebhook(ctx context.Context, assetID uuid.UUID, payload *muxtypes.MuxWebhook) error {
	metadata, err := s.getAssetMetadata(ctx, assetID)
	if err != nil {
		return err
	}
	if len(metadata.Owners) > 0 {
		if err := s.grpcForceDelete(ctx, &assetID); err != nil {
			return err
		}
	}
	if err := s.deleteAssetMetadata(ctx, assetID); err != nil {
		s.logger.Warn(
			"failed to delete asset metadata from webhook",
			zap.Error(err),
			zap.String("asset_id", assetID.String()),
			zap.String("event_id", payload.ID),
		)
		return err
	}
	return nil
}
