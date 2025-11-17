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
Package cloudinary provides the implementation of the gRPC
[assetpb.AssetServiceServer] interface and provides
various operations for cloudinary assets.
*/
package cloudinary

import (
	"context"

	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	metamodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/metadata"
	cldservice "github.com/mikhail5545/media-service-go/internal/services/cloudinary"
	"github.com/mikhail5545/media-service-go/internal/util/errors"
	"github.com/mikhail5545/media-service-go/internal/util/types"
	assetpb "github.com/mikhail5545/proto-go/proto/media_service/cloudinary/asset/v0"
	"google.golang.org/grpc"
)

// Server implements the gRPC [assetpb.AssetServiceServer] interface and provides
// operations for cloudinary assets. It acts as an adapter between the gRPC transport layer
// and the Server-layer business logic of microservice, defined in the [cloudinary.Service].
//
// See more information about [underlying gRPC services].
//
// [underlying gRPC services]: https://github.com/mikhail5545/proto-go
type Server struct {
	assetpb.UnimplementedAssetServiceServer
	svc cldservice.Service
}

// New creates a new instance of [cloudinary.Server]
func New(svc cldservice.Service) *Server {
	return &Server{svc: svc}
}

// Register registers the cloudinary server with a gRPC server instance.
func Register(s *grpc.Server, svc cldservice.Service) {
	assetpb.RegisterAssetServiceServer(s, New(svc))
}

// Get retrieves a single not soft-deleted asset record and it's metadata.
//
// Returns a `NotFound` gRPC error if the record is not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (s *Server) Get(ctx context.Context, req *assetpb.GetRequest) (*assetpb.GetResponse, error) {
	response, err := s.svc.Get(ctx, req.GetId())
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.GetResponse{Response: types.CldAssetResponseToProtobuf(response)}, nil
}

// GetWithDeleted retrieves a single asset record and it's metadata, including soft-deleted ones.
//
// Returns a `NotFound` gRPC error if the record is not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (s *Server) GetWithDeleted(ctx context.Context, req *assetpb.GetWithDeletedRequest) (*assetpb.GetWithDeletedResponse, error) {
	response, err := s.svc.GetWithDeleted(ctx, req.GetId())
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.GetWithDeletedResponse{Response: types.CldAssetResponseToProtobuf(response)}, nil
}

// List retrieves a paginated list of all not soft-deleted asset records with their metadata.
// The response also contains the total count of such records.
//
// Returns `InvalidArgument` gRPC error if the provided limit or offset are invalid.
func (s *Server) List(ctx context.Context, req *assetpb.ListRequest) (*assetpb.ListResponse, error) {
	responses, total, err := s.svc.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	var pbResponses []*assetpb.AssetResponse
	for _, response := range responses {
		pbResponses = append(pbResponses, types.CldAssetResponseToProtobuf(&response))
	}
	return &assetpb.ListResponse{Responses: pbResponses, Total: total}, nil
}

