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
	"strings"

	assetownerrepo "github.com/mikhail5545/media-service-go/internal/database/cloudinary/asset_owner"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	assetownermodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset_owner"
	metamodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/metadata"
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

func populateOwnersFromContext(customContext map[string]string, assetID string) ([]assetownermodel.AssetOwner, map[string][]string) {
	var ownersToCreate []assetownermodel.AssetOwner
	ownersByType := make(map[string][]string)
	for k, v := range customContext {
		if strings.HasSuffix(k, "_ids") {
			ownerType := strings.TrimSuffix(k, "_ids")
			ids := strings.Split(v, "|")
			ownersByType[ownerType] = ids
			for _, ownerID := range ids {
				if ownerID != "" {
					ownersToCreate = append(ownersToCreate, assetownermodel.AssetOwner{
						AssetID:   assetID,
						OwnerID:   ownerID,
						OwnerType: ownerType,
					})
				}
			}
		}
	}
	return ownersToCreate, ownersByType
}

func groupOwnersByType(owners []assetownermodel.AssetOwner) map[string]map[string]struct{} {
	grouped := make(map[string]map[string]struct{})
	for _, owner := range owners {
		if _, ok := grouped[owner.OwnerType]; !ok {
			grouped[owner.OwnerType] = make(map[string]struct{})
		}
		grouped[owner.OwnerType][owner.OwnerID] = struct{}{}
	}
	return grouped
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

func calculateNewOwnerState(currentState map[string]map[string]struct{}, resource *assetmodel.Resource) map[string]map[string]struct{} {
	// Deep copy the current state to avoid modifying the original map
	newState := make(map[string]map[string]struct{})
	for ownerType, ids := range currentState {
		newState[ownerType] = make(map[string]struct{})
		for id := range ids {
			newState[ownerType][id] = struct{}{}
		}
	}

	// Apply changes from the webhook
	for _, added := range resource.Added {
		if strings.HasSuffix(added.Name, "_ids") {
			ownerType := strings.TrimSuffix(added.Name, "_ids")
			if _, ok := newState[ownerType]; !ok {
				newState[ownerType] = make(map[string]struct{})
			}
			for id := range strings.SplitSeq(added.Value, "|") {
				if id != "" {
					newState[ownerType][id] = struct{}{}
				}
			}
		}
	}

	for _, removed := range resource.Removed {
		if strings.HasSuffix(removed.Name, "_ids") {
			ownerType := strings.TrimSuffix(removed.Name, "_ids")
			delete(newState, ownerType)
		}
	}

	for _, updated := range resource.Updated {
		if strings.HasSuffix(updated.Name, "_ids") {
			ownerType := strings.TrimSuffix(updated.Name, "_ids")
			newState[ownerType] = make(map[string]struct{}) // Clear old value
			for id := range strings.SplitSeq(updated.NewValue, "|") {
				if id != "" {
					newState[ownerType][id] = struct{}{}
				}
			}
		}
	}

	return newState
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

func (s *service) processOwnerChanges(ctx context.Context, repo assetownerrepo.Repository, assetID string, toAdd, toDelete map[string][]string) error {
	var allErrors []error

	// Process Deletions
	for ownerType, ids := range toDelete {
		if len(ids) > 0 {
			if _, err := repo.DeleteByOwnerTypeAndIDs(ctx, assetID, ownerType, ids); err != nil {
				allErrors = append(allErrors, fmt.Errorf("failed to delete old asset owner links for type %s: %w", ownerType, err))
			}
		}
	}

	// Process Additions
	if len(toAdd) > 0 {
		var ownersToCreate []assetownermodel.AssetOwner
		for ownerType, ids := range toAdd {
			for _, ownerID := range ids {
				ownersToCreate = append(ownersToCreate, assetownermodel.AssetOwner{
					AssetID:   assetID,
					OwnerID:   ownerID,
					OwnerType: ownerType,
				})
			}
		}
		if len(ownersToCreate) > 0 {
			if err := repo.CreateBatch(ctx, ownersToCreate); err != nil {
				allErrors = append(allErrors, fmt.Errorf("failed to create new asset owner links: %w", err))
			}
		}
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("failed to update asset owner links from context change: %w", errors.Join(allErrors...))
	}

	return nil
}

// combineAssetAndMetadata is a helper to merge an Asset and its metadata into an AssetResponse DTO.
func (s *service) combineAssetAndMetadata(asset *assetmodel.Asset, metadata *metamodel.AssetMetadata) *assetmodel.AssetResponse {
	response := &assetmodel.AssetResponse{
		Asset: asset,
	}

	if metadata != nil {
		response.Owners = metadata.Owners
	}

	return response
}
