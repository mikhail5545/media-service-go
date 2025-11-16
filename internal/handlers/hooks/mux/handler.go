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

// Package mux provides handler functionality to handle mux webhooks.
// It acts as an adapter between business logic and HTTP transport layer.
package mux

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/labstack/echo/v4"
	assetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	muxservice "github.com/mikhail5545/media-service-go/internal/services/mux"
)

type Handler struct {
	service muxservice.Service
}

func New(svc muxservice.Service) *Handler {
	return &Handler{service: svc}
}

func (h *Handler) ServeError(c echo.Context, code int, message string) error {
	return c.JSON(code, map[string]string{"error": message})
}

func (h *Handler) HandleWebhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Can't parse request body payload")
	}

	var payload *assetmodel.MuxWebhook
	if err := json.Unmarshal(body, &payload); err != nil {
		return h.ServeError(c, http.StatusBadRequest, "Can't unmarshal request body payload")
	}

	var webhookErr error
	switch payload.Type {
	case "video.asset.created":
		webhookErr = h.service.HandleAssetCreatedWebhook(c.Request().Context(), payload)
	case "video.asset.ready":
		webhookErr = h.service.HandleAssetReadyWebhook(c.Request().Context(), payload)
	case "video.asset.errored":
	case "video.asset.updated":
	case "video.asset.deleted":
	}

	if webhookErr != nil {
		return h.ServeError(c, http.StatusInternalServerError, webhookErr.Error())
	}
	return c.NoContent(http.StatusOK)
}
