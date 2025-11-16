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

// Package asset provides models, DTO models for [mux.Service] requests, webhooks and vdlitaion tools.
package asset

import (
	"time"

	"gorm.io/gorm"
)

// Asset represents local mux Asset model.
type Asset struct {
	// Internal unique identifier for the mux asset.
	ID        string         `gorm:"primaryKey;size:36" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at"`
	// Unique identifier for the direct upload (External mux API id). This field is
	// populated from the mux webhooks.
	MuxUploadID *string `gorm:"null" json:"mux_upload_id,omitempty"`
	// Unique identifier for the mux asset (External mux API id). This field is
	// populated from the mux webhooks.
	MuxAssetID *string `gorm:"null" json:"mux_asset_id,omitempty"`
	// Slice of Mux PlaybackID objects. This field is populated from the mux webhooks. Stored as a JSON array.
	PlaybackIDs []MuxWebhookPlaybackID `gorm:"type:jsonb" json:"playback_ids,omitempty"`
	// Slice of Mux Track objects. This field is populated from the mux webhooks. Stored as a JSON array.
	Tracks []MuxWebhookTrack `gorm:"type:jsonb" json:"tracks,omitempty"`
	// The detailed state of the asset ingest progress. This field is useful for
	// relaying more granular processin information to end users when a non-standard input
	// is encountered.
	//
	//	"ingesting", "transcoding", "completed", "live", "errored"
	State string `gorm:"null" json:"state,omitempty"`
	// The status of the primary mux asset track.
	//
	// 	"preparing", "ready", "errored", "deleted"
	Status *string `gorm:"null" json:"status,omitempty"`
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
	IngestType *string `gorm:"null" json:"ingest_type,omitempty"`
}
