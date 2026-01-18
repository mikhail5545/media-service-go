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

package mux

import (
	"fmt"
	"reflect"

	"github.com/google/uuid"
	assetrepo "github.com/mikhail5545/media-service-go/internal/database/postgres/mux/asset"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	muxtypes "github.com/mikhail5545/media-service-go/internal/models/mux/types"
	"github.com/mikhail5545/media-service-go/internal/util/memory"
	"github.com/mikhail5545/media-service-go/internal/util/parsing"
	"github.com/mikhail5545/media-service-go/internal/util/patch"
)

func retrieveAssetID(opt assetSearchOptions) (*assetrepo.GetOptions, error) {
	var assetID uuid.UUID
	var err error

	if opt.GetOptions != nil {
		return opt.GetOptions, nil
	}

	if opt.AssetUUID != nil {
		assetID = *opt.AssetUUID
	} else if opt.AssetID != "" {
		assetID, err = parsing.StrToUUID(opt.AssetID)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("either AssetID or AssetUUID must be provided")
	}
	return &assetrepo.GetOptions{
		ID: assetID,
	}, nil
}

func buildAssetUpdatesFromWebhook(existing *assetmodel.Asset, data *muxtypes.MuxWebhookData) map[string]any {
	updates := make(map[string]any)

	patch.UpdateIfChanged(updates, "asset_created_at", &data.CreatedAt, existing.AssetCreatedAt)
	patch.UpdateIfChanged(updates, "state", &data.Progress.State, memory.MakePtr(string(existing.State)))
	patch.UpdateIfChanged(updates, "upload_status", data.Status, memory.MakePtr(string(existing.UploadStatus)))
	patch.UpdateIfChanged(updates, "duration", data.Duration, existing.Duration)
	patch.UpdateIfChanged(updates, "resolution_tier", data.ResolutionTier, existing.ResolutionTier)
	patch.UpdateIfChanged(updates, "aspect_ratio", data.AspectRatio, existing.AspectRatio)
	patch.UpdateIfChanged(updates, "ingest_type", data.IngestType, memory.MakePtr(string(existing.IngestType)))

	if len(data.PlaybackIDs) > 0 && !reflect.DeepEqual(data.PlaybackIDs, existing.MuxPlaybackIDs) {
		updates["mux_playback_ids"] = data.PlaybackIDs
	}

	return updates
}
