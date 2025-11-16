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

package metadata

// AssetMetadata represents the metadata for a MUX asset stored in ArangoDB.
type AssetMetadata struct {
	// The _key field will be internal asset ID from PostgreSQL database.
	Key       string  `json:"_key,omitempty"`
	Title     string  `json:"title"`
	CreatorID string  `json:"creator_id"`
	Owners    []Owner `json:"owners"`
}

// Owner represents an entity that is associated with an asset.
type Owner struct {
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`
}
