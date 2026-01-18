package parsing

import (
	"strings"

	"github.com/google/uuid"
	serviceerrors "github.com/mikhail5545/media-service-go/internal/errors"
)

// ParseColumnTag extracts the column name from a GORM tag string.
// Example: "column:product_id;type:uuid" -> "product_id"
func ParseColumnTag(tag string) string {
	for part := range strings.SplitSeq(tag, ";") {
		part = strings.TrimSpace(part)
		// Check if this part starts with "column:"
		if strings.HasPrefix(part, "column:") {
			// Extract everything after "column:"
			return strings.TrimPrefix(part, "column:")
		}
	}
	return ""
}

func StrToUUIDs(strIDs []string) uuid.UUIDs {
	if len(strIDs) == 0 {
		return nil
	}
	dest := make(uuid.UUIDs, 0, len(strIDs))
	for i := range strIDs {
		if uid, err := uuid.Parse(strIDs[i]); err == nil {
			dest = append(dest, uid)
		}
	}
	return dest
}

func StrToUUID(strID string) (uuid.UUID, error) {
	if strID == "" {
		return uuid.Nil, nil
	}
	uid, err := uuid.Parse(strID)
	if err != nil {
		return uuid.Nil, serviceerrors.NewInvalidArgumentError(err)
	}
	return uid, nil
}

func CleanIDs(in uuid.UUIDs) uuid.UUIDs {
	cleaned := make(uuid.UUIDs, 0, len(in))
	for i := range in {
		if in[i] != uuid.Nil {
			cleaned = append(cleaned, in[i])
		}
	}
	return cleaned
}

func CleanStrings(in []string) []string {
	cleaned := make([]string, 0, len(in))
	for i := range in {
		if in[i] != "" {
			cleaned = append(cleaned, in[i])
		}
	}
	return cleaned
}
