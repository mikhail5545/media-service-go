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
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	"github.com/google/uuid"
	metamodel "github.com/mikhail5545/media-service-go/internal/models/mux/metadata"
)

// Validate validates fields of [asset.CreateUploadURLRequest].
// All request fields are required for this operation.
// Validation rules:
//
//   - OwnerID: required, valid UUID.
//   - CreatorID: required, valid UUID.
//   - OwnerType: required, min 3 characters, max 128 characters, one of: ["course_part"].
//   - Title: required, min 3 characters, max 512 characters.
func (req CreateUploadURLRequest) Validate() error {
	return validation.ValidateStruct(&req,
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
		validation.Field(
			&req.CreatorID,
			validation.Required,
			is.UUID,
		),
		validation.Field(
			&req.Title,
			validation.Required,
			validation.Length(3, 512),
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

// Validate validates fields of [asset.CreateUnownedUploadURLRequest].
// All request fields are required for this operation.
// Validation rules:
//
//   - Title: required, string, at least 3 characters, max 512 characters.
//   - CreatorID: required, valid UUID.
func (req CreateUnownedUploadURLRequest) Validate() error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.Title,
			validation.Required,
			validation.Length(3, 512),
		),
		validation.Field(
			&req.CreatorID,
			validation.Required,
			is.UUID,
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
						if owner, ok := value.(metamodel.Owner); ok {
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
