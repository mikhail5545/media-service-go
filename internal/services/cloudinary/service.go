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
	"net/url"
	"strconv"
	"time"

	"github.com/google/uuid"
	apiclient "github.com/mikhail5545/media-service-go/internal/apiclients/cloudinary"
	metadatarepo "github.com/mikhail5545/media-service-go/internal/database/mongo/cloudinary/metadata"
	assetrepo "github.com/mikhail5545/media-service-go/internal/database/postgres/cloudinary/asset"
	"github.com/mikhail5545/media-service-go/internal/database/types"
	serviceerrors "github.com/mikhail5545/media-service-go/internal/errors"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	metadatamodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/metadata"
	"github.com/mikhail5545/media-service-go/internal/util/parsing"
	"github.com/mikhail5545/product-service-client/client"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type AssetService interface {
	// Get retrieves an active asset based on the provided filter.
	Get(ctx context.Context, filter *assetmodel.GetFilter) (*assetmodel.Details, error)
	// GetWithArchived retrieves an asset that can be either active or archived based on the provided filter.
	GetWithArchived(ctx context.Context, filter *assetmodel.GetFilter) (*assetmodel.Details, error)
	// GetWithBroken retrieves an asset that can be active, archived, or broken based on the provided filter.
	GetWithBroken(ctx context.Context, filter *assetmodel.GetFilter) (*assetmodel.Details, error)
	// List retrieves a list of active assets based on the provided request.
	List(ctx context.Context, req *assetmodel.ListRequest) ([]*assetmodel.Details, string, error)
	// ListArchived retrieves a list of archived assets based on the provided request.
	ListArchived(ctx context.Context, req *assetmodel.ListRequest) ([]*assetmodel.Details, string, error)
	// ListBroken retrieves a list of broken assets based on the provided request.
	ListBroken(ctx context.Context, req *assetmodel.ListRequest) ([]*assetmodel.Details, string, error)
	// CreateSignedUploadURL generates a signed URL for uploading an asset to Cloudinary.
	// It returns the signed parameters required for the upload, end client must build signed upload
	// URL using generated parameters.
	CreateSignedUploadURL(ctx context.Context, req *assetmodel.CreateSignedUploadURLRequest) (*assetmodel.GeneratedSignedParams, error)
	// Archive marks an asset as archived.
	// Note that only assets without any owners can be archived.
	Archive(ctx context.Context, req *assetmodel.ChangeStateRequest) error
	// MarkAsBroken marks an asset as broken.
	// If the asset has owners, it notifies the product-service about the broken asset via [gRPC client].
	//
	// [gRPC client]: https://github.com/mikhail5545/product-service-client
	MarkAsBroken(ctx context.Context, req *assetmodel.ChangeStateRequest) error
	// AddOwner associates an external owner with an asset.
	// It updates the asset metadata in MongoDB to include the new owner.
	// Broken or archived assets cannot have owners added.
	AddOwner(ctx context.Context, req *assetmodel.ManageOwnerRequest) error
	// RemoveOwner disassociates an external owner from an asset.
	// It updates the asset metadata in MongoDB to remove the specified owner.
	RemoveOwner(ctx context.Context, req *assetmodel.ManageOwnerRequest) error
	// Restore restores an archived asset back to active status.
	// Only archived assets can be restored.
	Restore(ctx context.Context, req *assetmodel.ChangeStateRequest) error
	// Delete permanently deletes an archived asset along with its metadata.
	// It also deletes the asset from Cloudinary.
	// Note that only currently soft-deleted (archived) assets can be permanently deleted.
	Delete(ctx context.Context, req *assetmodel.ChangeStateRequest) error
	// HandleWebhook processes incoming webhook notifications from Cloudinary.
	// It validates the signature and routes the webhook to the appropriate handler based on its type.
	HandleWebhook(ctx context.Context, payload []byte, timestamp, signature string) error
}

