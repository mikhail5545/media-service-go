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

package types

import "time"

type Resource struct {
	ResourceType string `json:"resource_type"`
	Type         string `json:"type"`
	AssetID      string `json:"asset_id"`
	PublicID     string `json:"public_id"`
	Version      int64  `json:"version"`
	AssetFolder  string `json:"asset_folder"`
	DisplayName  string `json:"display_name"`
}

type NotificationContext struct {
	TriggeredAt time.Time   `json:"triggered_at"`
	TriggeredBy TriggeredBy `json:"triggered_by"`
}

type TriggeredBy struct {
	Source string `json:"source"`
	ID     string `json:"id"`
}

type CloudinaryDeleteWebhook struct {
	NotificationType    string              `json:"notification_type"`
	Resources           []Resource          `json:"resources"`
	NotificationContext NotificationContext `json:"notification_context"`
	SignatureKey        string              `json:"signature_key"`
}

// CloudinaryRenameWebhook represents Cloudinary API webhook triggered by an asset rename (public ID change).
type CloudinaryRenameWebhook struct {
	NotificationType    string              `json:"notification_type"`
	Timestamp           time.Time           `json:"timestamp"`
	RequestID           string              `json:"request_id"`
	ResourceType        string              `json:"resource_type"`
	Type                string              `json:"type"`
	AssetID             string              `json:"asset_id"`
	AssetFolder         string              `json:"asset_folder"`
	DisplayName         string              `json:"display_name"`
	FromPublicID        string              `json:"from_public_id"`
	ToPublicID          string              `json:"to_public_id"`
	NotificationContext NotificationContext `json:"notification_context"`
	SignatureKey        string              `json:"signature_key"`
}
