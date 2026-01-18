package mux

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	apiclient "github.com/mikhail5545/media-service-go/internal/apiclients/mux"
	assetmetadatarepo "github.com/mikhail5545/media-service-go/internal/database/mongo/mux/metadata"
	assetrepo "github.com/mikhail5545/media-service-go/internal/database/postgres/mux/asset"
	"github.com/mikhail5545/media-service-go/internal/database/types"
	serviceerrors "github.com/mikhail5545/media-service-go/internal/errors"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	metadatamodel "github.com/mikhail5545/media-service-go/internal/models/mux/metadata"
	muxtypes "github.com/mikhail5545/media-service-go/internal/models/mux/types"
	"github.com/mikhail5545/media-service-go/internal/util/parsing"
	"github.com/mikhail5545/product-service-client/client"
	muxgo "github.com/muxinc/mux-go/v6"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AssetService defines the interface for managing MUX assets.
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
	// CreateUploadURL generates a new upload URL for new MUX Direct Upload and creates a new asset.
	// After this step, the rest of the asset information will be populated via incoming MUX webhooks.
	CreateUploadURL(ctx context.Context, req *assetmodel.CreateUploadURLRequest) (*muxgo.UploadResponse, error)
	// Archive marks an asset as archived.
	// Note that only assets without any owners can be archived.
	Archive(ctx context.Context, req *assetmodel.ChangeStateRequest) error
	// MarkAsBroken marks an asset as broken.
	// If the asset has owners, it notifies the product-service about the broken asset via [gRPC client].
	//
	// [gRPC client]: https://github.com/mikhail5545/product-service-client
	MarkAsBroken(ctx context.Context, req *assetmodel.ChangeStateRequest) error
	// Delete permanently deletes an archived asset along with its metadata.
	// It also deletes the asset from MUX.
	// Note that only currently soft-deleted (archived) assets can be permanently deleted.
	Delete(ctx context.Context, req *assetmodel.ChangeStateRequest) error
	// HandleAssetWebhook processes incoming MUX asset webhooks based on their type.
	// It routes the webhook to the appropriate handler function.
	// It does not return any error, as we want to avoid retrying the webhook processing in case of failure.
	HandleAssetWebhook(ctx context.Context, payload *muxtypes.MuxWebhook) error
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
	// GeneratePlaybackToken generates a signed JWT playback token for secure video playback.
	GeneratePlaybackToken(ctx context.Context, req *assetmodel.GeneratePlaybackTokenRequest) (string, error)
}

// Service implements the AssetService interface for managing MUX assets.
type Service struct {
	repo                     *assetrepo.Repository
	metadataRepo             *assetmetadatarepo.Repository
	lessonVideoVersionClient *client.LessonVideoVersionServiceClient
	videoClient              *client.VideoServiceClient
	apiClient                *apiclient.Client
	logger                   *zap.Logger
}

var _ AssetService = (*Service)(nil)

type NewParams struct {
	Repo                     *assetrepo.Repository
	MetadataRepo             *assetmetadatarepo.Repository
	LessonVideoVersionClient *client.LessonVideoVersionServiceClient
	VideoClient              *client.VideoServiceClient
	ApiClient                *apiclient.Client
}

func New(
	params *NewParams,
	logger *zap.Logger,
) *Service {
	return &Service{
		repo:                     params.Repo,
		lessonVideoVersionClient: params.LessonVideoVersionClient,
		videoClient:              params.VideoClient,
		metadataRepo:             params.MetadataRepo,
		apiClient:                params.ApiClient,
		logger:                   logger.With(zap.String("layer", "service"), zap.String("service", "mux")),
	}
}

