// Package asset provides models, DTO models for [cloudinary.Service] requests and validation tools.
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
		validation.Field(&req.IDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&req.CloudinaryAssetIDs, validation.Each(validation.Length(1, 255))),
		validation.Field(&req.CloudinaryPublicIDs, validation.Each(validation.Length(1, 255))),
		validation.Field(&req.ResourceTypes, validation.Each(validation.Length(3, 50),
			validation.In("image", "video", "raw", "folder")),
		),
		validation.Field(&req.Formats, validation.Each(validation.Length(2, 20),
			validation.In("img", "jpg", "png", "mp4", "mov", "pdf", "docx", "zip")),
		),
		validation.Field(&req.OrderField, validation.In(OrderCreatedAt, OrderUpdatedAt, OrderFormat, OrderResourceType)),
		validation.Field(&req.OrderDir, validation.In(OrderAscending, OrderDescending)),
		validation.Field(&req.PageSize, validation.Min(1), validation.Max(1000)),
		validation.Field(&req.PageToken, validation.Length(1, 2048)),
	)
}

func (req ChangeStateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ID, validationutil.UUIDRule(true)...),
		validation.Field(&req.AdminID, validationutil.UUIDRule(true)...),
		validation.Field(&req.AdminName, validation.Length(1, 128)),
		validation.Field(&req.Note, validation.Length(10, 512)),
	)
}

func (req ManageOwnerRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ID, validationutil.UUIDRule(true)...),
		validation.Field(&req.OwnerID, validationutil.UUIDRule(true)...),
		validation.Field(&req.OwnerType, validation.Required, validation.Length(1, 50), validation.In("product")),
	)
}

func (req CreateSignedUploadURLRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.File, validation.Required, validation.Length(3, 0)),
		validation.Field(&req.PublicID, validation.Required, validation.Length(3, 1024)),
		validation.Field(&req.Eager, validation.Length(0, 255)),
		validation.Field(&req.AdminID, validationutil.UUIDRule(true)...),
		validation.Field(&req.AdminName, validation.Required, validation.Length(1, 128)),
		validation.Field(&req.Note, validation.Length(0, 512)),
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
