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
Package mux provides service-layer business logic for for mux asset model.
*/
package mux

import (
	"fmt"

	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	detailmodel "github.com/mikhail5545/media-service-go/internal/models/mux/detail"
	metamodel "github.com/mikhail5545/media-service-go/internal/models/mux/metadata"
	"google.golang.org/grpc/status"
)

// handleGRPCError is a helper function to handle gRPC client errors and return [mux.Error] with
// appropriate message and status code.
func handleGRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("unexpected error occurred: %w", err)
	}
	return fmt.Errorf("(gRPC call ended with code %d) %w: %s", st.Code(), st.Err(), st.Message())
}

func groupOwnersByTypeFromMetadata(owners []metamodel.Owner) map[string]map[string]struct{} {
	grouped := make(map[string]map[string]struct{})
	for _, owner := range owners {
		if _, ok := grouped[owner.OwnerType]; !ok {
			grouped[owner.OwnerType] = make(map[string]struct{})
		}
		grouped[owner.OwnerType][owner.OwnerID] = struct{}{}
	}
	return grouped
}

func diffOwnerMaps(old, new map[string]map[string]struct{}) (toAdd, toDelete map[string][]string) {
	toAdd = make(map[string][]string)
	toDelete = make(map[string][]string)

	// Find what to add
	for ownerType, newIDs := range new {
		oldIDs, oldTypeExists := old[ownerType]
		for id := range newIDs {
			if !oldTypeExists {
				toAdd[ownerType] = append(toAdd[ownerType], id)
				continue
			}
			if _, found := oldIDs[id]; !found {
				toAdd[ownerType] = append(toAdd[ownerType], id)
			}
		}
	}

	// Find what to delete
	for ownerType, oldIDs := range old {
		newIDs, newTypeExists := new[ownerType]
		if !newTypeExists { // whole type was removed
			for id := range oldIDs {
				toDelete[ownerType] = append(toDelete[ownerType], id)
			}
			continue
		}
		for id := range oldIDs {
			if _, found := newIDs[id]; !found {
				toDelete[ownerType] = append(toDelete[ownerType], id)
			}
		}
	}

	return toAdd, toDelete
}

// combineAssetAndMetadata is a helper to merge an Asset and its metadata into an AssetResponse DTO.
func (s *service) combineAssetAndMetadata(
	asset *assetmodel.Asset,
	metadata *metamodel.AssetMetadata,
	details *detailmodel.AssetDetail,
) *assetmodel.AssetResponse {
	response := &assetmodel.AssetResponse{
		Asset: asset,
	}

	if metadata != nil {
		response.Title = metadata.Title
		response.CreatorID = metadata.CreatorID
		response.Owners = metadata.Owners
	}

	if details != nil {
		response.Tracks = details.Tracks
	}

	return response
}
