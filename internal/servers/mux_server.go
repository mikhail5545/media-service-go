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

package servers

import (
	"context"

	"github.com/mikhail5545/media-service-go/internal/services"
	"github.com/mikhail5545/media-service-go/internal/utils"
	muxpb "github.com/mikhail5545/proto-go/proto/mux_upload/v0"
)

type MuxServer struct {
	muxpb.UnimplementedMuxUploadServiceServer
	muxService *services.MuxService
}

func NewMuxServer(muxService *services.MuxService) *MuxServer {
	return &MuxServer{
		muxService: muxService,
	}
}

func (s *MuxServer) GetMuxUpload(ctx context.Context, req *muxpb.GetMuxUploadRequest) (*muxpb.GetMuxUploadResponse, error) {
	upload, err := s.muxService.GetMuxUpload(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &muxpb.GetMuxUploadResponse{MuxUpload: utils.ConvertToMuxProtoBuf(upload)}, nil
}

func (s *MuxServer) DeleteMuxUpload(ctx context.Context, req *muxpb.DeleteMuxUploadRequest) (*muxpb.DeleteMuxUploadResponse, error) {
	err := s.muxService.DeleteMuxUpload(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return &muxpb.DeleteMuxUploadResponse{}, nil
}
