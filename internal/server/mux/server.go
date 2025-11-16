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
Package mux provides the implementation of the gRPC
[assetpb.MuxServiceServer] interface and provides
various operations for MUXUpload models.
*/
package mux

import (
	"context"

	muxservice "github.com/mikhail5545/media-service-go/internal/services/mux"
	"github.com/mikhail5545/media-service-go/internal/util/errors"
	"github.com/mikhail5545/media-service-go/internal/util/types"
	assetpb "github.com/mikhail5545/proto-go/proto/media_service/mux/asset/v0"
	"google.golang.org/grpc"
)

// Server implements the gRPC [assetpb.MuxServiceServer] interface and provides
// operations for MUXUpload models. It acts as an adapter between the gRPC transport layer
// and the Server-layer business logic of microservice, defined in the [mux.Service].
//
// For more information about underlying gRPC Server, see [github.com/mikhail5545/proto-go].
type Server struct {
	assetpb.UnimplementedAssetServiceServer
	service muxservice.Service
}

// New creates a new instance of [mux.Server]
func New(svc muxservice.Service) *Server {
	return &Server{
		service: svc,
	}
}

// Register registers the mux server with a gRPC server instance.
func Register(s *grpc.Server, svc muxservice.Service) {
	assetpb.RegisterAssetServiceServer(s, New(svc))
}

// CreateCoursePartUploadURL creates upload URL from course part with MUX direct upload API.
// It uses [muxclient.Client.CreateCoursePartUploadURL] method to access MUX direct upload API.
// Also, it checks if course part is already associated with [mux.Upload], if it is, method will update
// mux upload record to match new values without creating new mux upload instance or breaking it's association with the
// course part instance. Else, it will create new mux upload instance with desired values and associate it with course part
// via course part gRPC client.
//
// Returns an error if the partID is not a valid UUID `InvalidArgument`, the record is not found `NotFound`,
// MUX API client error `Unavailable`, any database/internal error `Internal` or gRPC
// course part client error (handled by handleGRPCError) gRPC errors.
// func (s *Server) CreateCoursePartUploadURL(ctx context.Context, req *assetpb.CreateUploadURLRequest) (*assetpb.CreateUploadURLResponse, error) {
// 	resp, err := s.service.CreateCoursePartUploadURL(ctx, req.GetPartId())
// 	if err != nil {
// 		return nil, errors.HandleServiceError(err)
// 	}
// 	return &assetpb.CreateUploadURLResponse{UploadUrl: resp.Data.Url}, nil
// }

// Get retrieves a single not soft-deleted mux upload record.
//
// Returns a `NotFound` gRPC error if the record is not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (s *Server) Get(ctx context.Context, req *assetpb.GetRequest) (*assetpb.GetResponse, error) {
	upload, err := s.service.Get(ctx, req.GetId())
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.GetResponse{MuxUpload: types.ConvertToMuxProtoBuf(upload)}, nil
}

// GetWithDeleted retrieves a single mux upload record, including soft-deleted ones.
//
// Returns a `NotFound` gRPC error if the record is not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (s *Server) GetWithDeleted(ctx context.Context, req *assetpb.GetWithDeletedRequest) (*assetpb.GetWithDeletedResponse, error) {
	upload, err := s.service.GetWithDeleted(ctx, req.GetId())
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.GetWithDeletedResponse{Asset: types.ConvertToMuxProtoBuf(upload)}, nil
}

// List retrieves a paginated list of all not soft-deleted mux upload records.
// The response also contains the total count of such records.
func (s *Server) List(ctx context.Context, req *assetpb.ListRequest) (*assetpb.ListResponse, error) {
	uploads, total, err := s.service.List(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	var pbuploads []*assetpb.MuxUpload
	for _, upload := range uploads {
		pbuploads = append(pbuploads, types.ConvertToMuxProtoBuf(&upload))
	}
	return &assetpb.ListResponse{MuxUploads: pbuploads, Total: total}, nil
}

// ListDeleted retrieves a paginated list of all soft-deleted mux upload records.
// The response also contains the total count of such records.
func (s *Server) ListDeleted(ctx context.Context, req *assetpb.ListDeletedRequest) (*assetpb.ListDeletedResponse, error) {
	uploads, total, err := s.service.ListDeleted(ctx, int(req.GetLimit()), int(req.GetOffset()))
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	var pbuploads []*assetpb.MuxUpload
	for _, upload := range uploads {
		pbuploads = append(pbuploads, types.ConvertToMuxProtoBuf(&upload))
	}
	return &assetpb.ListDeletedResponse{MuxUploads: pbuploads, Total: total}, nil
}

// Delete performs a soft-delete on a mux upload.
//
// Returns a `NotFound` gRPC error if any of the records are not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (s *Server) Delete(ctx context.Context, req *assetpb.DeleteRequest) (*assetpb.DeleteResponse, error) {
	err := s.service.Delete(ctx, req.GetId())
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.DeleteResponse{Id: req.GetId()}, nil
}

// DeletePermanent permanently deletes a mux upload from the database.
// This action is irreversible.
//
// Returns a `NotFound` gRPC error if any of the records are not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (s *Server) DeletePermanent(ctx context.Context, req *assetpb.DeletePermanentRequest) (*assetpb.DeletePermanentResponse, error) {
	err := s.service.DeletePermanent(ctx, req.GetId())
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.DeletePermanentResponse{Id: req.GetId()}, nil
}

// Restore restores a soft-deleted mux upload.
//
// Returns a `NotFound` gRPC error if any of the records are not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (s *Server) Restore(ctx context.Context, req *assetpb.RestoreRequest) (*assetpb.RestoreResponse, error) {
	err := s.service.Restore(ctx, req.GetId())
	if err != nil {
		return nil, errors.HandleServiceError(err)
	}
	return &assetpb.RestoreResponse{Id: req.GetId()}, nil
}
