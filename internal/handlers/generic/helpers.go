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

package generic

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
)

// HandleGet abstracts pattern of 'bind request with custom binder -> call service method -> return response'.
// binder must return echo.NewHTTPError in case of any error for abstraction to work properly.
func HandleGet[Req any, Res any](
	c echo.Context,
	binder func(echo.Context) (*Req, error),
	fn func(context.Context, *Req) (*Res, error),
	responseKey string,
) error {
	req, err := binder(c)
	if err != nil {
		return err
	}
	res, err := fn(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{responseKey: res})
}

// HandleList abstracts service List methods, 'bind request -> call service method -> return paginated response'.
func HandleList[Req any, Res any](
	c echo.Context,
	fn func(context.Context, *Req) ([]*Res, string, error),
	responseKey string,
) error {
	req := new(Req)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body payload")
	}
	res, nextPageToken, err := fn(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, map[string]any{
		responseKey:       res,
		"next_page_token": nextPageToken,
	})
}

// Handle abstracts the pattern: 'bind request -> call service operation -> return JSON response'.
func Handle[Req any, Res any](
	c echo.Context,
	fn func(context.Context, *Req) (Res, error),
	status int,
	responseKey string,
) error {
	req := new(Req)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	res, err := fn(c.Request().Context(), req)
	if err != nil {
		return err
	}
	return c.JSON(status, map[string]any{responseKey: res})
}

// HandleVoid abstracts the pattern: 'bind -> Service Call -> No Content Response'
func HandleVoid[Req any](
	c echo.Context,
	op func(context.Context, *Req) error,
	status int,
) error {
	req := new(Req)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body")
	}
	if err := op(c.Request().Context(), req); err != nil {
		return err
	}
	return c.NoContent(status)
}
