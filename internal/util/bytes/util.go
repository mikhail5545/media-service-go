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

package bytes

import (
	"encoding/hex"
	"fmt"

	"github.com/google/uuid"
)

func UUIDToBytes(id *uuid.UUID) ([]byte, error) {
	if id == nil {
		return nil, nil
	}
	return id.MarshalBinary()
}

func StrUUIDToBytes(id string) ([]byte, error) {
	if id == "" {
		return nil, nil
	}
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	return uid.MarshalBinary()
}

func ToUUID(id []byte) (uuid.UUID, error) {
	if len(id) == 0 {
		return uuid.Nil, nil
	}
	uid, err := uuid.FromBytes(id)
	if err != nil {
		return uuid.Nil, err
	}
	return uid, nil
}

// SliceToUUIDStrings converts a slice of byte slices (gRPC) to a slice of UUID strings (Internal).
// It uses the standard library validation to ensure data integrity.
func SliceToUUIDStrings(data [][]byte) ([]string, error) {
	if len(data) == 0 {
		return nil, nil // Return nil for empty input to avoid allocating empty slice
	}

	out := make([]string, 0, len(data))
	for i, b := range data {
		if len(b) != 16 {
			return nil, fmt.Errorf("invalid uuid length at index %d: %d", i, len(b))
		}
		uid, err := uuid.FromBytes(b)
		if err != nil {
			return nil, fmt.Errorf("invalid uuid at index %d: %w", i, err)
		}
		out = append(out, uid.String())
	}
	return out, nil
}

func uuidBytesToString(data []byte) (string, error) {
	if len(data) != 16 {
		return "", fmt.Errorf("invalid uuid length: %d", len(data))
	}
	var dst [36]byte
	hex.Encode(dst[0:8], data[0:4])
	dst[8] = '-'
	hex.Encode(dst[9:13], data[4:6])
	dst[13] = '-'
	hex.Encode(dst[14:18], data[6:8])
	dst[18] = '-'
	hex.Encode(dst[19:23], data[8:10])
	dst[23] = '-'
	hex.Encode(dst[24:36], data[10:16])
	return string(dst[:]), nil
}

// SliceToUUIDStringsFast converts a slice of byte slices (gRPC) to a slice of UUID strings (Internal).
// It uses a faster manual conversion method to improve performance. Implemented for future use cases where performance is critical, or
// SliceToUUIDStrings will cause bottlenecks.
func SliceToUUIDStringsFast(bs [][]byte) ([]string, error) {
	if len(bs) == 0 {
		return nil, nil // Return nil for empty input to avoid allocating empty slice
	}

	out := make([]string, 0, len(bs))
	for i, b := range bs {
		s, err := uuidBytesToString(b)
		if err != nil {
			return nil, fmt.Errorf("index %d: %w", i, err)
		}
		out = append(out, s)
	}
	return out, nil
}
