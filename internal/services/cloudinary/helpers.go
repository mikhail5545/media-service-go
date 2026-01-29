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
	"github.com/mikhail5545/media-service-go/internal/database/types"
	serviceerrors "github.com/mikhail5545/media-service-go/internal/errors"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	metadatamodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/metadata"
	"github.com/mikhail5545/media-service-go/internal/util/parsing"
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

func (s *Service) markAsBroken(ctx context.Context, txRepo *assetrepo.Repository, assetID uuid.UUID, req *assetmodel.ChangeStateRequest) error {
	adminID, err := parsing.StrToUUID(req.AdminID)
	if err != nil {
		return err
	}
	if _, err := txRepo.MarkAsBroken(ctx, assetrepo.StateOperationOptions{IDs: uuid.UUIDs{assetID}}, &types.AuditTrailOptions{
		AdminID:   adminID,
		AdminName: req.AdminName,
		Note:      req.Note,
	}); err != nil {
		s.logger.Error("failed to mark asset as broken", zap.Error(err), zap.String("asset_id", assetID.String()))
		return fmt.Errorf("failed to mark asset as broken: %w", err)
	}
	// make gRPC call to remove associations
	if err := s.grpcMarkAsBroken(ctx, &assetID, &adminID, req); err != nil {
		return err
	}
	return nil
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

func (s *Service) getByPublicID(ctx context.Context, txRepo *assetrepo.Repository, cloudinaryPublicID string) (*assetmodel.Asset, error) {
	asset, err := txRepo.Get(ctx, assetrepo.GetOptions{
		CloudinaryAssetID: cloudinaryPublicID,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve asset by Cloudinary Public ID", zap.Error(err), zap.String("cloudinary_public_id", cloudinaryPublicID))
		return nil, fmt.Errorf("failed to retrieve asset by Cloudinary Public ID: %w", err)
	}
	return asset, nil
}

func (s *Service) listByPublicIDs(ctx context.Context, txRepo *assetrepo.Repository, cloudinaryPublicIDs []string, scopes ...assetrepo.Scope) ([]*assetmodel.Asset, error) {
	if len(cloudinaryPublicIDs) == 0 {
		return []*assetmodel.Asset{}, nil
	}
	assets, err := txRepo.ListAll(ctx, assetrepo.ListAllOptions{
		CloudinaryPublicIDs: cloudinaryPublicIDs,
	}, scopes...)
	if err != nil {
		return nil, fmt.Errorf("failed to list assets by Cloudinary Public IDs: %w", err)
	}
	return assets, nil
}

func (s *Service) markAsBrokenAndClearOwners(ctx context.Context, assetID *uuid.UUID, metadata *metadatamodel.AssetMetadata, req *assetmodel.ChangeStateRequest) error {
	adminID, err := parsing.StrToUUID(req.AdminID)
	if err != nil {
		return err
	}
	if err := s.grpcMarkAsBroken(ctx, assetID, &adminID, req); err != nil {
		return err
	}
	metadata.Owners = []*metadatamodel.Owner{}
	return s.metadataRepo.Update(ctx, metadata.Key, metadata)
}
