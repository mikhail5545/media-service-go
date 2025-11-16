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

package assetowner

// AssetOwner represents the join table for the many-to-many relationship
// between assets and their owners (e.g., products, articles).
type AssetOwner struct {
	AssetID   string `gorm:"primaryKey;size:36" json:"asset_id"`
	OwnerID   string `gorm:"primaryKey;size:36" json:"owner_id"`
	OwnerType string `gorm:"primaryKey;varchar(128)" json:"owner_type"`
}
