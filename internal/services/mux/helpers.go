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
	"errors"
	"fmt"

	"github.com/google/uuid"
	assetrepo "github.com/mikhail5545/media-service-go/internal/database/postgres/mux/asset"
	"github.com/mikhail5545/media-service-go/internal/database/types"
	serviceerrors "github.com/mikhail5545/media-service-go/internal/errors"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	metadatamodel "github.com/mikhail5545/media-service-go/internal/models/mux/metadata"
	muxtypes "github.com/mikhail5545/media-service-go/internal/models/mux/types"
	"github.com/mikhail5545/media-service-go/internal/util/parsing"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

func (s *Service) getAsset(ctx context.Context, id uuid.UUID, scopes []assetrepo.Scope) (*assetmodel.Asset, error) {
	asset, err := s.repo.Get(ctx, assetrepo.GetOptions{
		ID: id,
	}, scopes...)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve asset", zap.Error(err), zap.String("asset_id", id.String()))
		return nil, fmt.Errorf("failed to retrieve asset: %w", err)
	}
	return asset, nil
}

func (s *Service) getAssetMetadata(ctx context.Context, id uuid.UUID) (*metadatamodel.AssetMetadata, error) {
	metadata, err := s.metadataRepo.Get(ctx, id.String())
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, serviceerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve asset metadata", zap.Error(err), zap.String("asset_id", id.String()))
		return nil, fmt.Errorf("failed to retrieve asset metadata: %w", err)
	}
	return metadata, nil
}

func (s *Service) get(ctx context.Context, filter *assetmodel.GetFilter, scopes []assetrepo.Scope) (*assetmodel.Details, error) {
	if err := filter.Validate(); err != nil {
		return nil, serviceerrors.NewValidationFailedError(err)
	}
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

type assetSearchOptions struct {
	AssetID    string
	AssetUUID  *uuid.UUID
	GetOptions *assetrepo.GetOptions
}

func (s *Service) getAssetFromWebhook(ctx context.Context, txRepo *assetrepo.Repository, payload *muxtypes.MuxWebhook) *assetmodel.Asset {
	var searchOpt assetSearchOptions

	if payload.Data.Meta.ExternalID != nil {
		searchOpt = assetSearchOptions{
			AssetID: *payload.Data.Meta.ExternalID,
		}
	} else if payload.Data.UploadID != nil {
		searchOpt = assetSearchOptions{
			GetOptions: &assetrepo.GetOptions{
				MuxUploadID: *payload.Data.UploadID,
			},
		}
	} else if payload.Data.ID != "" {
		searchOpt = assetSearchOptions{
			GetOptions: &assetrepo.GetOptions{
				MuxAssetID: payload.Data.ID,
			},
		}
	} else {
		s.logger.Warn("received webhook with no identifiable asset information", zap.String("event_type", payload.Type), zap.String("event_id", payload.ID))
		return nil
	}

	asset, err := s.getInTx(ctx, txRepo, []string{}, searchOpt)
	if err != nil {
		s.logger.Error("failed to get asset from webhook", zap.Error(err), zap.String("event_type", payload.Type), zap.String("event_id", payload.ID))
		return nil
	}
	return asset
}

func (s *Service) getInTx(ctx context.Context, txRepo *assetrepo.Repository, fields []string, opt assetSearchOptions) (*assetmodel.Asset, error) {
	getOpt, err := retrieveAssetID(opt)
	if err != nil {
		return nil, err
	}
	getOpt.Fields = fields

	asset, err := txRepo.Get(ctx, *getOpt)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, serviceerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve asset in transaction", zap.Error(err), zap.String("asset_id", getOpt.ID.String()))
		return nil, fmt.Errorf("failed to retrieve asset in transaction: %w", err)
	}
	return asset, nil
}

func (s *Service) list(
	ctx context.Context,
	req *assetmodel.ListRequest,
	scopes []assetrepo.Scope,
) ([]*assetmodel.Details, string, error) {
	if err := req.Validate(); err != nil {
		return nil, "", serviceerrors.NewValidationFailedError(err)
	}
	listOptions := assetrepo.ListOptions{
		MuxUploadIDs:    req.MuxUploadIDs,
		MuxAssetIDs:     req.MuxAssetIDs,
		AspectRatios:    req.AspectRatios,
		ResolutionTiers: req.ResolutionTiers,
		IngestTypes:     req.IngestTypes,
		OrderBy:         req.OrderBy,
		OrderDir:        req.OrderDir,
		PageSize:        req.PageSize,
		PageToken:       req.PageToken,
		UploadStatuses:  req.UploadStatuses,
	}
	listOptions.IDs = parsing.StrToUUIDs(req.MuxAssetIDs)

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

func validateBeforeArchive(asset *assetmodel.Asset) error {
	if asset.Status == assetmodel.StatusArchived {
		return serviceerrors.NewConflictError("asset is already archived")
	}

	if asset.Status != assetmodel.StatusBroken || asset.UploadStatus != assetmodel.UploadStatusErrored {
		return serviceerrors.NewConflictError("cannot archive asset that is not broken or errored")
	}

	if asset.Status == assetmodel.StatusUploadURLGenerated {
		return serviceerrors.NewConflictError("you should wait for upload to complete before archiving")
	}
	return nil
}

func (s *Service) archiveAsset(ctx context.Context, txRepo *assetrepo.Repository, req *assetmodel.ChangeStateRequest, assetID uuid.UUID) error {
	adminID, err := parsing.StrToUUID(req.AdminID)
	if err != nil {
		return err
	}

	if _, err := txRepo.Archive(ctx, assetrepo.StateOperationOptions{IDs: uuid.UUIDs{assetID}}, types.AuditTrailOptions{
		AdminID:   adminID,
		AdminName: req.AdminName,
		Note:      req.Note,
	}); err != nil {
		s.logger.Error("failed to archive asset", zap.Error(err), zap.String("asset_id", req.ID))
		return fmt.Errorf("failed to archive asset: %w", err)
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

func (s *Service) archiveAssetOnWebhook(ctx context.Context, txRepo *assetrepo.Repository, asset *assetmodel.Asset, eventID string) error {
	if asset.ArchiveEventID != nil && *asset.ArchiveEventID == eventID {
		// Already archived for this event
		return nil
	}
	if _, err := txRepo.Archive(ctx, assetrepo.StateOperationOptions{IDs: uuid.UUIDs{asset.ID}}, types.AuditTrailOptions{
		AdminName: "system",
		EventID:   eventID,
		Note: "Received 'video.asset.deleted' webhook from MUX. " +
			"Archiving asset in the system to keep consistency with MUX.",
	}); err != nil {
		s.logger.Warn(
			"failed to archive asset from webhook",
			zap.Error(err),
			zap.String("asset_id", asset.ID.String()),
			zap.String("event_id", eventID),
		)
		return err
	}
	return nil
}

func (s *Service) deleteMetadataAndMuxAsset(ctx context.Context, assetID *uuid.UUID, muxAssetID *string) error {
	switch {
	case assetID != nil:
		if err := s.deleteAssetMetadata(ctx, *assetID); err != nil {
			return err
		}
	case muxAssetID != nil:
		if err := s.apiClient.DeleteAsset(ctx, *muxAssetID); err != nil {
			s.logger.Error("failed to delete mux asset", zap.Error(err), zap.String("mux_asset_id", *muxAssetID))
			return fmt.Errorf("failed to delete mux asset: %w", err)
		}
	}
	return nil
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
