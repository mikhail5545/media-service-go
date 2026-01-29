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

package cloudinary

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	bytesutil "github.com/mikhail5545/media-service-go/internal/util/bytes"
	imagepbv1 "github.com/mikhail5545/product-service-client/pb/product_service/image/v1"
	"go.uber.org/zap"
)

func (s *Service) grpcMarkAsBroken(ctx context.Context, assetID, adminID *uuid.UUID, req *assetmodel.ChangeStateRequest) error {
	bytes, err := bytesutil.UUIDToBytes(assetID)
	if err != nil {
		return err
	}
	adminIDBytes, err := bytesutil.UUIDToBytes(adminID)
	if err != nil {
		return err
	}
	grpcReq := &imagepbv1.BrokenImageRequest{
		MediaServiceUuid: bytes,
		AdminUuid:        adminIDBytes,
		Reason:           req.Note,
		AdminName:        req.AdminName,
	}
	if _, err := s.imageServiceClient.BrokenImage(ctx, grpcReq); err != nil {
		s.logger.Error("failed to mark asset as broken via gRPC", zap.Error(err), zap.String("asset_id", assetID.String()))
		return fmt.Errorf("failed to mark asset as broken via gRPC: %w", err)
	}
	return nil
}

func (s *Service) grpcDelete(ctx context.Context, assetID *uuid.UUID) error {
	bytes, err := bytesutil.UUIDToBytes(assetID)
	if err != nil {
		return err
	}
	grpcReq := &imagepbv1.DeleteRequest{
		MediaServiceUuid: bytes,
	}
	if _, err := s.imageServiceClient.Delete(ctx, grpcReq); err != nil {
		s.logger.Error("failed to delete asset via gRPC", zap.Error(err), zap.String("asset_id", assetID.String()))
		return fmt.Errorf("failed to delete asset via gRPC: %w", err)
	}
	return nil
}
