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

	"github.com/google/uuid"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	bytesutil "github.com/mikhail5545/media-service-go/internal/util/bytes"
	errutil "github.com/mikhail5545/media-service-go/internal/util/errors"
	videopbv1 "github.com/mikhail5545/product-service-client/pb/product_service/video/v1"
	"go.uber.org/zap"
)

func (s *Service) grpcMarkAsBroken(ctx context.Context, assetID, adminID *uuid.UUID, req *assetmodel.ChangeStateRequest) error {
	assetIDBytes, err := bytesutil.UUIDToBytes(assetID)
	if err != nil {
		return err
	}
	adminIDBytes, err := bytesutil.UUIDToBytes(adminID)
	if err != nil {
		return err
	}
	if _, err := s.videoClient.BrokenVideo(ctx, &videopbv1.BrokenVideoRequest{
		MediaServiceUuid: assetIDBytes,
		AdminUuid:        adminIDBytes,
		AdminName:        req.AdminName,
		Reason:           req.Note,
	}); err != nil {
		s.logger.Error("failed to mark asset as broken via gRPC", zap.Error(err), zap.String("asset_id", req.ID))
		return errutil.HandleRPCError(err)
	}
	return nil
}

func (s *Service) grpcForceDelete(ctx context.Context, assetID *uuid.UUID) error {
	assetIDBytes, err := bytesutil.UUIDToBytes(assetID)
	if err != nil {
		return err
	}
	if _, err := s.videoClient.ForceDelete(ctx, &videopbv1.ForceDeleteRequest{
		MediaServiceUuid: assetIDBytes,
	}); err != nil {
		s.logger.Error("failed to force delete asset via gRPC", zap.Error(err), zap.String("asset_id", assetID.String()))
		return errutil.HandleRPCError(err)
	}
	return nil
}
