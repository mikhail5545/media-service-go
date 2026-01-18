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
	"errors"
	"fmt"

	"github.com/google/uuid"
	assetrepo "github.com/mikhail5545/media-service-go/internal/database/postgres/cloudinary/asset"
	serviceerrors "github.com/mikhail5545/media-service-go/internal/errors"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	metadatamodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/metadata"
	bytesutil "github.com/mikhail5545/media-service-go/internal/util/bytes"
	"github.com/mikhail5545/media-service-go/internal/util/parsing"
	imagepbv1 "github.com/mikhail5545/product-service-client/pb/proto/product_service/image/v1"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) getAsset(ctx context.Context, assetID uuid.UUID, scopes []assetrepo.Scope) (*assetmodel.Asset, error) {
	asset, err := s.repo.Get(ctx, assetrepo.GetOptions{
		ID: assetID,
	}, scopes...)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve asset", zap.Error(err), zap.String("asset_id", assetID.String()))
		return nil, fmt.Errorf("failed to retrieve asset: %w", err)
	}
	return asset, nil
}

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

func (s *Service) get(ctx context.Context, filter *assetmodel.GetFilter, scopes []assetrepo.Scope) (*assetmodel.Details, error) {
	assetID, err := parsing.StrToUUID(filter.ID)
	if err != nil {
		return nil, err
	}
	asset, err := s.getAsset(ctx, assetID, scopes)
	if err != nil {
		return nil, err
	}
	metadata, err := s.getAssetMetadata(ctx, assetID)
	if err != nil {
		return nil, err
	}
	return &assetmodel.Details{
		Asset:    asset,
		Metadata: metadata,
	}, nil
}

func (s *Service) list(ctx context.Context, req *assetmodel.ListRequest, scopes []assetrepo.Scope) ([]*assetmodel.Details, string, error) {
	listOptions := assetrepo.ListOptions{
		CloudinaryAssetIDs:  req.CloudinaryAssetIDs,
		CloudinaryPublicIDs: req.CloudinaryPublicIDs,
		ResourceTypes:       req.ResourceTypes,
		Formats:             req.Formats,
		OrderDir:            req.OrderDir,
		OrderField:          req.OrderField,
		PageSize:            req.PageSize,
		PageToken:           req.PageToken,
	}
	listOptions.IDs = parsing.StrToUUIDs(req.IDs)

	assets, nextPageToken, err := s.repo.List(ctx, listOptions, scopes...)
	if err != nil {
		s.logger.Error("failed to list assets", zap.Error(err))
		return nil, "", fmt.Errorf("failed to list assets: %w", err)
	}

	assetIDs := make([]string, len(assets))
	for i := range assets {
		assetIDs[i] = assets[i].ID.String()
	}
	metadataMap, err := s.metadataRepo.ListByKeys(ctx, assetIDs)
	if err != nil {
		s.logger.Error("failed to list asset metadata", zap.Error(err))
		return nil, "", fmt.Errorf("failed to list asset metadata: %w", err)
	}

	response := make([]*assetmodel.Details, 0, len(assets))
	for i := range assets {
		metadata, ok := metadataMap[assets[i].ID.String()]
		if !ok {
			s.logger.Warn("metadata not found for asset", zap.String("asset_id", assets[i].ID.String()))
			continue
		}
		response = append(response, &assetmodel.Details{
			Asset:    assets[i],
			Metadata: metadata,
		})
	}
	return response, nextPageToken, nil
}

func (s *Service) getInTx(ctx context.Context, txRepo *assetrepo.Repository, id string, fields []string) (*assetmodel.Asset, error) {
	assetID, err := parsing.StrToUUID(id)
	if err != nil {
		return nil, err
	}
	asset, err := txRepo.Get(ctx, assetrepo.GetOptions{
		ID:     assetID,
		Fields: fields,
	}, assetrepo.ScopeAll)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve asset in transaction", zap.Error(err), zap.String("asset_id", assetID.String()))
		return nil, fmt.Errorf("failed to retrieve asset in transaction: %w", err)
	}
	return asset, nil
}

func (s *Service) grpcRemoveAssociations(ctx context.Context, asset *assetmodel.Asset, metadata *metadatamodel.AssetMetadata) error {
	if len(metadata.Owners) > 0 {
		metadata.Owners = []metadatamodel.Owner{}
	}

	assetIDBytes, err := bytesutil.UUIDToBytes(&asset.ID)
	if err != nil {
		return err
	}
	if _, err := s.imageServiceClient.Delete(ctx, &imagepbv1.DeleteRequest{
		MediaServiceUuid: assetIDBytes,
	}); err != nil {
		s.logger.Error("failed to delete image associations via gRPC", zap.Error(err), zap.String("asset_id", asset.ID.String()))
		return fmt.Errorf("failed to delete image associations via gRPC: %w", err)
	}

	return s.metadataRepo.Update(ctx, metadata.Key, metadata)
}

func (s *Service) checkOwnership(ctx context.Context, owner *metadatamodel.Owner, assetID uuid.UUID) error {
	_, findErr := s.metadataRepo.GetByOwner(ctx, assetID.String(), owner)
	if findErr == nil {
		return serviceerrors.NewAlreadyExistsError("owner already exists for this asset")
	}
	if !errors.Is(findErr, mongo.ErrNoDocuments) {
		s.logger.Error(
			"failed to check existing owner in asset metadata",
			zap.Error(findErr),
			zap.String("asset_id", assetID.String()),
			zap.String("owner_id", owner.OwnerID),
			zap.String("owner_type", owner.OwnerType),
		)
		return fmt.Errorf("failed to check existing owner in asset metadata: %w", findErr)
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
	metadata.Owners = append(metadata.Owners, newOwner)

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

func (s *Service) deleteMetadata(ctx context.Context, assetID uuid.UUID) error {
	if err := s.metadataRepo.Delete(ctx, assetID.String()); err != nil {
		s.logger.Error("failed to delete asset metadata", zap.Error(err), zap.String("asset_id", assetID.String()))
		return fmt.Errorf("failed to delete asset metadata: %w", err)
	}
	return nil
}

func (s *Service) getByAssetID(ctx context.Context, txRepo *assetrepo.Repository, cloudinaryAssetID string) (*assetmodel.Asset, error) {
	asset, err := txRepo.Get(ctx, assetrepo.GetOptions{
		CloudinaryAssetID: cloudinaryAssetID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve asset by Cloudinary Asset ID", zap.Error(err), zap.String("cloudinary_asset_id", cloudinaryAssetID))
		return nil, fmt.Errorf("failed to retrieve asset by Cloudinary Asset ID: %w", err)
	}
	return asset, nil
}
