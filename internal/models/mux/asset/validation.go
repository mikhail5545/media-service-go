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
	"reflect"
	"sync"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/mikhail5545/media-service-go/internal/util/formatting"
	"github.com/mikhail5545/media-service-go/internal/util/parsing"
	validationutil "github.com/mikhail5545/media-service-go/internal/util/validation"
)

func (req GetFilter) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ID, validationutil.UUIDRule(true)...),
	)
}

func (req ListRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.IDs, validationutil.UUIDRule(false)...),
		validation.Field(&req.MuxUploadIDs, validation.Length(1, 255)),
		validation.Field(&req.MuxAssetIDs, validation.Length(1, 255)),
		validation.Field(&req.AspectRatios, validation.Each(validation.Length(1, 64), validation.In("16:9", "4:3", "1:1", "21:9", "3:2"))),
		validation.Field(&req.ResolutionTiers, validation.Each(validation.Length(1, 64), validation.In("audio-only", "720p", "1080p", "1440p", "2160p"))),
		validation.Field(&req.IngestTypes, validation.Each(validation.Length(1, 64), validation.In(
			IngestTypeOnDemandDirectUpload,
			IngestTypeLiveRTMP,
			IngestTypeOnDemandClip,
			IngestTypeLiveSRT,
			IngestTypeOnDemandURL,
		))),
		validation.Field(&req.OrderBy, validation.In(OrderCreatedAt, OrderUpdatedAt, OrderIngestType)),
		validation.Field(&req.OrderDir, validation.In(OrderAscending, OrderDescending)),
		validation.Field(&req.PageSize, validation.Min(1), validation.Max(1000)),
		validation.Field(&req.PageToken, validation.Length(1, 2048)),
	)
}

func (req ChangeStateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ID, validationutil.UUIDRule(true)...),
		validation.Field(&req.AdminID, validationutil.UUIDRule(true)...),
		validation.Field(&req.AdminName, validation.Required, validation.Length(1, 128)),
		validation.Field(&req.Note, validation.Required, validation.Length(10, 512)),
	)
}

func (req CreateUploadURLRequest) Validate() error {
	return validation.ValidateStruct(
		validation.Field(&req.AdminID, validationutil.UUIDRule(true)...),
		validation.Field(&req.AdminName, validation.Required, validation.Length(1, 128)),
		validation.Field(&req.Title, validation.Length(1, 256)),
	)
}

func (req ManageOwnerRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ID, validationutil.UUIDRule(true)...),
		validation.Field(&req.OwnerID, validationutil.UUIDRule(true)...),
		validation.Field(&req.OwnerType, validation.In("lesson")),
	)
}

func (req GeneratePlaybackTokenRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.AssetID, validationutil.UUIDRule(true)...),
		validation.Field(&req.UserID, validationutil.UUIDRule(true)...),
		validation.Field(&req.Expiration, validation.Required, validation.Min(int64(15*60))),
		validation.Field(&req.UserAgent, validation.Length(1, 256)),
		validation.Field(&req.SessionID, validationutil.UUIDRule(false)...),
	)
}

var (
	validFields     map[string]bool
	validFieldsOnce sync.Once
)

// ValidFields returns all valid field names for the Asset struct.
func ValidFields() map[string]bool {
	validFieldsOnce.Do(func() {
		validFields := make(map[string]bool)
		t := reflect.TypeOf(Asset{})

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if tag := field.Tag.Get("gorm"); tag != "" {
				if col := parsing.ParseColumnTag(tag); col != "" {
					validFields[col] = true
					continue
				}
			}
			validFields[formatting.ToSnakeCase(field.Name)] = true
		}
	})
	return validFields
}

// IsValidField checks if a field name is valid for selection.
func IsValidField(field string) bool {
	return ValidFields()[field]
}
