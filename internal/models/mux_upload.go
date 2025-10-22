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

package models

import (
	"time"
)

type MUXUpload struct {
	ID                    string     `gorm:"primaryKey;size:36" json:"id"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
	MUXUploadID           *string    `gorm:"null" json:"mux_upload_id,omitempty"`
	MUXAssetID            *string    `gorm:"null" json:"mux_asset_id,omitempty"`
	MUXPlaybackID         *string    `gorm:"null" json:"mux_playback_id,omitempty"`
	VideoProcessingStatus string     `gorm:"null" json:"video_processing_status,omitempty"`
	Duration              *float64   `gorm:"null" json:"duration,omitempty"`
	AspectRatio           *string    `gorm:"null" json:"aspect_ratio,omitempty"`
	MaxHeight             *int       `gorm:"null" json:"max_height,omitempty"`
	MaxWidth              *int       `gorm:"null" json:"max_width,omitempty"`
	AssetCreatedAt        *time.Time `gorm:"null" json:"asset_created_at,omitempty"`
}

// DTO models
type UpdateMUXUploadRequest struct {
	VideoProcessingStatus string    `json:"video_processing_status"`
	MUXUploadID           string    `json:"mux_upload_id"`
	MUXAssetID            string    `json:"mux_asset_id"`
	MUXPlaybackID         string    `json:"mux_playback_id"`
	Duration              float64   `json:"duration"`
	AspectRatio           string    `json:"aspect_ratio"`
	MaxHeight             int       `json:"max_height"`
	MaxWidth              int       `json:"max_width"`
	AssetCreatedAt        time.Time `json:"asset_created_at"`
}

type MUXVideo struct {
	VideoProcessingStatus string  `json:"video_processing_status"`
	MUXPlaybackID         string  `json:"mux_playback_id"`
	Duration              float64 `json:"duration"`
	AspectRatio           string  `json:"aspect_ratio"`
	MaxHeight             int     `json:"max_height"`
	MaxWidth              int     `json:"max_width"`
}
