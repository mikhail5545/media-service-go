/*
 * Copyright (c) 2026. Mikhail Kulik.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PageTokenPayload struct {
	CursorValue any    `json:"v"`
	LastID      string `json:"id"`
}

// EncodePageToken encodes a page token with the given cursor value and last ID.
func EncodePageToken(val any, id uuid.UUID) string {
	// Ensure time.Time values are in UTC
	if t, ok := val.(time.Time); ok {
		val = t.UTC()
	}
	p := PageTokenPayload{
		CursorValue: val,
		LastID:      id.String(),
	}
	b, _ := json.Marshal(p)
	return base64.RawURLEncoding.EncodeToString(b)
}

// DecodePageToken decodes a page token into the given cursor value and last ID.
func DecodePageToken(token string) (any, uuid.UUID, error) {
	if token == "" {
		return time.Time{}, uuid.Nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	var p PageTokenPayload
	if err := json.Unmarshal(b, &p); err != nil {
		return time.Time{}, uuid.Nil, err
	}
	id, err := uuid.Parse(p.LastID)
	if err != nil {
		return time.Time{}, uuid.Nil, err
	}
	return p.CursorValue, id, nil
}

func normalizeOrderDirection(dir string) string {
	if dir == "ASC" || dir == "asc" {
		return "ASC"
	}
	return "DESC"
}

type ApplyCursorParams struct {
	PageSize   int
	PageToken  string
	OrderField string
	OrderDir   string
}

func ApplyCursor(db *gorm.DB, params ApplyCursorParams) (*gorm.DB, error) {
	if params.PageSize < 0 {
		return nil, errors.New("page_size must be non-negative")
	}
	params.OrderDir = normalizeOrderDirection(params.OrderDir)

	cursorVal, lastID, err := DecodePageToken(params.PageToken)
	if err != nil {
		return nil, fmt.Errorf("invalid page token: %w", err)
	}

	orderExpr := fmt.Sprintf("%s %s, id %s", params.OrderField, params.OrderDir, params.OrderDir)
	db = db.Order(orderExpr).Limit(params.PageSize + 1) // Fetch one extra to check for next page

	if cursorVal != nil && lastID != uuid.Nil {
		op := ">"
		if params.OrderDir == "DESC" {
			op = "<"
		}
		db = db.Where(fmt.Sprintf("(%s, id) %s (?, ?)", params.OrderField, op), cursorVal, lastID)
	}
	return db, nil
}
