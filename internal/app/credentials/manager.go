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
	"context"
	"fmt"
	"slices"

	"github.com/1password/onepassword-sdk-go"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

type Manager struct {
	src         *Sources
	Credentials *Credentials
	opClient    *onepassword.Client
	logger      *zap.Logger
}

func New(ctx context.Context, src *Sources, token string, logger *zap.Logger) (*Manager, error) {
	opClient, err := onepassword.NewClient(ctx,
		onepassword.WithServiceAccountToken(token),
		onepassword.WithIntegrationInfo("product-service-go", "v0.1.0"),
	)
	if err != nil {
		return nil, err
	}
	return &Manager{
		src:         src,
		Credentials: &Credentials{},
		opClient:    opClient,
		logger:      logger.With(zap.String("component", "/app/credentials/manager.go")),
	}, nil
}

func (m *Manager) Source() *Sources {
	return m.src
}

func (m *Manager) OPClient() *onepassword.Client {
	return m.opClient
}

func (m *Manager) ResolveAll(ctx context.Context) error {
	if err := m.ResolvePostgresDBCredentials(ctx); err != nil {
		return err
	}
	if err := m.ResolveMongoDBCredentials(ctx); err != nil {
		return err
	}
	if err := m.ResolveGRPCServerCredentials(ctx); err != nil {
		return err
	}
	if err := m.ResolveGRPCClientCredentials(ctx); err != nil {
		return err
	}
	if err := m.ResolveMuxAPICredentials(ctx); err != nil {
		return err
	}
	if err := m.ResolveCloudinaryAPICredentials(ctx); err != nil {
		return err
	}
	return nil
}

// resolve resolves multiple secret references using the 1Password Secrets API.
func (m *Manager) resolve(ctx context.Context, references []string) (map[string]string, error) {
	resolved, err := m.opClient.SecretsAPI.ResolveAll(ctx, references)
	if err != nil {
		m.logger.Error("failed to resolve secrets", zap.Error(err))
		return nil, err
	}
	result := make(map[string]string)
	for _, ref := range references {
		resp := resolved.IndividualResponses[ref]
		if resp.Error != nil {
			return nil, fmt.Errorf("failed to resolve secret for reference %s: %v", ref, resp.Error)
		}
		result[ref] = resp.Content.Secret
	}
	return result, nil
}

// readItemFiles reads the specified files from a 1Password item.
func (m *Manager) readItemFiles(ctx context.Context, item onepassword.Item, nameIn []string) (result map[string][]byte, err error) {
	result = make(map[string][]byte)
	for _, file := range item.Files {
		if !slices.Contains(nameIn, file.Attributes.Name) {
			// If the file name is not in the requested list, skip it
			continue
		}
		result[file.Attributes.Name], err = m.opClient.Items().Files().Read(ctx, item.VaultID, item.ID, file.Attributes)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s from item %s: %v", file.Attributes.Name, item.ID, err)
		}
	}
	return result, nil
}

func (m *Manager) extractItem(ctx context.Context, vaultRef, itemRef string) (onepassword.Item, error) {
	resolved, err := m.resolve(ctx, []string{vaultRef, itemRef})
	if err != nil {
		m.logger.Error("failed to resolve vault and item references", zap.String("vaultRef", vaultRef), zap.String("itemRef", itemRef), zap.Error(err))
		return onepassword.Item{}, err
	}
	item, err := m.opClient.Items().Get(ctx, resolved[vaultRef], resolved[itemRef])
	if err != nil {
		m.logger.Error("failed to get item from 1Password", zap.String("vaultID", resolved[vaultRef]), zap.String("itemID", resolved[itemRef]), zap.Error(err))
		return onepassword.Item{}, err
	}
	return item, nil
}

func (m *Manager) ResolveMuxAPICredentials(ctx context.Context) error {
	resolved, err := m.resolve(ctx, []string{
		m.src.MuxAPI.APITokenRef, m.src.MuxAPI.SecretKeyRef,
		m.src.MuxAPI.PlaybackRestrictionIDRef,
		m.src.MuxAPI.SigningKeyIDRef, m.src.MuxAPI.SigningKeyPrivateRef,
	})
	if err != nil {
		m.logger.Error("failed to resolve Mux API credentials", zap.Error(err))
		return err
	}
	m.Credentials.MuxAPI = &MuxAPICredentials{
		APIToken:              resolved[m.src.MuxAPI.APITokenRef],
		SecretKey:             resolved[m.src.MuxAPI.SecretKeyRef],
		PlaybackRestrictionID: resolved[m.src.MuxAPI.PlaybackRestrictionIDRef],
		SigningKeyID:          resolved[m.src.MuxAPI.SigningKeyIDRef],
		SigningKeyPrivate:     resolved[m.src.MuxAPI.SigningKeyPrivateRef],
	}
	return nil
}

