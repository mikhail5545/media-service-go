package metadata

import "github.com/mikhail5545/media-service-go/internal/models/mux/types"

// AssetMetadata represents the metadata for a MUX asset stored in MongoDB.
type AssetMetadata struct {
	// The _key field will be internal asset ID from PostgreSQL database.
	Key         string                        `bson:"_id,omitempty" json:"_key,omitempty"`
	Title       string                        `bson:"title" json:"title"`
	CreatorID   string                        `bson:"creator_id" json:"creator_id"`
	Owners      []*Owner                      `bson:"owners" json:"owners"`
	Tracks      []*types.MuxWebhookTrack      `bson:"tracks" json:"tracks"`
	PlaybackIDs []*types.MuxWebhookPlaybackID `bson:"playback_ids" json:"playback_ids"`
}

// Owner represents an entity that is associated with an asset.
type Owner struct {
	OwnerID   string `bson:"owner_id" json:"owner_id"`
	OwnerType string `bson:"owner_type" json:"owner_type"`
}
