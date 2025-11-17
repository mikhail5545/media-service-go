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
	cldassetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	cldmetamodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/metadata"
	muxassetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	cldassetpb "github.com/mikhail5545/proto-go/proto/media_service/cloudinary/asset/v0"
	muxassetpb "github.com/mikhail5545/proto-go/proto/media_service/mux/asset/v0"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func MuxAssetResponseToProtofuf(response *muxassetmodel.AssetResponse) *muxassetpb.AssetResponse {
	pbResponse := &muxassetpb.AssetResponse{
		Asset: &muxassetpb.Asset{
			Id:             response.Asset.ID,
			CreatedAt:      timestamppb.New(response.CreatedAt),
			UpdatedAt:      timestamppb.New(response.UpdatedAt),
			MuxUploadId:    response.MuxUploadID,
			MuxAssetId:     response.MuxAssetID,
			State:          response.State,
			Status:         response.Status,
			AspectRatio:    response.AspectRatio,
			ResolutionTier: response.ResolutionTier,
			IngestType:     response.IngestType,
		},
		Title:     response.Title,
		CreatorId: response.CreatorID,
	}

	if response.DeletedAt.Valid {
		pbResponse.Asset.DeletedAt = timestamppb.New(response.DeletedAt.Time)
	}
	if response.Asset.Duration != nil {
		d := float32(*response.Duration)
		pbResponse.Asset.Duration = &d
	}
	if response.AssetCreatedAt != nil {
		pbResponse.Asset.AssetCreatedAt = timestamppb.New(*response.AssetCreatedAt)
	}

	for _, track := range response.Tracks {
		pbResponse.Tracks = append(pbResponse.Tracks, MuxTrackToProtobuf(&track))
	}
	for _, owner := range response.Owners {
		pbResponse.Owners = append(pbResponse.Owners, &muxassetpb.Owner{OwnerId: owner.OwnerID, OwnerType: owner.OwnerType})
	}
	for _, playbackID := range response.MuxPlaybackIDs {
		pbResponse.Asset.MuxPlaybackIds = append(pbResponse.Asset.MuxPlaybackIds, &muxassetpb.MuxPlaybackID{Id: playbackID.ID, Policy: playbackID.Policy, DrmConfigurationId: *playbackID.DrmConfigurationID})
	}

	return pbResponse
}

func MuxTrackToProtobuf(track *muxassetmodel.MuxWebhookTrack) *muxassetpb.MuxTrack {
	pbTrack := &muxassetpb.MuxTrack{
		Id:             track.ID,
		Type:           track.Type,
		MaxWidth:       track.MaxWidth,
		MaxHeight:      track.MaxHeight,
		MaxChannels:    track.MaxChannels,
		TextType:       track.TextType,
		TextSource:     track.TextSource,
		LanguageCode:   track.LanguageCode,
		Name:           track.Name,
		ClosedCaptions: track.ClosedCaptions,
		Passthrough:    track.Passthrough,
		Status:         track.Status,
		Primary:        track.Primary,
		Errors: &muxassetpb.MuxError{
			Type:     track.Errors.Type,
			Messages: track.Errors.Messages,
		},
	}
	if track.Duration != nil {
		d := float32(*track.Duration)
		pbTrack.Duration = &d
	}
	if track.MaxFrameRate != nil {
		d := float32(*track.MaxFrameRate)
		pbTrack.MaxFrameRate = &d
	}
	return pbTrack
}

func CldAssetResponseToProtobuf(response *cldassetmodel.AssetResponse) *cldassetpb.AssetResponse {
	pbResponse := &cldassetpb.AssetResponse{
		Asset: &cldassetpb.Asset{
			Id:                 response.ID,
			CreatedAt:          timestamppb.New(response.CreatedAt),
			UpdatedAt:          timestamppb.New(response.UpdatedAt),
			CloudinaryAssetId:  response.CloudinaryAssetID,
			CloudinaryPublicId: response.CloudinaryPublicID,
			Url:                response.URL,
			SecureUrl:          response.SecureURL,
			ResourceType:       response.ResourceType,
			Format:             response.Format,
			Tags:               response.Tags,
			AssetFolder:        response.AssetFolder,
			DisplayName:        response.DisplayName,
		},
	}
	if response.Width != nil {
		w := int32(*response.Width)
		pbResponse.Asset.Width = &w
	}
	if response.Height != nil {
		h := int32(*response.Height)
		pbResponse.Asset.Height = &h
	}
	if response.DeletedAt.Valid {
		pbResponse.Asset.DeletedAt = timestamppb.New(response.DeletedAt.Time)
	}

	for _, owner := range response.Owners {
		pbResponse.Owners = append(pbResponse.Owners, &cldassetpb.Owner{OwnerId: owner.OwnerID, OwnerType: owner.OwnerType})
	}
	return pbResponse
}

func SuccessRequestFromProtobuf(pbReq *cldassetpb.SuccessfulUploadRequest) *cldassetmodel.SuccessfulUploadRequest {
	req := &cldassetmodel.SuccessfulUploadRequest{
		CloudinaryAssetID:  pbReq.GetCloudinaryAssetId(),
		CloudinaryPublicID: pbReq.GetCloudinaryPublicId(),
		ResourceType:       pbReq.GetResourceType(),
		Format:             pbReq.GetFormat(),
		URL:                pbReq.GetUrl(),
		SecureURL:          pbReq.GetSecureUrl(),
		AssetFolder:        pbReq.GetAssetFolder(),
		DisplayName:        pbReq.GetDisplayName(),
	}
	for _, owner := range pbReq.GetOwners() {
		req.Owners = append(req.Owners, cldmetamodel.Owner{OwnerID: owner.GetOwnerId(), OwnerType: owner.GetOwnerType()})
	}

	if pbReq.Width != nil {
		w := int(*pbReq.Width)
		req.Width = &w
	}
	if pbReq.Height != nil {
		h := int(*pbReq.Height)
		req.Height = &h
	}
	return req
}
