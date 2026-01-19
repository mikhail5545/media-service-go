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

package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func fallbackLogDir(logCfg LogConfig, filename string) (*os.File, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home dir: %w", err)
	}
	fallback := filepath.Join(home, ".local", "share", logCfg.AppName)
	if err := os.MkdirAll(filepath.Dir(fallback), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create fallback log directory: %w", err)
	}
	fullPath := filepath.Join(fallback, filename)
	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open fallback log file: %w", err)
	}
	return file, nil
}

func openLogFile(logCfg LogConfig) (*os.File, error) {
	var filename string
	if logCfg.UseTimestamp {
		now := time.Now()
		filename = fmt.Sprintf("%02d-%s-%d.log", now.Day(), now.Month().String(), now.Year())
	} else {
		filename = "app.log"
	}
	// Ensure directory exists
	if err := os.MkdirAll(logCfg.Directory, 0o755); err != nil {
		if os.IsPermission(err) {
			// Try fallback directory if permission denied
			file, err := fallbackLogDir(logCfg, filename)
			if err != nil {
				return nil, err
			}
			return file, nil
		}
		return nil, fmt.Errorf("failed to create log directory '%s': %w", logCfg.Directory, err)
	}

	fullPath := filepath.Join(logCfg.Directory, filename)

	file, err := os.OpenFile(fullPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file '%q': %w", fullPath, err)
	}
	return file, nil
}

// newLogger creates a new zap.Logger based on the provided LogConfig.
// Make sure to call the returned cleanup function to close file handles to prevent potential recourse leak.
func newLogger(logCfg LogConfig) (*zap.Logger, func(), error) {
	f, err := openLogFile(logCfg)
	if err != nil {
		return nil, nil, err
	}

	// writers
	consoleWS := zapcore.Lock(os.Stdout)
	fileWS := zapcore.AddSync(f)

	// encoders
	consoleEncCfg := zap.NewDevelopmentEncoderConfig()
	consoleEncCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEnc := zapcore.NewConsoleEncoder(consoleEncCfg)

	fileEnc := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

	// cores: console for debug+; file for info+
	core := zapcore.NewTee(
		zapcore.NewCore(consoleEnc, consoleWS, zapcore.DebugLevel),
		zapcore.NewCore(fileEnc, fileWS, zapcore.InfoLevel),
	)

	logger := zap.New(core, zap.AddCaller())
	cleanup := func() {
		_ = logger.Sync()
		_ = f.Close()
	}
	return logger, cleanup, nil
}

type zapEchoWriter struct {
	s *zap.SugaredLogger
}

func newZapEchoWriter(s *zap.Logger) *zapEchoWriter {
	return &zapEchoWriter{s: s.Sugar()}
}

func detectLevel(msg string) string {
	levelRE := regexp.MustCompile(`(?i)\b(DEBUG|INFO|WARN(?:ING)?|ERROR|FATAL)\b`)
	m := levelRE.FindStringSubmatch(msg)
	if len(m) < 2 {
		return ""
	}
	l := strings.ToUpper(m[1])
	if l == "WARNING" {
		l = "WARN"
	}
	switch l {
	case "DEBUG", "INFO", "WARN", "ERROR", "FATAL":
		return l
	default:
		return ""
	}
}

func (w *zapEchoWriter) Write(p []byte) (n int, err error) {
	msg := strings.TrimSpace(string(p))
	if msg == "" {
		return len(p), nil
	}
	switch detectLevel(msg) {
	case "DEBUG":
		w.s.Debug(msg)
	case "INFO":
		w.s.Info(msg)
	case "WARN":
		w.s.Warn(msg)
	case "ERROR":
		w.s.Error(msg)
	case "FATAL":
		w.s.Fatal(msg)
	default:
		w.s.Info(msg)
	}
	return len(p), nil
}

func integrateWithEcho(e *echo.Echo, logger *zap.Logger) {
	e.Logger.SetOutput(newZapEchoWriter(logger))
}
