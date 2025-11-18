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

	"github.com/google/uuid"
	"github.com/mikhail5545/media-service-go/internal/clients/mux"
	metarepo "github.com/mikhail5545/media-service-go/internal/database/arango/mux/metadata"
	assetrepo "github.com/mikhail5545/media-service-go/internal/database/mux/asset"
	detailrepo "github.com/mikhail5545/media-service-go/internal/database/mux/detail"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	metamodel "github.com/mikhail5545/media-service-go/internal/models/mux/metadata"
	videoservice "github.com/mikhail5545/product-service-go/pkg/client/video"
	videopb "github.com/mikhail5545/proto-go/proto/product_service/video/v0"
	muxgo "github.com/muxinc/mux-go/v6"
	"gorm.io/gorm"
)

// Service provides service-layer business logic for mux assets.
// It acts as adapter between repository-layer CRUD logic [muxrepo.Repository] and
// handler layers/server layers.
type Service interface {
	// Get retrieves a single published and not soft-deleted mux upload record from the database along with it's metadata.
	//
	// Returns a [assetmodel.AssetResponse] struct containing the combined information.
	// Returns an error if the ID is invalid (ErrInvalidArgument), the record is not found (ErrNotFound),
	// or a database/internal error occurs.
	Get(ctx context.Context, id string) (*assetmodel.AssetResponse, error)
	// GetWithDeleted retrieves a single mux upload record from the database along with it's metadata, including soft-deleted ones.
	//
	// Returns a [assetmodel.AssetResponse] struct containing the combined information.
	// Returns an error if the ID is invalid (ErrInvalidArgument), the record is not found (ErrNotFound),
	// or a database/internal error occurs.
	GetWithDeleted(ctx context.Context, id string) (*assetmodel.AssetResponse, error)
	// List retrieves a paginated list of all published and not soft-deleted mux upload records along with their metadata.
	//
	// Returns a slice of [assetmodel.AssetResponse] structs containing the combined information, the total count of such records,
	// and an error if one occurs.
	// Returns an error if a database/internal error occurs.
	List(ctx context.Context, limit, offset int) ([]assetmodel.AssetResponse, int64, error)
	// ListUnowned retrieves a paginated list of all unowned mux upload records and their metadata.
	//
	// Returns a slice of [assetmodel.AssetResponse] structs containing the combined information, the total count of such records,
	// and an error if one occurs.
	// Returns an error if a database/internal error occurs.
	ListUnowned(ctx context.Context, limit, offset int) ([]assetmodel.AssetResponse, int64, error)
	// ListDeleted retrieves a paginated list of all soft-deleted mux upload records and their metadata.
	//
	// Returns a slice of [assetmodel.AssetResponse] structs containing the combined information, the total count of such records,
	// and an error if one occurs.
	// Returns an error if a database/internal error occurs.
	ListDeleted(ctx context.Context, limit, offset int) ([]assetmodel.AssetResponse, int64, error)
	// Delete performs a soft delete of an asset.
	// It should be called only for assets that don't have any owner or association.
	//
	// Returns an error if the ID is invalid (ErrInvalidArgument), the records are not found (ErrNotFound),
	// or a database/internal error occurs.
	Delete(ctx context.Context, id string) error
	// DeletePermanent performs a complete delete of a mux upload.
	// It also deletes mux asset via MUX Direct Upload API if `upload.MuxAssetId` is populated.
	//
	// Returns an error if the ID is invalid (ErrInvalidArgument), the records are not found (ErrNotFound),
	// delete of MUX asset failed (http.StatusServiceUnavailable),
	// or a database/internal error occurs.
	DeletePermanent(ctx context.Context, id string) error
	// Restore performs a restore of a mux upload record.
	// Mux upload record is not being published. This should be
	// done manually.
	//
	// Returns an error if the ID is invalid (ErrInvalidArgument), the records are not found (ErrNotFound),
	// or a database/internal error occurs.
	Restore(ctx context.Context, id string) error
	// CreateUploadURL creates upload URL for the direct upload using mux direct upload api.
	// It uses [mux.Client.CreateUploadURL] method to access MUX direct upload API.
	// If an owner already has an association with an asset, an error is returned.
	//
	// Returns a muxgo.UploadResponse struct on success.
	// Returns an error if the request payload is invalid (ErrInvalidArgument), if the owner already has an asset (ErrOwnerHasAsset),
	// or if a MUX API, database, or gRPC error occurs.
	CreateUploadURL(ctx context.Context, req *assetmodel.CreateUploadURLRequest) (*muxgo.UploadResponse, error)
	// CreateUnownedUploadURL creates an upload URL for a new asset without an initial owner.
	//
	// Returns a muxgo.UploadResponse struct on success.
	// Returns an error if the request payload is invalid (ErrInvalidArgument),
	// or a database/internal error occurs.
	CreateUnownedUploadURL(ctx context.Context, req *assetmodel.CreateUnownedUploadURLRequest) (*muxgo.UploadResponse, error)
	// Associate links an existing asset to an owner.
	// It also updates asset medatada.
	//
	// Returns an error if the request payload is invalid (ErrInvalidArgument), the records are not found (ErrNotFound),
	// or a database/internal error occurs.
	Associate(ctx context.Context, req *assetmodel.AssociateRequest) error
	// Deassociate removes the link between an asset and an owner.
	// It also deletes owner from asset metadata.
	//
	// Returns an error if the request payload is invalid (ErrInvalidArgument), the records are not found (ErrNotFound),
	// or a database/internal error occurs.
	Deassociate(ctx context.Context, req *assetmodel.DeassociateRequest) error
	// UpdateOwners processes asset ownership relations changes.
	// It recieves an updated list of asset owners, updates local DB metadata for asset (about it's owners),
	// processes the diff between old and new owners and notifies external services about this ownership
	// changes via gRPC connection.
	//
	// Returns an error if the request payload is invalid (ErrInvalidArgument), asset is not found (ErrNotFound),
	// or a database/internal error occures.
	UpdateOwners(ctx context.Context, req *assetmodel.UpdateOwnersRequest) error
	// HandleAssetCreatedWebhook processes an incoming Mux webhook with "video.asset.created" event type, finds the corresponding asset,
	// and updates it in a patch-like manner.
	HandleAssetCreatedWebhook(ctx context.Context, payload *assetmodel.MuxWebhook) error
	// HandleAssetReadyWebhook processes an incoming Mux webhook with "video.asset.ready" event type, finds the corresponding asset,
	// and updates it in a patch-like manner.
	HandleAssetReadyWebhook(ctx context.Context, payload *assetmodel.MuxWebhook) error
	// HandleAssetErroredWebhook processes an incoming Mux webhook with "video.asset.errored" event type, finds the corresponding asset,
	// and updates it in a patch-like manner. After update, it soft-deleted mux asset. If asset has owners, they will be deassociated and
	// all asset metadata about it's owners will be cleared.
	HandleAssetErroredWebhook(ctx context.Context, payload *assetmodel.MuxWebhook) error
}

