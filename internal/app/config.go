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

type Config struct {
	HTTP                           HTTPConfig
	GRPC                           GRPCConfig
	GRPCClient                     GRPCClientConfig
	Log                            LogConfig
	MongoDB                        MongoDBConfig
	GracefulShutdownTimeoutSeconds int
	Mux                            MuxAPIConfig
}

type HTTPConfig struct {
	Port int64
}

type GRPCConfig struct {
	Port int64
}

type LogConfig struct {
	Directory    string
	UseTimestamp bool
	AppName      string
}

type MuxAPIConfig struct {
	TestMode   bool
	CORSOrigin string
}

type MongoDBConfig struct {
	DbName string
}

type GRPCClientConfig struct {
	Address string
}
