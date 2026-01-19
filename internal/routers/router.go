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

package routers

import (
	"strings"

	"github.com/labstack/echo/v4"
)

type Router interface {
	Setup(group *echo.Group)
}

type Config struct {
	Api              string                // API prefix, e.g., /api
	Ver              string                // API version, e.g., /v1
	Use              []echo.MiddlewareFunc // Middlewares to use
	HTTPErrorHandler echo.HTTPErrorHandler
}

// Init initializes the Echo router with the given configuration and returns the versioned group.
func Init(e *echo.Echo, config Config) *echo.Group {
	if config.HTTPErrorHandler != nil {
		e.HTTPErrorHandler = config.HTTPErrorHandler
	}

	if len(config.Use) > 0 {
		e.Use(config.Use...)
	}

	apiPath := config.Api
	if !strings.HasPrefix(apiPath, "/") {
		apiPath = "/" + apiPath
	}

	verPath := config.Ver
	if !strings.HasPrefix(verPath, "/") {
		verPath = "/" + verPath
	}

	apiGroup := e.Group(apiPath)
	verGroup := apiGroup.Group(verPath)

	return verGroup
}
