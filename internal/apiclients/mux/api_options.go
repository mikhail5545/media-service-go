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

package mux

import (
	"encoding/base64"
	"fmt"
)

type config struct {
	corsOrigin            string
	test                  bool
	signingKeyID          string
	signingKeyPrivateKey  []byte
	playbackRestrictionID string
}

type Option func(*config) error

func WithCORSOrigin(origin string) Option {
	return func(c *config) error {
		c.corsOrigin = origin
		return nil
	}
}

func WithTestMode(test bool) Option {
	return func(c *config) error {
		c.test = test
		return nil
	}
}

func WithPlaybackRestrictionID(restrictionID string) Option {
	return func(c *config) error {
		c.playbackRestrictionID = restrictionID
		return nil
	}
}

func WithSigningKey(keyID string, b64key string) Option {
	return func(c *config) error {
		if keyID == "" || b64key == "" {
			return fmt.Errorf("missing mux signing key ID or private key")
		}

		privateKeyBytes, err := base64.StdEncoding.DecodeString(b64key)
		if err != nil {
			return fmt.Errorf("failed to decode mux signing private key: %w", err)
		}

		c.signingKeyID = keyID
		c.signingKeyPrivateKey = privateKeyBytes
		return nil
	}
}