// service provides service-layer business logic for mux assets.
// It acts as adapter between repository-layer CRUD logic [muxrepo.Repository] and
// handler layers/server layers.
type service struct {
	// Repo represents repository-layer logic for CRUD operations.
	Repo assetrepo.Repository
	// metaRepo represents repository-layer logic for asset's metadata CRUD opearatations.
	metaRepo metarepo.Repository
	// detailRepo represents repository-layer logic for asset's details CRUD operations.
	detailRepo detailrepo.Repository
	// Client represents MUX API client for direct asset management.
	Client         mux.MUX
	VideoSvcClient videoservice.Service
}

// New creates new instance of a [mux.service]
func New(
	repo assetrepo.Repository,
	mr metarepo.Repository,
	dr detailrepo.Repository,
	client mux.MUX,
	vsc videoservice.Service,
) Service {
	return &service{
		Repo:           repo,
		metaRepo:       mr,
		detailRepo:     dr,
		Client:         client,
		VideoSvcClient: vsc,
	}
}

// Get retrieves a single not soft-deleted asset record from the database along with it's metadata.
//
// Returns a [assetmodel.AssetResponse] struct containing the combined information.
// Returns an error if the ID is invalid (ErrInvalidArgument), the record is not found (ErrNotFound),
// or a database/internal error occurs.
func (s *service) Get(ctx context.Context, id string) (*assetmodel.AssetResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}
	asset, err := s.Repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
		}
		return nil, fmt.Errorf("failed to retrieve mux upload: %w", err)
	}

	metadata, err := s.metaRepo.Get(ctx, id)
	if err != nil && !errors.Is(err, metarepo.ErrNotFound) {
		return nil, fmt.Errorf("failed to retrieve asset metadata: %w", err)
	}

	details, err := s.detailRepo.Get(ctx, id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to retrieve asset details: %w", err)
	}

	response := s.combineAssetAndMetadata(asset, metadata, details)

	return response, nil
}

