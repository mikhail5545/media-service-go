package types

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	validationutil "github.com/mikhail5545/media-service-go/internal/util/validation"
)

type AuditTrailOptions struct {
	AdminID   uuid.UUID
	AdminName string
	Note      string
	EventID   string
}

func (a AuditTrailOptions) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.AdminID, validationutil.UUIDRule(false)...),
		validation.Field(&a.AdminName, validation.Required, validation.Length(2, 128)),
		validation.Field(&a.Note, validation.Required, validation.Length(10, 512)),
		validation.Field(&a.EventID, validation.Length(0, 256)),
	)
}