// Get retrieves an active asset based on the provided filter.
func (s *Service) Get(ctx context.Context, filter *assetmodel.GetFilter) (*assetmodel.Details, error) {
	return s.get(ctx, filter, []assetrepo.Scope{
		assetrepo.ScopeActive,
		assetrepo.ScopeUploadURLGenerated,
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

// CreateUploadURL generates a new upload URL for new MUX Direct Upload and creates a new asset.
// After this step, the rest of the asset information will be populated via incoming MUX webhooks.
func (s *Service) CreateUploadURL(ctx context.Context, req *assetmodel.CreateUploadURLRequest) (*muxgo.UploadResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, serviceerrors.NewValidationFailedError(err)
	}

	var resp *muxgo.UploadResponse
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		newAssetID, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("failed to generate new asset id: %w", err)
		}
		newAsset := &assetmodel.Asset{
			ID:           newAssetID,
			Status:       assetmodel.StatusUploadURLGenerated,
			UploadStatus: assetmodel.UploadStatusPreparing,
		}

		s.logger.Info("generating upload url", zap.String("asset_id", newAssetID.String()))

		muxMeta := &muxgo.AssetMetadata{
			Title:      req.Title,
			CreatorId:  req.AdminID,
			ExternalId: newAssetID.String(),
		}
		resp, err = s.apiClient.CreateDirectUploadURL(ctx, muxMeta, muxgo.SIGNED, muxgo.PUBLIC)
		if err != nil {
			s.logger.Error("failed to create direct upload url", zap.Error(err), zap.String("asset_id", newAssetID.String()))
			return fmt.Errorf("failed to create direct upload url: %w", err)
		}

		newAsset.MuxUploadID = &resp.Data.Id
		newAsset.MuxAssetID = &resp.Data.AssetId

		if err := txRepo.Create(ctx, newAsset); err != nil {
			s.logger.Error("failed to create mux asset record", zap.Error(err), zap.String("asset_id", newAssetID.String()))
			return fmt.Errorf("failed to create mux asset record: %w", err)
		}

		s.logger.Info("successfully generated upload url", zap.String("asset_id", newAssetID.String()), zap.String("upload_url", resp.Data.Url))

		metadata := &metadatamodel.AssetMetadata{
			Key:       newAssetID.String(),
			Title:     req.Title,
			CreatorID: req.AdminID,
			Owners:    []metadatamodel.Owner{},      // initialize empty owners slice
			Tracks:    []muxtypes.MuxWebhookTrack{}, // initialize empty tracks slice
		}

		if err := s.metadataRepo.Create(ctx, metadata); err != nil {
			s.logger.Error("failed to create asset metadata", zap.Error(err), zap.String("asset_id", newAssetID.String()))
			return fmt.Errorf("failed to create asset metadata: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Archive marks an asset as archived.
// Note that only assets without any owners can be archived.
func (s *Service) Archive(ctx context.Context, req *assetmodel.ChangeStateRequest) error {
	if err := req.Validate(); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}

	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		assetID, err := parsing.StrToUUID(req.ID)
		if err != nil {
			return err
		}
		asset, err := txRepo.Get(ctx, assetrepo.GetOptions{ID: assetID})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return serviceerrors.NewNotFoundError(err)
			}
			s.logger.Error("failed to retrieve asset for archiving", zap.Error(err), zap.String("asset_id", req.ID))
			return fmt.Errorf("failed to retrieve asset for archiving: %w", err)
		}

		if err := validateBeforeArchive(asset); err != nil {
			return err
		}

		metadata, err := s.getAssetMetadata(ctx, assetID)
		if err != nil {
			return err
		}
		if len(metadata.Owners) > 0 {
			return serviceerrors.NewConflictError("cannot archive asset that is associated with owners")
		}

		return s.archiveAsset(ctx, txRepo, req, assetID)
	})
}

