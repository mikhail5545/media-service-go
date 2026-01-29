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

package credentials

import (
	"google.golang.org/grpc/credentials"
)

type Credentials struct {
	PostgresDB    *PostgresDBCredentials
	MongoDB       *MongoDBCredentials
	GRPCServer    *GRPCServerCredentials
	GRPCClient    *GRPCClientCredentials
	MuxAPI        *MuxAPICredentials
	CloudinaryAPI *CloudinaryAPICredentials
}

type PostgresDBCredentials struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type MongoDBCredentials struct {
	ConnectionString string
	DBName           string
}

type GRPCServerCredentials struct {
	Credentials credentials.TransportCredentials
}

type GRPCClientCredentials struct {
	Address     string
	Credentials credentials.TransportCredentials
}

type MuxAPICredentials struct {
	APIToken              string
	SecretKey             string
	SigningKeyID          string
	SigningKeyPrivate     string
	PlaybackRestrictionID string
}

type CloudinaryAPICredentials struct {
	CloudName string
	APIKey    string
	APISecret string
}