// GetWithDeleted retrieves a single asset record from the database along with it's metadata, including soft-deleted ones.
//
// Returns a [assetmodel.AssetResponse] struct containing the combined information.
// Returns an error if the ID is invalid (ErrInvalidArgument), the record is not found (ErrNotFound),
// or a database/internal error occurs.
func (s *service) GetWithDeleted(ctx context.Context, id string) (*assetmodel.AssetResponse, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}
	asset, err := s.Repo.GetWithDeleted(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
		}
		return nil, fmt.Errorf("failed to retrieve mux upload: %w", err)
	}

	metadata, err := s.metaRepo.Get(ctx, id)
	if err != nil && !errors.Is(err, metarepo.ErrNotFound) {
		return nil, fmt.Errorf("failed to retrieve asset metadata: %w", err)
	}

	details, err := s.detailRepo.Get(ctx, id)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to retrieve asset details: %w", err)
	}

	response := s.combineAssetAndMetadata(asset, metadata, details)

	return response, nil
}

// List retrieves a paginated list of all not soft-deleted asset records along with their metadata.
//
// Returns a slice of [assetmodel.AssetResponse] structs containing the combined information, the total count of such records,
// and an error if one occurs.
// Returns an error if a database/internal error occurs.
func (s *service) List(ctx context.Context, limit, offset int) ([]assetmodel.AssetResponse, int64, error) {
	assets, err := s.Repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve mux assets: %w", err)
	}
	if len(assets) == 0 {
		return []assetmodel.AssetResponse{}, 0, nil
	}

	total, err := s.Repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count mux assets: %w", err)
	}

	assetIDs := make([]string, len(assets))
	for i, asset := range assets {
		assetIDs[i] = asset.ID
	}

	metadataMap, err := s.metaRepo.ListByKeys(ctx, assetIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve metadata for assets: %w", err)
	}

	detailMap, err := s.detailRepo.ListByAssetIDs(ctx, assetIDs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve details for assets: %w", err)
	}

	responses := make([]assetmodel.AssetResponse, len(assets))
	for i, asset := range assets {
		responses[i] = *s.combineAssetAndMetadata(&asset, metadataMap[asset.ID], detailMap[asset.ID])
	}

	return responses, total, nil
}

// ListUnowned retrieves a paginated list of all unowned asset records and their metadata.
//
// Returns a slice of [assetmodel.AssetResponse] structs containing the combined information, the total count of such records,
// and an error if one occurs.
// Returns an error if a database/internal error occurs.
func (s *service) ListUnowned(ctx context.Context, limit, offset int) ([]assetmodel.AssetResponse, int64, error) {
	unownedIDs, err := s.metaRepo.ListUnownedIDs(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve unowned asset IDs: %w", err)
	}
	if len(unownedIDs) == 0 {
		return []assetmodel.AssetResponse{}, 0, nil
	}

	assets, err := s.Repo.ListByIDs(ctx, limit, offset, unownedIDs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve unowned assets by IDs: %w", err)
	}

	assetIDs := make([]string, len(assets))
	for i := range assets {
		assetIDs[i] = assets[i].ID
	}

	metadataMap, err := s.metaRepo.ListByKeys(ctx, assetIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve metadata for assets: %w", err)
	}

	detailMap, err := s.detailRepo.ListByAssetIDs(ctx, assetIDs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve details for unowned assets: %w", err)
	}

	responses := make([]assetmodel.AssetResponse, len(assets))
	for i, asset := range assets {
		responses[i] = *s.combineAssetAndMetadata(&asset, metadataMap[asset.ID], detailMap[asset.ID])
	}

	return responses, int64(len(unownedIDs)), nil
}

