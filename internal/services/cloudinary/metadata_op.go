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

package cloudinary

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	serviceerrors "github.com/mikhail5545/media-service-go/internal/errors"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	metadatamodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/metadata"
	bytesutil "github.com/mikhail5545/media-service-go/internal/util/bytes"
	imagepbv1 "github.com/mikhail5545/product-service-client/pb/product_service/image/v1"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
)

func (s *Service) getAssetMetadata(ctx context.Context, assetID uuid.UUID) (*metadatamodel.AssetMetadata, error) {
	metadata, err := s.metadataRepo.Get(ctx, assetID.String())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, serviceerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve asset metadata", zap.Error(err), zap.String("asset_id", assetID.String()))
		return nil, fmt.Errorf("failed to retrieve asset metadata: %w", err)
	}
	return metadata, nil
}

func (s *Service) deleteAssetMetadata(ctx context.Context, assetID uuid.UUID) error {
	if err := s.metadataRepo.Delete(ctx, assetID.String()); err != nil {
		s.logger.Error("failed to delete asset metadata", zap.Error(err), zap.String("asset_id", assetID.String()))
		return fmt.Errorf("failed to delete asset metadata: %w", err)
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

func (s *Service) deleteMetadataOnDeleteWebhook(ctx context.Context, assets []*assetmodel.Asset) (int64, error) {
	if len(assets) == 0 {
		return 0, nil
	}
	assetIDs := make([]string, len(assets))
	for i := range assets {
		assetIDs[i] = assets[i].ID.String()
	}
	metadata, err := s.metadataRepo.ListByKeys(ctx, assetIDs)
	if err != nil {
		return 0, err
	}
	haveAssociations := make(uuid.UUIDs, 0, len(metadata))
	for i := range metadata {
		if len(metadata[i].Owners) > 0 {
			haveAssociations = append(haveAssociations, uuid.MustParse(metadata[i].Key))
		}
	}
	assetIDsBytes, err := bytesutil.SliceStringsToUUIDBytes(assetIDs)
	if err != nil {
		return 0, err
	}
	if _, err := s.imageServiceClient.ForceDeleteBatch(ctx, &imagepbv1.ForceDeleteBatchRequest{
		MediaServiceUuids: assetIDsBytes,
	}); err != nil {
		return 0, fmt.Errorf("failed to force delete images in product service: %w", err)
	}
	// After removing associations, delete unowned metadata
	deleted, err := s.metadataRepo.DeleteByKeys(ctx, assetIDs)
	if err != nil {
		return 0, err
	}
	return deleted, nil
}
