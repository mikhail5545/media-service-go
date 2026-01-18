package asset

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	cldassetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	validationutil "github.com/mikhail5545/media-service-go/internal/util/validation"
)

func (f Filter) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.IDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&f.CloudinaryAssetIDs, validation.Each(validation.Length(2, 255))),
		validation.Field(&f.CloudinaryPublicIDs, validation.Each(validation.Length(2, 255))),
		validation.Field(&f.ResourceTypes, validation.Each(validation.Length(2, 100))),
		validation.Field(&f.Formats, validation.Each(validation.Length(1, 50))),
		validation.Field(&f.OrderDir, validation.In(cldassetmodel.OrderAscending, cldassetmodel.OrderDescending)),
		validation.Field(&f.OrderField, validation.In(
			cldassetmodel.OrderCreatedAt,
			cldassetmodel.OrderUpdatedAt,
			cldassetmodel.OrderResourceType,
			cldassetmodel.OrderFormat,
		)),
		validation.Field(&f.PageSize, validation.Min(1), validation.Max(1000)),
		validation.Field(&f.PageToken, validation.Length(0, 2048)),
		validation.Field(&f.Fields, validation.Each(validation.By(validateField))),
	)
}

func validateField(field any) error {
	return validationutil.ValidateField(field, cldassetmodel.IsValidField)
}
