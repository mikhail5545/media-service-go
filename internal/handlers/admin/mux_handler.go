// github.com/mikhail5545/product-service-go
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

package admin

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mikhail5545/media-service-go/internal/models"
	"github.com/mikhail5545/media-service-go/internal/services"
)

type MUXHandler struct {
	muxService *services.MuxService
}

func NewMUXHandler(muxService *services.MuxService) *MUXHandler {
	return &MUXHandler{
		muxService: muxService,
	}
}

func (h *MUXHandler) GetCoursePartUploadURL(c echo.Context) error {
	var req models.GetCoursePartUploadURLRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request JSON payload."})
	}

	response, err := h.muxService.CreateCoursePartUploadURL(c.Request().Context(), req.PartID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, map[string]any{"upload_response": response})
}
