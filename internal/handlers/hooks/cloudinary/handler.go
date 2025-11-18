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

// Package mux provides handler functionality to handle cloudinary webhooks.
// It acts as an adapter between business logic and HTTP transport layer.
package cloudinary

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
	cldservice "github.com/mikhail5545/media-service-go/internal/services/cloudinary"
)

// Handler provides HTTP handlers. It holds [cldservice.Service] to perform service-layer logic.
type Handler struct {
	service cldservice.Service
}

// New creates a new instance of Handler.
func New(svc cldservice.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) ServeError(c echo.Context, code int, message string) error {
	return c.JSON(code, map[string]string{"error": message})
}

// HandleUploadWebhook processes an incoming Cloudinary upload webhook, finds the corresponding asset, and updates it in a patch-like manner.
//
// Method: POST
// Path: /webhooks/cloudinary/upload
func (h *Handler) UploadWebhook(c echo.Context) error {
	var body []byte
	n, err := c.Request().Body.Read(body)
	if n == 0 || err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Unable to read request body")
	}

	timestamp := c.Request().Header.Get("X-Cld-Timestamp")
	if timestamp == "" {
		return h.ServeError(c, http.StatusForbidden, "Missing X-Cld-Timestamp header")
	}
	signature := c.Request().Header.Get("X-Cld-Signature")
	if signature == "" {
		return h.ServeError(c, http.StatusForbidden, "Missing X-Cld-Signature header")
	}

	if err := h.service.HandleUploadWebhook(c.Request().Context(), body, timestamp, signature); err != nil {
		log.Printf("Failed to process cloudinary webhook: %s", err.Error())
	}
	return c.NoContent(http.StatusOK)
}