// ListDeleted retrieves a paginated list of all soft-deleted asset records and their metadata.
//
// Returns a slice of [assetmodel.AssetResponse] structs containing the combined information, the total count of such records,
// and an error if one occurs.
// Returns an error if a database/internal error occurs.
func (s *service) ListDeleted(ctx context.Context, limit, offset int) ([]assetmodel.AssetResponse, int64, error) {
	assets, err := s.Repo.ListDeleted(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve mux assets: %w", err)
	}
	if len(assets) == 0 {
		return []assetmodel.AssetResponse{}, 0, nil
	}

	total, err := s.Repo.CountDeleted(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count mux assets: %w", err)
	}

	assetIDs := make([]string, len(assets))
	for i := range assets {
		assetIDs[i] = assets[i].ID
	}

	metadataMap, err := s.metaRepo.ListByKeys(ctx, assetIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve metadata for assets: %w", err)
	}

	detailMap, err := s.detailRepo.ListByAssetIDs(ctx, assetIDs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve details for deleted assets: %w", err)
	}

	responses := make([]assetmodel.AssetResponse, len(assets))
	for i, asset := range assets {
		responses[i] = *s.combineAssetAndMetadata(&asset, metadataMap[asset.ID], detailMap[asset.ID])
	}

	return responses, total, nil
}

// DeletePermanent performs a complete delete of an asset.
// It also deletes mux asset via MUX Direct Upload API if `upload.MuxAssetId` is populated.
//
// Returns an error if the ID is invalid (ErrInvalidArgument), the records are not found (ErrNotFound),
// delete of MUX asset failed (http.StatusServiceUnavailable),
// or a database/internal error occurs.
func (s *service) DeletePermanent(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}
	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.Repo.WithTx(tx)

		asset, err := txRepo.Get(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: %w", ErrNotFound, err)
			}
			return fmt.Errorf("failed to retrieve mux upload: %w", err)
		}
		if asset.MuxAssetID != nil {
			if err := s.Client.DeleteAsset(*asset.MuxAssetID); err != nil {
				return err
			}
		}
		// Completely clear asset metadata in the ArangoDB.
		if err := s.metaRepo.Delete(ctx, asset.ID); err != nil {
			return fmt.Errorf("failed to delete asset metadata: %w", err)
		}
		// Delete asset from Postgres DB.
		if _, err := txRepo.DeletePermanent(ctx, id); err != nil {
			return fmt.Errorf("failed to delete mux upload: %w", err)
		}
		return nil
	})
}

// Delete performs a soft delete of an asset.
// If asset has any owners, they will be deassociated and local asset metadata about owhership will be deleted.
//
// Returns an error if the ID is invalid (ErrInvalidArgument), the records are not found (ErrNotFound),
// or a database/internal error occurs.
func (s *service) Delete(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}
	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.Repo.WithTx(tx)

		asset, err := txRepo.Get(ctx, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: %w", ErrNotFound, err)
			}
			return fmt.Errorf("failed to retrieve asset: %w", err)
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
		if _, err := s.Repo.WithTx(tx).Delete(ctx, id); err != nil {
			return fmt.Errorf("failed to delete mux upload: %w", err)
		}
		return nil
	})
}

// Restore performs a restore of an asset.
//
// Returns an error if the ID is invalid (ErrInvalidArgument), the records are not found (ErrNotFound),
// or a database/internal error occurs.
func (s *service) Restore(ctx context.Context, id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}
	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		ra, err := s.Repo.WithTx(tx).Restore(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to restore mux upload: %w", err)
		} else if ra == 0 {
			return fmt.Errorf("%w: %w", ErrNotFound, err)
		}
		return nil
	})
}

