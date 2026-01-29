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

import "os"

type Sources struct {
	GRPCServer    GRPCServerRefs
	GRPCClient    GRPCClientRefs
	PostgresDB    PostgresDBRefs
	MongoDB       MongoDBRefs
	MuxAPI        MuxAPIRefs
	CloudinaryAPI CloudinaryAPRefs
}

type GRPCServerRefs struct {
	CertVaultRef string
	CertItemRef  string
}

type GRPCClientRefs struct {
	AddressRef   string
	CertVaultRef string
	CertItemRef  string
}

type PostgresDBRefs struct {
	HostRef     string
	PortRef     string
	UserRef     string
	PasswordRef string
	DBNameRef   string
}

type MongoDBRefs struct {
	ConnectionStringRef string
}

type MuxAPIRefs struct {
	APITokenRef              string
	SecretKeyRef             string
	SigningKeyIDRef          string
	SigningKeyPrivateRef     string
	PlaybackRestrictionIDRef string
}

type CloudinaryAPRefs struct {
	CloudNameRef string
	APIKeyRef    string
	APISecretRef string
}

func LoadSources() *Sources {
	return &Sources{
		GRPCServer: GRPCServerRefs{
			CertVaultRef: os.Getenv("GRPC_SERVER_CERT_VAULT_REF"),
			CertItemRef:  os.Getenv("GRPC_SERVER_CERT_ITEM_REF"),
		},
		GRPCClient: GRPCClientRefs{
			AddressRef:   os.Getenv("GRPC_CLIENT_ADDRESS_REF"),
			CertVaultRef: os.Getenv("GRPC_CLIENT_CERT_VAULT_REF"),
			CertItemRef:  os.Getenv("GRPC_CLIENT_CERT_ITEM_REF"),
		},
		PostgresDB: PostgresDBRefs{
			HostRef:     os.Getenv("POSTGRES_HOST_REF"),
			PortRef:     os.Getenv("POSTGRES_PORT_REF"),
			UserRef:     os.Getenv("POSTGRES_USER_REF"),
			PasswordRef: os.Getenv("POSTGRES_PASSWORD_REF"),
			DBNameRef:   os.Getenv("POSTGRES_DBNAME_REF"),
		},
		MongoDB: MongoDBRefs{
			ConnectionStringRef: os.Getenv("MONGO_CONNECTION_STRING_REF"),
		},
		MuxAPI: MuxAPIRefs{
			APITokenRef:              os.Getenv("MUX_API_TOKEN_REF"),
			SecretKeyRef:             os.Getenv("MUX_SECRET_KEY_REF"),
			SigningKeyIDRef:          os.Getenv("MUX_SIGNING_KEY_ID_REF"),
			SigningKeyPrivateRef:     os.Getenv("MUX_SIGNING_KEY_PRIVATE_REF"),
			PlaybackRestrictionIDRef: os.Getenv("MUX_PLAYBACK_RESTRICTION_ID_REF"),
		},
		CloudinaryAPI: CloudinaryAPRefs{
			CloudNameRef: os.Getenv("CLD_CLOUD_NAME_REF"),
			APIKeyRef:    os.Getenv("CLD_API_KEY_REF"),
			APISecretRef: os.Getenv("CLD_API_SECRET_REF"),
		},
	}
}
