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
	"net/http"

	"github.com/labstack/echo/v4"
	cldservice "github.com/mikhail5545/media-service-go/internal/services/cloudinary"
)

type WebhookHandler struct {
	service *cldservice.Service
}

func New(svc *cldservice.Service) *WebhookHandler {
	return &WebhookHandler{
		service: svc,
	}
}

func (h *WebhookHandler) Handle(c echo.Context) error {
	var body []byte
	n, err := c.Request().Body.Read(body)
	if n == 0 || err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}

	timestamp := c.Request().Header.Get("X-Cld-Timestamp")
	if timestamp == "" {
		return echo.NewHTTPError(http.StatusForbidden, "missing X-Cld-Timestamp header")
	}
	signature := c.Request().Header.Get("X-Cld-Signature")
	if signature == "" {
		return echo.NewHTTPError(http.StatusForbidden, "missing X-Cld-Signature header")
	}

	return h.service.HandleUploadWebhook(c.Request().Context(), body, timestamp, signature)
}
