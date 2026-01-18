package metadata

// AssetMetadata represents the metadata for a Cloudinary asset stored in ArangoDB.
type AssetMetadata struct {
	// The _key field will be internal asset ID from PostgreSQL database.
	Key    string  `json:"_key,omitempty"`
	Owners []Owner `json:"owners"`
}

// Owner represents an entity that is associated with an asset.
type Owner struct {
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`
}
