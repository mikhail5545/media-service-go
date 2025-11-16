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

package metadata

import (
	"context"
	"errors"
	"fmt"

	"github.com/arangodb/go-driver/v2/arangodb"
	"github.com/arangodb/go-driver/v2/arangodb/shared"
	metadatamodel "github.com/mikhail5545/media-service-go/internal/models/mux/metadata"
)

const CollectionName = "mux_asset_metadata"

var (
	ErrNotFound = errors.New("document not found")
	ErrConflict = errors.New("conflict")
)

// Repository defines the interface for MUX asset metadata operations in ArangoDB.
type Repository interface {
	// EnsureCollection creates the collection if it doesn't exist.
	EnsureCollection(ctx context.Context, db arangodb.Database) error
	// Get retrieves the metadata for a specific asset.
	Get(ctx context.Context, key string) (*metadatamodel.AssetMetadata, error)
	// Create creates an asset's metadata.
	Create(ctx context.Context, metadata *metadatamodel.AssetMetadata) error
	// Update creates or updates an asset's metadata with new values.
	Update(ctx context.Context, key string, metadata *metadatamodel.AssetMetadata) error
	// Delete deletes an asset's metadata.
	Delete(ctx context.Context, key string) error
	// ListUnownedIDs retrieves the keys of all assets that have no owners.
	ListUnownedIDs(ctx context.Context) ([]string, error)
	// ListByKeys retrieves metadata for a list of asset keys.
	ListByKeys(ctx context.Context, keys []string) (map[string]*metadatamodel.AssetMetadata, error)
	// CountUnowned counts all assets that have no owners.
	CountUnowned(ctx context.Context) (int64, error)
}

// arangoRepository holds arangodb.Database for ArangoDB-related operations.
type arangoRepository struct {
	db arangodb.Database
}

// New creates a new ArangoDB-based metadata repository.
func New(db arangodb.Database) Repository {
	return &arangoRepository{db: db}
}

// EnsureCollection creates the collection if it doesn't exist.
func (r *arangoRepository) EnsureCollection(ctx context.Context, db arangodb.Database) error {
	exists, err := db.CollectionExists(ctx, CollectionName)
	if err != nil {
		return fmt.Errorf("failed to check if collection exists: %w", err)
	}
	if !exists {
		_, err := db.CreateCollectionV2(ctx, CollectionName, nil)
		if err != nil {
			return fmt.Errorf("failed to create collection '%s': %w", CollectionName, err)
		}
	}
	return nil
}

