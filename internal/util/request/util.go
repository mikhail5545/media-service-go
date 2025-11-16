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

// Package request provides shared utility functions for HTTP handlers.
//
// It works with requests, parsing and validating various data from echo.Context:
//
//   - Prams
//
//   - Query Search Params
//
//   - Request JSON payload
package request

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// BindAndValidateJSON binds the request body to the given struct and handles validation errors.
func BindAndValidateJSON(c echo.Context, req any) error {
	if err := c.Bind(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request JSON payload."})
	}
	return nil
}

// GetIDParam extracts a required ID from the path parameters.
func GetIDParam(c echo.Context, paramName, errorMsg string) (string, error) {
	id := c.Param(paramName)
	if _, err := uuid.Parse(id); err != nil {
		return "", echo.NewHTTPError(http.StatusBadRequest, map[string]string{"error": errorMsg})
	}
	return id, nil
}

// GetPaginationParams extracts 'limit' and 'offset' from query parameters with default values.
func GetPaginationParams(c echo.Context, defaultLimit, defaultOffset int) (int, int, error) {
	limitStr := c.QueryParam("limit")
	limit := defaultLimit
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < -1 {
			return 0, 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid pagination parameters.")
		}
	}

	offsetStr := c.QueryParam("offset")
	offset := defaultOffset
	if offsetStr != "" {
		var err error
		offset, err = strconv.Atoi(offsetStr)
		if err != nil || offset < 0 {
			return 0, 0, echo.NewHTTPError(http.StatusBadRequest, "Invalid pagination parameters.")
		}
	}

	return limit, offset, nil
}
