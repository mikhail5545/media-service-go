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

// Package asset provides models, DTO models for [cloudinary.Service] requests and validation tools.
package asset

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
	"github.com/mikhail5545/media-service-go/internal/models/cloudinary/metadata"
)

// Validate validates fields of [asset.CreateSignedUploadURLRequest].
// All request fields except eager are required for this operation.
// Validation rules:
//
//   - Eager: optional.
//   - File: required, at least 3 characters.
//   - PublicID: required, at least 3 characters.
func (req CreateSignedUploadURLRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.File,
			validation.Required,
			validation.Length(3, 0),
		),
		validation.Field(
			&req.PublicID,
			validation.Required,
			validation.Length(3, 0),
		),
	)
}

// Validate validates fields of [asset.AssociateRequest].
// All request fields are required for this operation.
// Validation rules:
//
//   - ID: required, valid UUID.
//   - OwnerID: required, valid UUID.
//   - OwnerType: required, min 3 characters, max 128 characters, one of: ["course_part"].
func (req AssociateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.ID,
			validation.Required,
			is.UUID,
		),
		validation.Field(
			&req.OwnerID,
			validation.Required,
			is.UUID,
		),
		validation.Field(
			&req.OwnerType,
			validation.Required,
			validation.Length(1, 128),
			validation.In("course_part"),
		),
	)
}

// Validate validates fields of [asset.DeassociateRequest].
// All request fields are required for this operation.
// Validation rules:
//
//   - ID: required, valid UUID.
//   - OwnerID: required, valid UUID.
//   - OwnerType: required, min 3 characters, max 128 characters, one of: ["course_part"].
func (req DeassociateRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.ID,
			validation.Required,
			is.UUID,
		),
		validation.Field(
			&req.OwnerID,
			validation.Required,
			is.UUID,
		),
		validation.Field(
			&req.OwnerType,
			validation.Required,
			validation.Length(1, 128),
			validation.In("course_part"),
		),
	)
}

// Validate validates fields of [asset.UpdateOwnersRequest].
// All request fields are required for this operation.
// Validation rules:
//
//   - ID: required, valid UUID.
//   - Owners: required, slice of [metamodel.Owner], each must have a valid UUID and valid OwnerType.
func (req UpdateOwnersRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.ID,
			validation.Required,
			is.UUID,
		),
		validation.Field(
			&req.Owners,
			validation.Required,
			validation.Length(1, 0),
			validation.Each(
				validation.By(
					func(value interface{}) error {
						if owner, ok := value.(metadata.Owner); ok {
							if _, err := uuid.Parse(owner.OwnerID); err != nil {
								return errors.New("must be a valid uuid")
							}
							if len(owner.OwnerType) <= 3 {
								return errors.New("must be at least 4 characters long")
							}
						}
						return nil
					},
				),
			),
		),
	)
}

func (req DestroyAssetRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.ID,
			validation.Required,
			is.UUID,
		),
		validation.Field(
			&req.ResourceType,
			validation.Required,
			validation.Length(3, 0),
		),
	)
}

func (req CleanupOrphanAssetsRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.Folder,
			validation.Required,
			validation.Length(1, 255),
		),
		validation.Field(
			&req.AssetType,
			validation.Required,
			validation.Length(3, 0),
		),
	)
}

// Validate validates fields of [asset.SuccessfulUploadRequest].
// All request fields except Owners are required for this operation.
// Validation rules:
//
//   - CloudinaryAssetID: required.
//   - CloudinaryPublicID: required, at least 3 characters, max 255 characters.
//   - SecureURL: required, valid URL.
//   - AssetFolder: required, at least 3 characters, max 255 characters.
//   - DisplayName: required, at least 3 characters, max 255 characters.
//   - Owners: optional, slice of [metamodel.Owner], if populated, each must have a valid UUID and valid OwnerType.
func (req SuccessfulUploadRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.CloudinaryAssetID, validation.Required),
		validation.Field(&req.CloudinaryPublicID, validation.Required, validation.Length(3, 255)),
		validation.Field(&req.SecureURL, validation.Required, is.URL),
		validation.Field(&req.AssetFolder, validation.Required, validation.Length(3, 255)),
		validation.Field(&req.DisplayName, validation.Required, validation.Length(3, 255)),
		validation.Field(
			&req.Owners,
			validation.When(len(req.Owners) > 0,
				validation.Each(
					validation.By(
						func(value any) error {
							if owner, ok := value.(metadata.Owner); ok {
								if _, err := uuid.Parse(owner.OwnerID); err != nil {
									return errors.New("must be a valid uuid")
								}
								if len(owner.OwnerType) <= 3 {
									return errors.New("must be at least 4 characters long")
								}
							}
							return nil
						},
					),
				),
			),
		),
	)
}
