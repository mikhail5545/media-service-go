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

package mux

import (
	"context"
	"time"

	"github.com/mikhail5545/media-service-go/internal/grpc/common"
	muxconv "github.com/mikhail5545/media-service-go/internal/grpc/conversion/mux"
	muxservice "github.com/mikhail5545/media-service-go/internal/services/mux"
	errutil "github.com/mikhail5545/media-service-go/internal/util/errors"
	muxassetpbv1 "github.com/mikhail5545/media-service-go/pb/media_service/mux/asset/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	muxassetpbv1.UnimplementedAssetServiceServer
	service   *muxservice.Service
	converter *muxconv.Converter
	logger    *zap.Logger
}

var _ muxassetpbv1.AssetServiceServer = (*Server)(nil)

func New(svc *muxservice.Service, logger *zap.Logger) *Server {
	return &Server{
		service:   svc,
		converter: muxconv.New(logger),
		logger:    logger.With(zap.String("component", "grpc/mux/Server")),
	}
}

func Register(grpcServer *grpc.Server, svc *muxservice.Service, logger *zap.Logger) {
	muxassetpbv1.RegisterAssetServiceServer(grpcServer, New(svc, logger))
}

func (s *Server) Ping(ctx context.Context, req *muxassetpbv1.PingRequest) (*muxassetpbv1.PingResponse, error) {
	return &muxassetpbv1.PingResponse{
		Timestamp: time.Now().Unix(),
	}, nil
}

func (s *Server) Get(ctx context.Context, req *muxassetpbv1.GetRequest) (*muxassetpbv1.GetResponse, error) {
	return common.Handle(ctx, s.converter.ConvertGetRequest, s.converter.ConvertGetResponse, s.service.Get, req)
}

func (s *Server) GetWithArchived(ctx context.Context, req *muxassetpbv1.GetWithArchivedRequest) (*muxassetpbv1.GetWithArchivedResponse, error) {
	return common.Handle(ctx, s.converter.ConvertGetRequest, s.converter.ConvertGetWithArchivedResponse, s.service.GetWithArchived, req)
}

func (s *Server) GetWithBroken(ctx context.Context, req *muxassetpbv1.GetWithBrokenRequest) (*muxassetpbv1.GetWithBrokenResponse, error) {
	return common.Handle(ctx, s.converter.ConvertGetRequest, s.converter.ConvertGetWithBrokenResponse, s.service.GetWithBroken, req)
}

func (s *Server) List(ctx context.Context, req *muxassetpbv1.ListRequest) (*muxassetpbv1.ListResponse, error) {
	return common.HandleList(ctx, s.converter.ConvertListRequest, s.converter.ConvertListResponse, s.service.List, req)
}

func (s *Server) ListArchived(ctx context.Context, req *muxassetpbv1.ListArchivedRequest) (*muxassetpbv1.ListArchivedResponse, error) {
	return common.HandleList(ctx, s.converter.ConvertListRequest, s.converter.ConvertListArchivedResponse, s.service.ListArchived, req)
}

func (s *Server) ListBroken(ctx context.Context, req *muxassetpbv1.ListBrokenRequest) (*muxassetpbv1.ListBrokenResponse, error) {
	return common.HandleList(ctx, s.converter.ConvertListRequest, s.converter.ConvertListBrokenResponse, s.service.ListBroken, req)
}

func (s *Server) CreateUploadURL(ctx context.Context, req *muxassetpbv1.CreateUploadURLRequest) (*muxassetpbv1.CreateUploadURLResponse, error) {
	return common.Handle(ctx, s.converter.ConvertCreateUploadURLRequest, s.converter.ConvertCreateUploadURLResponse, s.service.CreateUploadURL, req)
}

func (s *Server) MarkAsBroken(ctx context.Context, req *muxassetpbv1.MarkAsBrokenRequest) (*muxassetpbv1.MarkAsBrokenResponse, error) {
	return common.HandleEmpty(ctx, s.converter.ConvertChangeStateRequest, s.service.MarkAsBroken, req, &muxassetpbv1.MarkAsBrokenResponse{})
}

func (s *Server) Archive(ctx context.Context, req *muxassetpbv1.ArchiveRequest) (*muxassetpbv1.ArchiveResponse, error) {
	return common.HandleEmpty(ctx, s.converter.ConvertChangeStateRequest, s.service.Archive, req, &muxassetpbv1.ArchiveResponse{})
}

func (s *Server) Restore(ctx context.Context, req *muxassetpbv1.RestoreRequest) (*muxassetpbv1.RestoreResponse, error) {
	return common.HandleEmpty(ctx, s.converter.ConvertChangeStateRequest, s.service.Restore, req, &muxassetpbv1.RestoreResponse{})
}

func (s *Server) Delete(ctx context.Context, req *muxassetpbv1.DeleteRequest) (*muxassetpbv1.DeleteResponse, error) {
	return common.HandleEmpty(ctx, s.converter.ConvertChangeStateRequest, s.service.Delete, req, &muxassetpbv1.DeleteResponse{})
}

func (s *Server) AddOwner(ctx context.Context, req *muxassetpbv1.AddOwnerRequest) (*muxassetpbv1.AddOwnerResponse, error) {
	return common.HandleEmpty(ctx, s.converter.ConvertManageOwnerRequest, s.service.AddOwner, req, &muxassetpbv1.AddOwnerResponse{})
}

func (s *Server) RemoveOwner(ctx context.Context, req *muxassetpbv1.RemoveOwnerRequest) (*muxassetpbv1.RemoveOwnerResponse, error) {
	return common.HandleEmpty(ctx, s.converter.ConvertManageOwnerRequest, s.service.RemoveOwner, req, &muxassetpbv1.RemoveOwnerResponse{})
}

func (s *Server) GeneratePlaybackToken(ctx context.Context, req *muxassetpbv1.GeneratePlaybackTokenRequest) (*muxassetpbv1.GeneratePlaybackTokenResponse, error) {
	genReq, err := s.converter.ConvertGeneratePlaybackTokenRequest(req)
	if err != nil {
		return nil, err
	}
	token, err := s.service.GeneratePlaybackToken(ctx, genReq)
	if err != nil {
		return nil, errutil.ToGRPCCode(err)
	}
	return &muxassetpbv1.GeneratePlaybackTokenResponse{
		Token: token,
	}, nil
}
