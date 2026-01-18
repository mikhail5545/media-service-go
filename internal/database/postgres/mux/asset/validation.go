package asset

import (
	"github.com/go-ozzo/ozzo-validation/v4"
	muxassetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	validationutil "github.com/mikhail5545/media-service-go/internal/util/validation"
)

func (f Filter) Validate() error {
	return validation.ValidateStruct(&f,
		validation.Field(&f.IDs, validation.Each(validationutil.UUIDRule(false)...)),
		validation.Field(&f.MuxUploadIDs, validation.Each(validation.Length(1, 255))),
		validation.Field(&f.MuxAssetIDs, validation.Each(validation.Length(1, 255))),
		validation.Field(&f.AspectRatios, validation.Each(validation.Length(1, 50))),
		validation.Field(&f.ResolutionTiers, validation.Each(validation.Length(1, 50))),
		validation.Field(&f.UploadStatuses, validation.Each(validation.In(
			muxassetmodel.UploadStatusPreparing,
			muxassetmodel.UploadStatusReady,
			muxassetmodel.UploadStatusErrored,
			muxassetmodel.UploadStatusDeleted,
		))),
		validation.Field(&f.IngestTypes, validation.Each(validation.In(
			muxassetmodel.IngestTypeLiveSRT,
			muxassetmodel.IngestTypeOnDemandURL,
			muxassetmodel.IngestTypeOnDemandClip,
			muxassetmodel.IngestTypeLiveRTMP,
			muxassetmodel.IngestTypeOnDemandDirectUpload,
		))),
		validation.Field(&f.OrderDir, validation.In(muxassetmodel.OrderAscending, muxassetmodel.OrderDescending)),
		validation.Field(&f.OrderBy, validation.In(muxassetmodel.OrderUpdatedAt, muxassetmodel.OrderCreatedAt, muxassetmodel.OrderIngestType)),
		validation.Field(&f.Fields, validation.Each(validation.By(validateField))),
		validation.Field(&f.PageSize, validation.Min(1), validation.Max(1000)),
		validation.Field(&f.PageToken, validation.Length(1, 2048)),
	)
}

func validateField(value any) error {
	return validationutil.ValidateField(value, muxassetmodel.IsValidField)
}
