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

package metadata

import (
	"github.com/mikhail5545/media-service-go/internal/grpc/conversion/common"
	"github.com/mikhail5545/media-service-go/internal/grpc/conversion/mux/webhooks"
	metamodel "github.com/mikhail5545/media-service-go/internal/models/mux/metadata"
	bytesutil "github.com/mikhail5545/media-service-go/internal/util/bytes"
	muxmetapbv1 "github.com/mikhail5545/media-service-go/pb/media_service/mux/metadata/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Converter struct {
	logger *zap.Logger
}

func New(logger *zap.Logger) *Converter {
	return &Converter{
		logger: logger.With(zap.String("component", "grpc/mux/metadata/Converter")),
	}
}

func (c *Converter) OwnerToProto(owner *metamodel.Owner) (*muxmetapbv1.Owner, error) {
	bytes, err := bytesutil.StrUUIDToBytes(owner.OwnerID)
	c.logger.Error("failed to convert owner ID to bytes", zap.Error(err))
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &muxmetapbv1.Owner{
		OwnerUuid: bytes,
		OwnerType: owner.OwnerType,
	}, nil
}

func (c *Converter) convertAssociations(meta *metamodel.AssetMetadata, pbMeta *muxmetapbv1.AssetMetadata) (err error) {
	pbMeta.Owners, err = common.ConvertList(meta.Owners, c.OwnerToProto)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	pbMeta.PlaybackIds, err = common.ConvertList(meta.PlaybackIDs, webhooks.MuxWebhookPlaybackIDToProto)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	pbMeta.Tracks, err = common.ConvertList(meta.Tracks, webhooks.MuxWebhookTrackToPorto)
	if err != nil {
		return status.Error(codes.Internal, err.Error())
	}
	return nil
}

func (c *Converter) ToProto(metadata *metamodel.AssetMetadata) (*muxmetapbv1.AssetMetadata, error) {
	pbMeta := &muxmetapbv1.AssetMetadata{
		Key:       metadata.Key,
		Title:     metadata.Title,
		CreatorId: metadata.CreatorID,
	}
	if err := c.convertAssociations(metadata, pbMeta); err != nil {
		return nil, err
	}
	return pbMeta, nil
}