// MarkAsBroken marks an asset as broken.
// If the asset has owners, it notifies the product-service about the broken asset via [gRPC client].
//
// [gRPC client]: https://github.com/mikhail5545/product-service-client
func (s *Service) MarkAsBroken(ctx context.Context, req *assetmodel.ChangeStateRequest) error {
	if err := req.Validate(); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}

	return s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		asset, err := s.getInTx(ctx, txRepo, []string{
			"id", "status", "upload_status",
		}, assetSearchOptions{
			AssetID: req.ID,
		})
		if asset.Status == assetmodel.StatusBroken {
			return serviceerrors.NewConflictError("asset is already marked as broken")
		}

		metadata, err := s.getAssetMetadata(ctx, asset.ID)
		if err != nil {
			return err
		}

		adminID, err := parsing.StrToUUID(req.AdminID)
		if err != nil {
			return err
		}
		if len(metadata.Owners) > 0 {
			if err := s.grpcMarkAsBroken(ctx, asset.ID, adminID, req); err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete permanently deletes an archived asset along with its metadata.
// It also deletes the asset from MUX.
// Note that only currently soft-deleted (archived) assets can be permanently deleted.
func (s *Service) Delete(ctx context.Context, req *assetmodel.ChangeStateRequest) error {
	if err := req.Validate(); err != nil {
		return serviceerrors.NewValidationFailedError(err)
	}
	var assetIDtoDelete *uuid.UUID
	var muxAssetIDtoDelete *string
	err := s.repo.DB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.repo.WithTx(tx)

		asset, err := s.getInTx(ctx, txRepo, []string{
			"id", "status", "upload_status",
		}, assetSearchOptions{
			AssetID: req.ID,
		})
		if err != nil {
			return err
		}
		if asset.Status != assetmodel.StatusArchived {
			return serviceerrors.NewConflictError("only archived assets can be deleted")
		}

		// Delete asset record from Postgres
		if _, err := txRepo.Delete(ctx, assetrepo.StateOperationOptions{IDs: uuid.UUIDs{asset.ID}}); err != nil {
			s.logger.Error("failed to delete mux asset record", zap.Error(err), zap.String("asset_id", asset.ID.String()))
			return fmt.Errorf("failed to delete mux asset record: %w", err)
		}
		assetIDtoDelete = &asset.ID
		muxAssetIDtoDelete = asset.MuxAssetID
		return nil
	})
	if err != nil {
		return err
	}
	if err := s.deleteMetadataAndMuxAsset(ctx, assetIDtoDelete, muxAssetIDtoDelete); err != nil {
		return err
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

		asset, err := s.getInTx(ctx, txRepo, []string{
			"id", "status", "upload_status",
		}, assetSearchOptions{
			AssetID: req.ID,
		})
		if err != nil {
			return err
		}
		if asset.Status == assetmodel.StatusArchived || asset.Status == assetmodel.StatusBroken {
			return serviceerrors.NewConflictError("cannot add owner to archived or broken asset")
		}
		if asset.UploadStatus == assetmodel.UploadStatusErrored || asset.UploadStatus == assetmodel.UploadStatusDeleted {
			return serviceerrors.NewConflictError("cannot add owner to asset with errored or deleted upload status")
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

		asset, err := s.getInTx(ctx, txRepo, []string{
			"id", "status", "upload_status",
		}, assetSearchOptions{
			AssetID: req.ID,
		})
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

		asset, err := s.getInTx(ctx, txRepo, []string{
			"id", "status", "upload_status",
		}, assetSearchOptions{
			AssetID: req.ID,
		})
		if err != nil {
			return err
		}
		if asset.Status != assetmodel.StatusArchived {
			return serviceerrors.NewConflictError("only archived assets can be restored")
		}

		adminID, err := parsing.StrToUUID(req.AdminID)
		if err != nil {
			return err
		}

		if _, err := txRepo.Restore(ctx, assetrepo.StateOperationOptions{IDs: uuid.UUIDs{asset.ID}}, types.AuditTrailOptions{
			AdminID:   adminID,
			AdminName: req.AdminName,
			Note:      req.Note,
		}); err != nil {
			s.logger.Error("failed to restore asset", zap.Error(err), zap.String("asset_id", req.ID))
			return fmt.Errorf("failed to restore asset: %w", err)
		}
		return nil
	})
}

// GeneratePlaybackToken generates a signed JWT playback token for secure video playback.
func (s *Service) GeneratePlaybackToken(ctx context.Context, req *assetmodel.GeneratePlaybackTokenRequest) (string, error) {
	if err := req.Validate(); err != nil {
		return "", serviceerrors.NewValidationFailedError(err)
	}
	asset, err := s.repo.Get(ctx, assetrepo.GetOptions{
		ID: req.AssetID,
		Fields: []string{
			"id", "status", "upload_status", "primary_signed_playback_id",
		},
	}, assetrepo.ScopeAll)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", serviceerrors.NewNotFoundError(err)
		}
		s.logger.Error("failed to retrieve asset for playback token generation", zap.Error(err), zap.String("asset_id", req.AssetID.String()))
		return "", fmt.Errorf("failed to retrieve asset for playback token generation: %w", err)
	}
	if asset.Status != assetmodel.StatusActive {
		return "", serviceerrors.NewConflictError("playback token can only be generated for active assets")
	}
	if asset.UploadStatus != assetmodel.UploadStatusReady {
		return "", serviceerrors.NewConflictError("playback token can only be generated for assets with ready upload status")
	}
	if asset.PrimarySignedPlaybackID == nil {
		s.logger.Error("asset does not have a signed playback ID for token generation", zap.String("asset_id", req.AssetID.String()))
		return "", serviceerrors.NewConflictError("asset does not have a signed playback ID for token generation")
	}
	return s.apiClient.GeneratePlaybackJWTToken(apiclient.GeneratePlaybackTokenOptions{
		UserID:     req.UserID,
		PlaybackID: *asset.PrimarySignedPlaybackID,
		Expiration: req.Expiration,
		UserAgent:  req.UserAgent,
		SessionID:  req.SessionID,
	})
}
