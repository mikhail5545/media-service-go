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
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/mikhail5545/media-service-go/internal/clients/cloudinary"
	metarepo "github.com/mikhail5545/media-service-go/internal/database/arango/cloudinary/metadata"
	assetrepo "github.com/mikhail5545/media-service-go/internal/database/cloudinary/asset"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	metamodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/metadata"
	imageclient "github.com/mikhail5545/product-service-go/pkg/client/image"
	imagepb "github.com/mikhail5545/proto-go/proto/product_service/image/v0"
	"gorm.io/gorm"
)

// Service provides service-layer logic for Cloudinary asset management and asset models.
type Service interface {
	// Get retrieves a single not soft-deleted asset record from the database along with it's metadata.
	//
	// Returns a [assetmodel.AssetResponse] struct containing the combined information.
	// Returns an error if the ID is invalid (ErrInvalidArgument), the record is not found (ErrNotFound),
	// or a database/internal error occurs.
	Get(ctx context.Context, id string) (*assetmodel.AssetResponse, error)
	// GetWithDeleted retrieves a single asset record from the database along with it's metadata, including soft-deleted ones.
	//
	// Returns a [assetmodel.AssetResponse] struct containing the combined information.
	// Returns an error if the ID is invalid (ErrInvalidArgument), the record is not found (ErrNotFound),
	// or a database/internal error occurs.
	GetWithDeleted(ctx context.Context, id string) (*assetmodel.AssetResponse, error)
	// List retrieves a paginated list of all not soft-deleted asset records along with their metadata.
	//
	// Returns a slice of [assetmodel.AssetResponse] structs containing the combined information, the total count of such records,
	// and an error if one occurs.
	// Returns an error if a database/internal error occurs.
	List(ctx context.Context, limit, offset int) ([]assetmodel.AssetResponse, int64, error)
	// ListUnowned retrieves a paginated list of all unowned asset records along with their metadata.
	//
	// Returns a slice of [assetmodel.AssetResponse] structs containing the combined information, the total count of such records,
	// and an error if one occurs.
	// Returns an error if a database/internal error occurs.
	ListUnowned(ctx context.Context, limit, offset int) ([]assetmodel.AssetResponse, int64, error)
	// ListDeleted retrieves a paginated list of all soft-deleted asset records along with their metadata.
	//
	// Returns a slice of [assetmodel.AssetResponse] structs containing the combined information, the total count of such records,
	// and an error if one occurs.
	// Returns an error if a database/internal error occurs.
	ListDeleted(ctx context.Context, limit, offset int) ([]assetmodel.AssetResponse, int64, error)
	// CreateSignedUploadURL creates a signature for a direct frontend upload.
	// Direct upload url should be constructed using this params, this function only creates
	// signature for signed upload.
	//
	// Returns a map representation of upload params used during signature creation along with the signature itself.
	// Example: {"signature": "generated_signature", public_id: "asset_public_id", "timestamp": "unix_time", "api_key": "cloudinary_api_key"}.
	// Returns an error if request is invalid (http.StatusBadRequest) or internal error occures (http.StatusInternalServerError).
	CreateSignedUploadURL(ctx context.Context, req *assetmodel.CreateSignedUploadURLRequest) (map[string]string, error)
	// UpdateOwners processes asset ownership relations changes.
	// It recieves an updated list of asset owners, updates local DB metadata for asset (about it's owners),
	// processes the diff between old and new owners and notifies external services about this ownership
	// changes via gRPC connection.
	//
	// Returns an error if the request payload is invalid (ErrInvalidArgument), asset is not found (ErrNotFound),
	// or a database/internal error occures.
	UpdateOwners(ctx context.Context, req *assetmodel.UpdateOwnersRequest) error
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
	// SuccessfulUpload creates a new asset with provided information and creates owner relations for it.
	// It saves asset metadata about owner relations in the local noSQL db and notifies external services about
	// ownership changes via gRPC connection. This method should be called after successful cloudinary image upload.
	//
	// Returns newly created asset.
	// Returns an error if the request payload is invalid (ErrInvalidArgument) or a database/internal error occures.
	SuccessfulUpload(ctx context.Context, req *assetmodel.SuccessfulUploadRequest) (*assetmodel.AssetResponse, error)
	// CleanupOrphanAssets finds and deletes assets that exist in Cloudinary but not in the local database.
	//
	// Returns the number of cleaned assets.
	// Returns an error if the request payload is invalid (ErrInvalidArgument) or a database/internal error occures.
	CleanupOrphanAssets(ctx context.Context, req *assetmodel.CleanupOrphanAssetsRequest) (int, error)
	// Delete performs a soft-delete of an asset. It does not delete Cloudinary asset.
	// If assset has owners, it will be deassociated from them first.
	//
	// Returns an error if the ID is not a valid UUID (ErrInvalidArgument), asset not found (ErrNotFound)
	// or detabase/internal error occurs.
	Delete(ctx context.Context, assetID string) error
	// DeletePermanent performs a complete delete of an asset. It also deletes Cloudinary asset.
	// By this time, asset shouldn't have any owners. They should be deleted when asset is being soft-deleted.
	// This action is irreversable.
	//
	// Returns an error if the request payload is invalid (ErrInvalidArgument), asset not found (ErrNotFound),
	// or detabase/internal error occurs.
	DeletePermanent(ctx context.Context, req *assetmodel.DestroyAssetRequest) error
	// Restore performs a restore of an asset.
	//
	// Returns an error if the ID is not a valid UUID (ErrInvalidArgument), asset not found (ErrNotFound)
	// or detabase/internal error occurs.
	Restore(ctx context.Context, assetID string) error
	// HandleUploadWebhook processes an incoming Cloudinary upload webhook, finds the corresponding asset,
	// and updates it in a patch-like manner.
	HandleUploadWebhook(ctx context.Context, payload []byte, recievedTimestamp, recievedSignature string) error
}

