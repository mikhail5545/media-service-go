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

package errors

import (
	"net/http"

	"github.com/labstack/echo/v4"
	serviceerrors "github.com/mikhail5545/media-service-go/internal/errors"
	errutil "github.com/mikhail5545/media-service-go/internal/util/errors"
)

func HTTPErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}

	if he, ok := err.(*echo.HTTPError); ok {
		code := he.Code
		message := he.Message
		if msg, ok := message.(string); ok && msg != "" {
			message = http.StatusText(code)
		}

		internalCode := "INTERNAL_SERVER_ERROR"
		switch code {
		case http.StatusNotFound:
			internalCode = serviceerrors.ErrorAliases[serviceerrors.ErrNotFound]
		case http.StatusMethodNotAllowed:
			internalCode = serviceerrors.ErrorAliases[serviceerrors.ErrInvalidArgument]
		case http.StatusBadRequest:
			internalCode = serviceerrors.ErrorAliases[serviceerrors.ErrInvalidArgument]
		}

		resp := errutil.ErrorResponse{}
		resp.Error.Code = internalCode
		resp.Error.Message = message.(string)

		c.JSON(code, resp)
		return
	}

	statusCode, payload := errutil.MapServiceError(err)
	c.JSON(statusCode, payload)
}