// ListUnowned retrieves a paginated list of all unowned asset records with their metadata.
// The response also contains the total count of such records.
//
// Returns `InvalidArgument` gRPC error if the provided limit or offset are invalid.
func (s *Server) ListUnowned(ctx context.Context, req *assetpb.ListUnownedRequest) (*assetpb.ListUnownedResponse, error) {
	responses, total, err := s.svc.ListUnowned(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	var pbResponses []*assetpb.AssetResponse
	for _, response := range responses {
		pbResponses = append(pbResponses, types.CldAssetResponseToProtobuf(&response))
	}
	return &assetpb.ListUnownedResponse{Responses: pbResponses, Total: total}, nil
}

// ListDeleted retrieves a paginated list of all soft-deleted asset records with their metadata.
// The response also contains the total count of such records.
//
// Returns `InvalidArgument` gRPC error if the provided limit or offset are invalid.
func (s *Server) ListDeleted(ctx context.Context, req *assetpb.ListDeletedRequest) (*assetpb.ListDeletedResponse, error) {
	responses, total, err := s.svc.ListDeleted(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	var pbResponses []*assetpb.AssetResponse
	for _, response := range responses {
		pbResponses = append(pbResponses, types.CldAssetResponseToProtobuf(&response))
	}
	return &assetpb.ListDeletedResponse{Responses: pbResponses, Total: total}, nil
}

// CreateSignedUploadURL CreateSignedUploadURL creates a signature for a direct frontend upload. Direct upload url should be
// constructed using this params, this function only creates signature for signed upload.
//
// Returns a `InvalidArgument` gRPC error if the request payload is invalid.
// Returns an `Unavailable` gRPC error if any of Cloudinary API calls fails.
func (s *Server) CreateSignedUploadURL(ctx context.Context, req *assetpb.CreateSignedUploadURLRequest) (*assetpb.CreateSignedUploadURLResponse, error) {
	createReq := &assetmodel.CreateSignedUploadURLRequest{
		Eager:    req.Eager,
		File:     req.GetFile(),
		PublicID: req.GetPublicId(),
	}
	res, err := s.svc.CreateSignedUploadURL(ctx, createReq)
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	eager := res["eager"]
	return &assetpb.CreateSignedUploadURLResponse{
		Signature: res["signature"],
		PublicId:  res["public_id"],
		Timestamp: res["timestamp"],
		ApiKey:    res["api_key"],
		Eager:     &eager,
	}, nil
}

// UpdateOwners processes asset ownership relations changes. It recieves an updated list of asset owners, updates local metadata
// for asset (about it's owners), processes the diff between old and new owners and notifies external services about this ownership
// changes via gRPC connection.
//
// Returns `NotFound` gRPC error if an asset not found.
// Returns `InvalidArgument` gRPC error if the request payload is invalid.
func (s *Server) UpdateOwners(ctx context.Context, req *assetpb.UpdateOwnersRequest) (*assetpb.UpdateOwnersResponse, error) {
	updateReq := &assetmodel.UpdateOwnersRequest{
		ID: req.GetId(),
	}
	for _, owner := range req.Owners {
		updateReq.Owners = append(updateReq.Owners, metamodel.Owner{OwnerID: owner.GetOwnerId(), OwnerType: owner.GetOwnerType()})
	}
	if err := s.svc.UpdateOwners(ctx, updateReq); err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.UpdateOwnersResponse{Id: req.GetId()}, nil
}

// Associate associates an asset with a single owner.
// It updates local metadata and notifies another services via gRPC calls.
//
// Returns `NotFound` gRPC error if an asset/owner not found.
// Returns `InvalidArgument` gRPC error if the request payload is invalid or owner aleady associated with another asset.
func (s *Server) Associate(ctx context.Context, req *assetpb.AssociateRequest) (*assetpb.AssociateResponse, error) {
	associateReq := &assetmodel.AssociateRequest{
		ID:        req.GetId(),
		OwnerID:   req.GetOwnerId(),
		OwnerType: req.GetOwnerType(),
	}
	if err := s.svc.Associate(ctx, associateReq); err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.AssociateResponse{Id: req.GetId()}, nil
}

// Deassociate deassociates an asset from a single owner.
// It updates local metadata and notifies another services via gRPC calls.
//
// Returns a` NotFound` gRPC error if an asset/owner not found.
// Returns an `InvalidArgument` gRPC error if the request payload is invalid.
func (s *Server) Deassociate(ctx context.Context, req *assetpb.DeassociateRequest) (*assetpb.DeassociateResponse, error) {
	deassociateReq := &assetmodel.DeassociateRequest{
		ID:        req.GetId(),
		OwnerID:   req.GetOwnerId(),
		OwnerType: req.GetOwnerType(),
	}
	if err := s.svc.Deassociate(ctx, deassociateReq); err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.DeassociateResponse{Id: req.GetId()}, nil
}

// SuccessfulUpload creates a new asset with provided information and creates owner relations for it.
// It saves asset metadata about owner relations in the local noSQL db and notifies external services about ownership
// changes via gRPC connection. This method should be called after successful cloudinary image upload.
//
// Returns newly created asset.
// Returns an `InvalidArgument` gRPC error if the request payload is invalid.
func (s *Server) SuccessfulUpload(ctx context.Context, req *assetpb.SuccessfulUploadRequest) (*assetpb.SuccessfulUploadResponse, error) {
	successReq := types.SuccessRequestFromProtobuf(req)
	response, err := s.svc.SuccessfulUpload(ctx, successReq)
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.SuccessfulUploadResponse{Response: types.CldAssetResponseToProtobuf(response)}, nil
}

// CleanupOrphanAssets finds and deletes assets that exist in Cloudinary but not in the local database.
//
// Returns the number of cleaned assets.
// Returns an `InvalidArgument` gRPC error if the request payload is invalid.
func (s *Server) CleanupOrphanAssets(ctx context.Context, req *assetpb.CleanupOrphanAssetsRequest) (*assetpb.CleanupOrphanAssetsResponse, error) {
	cleanupReq := &assetmodel.CleanupOrphanAssetsRequest{
		Folder:    req.GetFolder(),
		AssetType: req.GetAssetType(),
	}
	count, err := s.svc.CleanupOrphanAssets(ctx, cleanupReq)
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.CleanupOrphanAssetsResponse{Count: int64(count)}, nil
}

// Delete performs a soft-delete on an asset. If asset is assocaited with owners, they will be deassociated and
// all asset metadata about its owners will be cleared.
//
// Returns a `NotFound` gRPC error if any of the records are not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (s *Server) Delete(ctx context.Context, req *assetpb.DeleteRequest) (*assetpb.DeleteResponse, error) {
	if err := s.svc.Delete(ctx, req.GetId()); err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.DeleteResponse{Id: req.GetId()}, nil
}

// DeletePermanent permanently deletes an asset and it's metadata.
// It also deletes an asset from the Cloudinary.
// This action is irreversible.
//
// Returns a `NotFound` gRPC error if any of the records are not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
// Returns an `Unavailable` gRPC error if any of Cloudinary API calls fails.
func (s *Server) DeletePermanent(ctx context.Context, req *assetpb.DeletePermanentRequest) (*assetpb.DeletePermanentResponse, error) {
	deleteReq := &assetmodel.DestroyAssetRequest{
		ID: req.GetId(),
	}
	if err := s.svc.DeletePermanent(ctx, deleteReq); err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.DeletePermanentResponse{Id: req.GetId()}, nil
}

// Restore restores a soft-deleted asset.
//
// Returns a `NotFound` gRPC error if any of the records are not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (s *Server) Restore(ctx context.Context, req *assetpb.RestoreRequest) (*assetpb.RestoreResponse, error) {
	if err := s.svc.Restore(ctx, req.GetId()); err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.RestoreResponse{Id: req.GetId()}, nil
}