// CreateUploadURL creates new signed upload url to upload a new asset.
// It uses [muxclient.Client.CreateUploadURL] method to access MUX direct upload API.
// This method can be called only for owners without associated asset. If owner has an asset,
// they should be deassociated first.
//
// Returns muxgo.UploadResponse struct on success.
// Returns an error if the request payload is invalid (ErrInvalidArgument), if owner already associated with an asset (ErrOwnerHasAsset),
// or a database/internal error occures.
func (s *service) CreateUploadURL(ctx context.Context, req *assetmodel.CreateUploadURLRequest) (*muxgo.UploadResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	var response *muxgo.UploadResponse
	err := s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.Repo.WithTx(tx)

		getResponse, err := s.VideoSvcClient.GetOwner(ctx, &videopb.GetOwnerRequest{OwnerId: req.OwnerID, OwnerType: req.OwnerType})
		if err != nil {
			return handleGRPCError(err)
		}

		if getResponse.Owner.VideoId != nil {
			return ErrOwnerHasAsset
		}

		data, err := s.Client.CreateUploadURL(req.CreatorID, req.Title)
		if err != nil {
			return err
		}
		response = data

		newAsset := &assetmodel.Asset{
			ID:          uuid.New().String(),
			MuxUploadID: &data.Data.Id,
			MuxAssetID:  &data.Data.AssetId,
			State:       "url_upload_created",
		}

		if err := txRepo.Create(ctx, newAsset); err != nil {
			return fmt.Errorf("failed to create new asset: %w", err)
		}

		newOwners := []metamodel.Owner{{OwnerID: req.OwnerID, OwnerType: req.OwnerType}}

		newMetadata := &metamodel.AssetMetadata{
			Key:       newAsset.ID,
			CreatorID: req.CreatorID,
			Title:     req.Title,
			Owners:    newOwners,
		}

		if err := s.metaRepo.Create(ctx, newMetadata); err != nil {
			return fmt.Errorf("failed to create new asset metadata: %w", err)
		}

		if _, err := s.VideoSvcClient.Add(ctx, &videopb.AddRequest{
			OwnerId:        req.OwnerID,
			OwnerType:      req.OwnerType,
			MediaServiceId: newAsset.ID,
		}); err != nil {
			return handleGRPCError(err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

// CreateUnownedUploadURL creates an upload URL for a new asset without an initial owner.
//
// Returns a muxgo.UploadResponse struct on success.
// Returns an error if the request payload is invalid (ErrInvalidArgument),
// or a database/internal error occurs.
func (s *service) CreateUnownedUploadURL(ctx context.Context, req *assetmodel.CreateUnownedUploadURLRequest) (*muxgo.UploadResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	var response *muxgo.UploadResponse
	err := s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.Repo.WithTx(tx)

		data, err := s.Client.CreateUploadURL(req.CreatorID, req.Title)
		if err != nil {
			return err
		}
		response = data

		newAsset := &assetmodel.Asset{
			ID:          uuid.New().String(),
			MuxUploadID: &data.Data.Id,
			MuxAssetID:  &data.Data.AssetId,
			State:       "url_upload_created",
		}

		if err := txRepo.Create(ctx, newAsset); err != nil {
			return fmt.Errorf("failed to create new asset: %w", err)
		}

		newMetadata := &metamodel.AssetMetadata{
			Key:       newAsset.ID,
			CreatorID: req.CreatorID,
			Title:     req.Title,
			Owners:    []metamodel.Owner{}, // No owners initially
		}

		if err := s.metaRepo.Create(ctx, newMetadata); err != nil {
			return fmt.Errorf("failed to create new asset metadata: %w", err)
		}

		return nil
	})

	return response, err
}

// Associate links an existing asset to an owner.
// It also updates asset medatada.
//
// Returns an error if the request payload is invalid (ErrInvalidArgument), the records are not found (ErrNotFound),
// or a database/internal error occurs.
func (s *service) Associate(ctx context.Context, req *assetmodel.AssociateRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.Repo.WithTx(tx)

		getResponse, err := s.VideoSvcClient.GetOwner(ctx, &videopb.GetOwnerRequest{OwnerId: req.OwnerID, OwnerType: req.OwnerType})
		if err != nil {
			return handleGRPCError(err)
		}

		if getResponse.Owner.VideoId != nil {
			return ErrOwnerHasAsset
		}

		// Retrieve asset from the database
		asset, err := txRepo.Get(ctx, req.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: %w", ErrNotFound, err)
			}
			return fmt.Errorf("failed to retrieve asset: %w", err)
		}

		// Retrieve asset metadata from ArangoDB
		currentMetadata, err := s.metaRepo.Get(ctx, asset.ID)
		if err != nil {
			return fmt.Errorf("failed to retrieve asset metadata: %w", err)
		}

		var newOwners []metamodel.Owner
		newOwners = append(newOwners, currentMetadata.Owners...)
		newOwners = append(newOwners, metamodel.Owner{
			OwnerID:   req.OwnerID,
			OwnerType: req.OwnerType,
		})

		if err := s.metaRepo.Update(ctx, asset.ID, &metamodel.AssetMetadata{
			Owners: newOwners,
		}); err != nil {
			return fmt.Errorf("failed to update asset metadata: %w", err)
		}

		// Associate owner with the asset
		if _, err = s.VideoSvcClient.Add(ctx, &videopb.AddRequest{
			OwnerId:        req.OwnerID,
			OwnerType:      req.OwnerType,
			MediaServiceId: req.ID,
		}); err != nil {
			return handleGRPCError(err)
		}
		return nil
	})
}

