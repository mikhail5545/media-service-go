package validation

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/google/uuid"
	"github.com/gosimple/slug"
)

func composeRules(required bool, additionalRules ...validation.Rule) []validation.Rule {
	rules := additionalRules
	if required {
		rules = append([]validation.Rule{validation.Required}, rules...)
	}
	return rules
}

func extractValue[T any](dest *T, src any) error {
	if src == nil {
		return nil
	}
	switch v := src.(type) {
	case T:
		*dest = v
	case *T:
		if v == nil {
			return nil
		}
		*dest = *v
	default:
		return fmt.Errorf("must be a %T or %T", *dest, dest)
	}
	return nil
}

func IsValidSlug(value any) error {
	var strSlug string
	if err := extractValue(&strSlug, value); err != nil {
		return err
	}
	if strSlug == "" {
		return nil
	}
	if !slug.IsSlug(strSlug) {
		return fmt.Errorf("must be a valid slug")
	}
	return nil
}

func isValidStringUUIDv7(strID string) error {
	if strID == "" {
		return nil
	}
	uid, err := uuid.Parse(strID)
	if err != nil {
		return fmt.Errorf("must be a valid UUIDv7")
	}
	if uid.Version() != uuid.Version(7) {
		return fmt.Errorf("must be a valid UUIDv7")
	}
	return nil
}

func IsValidUUIDv7(value any) error {
	if value == nil {
		return nil
	}
	var strID string
	var uid uuid.UUID
	strErr := extractValue(&strID, value)
	uidErr := extractValue(&uid, value)
	if strErr != nil && uidErr != nil {
		return fmt.Errorf("must be a valid UUIDv7 string or uuid.UUID")
	}
	if strErr == nil {
		return isValidStringUUIDv7(strID)
	}
	if uid == uuid.Nil {
		return nil // empty value is considered valid
	}
	if uid.Version() != uuid.Version(7) {
		return fmt.Errorf("must be a valid UUIDv7")
	}
	return nil
}

type fieldValidator func(field string) bool

func ValidateField(value any, validator fieldValidator) error {
	if value == nil {
		return fmt.Errorf("must be a valid field name")
	}
	field, ok := value.(string)
	if !ok {
		return fmt.Errorf("must be a valid field name")
	}
	if !validator(field) {
		return fmt.Errorf("must be a valid field name")
	}
	return nil
}

// SlugRule returns the ozzo-validation rules for a slug.
func SlugRule(required bool) []validation.Rule {
	return composeRules(required, validation.Length(2, 255), validation.By(IsValidSlug))
}

// UUIDRule returns the ozzo-validation rules for a UUID.
func UUIDRule(required bool) []validation.Rule {
	return composeRules(required, validation.By(IsValidUUIDv7))
}
