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

// Package types provides function to convert data types.
package types

import (
	"time"

	muxpb "github.com/mikhail5545/proto-go/proto/media_service/mux/asset/v0"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertToMuxProtoBuf(muxUpload *mux.Upload) *muxpb.MuxUpload {
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
	if muxUpload.Width != nil {
		mw32 := int32(*muxUpload.Width)
		muxUploadpb.Width = &mw32
	}
	if muxUpload.Height != nil {
		mh32 := int32(*muxUpload.Height)
		muxUploadpb.Height = &mh32
	}

	return muxUploadpb
}

func MuxUploadToProtobufUpdate(resp *muxpb.UpdateResponse, updates map[string]any) *muxpb.UpdateResponse {
	resp.Updated = &fieldmaskpb.FieldMask{}
	for k, v := range updates {
		switch k {
		case "mux_upload_id":
			if val, ok := v.(string); ok {
				resp.MuxUploadId = &val
				resp.Updated.Paths = append(resp.Updated.Paths, "updateresponse.mux_upload_id")
			}
		case "mux_asset_id":
			if val, ok := v.(string); ok {
				resp.MuxAssetId = &val
				resp.Updated.Paths = append(resp.Updated.Paths, "updateresponse.mux_asset_id")
			}
		case "mux_playback_id":
			if val, ok := v.(string); ok {
				resp.MuxPlaybackId = &val
				resp.Updated.Paths = append(resp.Updated.Paths, "updateresponse.mux_playback_id")
			}
		case "video_processing_status":
			if val, ok := v.(string); ok {
				resp.VideoProcessingStatus = &val
				resp.Updated.Paths = append(resp.Updated.Paths, "updateresponse.video_processing_status")
			}
		case "duration":
			if val, ok := v.(float32); ok {
				resp.Duration = &val
				resp.Updated.Paths = append(resp.Updated.Paths, "updateresponse.duration")
			}
		case "aspect_ratio":
			if val, ok := v.(string); ok {
				resp.AspectRatio = &val
				resp.Updated.Paths = append(resp.Updated.Paths, "updateresponse.aspect_ratio")
			}
		case "width":
			if val, ok := v.(int32); ok {
				resp.Width = &val
				resp.Updated.Paths = append(resp.Updated.Paths, "updateresponse.width")
			}
		case "height":
			if val, ok := v.(int32); ok {
				resp.Height = &val
				resp.Updated.Paths = append(resp.Updated.Paths, "updateresponse.height")
			}
		case "asset_created_at":
			if val, ok := v.(time.Time); ok {
				resp.AssetCreatedAt = timestamppb.New(val)
				resp.Updated.Paths = append(resp.Updated.Paths, "updateresponse.asset_created_at")
			}
		}
	}
	return resp
}
