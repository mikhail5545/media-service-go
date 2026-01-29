/*
 * Copyright (c) 2026. Mikhail Kulik
 *
 * This program is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Affero General Public License as published
 *  by the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *  GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 *  along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package webhooks

import (
	"github.com/mikhail5545/media-service-go/internal/models/mux/types"
	muxwebhookpbv1 "github.com/mikhail5545/media-service-go/pb/media_service/mux/webhook/v1"
)

func MuxWebhookErrorToProto(err *types.MuxWebhookError) *muxwebhookpbv1.MuxWebhookError {
	return &muxwebhookpbv1.MuxWebhookError{
		Type:     err.Type,
		Messages: err.Messages,
	}
}

func MuxWebhookPlaybackIDToProto(playbackID *types.MuxWebhookPlaybackID) (*muxwebhookpbv1.MuxWebhookPlaybackID, error) {
	return &muxwebhookpbv1.MuxWebhookPlaybackID{
		Id:                 playbackID.ID,
		Policy:             playbackID.Policy,
		DrmConfigurationId: playbackID.DrmConfigurationID,
	}, nil
}

func MuxWebhookTrackToPorto(track *types.MuxWebhookTrack) (*muxwebhookpbv1.MuxWebhookTrack, error) {
	var pbError *muxwebhookpbv1.MuxWebhookError = nil
	if track.Errors != nil {
		pbError = MuxWebhookErrorToProto(track.Errors)
	}
	return &muxwebhookpbv1.MuxWebhookTrack{
		Id:             track.ID,
		Type:           track.Type,
		Duration:       track.Duration,
		MaxWidth:       track.MaxWidth,
		MaxHeight:      track.MaxHeight,
		MaxFrameRate:   track.MaxFrameRate,
		MaxChannels:    track.MaxChannels,
		TextType:       track.TextType,
		TextSource:     track.TextSource,
		LanguageCode:   track.LanguageCode,
		Name:           track.Name,
		ClosedCaptions: track.ClosedCaptions,
		Status:         track.Status,
		Primary:        track.Primary,
		Errors:         pbError,
	}, nil
}