// Service provides service-layer logic for Cloudinary asset management and asset models.
// It holds an instance of cloudinary API client to perform external API operations and
// instances of [assetrepo.Repository] to perform database operations.
type service struct {
	Client         cloudinary.Cloudinary
	Repo           assetrepo.Repository
	metaRepo       metarepo.Repository
	ImageSvcClient imageclient.Service
}

// New creates a new Service instance using provided cloudinary API client, asset and asset owner repositories.
func New(cnt cloudinary.Cloudinary, repo assetrepo.Repository, mr metarepo.Repository, img imageclient.Service) Service {
	return &service{
		Client:         cnt,
		Repo:           repo,
		metaRepo:       mr,
		ImageSvcClient: img,
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
		return nil, fmt.Errorf("failed to retrieve asset: %w", err)
	}

	metadata, err := s.metaRepo.Get(ctx, id)
	if err != nil && !errors.Is(err, metarepo.ErrNotFound) {
		return nil, fmt.Errorf("failed to retrieve asset metadata: %w", err)
	}

	response := s.combineAssetAndMetadata(asset, metadata)

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
		return nil, fmt.Errorf("failed to retrieve asset: %w", err)
	}

	metadata, err := s.metaRepo.Get(ctx, id)
	if err != nil && !errors.Is(err, metarepo.ErrNotFound) {
		return nil, fmt.Errorf("failed to retrieve asset metadata: %w", err)
	}

	response := s.combineAssetAndMetadata(asset, metadata)

	return response, nil
}

// List retrieves a paginated list of all not soft-deleted asset records along with their metadata.
//
// Returns a slice of [assetmodel.AssetResponse] structs containing the combined information, the total count of such records,
// and an error if one occurs.
// Returns an error if a database/internal error occurs.
func (s *service) List(ctx context.Context, limit, offset int) ([]assetmodel.AssetResponse, int64, error) {
	if limit < -1 || offset < 0 {
		return nil, 0, fmt.Errorf("%w: limit cannot be less then -1, offset cannot be less then 0", ErrInvalidArgument)
	}

	assets, err := s.Repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve assets: %w", err)
	}
	if len(assets) == 0 {
		return []assetmodel.AssetResponse{}, 0, nil
	}

	total, err := s.Repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count assets: %w", err)
	}

	assetIDs := make([]string, len(assets))
	for i, asset := range assets {
		assetIDs[i] = asset.ID
	}

	metadataMap, err := s.metaRepo.ListByKeys(ctx, assetIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve metadata for assets: %w", err)
	}

	responses := make([]assetmodel.AssetResponse, len(assets))
	for i, asset := range assets {
		responses[i] = *s.combineAssetAndMetadata(&asset, metadataMap[asset.ID])
	}

	return responses, total, nil
}

