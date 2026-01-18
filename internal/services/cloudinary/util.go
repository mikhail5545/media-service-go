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

package cloudinary

import (
	"reflect"

	assetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	"github.com/mikhail5545/media-service-go/internal/util/patch"
)

func buildUpdatesFromWebhook(existing *assetmodel.Asset, webhook *assetmodel.CloudinaryUploadWebhook) map[string]any {
	updates := make(map[string]any)

	patch.UpdateIfChanged(updates, "display_name", &webhook.DisplayName, &existing.DisplayName)
	patch.UpdateIfChanged(updates, "asset_folder", &webhook.AssetFolder, &existing.AssetFolder)
	patch.UpdateIfChanged(updates, "url", &webhook.Url, &existing.URL)
	patch.UpdateIfChanged(updates, "secure_url", &webhook.SecureUrl, &existing.SecureURL)
	patch.UpdateIfChanged(updates, "format", &webhook.Format, &existing.Format)
	patch.UpdateIfChanged(updates, "width", &webhook.Width, existing.Width)
	patch.UpdateIfChanged(updates, "height", &webhook.Height, existing.Height)
	patch.UpdateIfChanged(updates, "public_id", &webhook.PublicID, &existing.CloudinaryPublicID)
	patch.UpdateIfChanged(updates, "resource_type", &webhook.ResourceType, &existing.ResourceType)

	if len(webhook.Tags) > 0 && !reflect.DeepEqual(webhook.Tags, existing.Tags) {
		updates["tags"] = webhook.Tags
	}
	return updates
}