// Deassociate removes the link between an asset and an owner.
// It also deletes owner from asset metadata.
//
// Returns an error if the request payload is invalid (ErrInvalidArgument), the records are not found (ErrNotFound),
// or a database/internal error occurs.
func (s *service) Deassociate(ctx context.Context, req *assetmodel.DeassociateRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		// Ensure asset exists
		_, err := s.Repo.WithTx(tx).Get(ctx, req.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: %w", ErrNotFound, err)
			}
			return fmt.Errorf("failed to retrieve asset: %w", err)
		}

		// Get current owners
		currentMetadata, err := s.metaRepo.Get(ctx, req.ID)
		if err != nil {
			if errors.Is(err, metarepo.ErrNotFound) {
				// Asset has no owners, nothing to deassociate.
				return nil
			}
			return fmt.Errorf("failed to retrieve asset metadata: %w", err)
		}

		// Remove the specified owner from the list
		var newOwners []metamodel.Owner
		for _, owner := range currentMetadata.Owners {
			if owner.OwnerID == req.OwnerID && owner.OwnerType == req.OwnerType {
				continue // Skip the owner to be removed
			}
			newOwners = append(newOwners, owner)
		}

		// Update metadata in ArangoDB
		if err := s.metaRepo.Update(ctx, req.ID, &metamodel.AssetMetadata{Owners: newOwners}); err != nil {
			return fmt.Errorf("failed to update asset metadata: %w", err)
		}

		// Notify other services
		if _, err = s.VideoSvcClient.Remove(ctx, &videopb.RemoveRequest{
			OwnerId: req.OwnerID, OwnerType: req.OwnerType,
		}); err != nil {
			return handleGRPCError(err)
		}
		return nil
	})
}

// UpdateOwners processes asset ownership relations changes.
// It recieves an updated list of asset owners, updates local DB metadata for asset (about it's owners),
// processes the diff between old and new owners and notifies external services about this ownership
// changes via gRPC connection.
//
// Returns an error if the request payload is invalid (ErrInvalidArgument), asset is not found (ErrNotFound),
// or a database/internal error occures.
func (s *service) UpdateOwners(ctx context.Context, req *assetmodel.UpdateOwnersRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	// Ensure asset exists in Postgres before updating metadata in ArangoDB
	asset, err := s.Repo.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: %w", ErrNotFound, err)
		}
		return fmt.Errorf("failed to retrieve asset: %w", err)
	}

	currentMetadata, err := s.metaRepo.Get(ctx, req.ID)
	var currentOwners []metamodel.Owner
	if err != nil && !errors.Is(err, metarepo.ErrNotFound) {
		return fmt.Errorf("failed to get asset owners metadata: %w", err)
	} else if errors.Is(err, metarepo.ErrNotFound) {
		// Not found is a valid case, it just means there are no owners yet.
	} else if currentMetadata != nil {
		currentOwners = currentMetadata.Owners
	}

	currentOwnerMap := groupOwnersByTypeFromMetadata(currentOwners)
	newOwnerMap := groupOwnersByTypeFromMetadata(req.Owners)

	// Calculate what to add and what to delete
	toAdd, toDelete := diffOwnerMaps(currentOwnerMap, newOwnerMap)

	// Update assest metadata (owners) in ArangoDB
	if err := s.metaRepo.Update(ctx, req.ID, &metamodel.AssetMetadata{
		Owners: req.Owners,
	}); err != nil {
		return fmt.Errorf("failed to update asset metadata in ArangoDB: %w", err)
	}

	// After successful DB update, notify other services via gRPC
	if err := s.processChanges(ctx, asset, toAdd, toDelete); err != nil {
		return fmt.Errorf("failed to notify external services: %w", err)
	}
	return nil
}