// ListUnowned retrieves a paginated list of all unowned asset records along with their metadata.
//
// Returns a slice of [assetmodel.AssetResponse] structs containing the combined information, the total count of such records,
// and an error if one occurs.
// Returns an error if a database/internal error occurs.
func (s *service) ListUnowned(ctx context.Context, limit, offset int) ([]assetmodel.AssetResponse, int64, error) {
	if limit < -1 || offset < 0 {
		return nil, 0, fmt.Errorf("%w: limit cannot be less then -1, offset cannot be less then 0", ErrInvalidArgument)
	}

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

	responses := make([]assetmodel.AssetResponse, len(assets))
	for i, asset := range assets {
		responses[i] = *s.combineAssetAndMetadata(&asset, metadataMap[asset.ID])
	}

	return responses, int64(len(unownedIDs)), nil
}

// ListDeleted retrieves a paginated list of all soft-deleted asset records along with their metadata.
//
// Returns a slice of [assetmodel.AssetResponse] structs containing the combined information, the total count of such records,
// and an error if one occurs.
// Returns an error if a database/internal error occurs.
func (s *service) ListDeleted(ctx context.Context, limit, offset int) ([]assetmodel.AssetResponse, int64, error) {
	if limit < -1 || offset < 0 {
		return nil, 0, fmt.Errorf("%w: limit cannot be less then -1, offset cannot be less then 0", ErrInvalidArgument)
	}

	assets, err := s.Repo.ListDeleted(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve assets: %w", err)
	}
	if len(assets) == 0 {
		return []assetmodel.AssetResponse{}, 0, nil
	}

	total, err := s.Repo.CountDeleted(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count assets: %w", err)
	}

	assetIDs := make([]string, len(assets))
	for i, asset := range assets {
		assetIDs[i] = asset.ID
	}

	metadataMap, err := s.metaRepo.ListByKeys(ctx, assetIDs)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to retrieve metadata for assets: %w", err)
	}

	responses := make([]assetmodel.AssetResponse, len(assets))
	for i, asset := range assets {
		responses[i] = *s.combineAssetAndMetadata(&asset, metadataMap[asset.ID])
	}

	return responses, total, nil
}

// Delete performs a soft-delete of an asset. It does not delete Cloudinary asset.
// If assset has owners, it will be deassociated from them first.
//
// Returns an error if the ID is not a valid UUID (ErrInvalidArgument), asset not found (ErrNotFound)
// or detabase/internal error occurs.
func (s *service) Delete(ctx context.Context, assetID string) error {
	if _, err := uuid.Parse(assetID); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.Repo.WithTx(tx)

		asset, err := txRepo.Get(ctx, assetID)
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
			toDelete := make(map[string][]string)
			for _, owner := range meta.Owners {
				toDelete[owner.OwnerType] = append(toDelete[owner.OwnerType], owner.OwnerID)
			}

			if err := s.processChanges(ctx, asset, nil, toDelete); err != nil {
				return fmt.Errorf("failed to notify external services about changes: %w", err)
			}

			if err := s.metaRepo.DeleteOwners(ctx, asset.ID); err != nil && !errors.Is(err, metarepo.ErrNotFound) {
				return fmt.Errorf("failed to delete asset owners metadata: %w", err)
			}
		}

		_, err = txRepo.Delete(ctx, assetID)
		return err
	})
}

