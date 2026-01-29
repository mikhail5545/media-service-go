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

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	errorhandler "github.com/mikhail5545/media-service-go/internal/handlers/errors"
	"github.com/mikhail5545/media-service-go/internal/routers"
	"github.com/mikhail5545/media-service-go/internal/routers/admin"
	"go.uber.org/zap"
)

func setupRouters(e *echo.Echo, services *Services) {
	baseGroup := routers.Init(e, routers.Config{
		Api: "/api",
		Ver: "/v1",
		Use: []echo.MiddlewareFunc{
			middleware.Logger(),
			middleware.Recover(),
			middleware.ContextTimeout(60 * time.Second),
		},
		HTTPErrorHandler: errorhandler.HTTPErrorHandler,
	})

	adminRtr := admin.New(admin.Dependencies{
		CldSvc: services.CldSvc,
		MuxSvc: services.MuxSvc,
	})
	adminRtr.Setup(baseGroup)
}

func runHTTPServer(e *echo.Echo, port int64, logger *zap.Logger, errChan chan<- error) {
	httpListenAddr := fmt.Sprintf(":%d", port)
	logger.Info("Starting HTTP server", zap.String("address", httpListenAddr))
	if err := e.Start(httpListenAddr); err != nil {
		errChan <- err
	} else {
		errChan <- nil
	}
}

func shutdownHTTPServer(shutdownContext context.Context, e *echo.Echo, logger *zap.Logger) {
	if err := e.Shutdown(shutdownContext); err != nil {
		logger.Error("Error during HTTP server shutdown", zap.Error(err))
	} else {
		logger.Info("HTTP server shutdown completed")
	}
}