func (m *Manager) ResolveCloudinaryAPICredentials(ctx context.Context) error {
	resolved, err := m.resolve(ctx, []string{m.src.CloudinaryAPI.CloudNameRef, m.src.CloudinaryAPI.APIKeyRef, m.src.CloudinaryAPI.APISecretRef})
	if err != nil {
		m.logger.Error("failed to resolve Cloudinary API credentials", zap.Error(err))
		return err
	}
	m.Credentials.CloudinaryAPI = &CloudinaryAPICredentials{
		CloudName: resolved[m.src.CloudinaryAPI.CloudNameRef],
		APIKey:    resolved[m.src.CloudinaryAPI.APIKeyRef],
		APISecret: resolved[m.src.CloudinaryAPI.APISecretRef],
	}
	return nil
}

func (m *Manager) ResolvePostgresDBCredentials(ctx context.Context) error {
	resolved, err := m.resolve(ctx, []string{
		m.src.PostgresDB.HostRef, m.src.PostgresDB.PortRef,
		m.src.PostgresDB.UserRef, m.src.PostgresDB.PasswordRef,
		m.src.PostgresDB.DBNameRef,
	})
	if err != nil {
		m.logger.Error("failed to resolve Postgres DB credentials", zap.Error(err))
		return err
	}
	m.Credentials.PostgresDB = &PostgresDBCredentials{
		Host:     resolved[m.src.PostgresDB.HostRef],
		Port:     resolved[m.src.PostgresDB.PortRef],
		User:     resolved[m.src.PostgresDB.UserRef],
		Password: resolved[m.src.PostgresDB.PasswordRef],
		DBName:   resolved[m.src.PostgresDB.DBNameRef],
	}
	return nil
}

func (m *Manager) ResolveMongoDBCredentials(ctx context.Context) error {
	connString, err := m.opClient.SecretsAPI.Resolve(ctx, m.src.MongoDB.ConnectionStringRef)
	if err != nil {
		m.logger.Error("failed to resolve MongoDB credentials", zap.Error(err))
		return err
	}
	m.Credentials.MongoDB = &MongoDBCredentials{
		ConnectionString: connString,
	}
	return nil
}

func (m *Manager) ResolveGRPCServerCredentials(ctx context.Context) error {
	item, err := m.extractItem(ctx, m.src.GRPCServer.CertVaultRef, m.src.GRPCServer.CertItemRef)
	if err != nil {
		m.logger.Error("failed to extract gRPC server cert item", zap.Error(err))
		return err
	}
	files, err := m.readItemFiles(ctx, item, []string{"ca.pem", "server.crt", "server.key"})
	if err != nil {
		m.logger.Error("failed to read gRPC server cert files", zap.Error(err))
		return err
	}
	tlsConfig, err := buildTLSConfig(files["ca.pem"], files["server.crt"], files["server.key"])
	if err != nil {
		m.logger.Error("failed to create TLS config for gRPC server", zap.Error(err))
		return err
	}
	m.Credentials.GRPCServer = &GRPCServerCredentials{
		Credentials: credentials.NewTLS(tlsConfig),
	}
	return nil
}

func (m *Manager) ResolveGRPCClientCredentials(ctx context.Context) error {
	item, err := m.extractItem(ctx, m.src.GRPCClient.CertVaultRef, m.src.GRPCClient.CertItemRef)
	if err != nil {
		m.logger.Error("failed to extract gRPC client cert item", zap.Error(err))
		return err
	}
	files, err := m.readItemFiles(ctx, item, []string{"ca.pem", "server.crt", "server.key"})
	if err != nil {
		m.logger.Error("failed to read gRPC client cert files", zap.Error(err))
		return err
	}
	tlsConfig, err := buildTLSConfig(files["ca.pem"], files["server.crt"], files["server.key"])
	if err != nil {
		m.logger.Error("failed to create TLS config for gRPC client", zap.Error(err))
		return err
	}
	address, err := m.opClient.SecretsAPI.Resolve(ctx, m.src.GRPCClient.AddressRef)
	if err != nil {
		m.logger.Error("failed to resolve gRPC client address", zap.Error(err))
		return err
	}
	m.Credentials.GRPCClient = &GRPCClientCredentials{
		Address:     address,
		Credentials: credentials.NewTLS(tlsConfig),
	}
	return nil
}