type Service struct {
	repo               *assetrepo.Repository
	metadataRepo       *metadatarepo.Repository
	imageServiceClient *client.ImageServiceClient
	apiClient          *apiclient.Client
	logger             *zap.Logger
}

var _ AssetService = (*Service)(nil)

type NewParams struct {
	Repo               *assetrepo.Repository
	MetadataRepo       *metadatarepo.Repository
	ImageServiceClient *client.ImageServiceClient
	ApiClient          *apiclient.Client
}

func New(params *NewParams, logger *zap.Logger) *Service {
	return &Service{
		repo:               params.Repo,
		metadataRepo:       params.MetadataRepo,
		imageServiceClient: params.ImageServiceClient,
		apiClient:          params.ApiClient,
		logger:             logger.With(zap.String("layer", "service"), zap.String("service", "Cloudinary")),
	}
}

// Get retrieves an active asset based on the provided filter.
func (s *Service) Get(ctx context.Context, filter *assetmodel.GetFilter) (*assetmodel.Details, error) {
	return s.get(ctx, filter, []assetrepo.Scope{
		assetrepo.ScopeActive,
	})
}

// GetWithArchived retrieves an asset that can be either active or archived based on the provided filter.
func (s *Service) GetWithArchived(ctx context.Context, filter *assetmodel.GetFilter) (*assetmodel.Details, error) {
	return s.get(ctx, filter, []assetrepo.Scope{
		assetrepo.ScopeActive,
		assetrepo.ScopeArchived,
	})
}

// GetWithBroken retrieves an asset that can be active, archived, or broken based on the provided filter.
func (s *Service) GetWithBroken(ctx context.Context, filter *assetmodel.GetFilter) (*assetmodel.Details, error) {
	return s.get(ctx, filter, []assetrepo.Scope{
		assetrepo.ScopeActive,
		assetrepo.ScopeArchived,
		assetrepo.ScopeBroken,
	})
}

// List retrieves a list of active assets based on the provided request.
func (s *Service) List(ctx context.Context, req *assetmodel.ListRequest) ([]*assetmodel.Details, string, error) {
	return s.list(ctx, req, []assetrepo.Scope{
		assetrepo.ScopeActive,
	})
}

// ListArchived retrieves a list of archived assets based on the provided request.
func (s *Service) ListArchived(ctx context.Context, req *assetmodel.ListRequest) ([]*assetmodel.Details, string, error) {
	return s.list(ctx, req, []assetrepo.Scope{
		assetrepo.ScopeArchived,
	})
}

// ListBroken retrieves a list of broken assets based on the provided request.
func (s *Service) ListBroken(ctx context.Context, req *assetmodel.ListRequest) ([]*assetmodel.Details, string, error) {
	return s.list(ctx, req, []assetrepo.Scope{
		assetrepo.ScopeBroken,
	})
}

// CreateSignedUploadURL generates a signed URL for uploading an asset to Cloudinary.
// It returns the signed parameters required for the upload, end client must build signed upload
// URL using generated parameters.
func (s *Service) CreateSignedUploadURL(ctx context.Context, req *assetmodel.CreateSignedUploadURLRequest) (*assetmodel.GeneratedSignedParams, error) {
	if err := req.Validate(); err != nil {
		return nil, serviceerrors.NewValidationFailedError(err)
	}

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		adminID, err := parsing.StrToUUID(req.AdminID)
		if err != nil {
			return err
		}
		asset := &assetmodel.Asset{
			CloudinaryPublicID: req.PublicID,
			Status:             assetmodel.StatusUploadURLGenerated,
			CreatedByName:      &req.AdminName,
			CreatedBy:          &adminID,
		}

		if err := txRepo.Create(ctx, asset); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return serviceerrors.NewAlreadyExistsError("asset with the given public ID already exists")
			}
			s.logger.Error("failed to create asset record for signed upload URL",
				zap.String("public_id", req.PublicID),
				zap.String("admin_id", req.AdminID),
				zap.String("admin_name", req.AdminName),
				zap.Error(err),
			)
			return fmt.Errorf("failed to create asset record for signed upload URL: %w", err)
		}

		return nil
	})

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	params := make(url.Values)

	if req.Eager != nil {
		params.Set("eager", *req.Eager)
	}
	params.Set("timestamp", timestamp)
	params.Set("public_id", req.PublicID)

	signature, err := s.apiClient.SignUploadParams(ctx, params)
	if err != nil {
		s.logger.Error("failed to sign upload params",
			zap.String("timestamp", timestamp),
			zap.String("public_id", req.PublicID),
			zap.Error(err),
		)
		return nil, fmt.Errorf("failed to sign upload params: %w", err)
	}

	return &assetmodel.GeneratedSignedParams{
		Signature: signature,
		ApiKey:    s.apiClient.GetApiKey(),
		PublicID:  req.PublicID,
		Timestamp: timestamp,
		Eager:     req.Eager,
	}, nil
}