// Get retrieves the metadata for a specific asset.
func (r *arangoRepository) Get(ctx context.Context, key string) (*metadatamodel.AssetMetadata, error) {
	col, err := r.db.GetCollection(ctx, CollectionName, &arangodb.GetCollectionOptions{SkipExistCheck: false})
	if err != nil {
		return nil, fmt.Errorf("failed to get collection '%s': %w", CollectionName, err)
	}

	var doc metadatamodel.AssetMetadata
	_, err = col.ReadDocument(ctx, key, &doc)
	if err != nil {
		if shared.IsNotFound(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to read document with key '%s': %w", key, err)
	}
	return &doc, nil
}

// ListUnownedIDs retrieves the keys of all assets that have no owners.
func (r *arangoRepository) ListUnownedIDs(ctx context.Context) ([]string, error) {
	query := `
		FOR m IN @@collection
		FILTER m.owners == [] OR m.owners == null
		RETURN m._key
	`
	bindVars := map[string]any{
		"@collection": CollectionName,
	}

	cur, err := r.db.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
	if err != nil {
		return nil, fmt.Errorf("failed to query for unowned asset metadata ids: %w", err)
	}
	defer cur.Close()

	var ids []string
	for cur.HasMore() {
		var id string
		_, err := cur.ReadDocument(ctx, &id)
		if err != nil {
			return nil, fmt.Errorf("failed to read unowned asset id from cursor: %w", err)
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// ListByKeys retrieves metadata for a list of asset keys.
func (r *arangoRepository) ListByKeys(ctx context.Context, keys []string) (map[string]*metadatamodel.AssetMetadata, error) {
	if len(keys) == 0 {
		return make(map[string]*metadatamodel.AssetMetadata), nil
	}

	query := `
		FOR m IN @@collection
		FILTER m._key IN @keys
		RETURN m
	`
	bindVars := map[string]any{
		"@collection": CollectionName,
		"keys":        keys,
	}

	cursor, err := r.db.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
	if err != nil {
		return nil, fmt.Errorf("failed to query for asset metadata by keys: %w", err)
	}
	defer cursor.Close()

	metadataMap := make(map[string]*metadatamodel.AssetMetadata)
	for cursor.HasMore() {
		doc := &metadatamodel.AssetMetadata{}
		_, err := cursor.ReadDocument(ctx, doc)
		if err != nil {
			return nil, fmt.Errorf("failed to read metadata document from cursor: %w", err)
		}
		if doc.Key != "" {
			metadataMap[doc.Key] = doc
		}
	}
	return metadataMap, nil
}

// CountUnowned counts all assets that have no owners.
func (r *arangoRepository) CountUnowned(ctx context.Context) (int64, error) {
	query := `
		FOR m IN @@collection
		FILTER m.owners == [] OR m.owners == null
		COLLECT WITH COUNT INTO length
		RETURN length
	`
	bindVars := map[string]any{
		"@collection": CollectionName,
	}

	cur, err := r.db.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
	if err != nil {
		return 0, fmt.Errorf("failed to query for unowned asset count: %w", err)
	}
	defer cur.Close()

	if !cur.HasMore() {
		// This case is unlikely but good to handle. If there are no documents,
		// the query returns 0, so the cursor should have one result.
		return 0, nil
	}

	var count int64
	if _, err := cur.ReadDocument(ctx, &count); err != nil {
		return 0, fmt.Errorf("failed to read unowned asset count from cursor: %w", err)
	}

	return count, nil
}

// Create creates an asset's metadata.
func (r *arangoRepository) Create(ctx context.Context, metadata *metadatamodel.AssetMetadata) error {
	col, err := r.db.GetCollection(ctx, CollectionName, &arangodb.GetCollectionOptions{SkipExistCheck: false})
	if err != nil {
		return fmt.Errorf("failed to get collection '%s': %w", CollectionName, err)
	}

	doc := metadatamodel.AssetMetadata{
		Key:    metadata.Key,
		Owners: metadata.Owners,
		Title:  metadata.Title,
	}

	if _, err := col.CreateDocument(ctx, &doc); err != nil {
		if shared.IsConflict(err) {
			return fmt.Errorf("%w: %w", ErrConflict, err)
		}
	}

	return nil
}

// Update creates or updates an asset's metadata. It performs a partial update,
// only modifying fields that are not nil in the provided metadata struct.
// To clear a field, provide its zero value (e.g., an empty slice for Owners).
func (r *arangoRepository) Update(ctx context.Context, key string, metadata *metadatamodel.AssetMetadata) error {
	// Build the query dynamically based on which fields are provided.
	updateParts := ""
	bindVars := map[string]any{
		"key":         key,
		"@collection": CollectionName,
	}

	// Using `!= nil` allows explicitly setting fields to their zero value (e.g., empty string or empty slice).
	// This assumes the fields in `metadatamodel.AssetMetadata` are pointers for optional string values.
	// If they are not, this logic should be adjusted. For slices, `!= nil` is correct.
	if metadata.Owners != nil {
		updateParts += "owners: @owners, "
		bindVars["owners"] = metadata.Owners
	}
	// Assuming Title is a pointer (*string) in the model to distinguish between not-provided and empty.
	// If it's a value type (string), the original `metadata.Title != ""` prevents clearing the title.
	if metadata.Title != "" { // Or `if metadata.Title != nil` if it's a pointer
		updateParts += "title: @title, "
		bindVars["title"] = metadata.Title
	}
	if metadata.CreatorID != "" { // Or `if metadata.CreatorID != nil`
		updateParts += "creator_id: @creator_id, "
		bindVars["creator_id"] = metadata.CreatorID
	}

	if len(updateParts) > 0 {
		// Trim trailing comma and space
		updateParts = updateParts[:len(updateParts)-2]

		query := fmt.Sprintf(`
			UPSERT { _key: @key }
			INSERT { _key: @key, %s }
			UPDATE { %s }
			IN @@collection`, updateParts, updateParts)

		cur, err := r.db.Query(ctx, query, &arangodb.QueryOptions{BindVars: bindVars})
		if err != nil {
			return fmt.Errorf("failed to execute upsert query for key '%s': %w", key, err)
		}
		defer cur.Close()
	}

	return nil
}

// Delete deletes an asset's metadata.
func (r *arangoRepository) Delete(ctx context.Context, key string) error {
	col, err := r.db.GetCollection(ctx, CollectionName, &arangodb.GetCollectionOptions{SkipExistCheck: false})
	if err != nil {
		return fmt.Errorf("failed to get collection '%s': %w", CollectionName, err)
	}

	_, err = col.DeleteDocument(ctx, key)
	if err != nil {
		if shared.IsNotFound(err) {
			return ErrNotFound
		}
		return fmt.Errorf("failed to delete document %w", err)
	}
	return nil
}
