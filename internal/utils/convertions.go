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

package utils

import (
	"github.com/mikhail5545/media-service-go/internal/models"
	muxpb "github.com/mikhail5545/proto-go/proto/mux_upload/v0"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertToMuxProtoBuf(muxUpload *models.MUXUpload) *muxpb.MuxUpload {
	muxUploadpb := &muxpb.MuxUpload{
		Id:                    muxUpload.ID,
		CreatedAt:             timestamppb.New(muxUpload.CreatedAt),
		UpdatedAt:             timestamppb.New(muxUpload.UpdatedAt),
		MuxUploadId:           muxUpload.MUXUploadID,
		MuxAssetId:            muxUpload.MUXAssetID,
		MuxPlaybackId:         muxUpload.MUXPlaybackID,
		VideoProcessingStatus: muxUpload.VideoProcessingStatus,
		AspectRatio:           muxUpload.AspectRatio,
		AssetCreatedAt:        timestamppb.New(*muxUpload.AssetCreatedAt),
	}
	if muxUpload.MaxWidth != nil {
		mw32 := int32(*muxUpload.MaxWidth)
		muxUploadpb.Width = &mw32
	}
	if muxUpload.MaxHeight != nil {
		mh32 := int32(*muxUpload.MaxHeight)
		muxUploadpb.Height = &mh32
	}

	return muxUploadpb
}