// Archive marks an asset as archived.
// Note that only assets without any owners can be archived.
func (s *Service) Archive(ctx context.Context, req *assetmodel.ChangeStateRequest) error {
	if err := req.Validate(); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}

	var toDelete *uuid.UUID
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		asset, err := s.getInTx(ctx, txRepo, req.ID, []string{"id", "status"})
		if err != nil {
			return err
		}
		if asset.Status == assetmodel.StatusArchived {
			return serviceerrors.NewConflictError("asset is already archived")
		}

		metadata, err := s.getAssetMetadata(ctx, asset.ID)
		if err != nil {
			return err
		}

		if len(metadata.Owners) > 0 {
			return serviceerrors.NewConflictError("cannot archive asset with owners")
		}
		toDelete = &asset.ID
		return nil
	})
	if err != nil {
		return err
	}
	// If transaction commits successfully, delete gRPC relations outside transaction
	if toDelete != nil {
		return s.grpcDelete(ctx, toDelete)
	}
	return nil
}

// MarkAsBroken marks an asset as broken.
// If the asset has owners, it notifies the product-service about the broken asset via [gRPC client].
//
// [gRPC client]: https://github.com/mikhail5545/product-service-client
func (s *Service) MarkAsBroken(ctx context.Context, req *assetmodel.ChangeStateRequest) error {
	if err := req.Validate(); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}
	var toDelete *uuid.UUID
	var metadataToDelete *metadatamodel.AssetMetadata

	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		asset, err := s.getInTx(ctx, txRepo, req.ID, []string{"id", "status"})
		if err != nil {
			return err
		}
		if asset.Status == assetmodel.StatusBroken {
			return serviceerrors.NewConflictError("asset is already marked as broken")
		}

		metadata, err := s.getAssetMetadata(ctx, asset.ID)
		if err != nil {
			return err
		}

		if len(metadata.Owners) > 0 {
			metadataToDelete = metadata
			toDelete = &asset.ID
		}

		return nil
	})
	if err != nil {
		return err
	}
	// If transaction commits successfully, delete gRPC relations outside transaction
	if toDelete != nil {
		return s.markAsBrokenAndClearOwners(ctx, toDelete, metadataToDelete, req)
	}
	return nil
}

// AddOwner associates an external owner with an asset.
// It updates the asset metadata in MongoDB to include the new owner.
// Broken or archived assets cannot have owners added.
func (s *Service) AddOwner(ctx context.Context, req *assetmodel.ManageOwnerRequest) error {
	if err := req.Validate(); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}

	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		asset, err := s.getInTx(ctx, txRepo, req.ID, []string{"id", "status"})
		if err != nil {
			return err
		}
		if asset.Status == assetmodel.StatusArchived || asset.Status == assetmodel.StatusBroken {
			return serviceerrors.NewConflictError("cannot add owner to archived or broken asset")
		}

		return s.addOwner(ctx, asset.ID, req)
	})
}

