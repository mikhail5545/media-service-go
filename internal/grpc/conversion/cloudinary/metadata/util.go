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
	metamodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/metadata"
	bytesutil "github.com/mikhail5545/media-service-go/internal/util/bytes"
	cldmetapbv1 "github.com/mikhail5545/media-service-go/pb/media_service/cloudinary/metadata/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Converter struct {
	logger *zap.Logger
}

func New(logger *zap.Logger) *Converter {
	return &Converter{
		logger: logger.With(zap.String("component", "grpc/conversion/cloudinary/metadata")),
	}
}

func (c *Converter) OwnerToProto(owner *metamodel.Owner) (*cldmetapbv1.Owner, error) {
	bytes, err := bytesutil.StrUUIDToBytes(owner.OwnerID)
	if err != nil {
		c.logger.Warn("Failed to convert uuid string to bytes", zap.Error(err))
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &cldmetapbv1.Owner{
		OwnerUuid: bytes,
		OwnerType: owner.OwnerType,
	}, nil
}

func (c *Converter) ToProto(metadata *metamodel.AssetMetadata) (*cldmetapbv1.AssetMetadata, error) {
	meta := &cldmetapbv1.AssetMetadata{
		Key: metadata.Key,
	}
	var err error
	meta.Owners, err = common.ConvertList(metadata.Owners, c.OwnerToProto)
	if err != nil {
		return nil, err
	}
	return meta, nil
}
