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

package asset

import (
	"time"

	metamodel "github.com/mikhail5545/media-service-go/internal/models/mux/metadata"
)

// AssetResponse is a DTO that combines the core Asset model with its metadata.
type AssetResponse struct {
	*Asset
	// Title is populated from a separate ArangoDB table.
	Title string `json:"title,omitempty"`
	// CreatorID is populated from a separate ArangoDB table.
	CreatorID string `json:"creator_id,omitempty"`
	// Owners is populated from a separate ArangoDB table.
	Owners []metamodel.Owner `json:"owners,omitempty"`
	// Tracks are populated from a separate details PostgreSQL table.
	Tracks []MuxWebhookTrack `json:"tracks,omitempty"`
}

type UpdateOwnersRequest struct {
	ID     string            `json:"id"`
	Owners []metamodel.Owner `json:"owners"`
}

type CreateUploadURLRequest struct {
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`
	Title     string `json:"title"`
	CreatorID string `json:"creator_id"`
}

type CreateUnownedUploadURLRequest struct {
	Title     string `json:"title"`
	CreatorID string `json:"creator_id"`
}

type AssociateRequest struct {
	ID        string `json:"id"`
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`
}

type DeassociateRequest struct {
	ID        string `json:"id"`
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`
}

type UpdateMetadataRequest struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// MuxWebhook represents the mux webhook payload.
// See mux API [webhook reference] for more details.
//
// [webhook reference]: https://www.mux.com/docs/webhook-reference
type MuxWebhook struct {
	// Type for the webhook event
	Type string `json:"type"`
	// Unique identifier for the event
	ID string `json:"id"`
	// Time the event was created
	CreatedAt   time.Time             `json:"created_at"`
	Object      MuxWebhookObject      `json:"object"`
	Environment MuxWebhookEnvironment `json:"environment"`
	Data        MuxWebhookData        `json:"data"`
}

type MuxWebhookObject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

// MuxWebhookEnvironment represents `environment` object inside of the mux webhook event.
type MuxWebhookEnvironment struct {
	// Name for the environment
	Name string `json:"name"`
	// Unique identifier for the environment
	ID string `json:"id"`
}

// MuxWebhookData represents the mux webhook data object.
type MuxWebhookData struct {
	// Unique identifier for the asset. Max 255 characters.
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	// The status of the asset
	//
	// 	"created", "ready", "errored"
	Status *string `json:"status,omitempty"`
	// The duration of the asset in seconds (max duration for a single asset is 12 hours).
	Duration *float32 `json:"duration,omitempty"`
	// The resolution tier that the asset was ingested at, affecting billing for ingest & storage.
	// The asset may be delivered at lower resolutions depending on the device and bandwidth, however
	// it cannot be delivered at a higher value than is stored.
	//
	//	"audio-only", "720p", "1080p", "1440p", "2160p"
	ResolutionTier *string `json:"resolution_tier,omitempty"`
	// Max resolution tier can be used to control the maximum `resolution_tier` your asset is encoded,
	// stored and streamed at. If not set, this defauls to `1080p`.
	MaxResolutionTier *string `json:"max_resolution_tier,omitempty"`
	// The video quality controls the cost, quality, and available platform features for the asset. The
	// default video quality for an account can be set in the Mux Dashboard.
	VideoQuality *string `json:"video_quality,omitempty"`
	// The maximum frame rate that has been stored for the asset. The asset may be delivered
	// at lower frame rates depending on the device and bandwidth, however it cannot be delivered at a higher
	// value than is stored. This field may return `-1` if the frame rate of the input cannot be reliably
	// determined.
	MaxStoredFrameRate *string `json:"max_stored_frame_rate,omitempty"`
	// The aspect ratio of the asset.
	//
	// 	"width:height" -> "16:9"
	AspectRatio *string `json:"aspect_ratio,omitempty"`
	// An array of Playback ID objects. Use these to create HLS playback URLs.
	// See [play_your_videos] for more details.
	//
	// [play_your_videos]: https://docs.mux.com/guides/play-your-videos
	PlaybackIDs []MuxWebhookPlaybackID `json:"playback_ids"`
	// The individual media tracks that make up an asset.
	Tracks []MuxWebhookTrack `json:"tracks,omitempty"`
	// Object that describes any errors that happened when processing this asset.
	Errors *MuxWebhookError `json:"errors,omitempty"`
	// Unique identifier for the direct upload. This is an optional parameter added when the asset is created from a direct upload.
	UploadID *string `json:"upload_id,omitempty"`
	// This field can be set to anything. It will be included in the asset details
	// and related webhooks. If you're looking for more structured metadata, such as `title` or `external_id`, you can set
	// the `meta` object instead. Max 255 characters.
	Passthrough string `json:"passthrough"`
	// The type of ingest used to create the asset.
	//
	//	"on_demand_url", "on_demand_direct_upload", "on_demand_clip", "live_rtmp", "live_srt"
	IngestType *string `json:"ingest_type,omitempty"`
	// Customer provided metadata about this asset.
	//
	// Note: this metadata may be publicly available via the video player. Do not include PII or sensitive information.
	Meta *MuxWebhookMeta `json:"meta,omitempty"`
	// Detailed state information about the asset ingest process.
	Progress MuxWebhookProgress `json:"progress"`
}

// MuxWebhookMeta represents mux webhook meta object.
// Customer provided metadata about this asset.
//
// Note: this metadata may be publicly available via the video player. Do not include PII or sensitive information.
type MuxWebhookMeta struct {
	// The asset title. Max 512 code points.
	Title *string `json:"title,omitempty"`
	// This is an identifier you provide to keep track of the creator of the asset. Max 128 code points.
	CreatorID *string `json:"creator_id,omitempty"`
	// This is an identifier you provide to link the asset to your own data. Max 128 code points.
	ExternalID *string `json:"external_id"`
}

// MuxWebhookPlaybackID represents Mux webhook Playback ID object.
type MuxWebhookPlaybackID struct {
	// Unique identifier for the PlaybackID.
	ID string `json:"id"`
	// Possible values: "public", "signed", "drm"
	//
	//	- public IDs are accessable by constructing an HLS URL like https://stream.mux.com/${PLAYBACK_ID}
	//	- signed playback IDs should be used with tokens https://stream.mux.com/${PLAYBACK_ID}?token={TOKEN}
	//		see [secure video playback] for details about creating tokens.
	//	- drm playback IDs are protected with DRM technologies. See [DRM documentation] for more details.
	//
	// [secure video playback]: https://docs.mux.com/guides/secure-video-playback
	// [DRM documentation]: https://docs.mux.com/guides/protect-videos-with-drm
	Policy string `json:"policy"`
	// The DRM configuration used by this playback ID. Must only be set when policy is set to drm.
	DrmConfigurationID *string `json:"drm_configuration_id,omitempty"`
}

type MuxWebhookTrack struct {
	// Unique identifier for the Track.
	ID string `json:"id"`
	// The type of the track.
	//
	//	"video", "audio", "text"
	Type string `json:"type"`
	// The duration in seconds for the track media. This parameter
	// is not set for "text" type tracks. This field is optional and may not be set. The top level
	// `duration` field of an asset will always be set.
	Duration *float64 `json:"duration,omitempty"`
	// The maximum width in pixels available for the track. Only set for the "video" type tracks.
	MaxWidth *int64 `json:"max_width,omitempty"`
	// the maximum height in pixels available for the track. Only set for the "video" type tracks.
	MaxHeight *int64 `json:"max_height,omitempty"`
	// The maximum frame rate available for the track. Only set for the "video" type tracks. This field
	// may return -1 if the frame rate of the input cannot be reliably determined.
	MaxFrameRate *float64 `json:"max_frame_rate,omitempty"`
	// The maximum number of audio channels the track supports. Only set for the "audio" type track.
	MaxChannels *int64 `json:"max_channels,omitempty"`
	// This parameter is set for "text" type tracks.
	//
	//	"subtitles"
	TextType *string `json:"text_type,omitempty"`
	// The source of the text contained in a Track of type "text".
	//
	//	"uploaded", "embedded", "generated_live", "generated_live_final", "generated_vod"
	TextSource *string `json:"text_source,omitempty"`
	// The language code value represents BCP 47 specification compliant value. For examle, "en" for English or "en-US" for the
	// US version of English. This parameter is only set for the "text" and "audio" track types.
	LanguageCode *string `json:"language_code,omitempty"`
	// The name of the track containing a human-readable description. The HLS manifest will associate a subtitle "text"
	// or "audio" track with this value. For example, the value should be "English" for a subtitle text track for the `language_code`
	// value of "en-US". This parameter is only set for "text" and "audio" tracks.
	Name *string `json:"name,omitempty"`
	// Indicates the track provides Subtitles for the Deaf or Hard-Of-Hearing. This parameter is set tracks where
	// `type` is "text" and `text_type` is subtitles.
	ClosedCaptions *bool `json:"closed_captions,omitempty"`
	// Arbitrary user-supplied metadata set for the track either when creating the asset or track. This parameter
	// is only set for "text" type tracks. Max 255 characters.
	Passthrough *string `json:"passthrough,omitempty"`
	// The status of the track. This parameter os only set for "text" type tracks.
	//
	//	"preparing", "ready", "errored", "deleted"
	Status *string `json:"status,omitempty"`
	// For an audio track, indicates that this is the primary audio track, ingested from the main input
	// of this asset. The primary audio track cannot be deleted.
	Primary *bool `json:"primary,omitempty"`
	// Object that describes any errors that happened when processing this asset.
	Errors *MuxWebhookError `json:"errors,omitempty"`
}

// MuxWebhookError represents mux webhook errors object.
// Object that describes any errors that happened when processing this asset.
type MuxWebhookError struct {
	// The type of error that occurred for this asset.
	Type string `json:"type"`
	// Error messages with more details.
	Messages []string `json:"messages"`
}

// MuxWebhookProgress represents mux webhook progress object.
// Detailed state information about the asset ingest process.
type MuxWebhookProgress struct {
	// The detailed state of the asset ingest progress. This field
	// is useful for relaying more granular processing information to
	// end users when a non-standard input is encountered.
	//
	//	"ingesting", "transcoding", "completed", "live", "errored"
	State string `json:"state"`
	// Represents the estimated completion percentage. Returns 0-100 when in "ingesting", "transcoding",
	// or "completed" state, and -1 when in "live" or "errored" state.
	Progress *float64 `json:"progress,omitempty"`
}
