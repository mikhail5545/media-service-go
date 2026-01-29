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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/mikhail5545/media-service-go/internal/app"
	"github.com/spf13/pflag"
)

func main() {
	ctx := context.Background()
	cfg := parseArguments()

	cfg.Log.AppName = "media-service"

	application, err := app.New(ctx, &cfg)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to initialize application: %v\n", err)
		os.Exit(1)
	}

	if err := application.Init(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "failed to initialize application components: %v\n", err)
	}

	if err := application.Run(ctx); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "application runtime error: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := application.Close(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "failed to close application: %v\n", err)
		}
	}()
}

func parseArguments() app.Config {
	cfg := app.Config{}

	pflag.Int64VarP(&cfg.GRPC.Port, "grpc-port", "g", 50052, "gRPC server port")
	pflag.Int64VarP(&cfg.HTTP.Port, "http-port", "p", 8082, "HTTP server port")
	pflag.IntVarP(&cfg.GracefulShutdownTimeoutSeconds, "graceful-shutdown-timeout", "t", 15, "Graceful shutdown timeout in seconds")
	pflag.StringVarP(&cfg.Log.Directory, "log-directory", "l", "./logs", "Directory to store log files")
	pflag.BoolVarP(&cfg.Log.UseTimestamp, "log-use-timestamp", "", true, "Whether to use timestamp in log file names")
	pflag.BoolVarP(&cfg.Mux.TestMode, "mux-test-mode", "", false, "Enable Mux test mode")
	pflag.StringVarP(&cfg.Mux.CORSOrigin, "mux-cors-origin", "", "", "Mux CORS origin")
	pflag.Parse()

	return cfg
}
