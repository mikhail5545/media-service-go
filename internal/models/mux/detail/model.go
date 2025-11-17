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
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
)

// AssetDetail holds bulky, less-frequently queried data related to an Asset.
type AssetDetail struct {
	AssetID string                       `gorm:"primaryKey;size:36"`
	Tracks  []assetmodel.MuxWebhookTrack `gorm:"type:jsonb"`
}