// DeletePermanent performs a complete delete of an asset. It also deletes Cloudinary asset.
// By this time, asset shouldn't have any owners. They should be deleted when asset is being soft-deleted.
// This action is irreversable.
//
// Returns an error if the request payload is invalid (ErrInvalidArgument), asset not found (ErrNotFound),
// or detabase/internal error occurs.
func (s *service) DeletePermanent(ctx context.Context, req *assetmodel.DestroyAssetRequest) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.Repo.WithTx(tx)

		asset, err := txRepo.Get(ctx, req.ID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: %w", ErrNotFound, err)
			}
			return fmt.Errorf("failed to retrieve asset: %w", err)
		}

		if _, err := txRepo.DeletePermanent(ctx, req.ID); err != nil {
			return fmt.Errorf("failed to delete asset: %w", err)
		}

		if err := s.Client.DeleteAsset(ctx, asset.CloudinaryPublicID, req.ResourceType); err != nil {
			return fmt.Errorf("failed to delete cloudinary asset: %w", err)
		}
		return nil
	})
}

// Restore performs a restore of an asset.
//
// Returns an error if the ID is not a valid UUID (ErrInvalidArgument), asset not found (ErrNotFound)
// or detabase/internal error occurs.
func (s *service) Restore(ctx context.Context, assetID string) error {
	if _, err := uuid.Parse(assetID); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	return s.Repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.Repo.WithTx(tx)
		ra, err := txRepo.Restore(ctx, assetID)
		if err != nil {
			return fmt.Errorf("failed to restore asset: %w", err)
		}
		if ra == 0 {
			return fmt.Errorf("%w: %w", ErrNotFound, err)
		}
		return nil
	})
}

// CreateSignedUploadURL creates a signature for a direct frontend upload.
// Direct upload url should be constructed using this params, this function only creates
// signature for signed upload.
//
// Returns a map representation of upload params used during signature creation along with the signature itself.
// Example: {"signature": "generated_signature", public_id: "asset_public_id", "timestamp": "unix_time", "api_key": "cloudinary_api_key"}.
// Returns an error if request is invalid (cloudinary.ErrInvalidArgument), Cloudinary API error occures (cloudinary.ErrCloudinaryAPI)
// or internal error occures.
func (s *service) CreateSignedUploadURL(ctx context.Context, req *assetmodel.CreateSignedUploadURLRequest) (map[string]string, error) {
	signedParams := make(map[string]string)

	timestamp := fmt.Sprintf("%d", time.Now().Unix())
	params := make(url.Values)
	if req.Eager != nil {
		params.Set("eager", *req.Eager)
		signedParams["eager"] = *req.Eager
	}
	params.Set("public_id", req.PublicID)
	params.Set("timestamp", timestamp)
	signature, err := s.Client.SignUploadParams(ctx, params)
	if err != nil {
		return nil, err
	}
	apiKey := s.Client.GetApiKey()

	signedParams["signature"] = signature
	signedParams["public_id"] = req.PublicID
	signedParams["timestamp"] = timestamp
	signedParams["api_key"] = apiKey
	return signedParams, nil
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

	// Check if asset exists in Postgres
	asset, err := s.Repo.Get(ctx, req.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: %w", ErrNotFound, err)
		}
		return fmt.Errorf("failed to retrieve asset: %w", err)
	}

	// Get current owners from ArangoDB
	currentMetadata, err := s.metaRepo.Get(ctx, asset.ID)
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
	if err := s.metaRepo.UpdateOwners(ctx, asset.ID, req.Owners); err != nil {
		return fmt.Errorf("failed to update asset metadata in ArangoDB: %w", err)
	}

	// After successful DB update, notify other services via gRPC
	if err := s.processChanges(ctx, asset, toAdd, toDelete); err != nil {
		return fmt.Errorf("failed to notify external services: %w", err)
	}
	return nil
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

		if err := s.metaRepo.UpdateOwners(ctx, asset.ID, newOwners); err != nil {
			return fmt.Errorf("failed to update asset metadata: %w", err)
		}

		// Associate owner with the asset
		if _, err := s.ImageSvcClient.Add(ctx, &imagepb.AddRequest{
			PublicId:       asset.CloudinaryPublicID,
			Url:            asset.URL,
			SecureUrl:      asset.SecureURL,
			MediaServiceId: asset.ID,
			OwnerId:        req.OwnerID,
			OwnerType:      req.OwnerType,
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
		txRepo := s.Repo.WithTx(tx)

		// Ensure asset exists
		_, err := txRepo.Get(ctx, req.ID)
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
		if err := s.metaRepo.UpdateOwners(ctx, req.ID, newOwners); err != nil {
			return fmt.Errorf("failed to update asset metadata: %w", err)
		}

		// Notify other services
		if _, err := s.ImageSvcClient.Delete(ctx, &imagepb.DeleteRequest{
			MediaServiceId: req.ID,
			OwnerId:        req.OwnerID,
			OwnerType:      req.OwnerType,
		}); err != nil {
			return handleGRPCError(err)
		}

		return nil
	})
}

