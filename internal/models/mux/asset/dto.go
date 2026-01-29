package asset

import (
	"github.com/google/uuid"
	"github.com/mikhail5545/media-service-go/internal/models/mux/metadata"
)

type OrderField string

const (
	OrderCreatedAt  OrderField = "created_at"
	OrderUpdatedAt  OrderField = "updated_at"
	OrderIngestType OrderField = "ingest_type"
)

type OrderDirection string

const (
	OrderAscending  OrderDirection = "ASC"
	OrderDescending OrderDirection = "DESC"
)

type GetFilter struct {
	ID           string       `param:"id" json:"-"`
	UploadStatus UploadStatus `query:"upload_status" json:"upload_status"`
}

type ListRequest struct {
	IDs          []string `query:"ids"`
	MuxUploadIDs []string `query:"mux_upload_ids"`
	MuxAssetIDs  []string `query:"mux_asset_ids"`

	AspectRatios    []string       `query:"aspect_ratios"`
	ResolutionTiers []string       `query:"resolution_tiers"`
	IngestTypes     []IngestType   `query:"ingest_types"`
	UploadStatuses  []UploadStatus `query:"upload_statuses"`

	OrderBy  OrderField     `query:"order_by"`
	OrderDir OrderDirection `query:"order_dir"`

	PageSize  int    `query:"page_size"`
	PageToken string `query:"page_token"`
}

// Details is a DTO that combines the core Asset model with its metadata.
type Details struct {
	Asset    *Asset
	Metadata *metadata.AssetMetadata
}

type CreateUploadURLRequest struct {
	Title     string `json:"title"`
	AdminID   string `json:"admin_id"`
	AdminName string `json:"admin_name"`
}

type ChangeStateRequest struct {
	ID        string `param:"id" json:"-"`
	AdminID   string `json:"admin_id"`
	AdminName string `json:"admin_name"`
	Note      string `json:"note"`
}

type ManageOwnerRequest struct {
	ID        string `param:"id" json:"-"`
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`
}

// GeneratePlaybackTokenRequest represents a request to generate a playback token for a MUX asset.
// Not intended to be used within HTTP handlers, only for service-to-service communication over gRPC.
type GeneratePlaybackTokenRequest struct {
	AssetID    uuid.UUID
	UserID     uuid.UUID
	Expiration int64      // in seconds
	SessionID  *uuid.UUID // optional
	UserAgent  *string    // optional
}