// RemoveOwner disassociates an external owner from an asset.
// It updates the asset metadata in MongoDB to remove the specified owner.
func (s *Service) RemoveOwner(ctx context.Context, req *assetmodel.ManageOwnerRequest) error {
	if err := req.Validate(); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}
	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		asset, err := s.getInTx(ctx, txRepo, req.ID, []string{"id", "status"})
		if err != nil {
			return err
		}

		toRemove := metadatamodel.Owner{
			OwnerID:   req.OwnerID,
			OwnerType: req.OwnerType,
		}
		metadata, err := s.metadataRepo.GetByOwner(ctx, asset.ID.String(), &toRemove)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return serviceerrors.NewNotFoundError(err)
			}
			s.logger.Error("failed to retrieve asset metadata for removing owner",
				zap.Error(err), zap.String("owner_id", req.OwnerID), zap.String("owner_type", req.OwnerType),
			)
			return fmt.Errorf("failed to retrieve asset metadata for removing owner: %w", err)
		}
		return s.removeOwner(ctx, metadata, req)
	})
}

// Restore restores an archived asset back to active status.
// Only archived assets can be restored.
func (s *Service) Restore(ctx context.Context, req *assetmodel.ChangeStateRequest) error {
	if err := req.Validate(); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}

	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		asset, err := s.getInTx(ctx, txRepo, req.ID, []string{"id", "status"})
		if err != nil {
			return err
		}
		if asset.Status != assetmodel.StatusArchived {
			return serviceerrors.NewConflictError("asset is not archived")
		}

		adminID, err := parsing.StrToUUID(req.AdminID)
		if err != nil {
			return err
		}
		if _, err := txRepo.Restore(ctx, assetrepo.StateOperationOptions{IDs: uuid.UUIDs{asset.ID}}, &types.AuditTrailOptions{
			AdminID:   adminID,
			AdminName: req.AdminName,
			Note:      req.Note,
		}); err != nil {
			s.logger.Error("failed to restore archived asset", zap.Error(err), zap.String("asset_id", asset.ID.String()))
			return fmt.Errorf("failed to restore archived asset: %w", err)
		}

		return nil
	})
}

// Delete permanently deletes an archived asset along with its metadata.
// It also deletes the asset from Cloudinary.
// Note that only currently soft-deleted (archived) assets can be permanently deleted.
func (s *Service) Delete(ctx context.Context, req *assetmodel.ChangeStateRequest) error {
	if err := req.Validate(); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}
	var toDelete *uuid.UUID
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		asset, err := s.getInTx(ctx, txRepo, req.ID, []string{"id", "status", "cloudinary_public_id", "recourse_type"})
		if err != nil {
			return err
		}
		if asset.Status != assetmodel.StatusArchived {
			return serviceerrors.NewConflictError("only archived assets can be deleted")
		}

		if asset.CloudinaryPublicID != "" {
			// Delete asset from Cloudinary
			if err := s.apiClient.DeleteAsset(ctx, asset.CloudinaryPublicID, asset.ResourceType); err != nil {
				s.logger.Error("failed to delete asset from Cloudinary", zap.Error(err), zap.String("asset_id", asset.ID.String()))
				return fmt.Errorf("failed to delete asset from Cloudinary: %w", err)
			}
		}

		// Delete asset record from Postgres
		if _, err := txRepo.Delete(ctx, assetrepo.StateOperationOptions{IDs: uuid.UUIDs{asset.ID}}); err != nil {
			s.logger.Error("failed to delete asset record from Postgres", zap.Error(err), zap.String("asset_id", asset.ID.String()))
			return fmt.Errorf("failed to delete asset record from Postgres: %w", err)
		}
		toDelete = &asset.ID
		return nil
	})
	if err != nil {
		return err
	}
	if toDelete != nil {
		return s.deleteAssetMetadata(ctx, *toDelete)
	}
	return nil
}
