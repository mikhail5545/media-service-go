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
Package cloudinary provides service-layer logic for Cloudinary asset management and asset models.
*/
package cloudinary

import (
	"context"
	"errors"
	"fmt"
	"log"

	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	imagepb "github.com/mikhail5545/proto-go/proto/product_service/image/v0"
)

// processBatchGRPC is a generic helper to process batch gRPC calls for different owner types.
func processBatchGRPC[T any](
	ctx context.Context,
	owners map[string][]string,
	callGRPC func(ctx context.Context, ownerType string, ids []string) (int64, error),
) error {
	var allErrors []error
	for ownerType, ids := range owners {
		ownersAffected, err := callGRPC(ctx, ownerType, ids)
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("owner type %s: %w", ownerType, handleGRPCError(err)))
			continue
		}

		if int(ownersAffected) != len(ids) {
			log.Printf("For owner type '%s', owners affected: %d out of %d", ownerType, ownersAffected, len(ids))
		}
	}

	if len(allErrors) > 0 {
		return errors.Join(allErrors...)
	}
	return nil
}

func (s *service) processAddBatch(ctx context.Context, asset *assetmodel.Asset, owners map[string][]string) error {
	addFunc := func(ctx context.Context, ownerType string, ids []string) (int64, error) {
		resp, err := s.ImageSvcClient.AddBatch(ctx, &imagepb.AddBatchRequest{
			PublicId:       asset.CloudinaryPublicID,
			Url:            asset.URL,
			SecureUrl:      asset.SecureURL,
			MediaServiceId: asset.ID,
			OwnerIds:       ids,
			OwnerType:      ownerType,
		})
		if err != nil {
			return 0, err
		}
		return resp.GetOwnersAffected(), nil
	}
	return processBatchGRPC[imagepb.AddBatchRequest](ctx, owners, addFunc)
}

func (s *service) processDeleteBatch(ctx context.Context, asset *assetmodel.Asset, owners map[string][]string) error {
	deleteFunc := func(ctx context.Context, ownerType string, ids []string) (int64, error) {
		resp, err := s.ImageSvcClient.DeleteBatch(ctx, &imagepb.DeleteBatchRequest{
			MediaServiceId: asset.ID,
			OwnerIds:       ids,
			OwnerType:      ownerType,
		})
		if err != nil {
			return 0, err
		}
		return resp.GetOwnersAffected(), nil
	}
	return processBatchGRPC[imagepb.DeleteBatchRequest](ctx, owners, deleteFunc)
}

func (s *service) processChanges(ctx context.Context, asset *assetmodel.Asset, toAdd, toDelete map[string][]string) error {
	var allErrors []error
	if len(toAdd) > 0 {
		if err := s.processAddBatch(ctx, asset, toAdd); err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed to process additions: %w", err))
		}
	}
	if len(toDelete) > 0 {
		if err := s.processDeleteBatch(ctx, asset, toDelete); err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed to process deletions: %w", err))
		}
	}
	return errors.Join(allErrors...)
}
