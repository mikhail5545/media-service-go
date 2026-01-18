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
	"net/http"

	"github.com/labstack/echo/v4"
	muxtypes "github.com/mikhail5545/media-service-go/internal/models/mux/types"
	muxservice "github.com/mikhail5545/media-service-go/internal/services/mux"
)

type WebhookHandler struct {
	service *muxservice.Service
}

func New(svc *muxservice.Service) *WebhookHandler {
	return &WebhookHandler{
		service: svc,
	}
}

func (h *WebhookHandler) Handle(c echo.Context) error {
	var payload *muxtypes.MuxWebhook
	if err := c.Bind(payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return h.service.HandleAssetWebhook(c.Request().Context(), payload)
}
