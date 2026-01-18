// Package asset provides models, DTO models for [mux.Service] requests, webhooks and validation tools.
package asset

import (
	"time"

	"github.com/google/uuid"
	"github.com/mikhail5545/media-service-go/internal/models/mux/types"
	"gorm.io/gorm"
)

// UploadStatus represents the status of the external VOD provider upload.
type UploadStatus string

const (
	UploadStatusPreparing UploadStatus = "preparing"
	UploadStatusReady     UploadStatus = "ready"
	UploadStatusErrored   UploadStatus = "errored"
	UploadStatusDeleted   UploadStatus = "deleted"
)

// Status represents the internal status of the mux asset.
type Status string

const (
	StatusUploadURLGenerated Status = "upload_url_generated"
	StatusActive             Status = "active"
	StatusArchived           Status = "archived"
	StatusBroken             Status = "broken"
)

type State string

const (
	StateIngesting   State = "ingesting"
	StateTranscoding State = "transcoding"
	StateCompleted   State = "completed"
	StateLive        State = "live"
	StateErrored     State = "errored"
)

type IngestType string

const (
	IngestTypeOnDemandURL          IngestType = "on_demand_url"
	IngestTypeOnDemandDirectUpload IngestType = "on_demand_direct_upload"
	IngestTypeOnDemandClip         IngestType = "on_demand_clip"
	IngestTypeLiveRTMP             IngestType = "live_rtmp"
	IngestTypeLiveSRT              IngestType = "live_srt"
)

// Asset represents local mux Asset model.
type Asset struct {
	// Internal unique identifier for the mux asset.
	ID        uuid.UUID      `gorm:"primaryKey;type:uuid" json:"id"` // UUIDv7
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
	// Unique identifier for the direct upload (External mux API id). This field is
	// populated from the mux webhooks.
	MuxUploadID *string `gorm:"null" json:"mux_upload_id,omitempty"`
	// Unique identifier for the mux asset (External mux API id). This field is
	// populated from the mux webhooks.
	MuxAssetID *string `gorm:"null" json:"mux_asset_id,omitempty"`
	// The detailed state of the asset ingest progress. This field is useful for
	// relaying more granular processing information to end users when a non-standard input
	// is encountered.
	//
	//	"ingesting", "transcoding", "completed", "live", "errored"
	State State `gorm:"null" json:"state,omitempty"`
	// The status of the primary mux asset track.
	//
	// 	"preparing", "ready", "errored", "deleted"
	UploadStatus UploadStatus `gorm:"null;type:varchar(50)" json:"upload_status,omitempty"`
	Status       Status       `gorm:"type:varchar(50);default:'active';not null" json:"status"`
	// The duration of the asset in seconds (max duration for a single asset is 12 hours).
	Duration *float32 `gorm:"null" json:"duration,omitempty"`
	// The aspect ratio of the asset.
	//
	// 	"width:height" -> "16:9"
	AspectRatio    *string    `gorm:"null" json:"aspect_ratio,omitempty"`
	AssetCreatedAt *time.Time `gorm:"null" json:"asset_created_at,omitempty"`
	// The resolution tier that the asset was ingested at, affecting billing for ingest & storage.
	// The asset may be delivered at lower resolutions depending on the device and bandwidth, however
	// it cannot be delivered at a higher value than is stored.
	//
	//	"audio-only", "720p", "1080p", "1440p", "2160p"
	ResolutionTier *string `gorm:"null" json:"resolution_tier,omitempty"`
	// The type of ingest used to create the asset.
	//
	//	"on_demand_url", "on_demand_direct_upload", "on_demand_clip", "live_rtmp", "live_srt"
	IngestType IngestType `gorm:"null" json:"ingest_type,omitempty"`

	// --- PlaybackIDs ---

	// PrimarySignedPlaybackID facilitates quick lookup for token generation without querying Mongo.
	// Populated from the first 'signed' policy playbackID found in the webhook metadata.
	PrimarySignedPlaybackID *string `gorm:"type:varchar(255);null;index" json:"primary_signed_playback_id,omitempty"`
	// PrimaryPublicPlaybackID facilitates quick lookup for public playback without querying Mongo.
	// Populated from the first 'public' policy playbackID found in the webhook metadata.
	PrimaryPublicPlaybackID *string `gorm:"type:varchar(255);null;index" json:"primary_public_playback_id,omitempty"`

	// --- Audit fields ---

	CreatedBy        *uuid.UUID `gorm:"type:uuid;null" json:"created_by,omitempty"`
	ArchivedBy       *uuid.UUID `gorm:"type:uuid;null" json:"archived_by,omitempty"`
	RestoredBy       *uuid.UUID `gorm:"type:uuid;null" json:"restored_by,omitempty"`
	MarkedAsBrokenBy *uuid.UUID `gorm:"type:uuid;null" json:"marked_as_broken_by,omitempty"`

	CreatedByName        *string `gorm:"type:varchar(128);null" json:"created_by_name,omitempty"`
	ArchivedByName       *string `gorm:"type:varchar(128);null" json:"archived_by_name,omitempty"`
	RestoredByName       *string `gorm:"type:varchar(128);null" json:"restored_by_name,omitempty"`
	MarkedAsBrokenByName *string `gorm:"type:varchar(128);null" json:"marked_as_broken_by_name,omitempty"`

	Note          *string `gorm:"type:varchar(512)" json:"note,omitempty"`
	ArchiveReason *string `gorm:"type:varchar(512)" json:"archive_reason,omitempty"`

	// ArchiveEventID is the MUX webhook event ID that caused the asset to be archived.
	// Used for idempotency to avoid archiving the same asset multiple times on repeated webhooks.
	ArchiveEventID *string                `gorm:"type:varchar(255);null" json:"archive_event_id,omitempty"`
	MuxError       *types.MuxWebhookError `gorm:"type:jsonb;null" json:"mux_error,omitempty"`
}

func (*Asset) TableName() string {
	return "mux_assets"
}

// BeforeCreate is a GORM hook that is called before a record is created.
// It generates ID.
func (a *Asset) BeforeCreate(tx *gorm.DB) (err error) {
	if a.ID == uuid.Nil {
		a.ID, err = uuid.NewV7()
		if err != nil {
			return err
		}
	}
	return nil
}

// BeforeDelete is a GORM hook that is called before a record is deleted.
// It sets the status to 'archived'.
func (a *Asset) BeforeDelete(tx *gorm.DB) (err error) {
	return tx.Model(a).Update("status", StatusArchived).Error
}