// SuccessfulUpload creates a new asset with provided information and creates owner relations for it.
// It saves asset metadata about owner relations in the local noSQL db and notifies external services about
// ownership changes via gRPC connection. This method should be called after successful cloudinary image upload.
//
// Returns newly created asset.
// Returns an error if the request payload is invalid (ErrInvalidArgument) or a database/internal error occures.
func (s *service) SuccessfulUpload(ctx context.Context, req *assetmodel.SuccessfulUploadRequest) (*assetmodel.AssetResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	newAsset := &assetmodel.Asset{
		ID:                 uuid.New().String(),
		CloudinaryAssetID:  req.CloudinaryAssetID,
		CloudinaryPublicID: req.CloudinaryPublicID,
		ResourceType:       req.ResourceType,
		Format:             req.Format,
		Width:              req.Width,
		Height:             req.Height,
		URL:                req.URL,
		SecureURL:          req.SecureURL,
		AssetFolder:        req.AssetFolder,
		DisplayName:        req.DisplayName,
	}

	if err := s.Repo.Create(ctx, newAsset); err != nil {
		return nil, fmt.Errorf("failed to create asset record: %w", err)
	}

	// Asset may be created without owners initially.
	if len(req.Owners) > 0 {
		if err := s.metaRepo.UpdateOwners(ctx, newAsset.ID, req.Owners); err != nil {
			return nil, fmt.Errorf("failed to create asset owners metadata: %w", err)
		}
	}

	toAdd := make(map[string][]string)
	for _, owner := range req.Owners {
		toAdd[owner.OwnerType] = append(toAdd[owner.OwnerType], owner.OwnerID)
	}

	if err := s.processChanges(ctx, newAsset, toAdd, nil); err != nil {
		return nil, fmt.Errorf("failed to notify external services: %w", err)
	}

	response := s.combineAssetAndMetadata(newAsset, &metamodel.AssetMetadata{Key: newAsset.ID, Owners: req.Owners})

	return response, nil
}

// CleanupOrphanAssets finds and deletes assets that exist in Cloudinary but not in the local database.
//
// Returns the number of cleaned assets.
// Returns an error if the request payload is invalid (ErrInvalidArgument) or a database/internal error occures.
func (s *service) CleanupOrphanAssets(ctx context.Context, req *assetmodel.CleanupOrphanAssetsRequest) (int, error) {
	if err := req.Validate(); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrInvalidArgument, err)
	}

	cldAssets, err := s.Client.ListAssetsByFolder(ctx, req.Folder)
	if err != nil {
		return 0, err
	}

	localAssetIDs, err := s.Repo.ListAllCloudinaryAssetIDs(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to list assets from database: %w", err)
	}

	var orphansToDelete []string
	for _, asset := range cldAssets {
		if _, exists := localAssetIDs[asset.AssetID]; !exists { // If asset not exists
			orphansToDelete = append(orphansToDelete, asset.PublicID)
		}
	}

	if len(orphansToDelete) == 0 {
		log.Println("Orphan asset cleanup: No orphan assets found.")
		return 0, nil
	}

	log.Printf("Orphan asset cleanup: Found %d orphan(s) to delete.", len(orphansToDelete))

	if err := s.Client.DeleteAssets(ctx, "image", orphansToDelete); err != nil {
		return 0, fmt.Errorf("failed to delete assets: %w", err)
	}
	return len(orphansToDelete), nil
}
