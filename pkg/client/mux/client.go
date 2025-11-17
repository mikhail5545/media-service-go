// github.com/mikhail5545/media-service-go
// microservice for vitianmove project family
// Copyright (C) 2025  Mikhail Kulik

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.

// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

/*
Package mux provides the client-side implementation for gRPC [assetpb.AssetServiceClient].
It provides all client-side methods to call server-side business-logic.
*/
package mux

import (
	"context"
	"fmt"
	"log"

	assetpb "github.com/mikhail5545/proto-go/proto/media_service/mux/asset/v0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Service provides the client-side implementation for gRPC [assetpb.AssetServiceClient].
// It acts as an adapter between server-side [assetpb.AssetServiceServer] and
// client-side [assetpb.AssetServiceClient] to communicate and transport information.
// See more details about [underlying protobuf services].
//
// [underlying protobuf services]: https://github.com/mikhail5545/proto-go
type Service interface {
	// Get calls [AssetServiceClient.Get] via gRPC client connection
	// to retrieve a single not soft-deleted asset record and it's metadata.
	//
	// Returns a `NotFound` gRPC error if the record is not found.
	// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
	Get(ctx context.Context, req *assetpb.GetRequest) (*assetpb.GetResponse, error)
	// GetWithDeleted calls [AssetServiceClient.GetWithDeleted] via gRPC client connection
	// to retrieve a single asset record and it's metadata, including soft-deleted ones.
	//
	// Returns a `NotFound` gRPC error if the record is not found.
	// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
	GetWithDeleted(ctx context.Context, req *assetpb.GetWithDeletedRequest) (*assetpb.GetWithDeletedResponse, error)
	// List calls [AssetServiceClient.List] via gRPC client connection
	// to retrieve a paginated list of all not soft-deleted asset records with their metadata.
	// The response also contains the total count of such records.
	//
	// Returns `InvalidArgument` gRPC error if the provided limit or offset are invalid.
	List(ctx context.Context, req *assetpb.ListRequest) (*assetpb.ListResponse, error)
	// ListUnowned calls [AssetServiceClient.ListUnowned] via gRPC client connection
	// to retrieve a paginated list of all unowned asset records with their metadata.
	// The response also contains the total count of such records.
	//
	// Returns `InvalidArgument` gRPC error if the provided limit or offset are invalid.
	ListUnowned(ctx context.Context, req *assetpb.ListUnownedRequest) (*assetpb.ListUnownedResponse, error)
	// ListDeleted calls [AssetServiceClient.ListDeleted] via gRPC client connection
	// to retrieve a paginated list of all soft-deleted asset records with their metadata.
	// The response also contains the total count of such records.
	//
	// Returns `InvalidArgument` gRPC error if the provided limit or offset are invalid.
	ListDeleted(ctx context.Context, req *assetpb.ListDeletedRequest) (*assetpb.ListDeletedResponse, error)
	// Delete calls [AssetServiceClient.Delete] via gRPC client connection
	// to perform a soft-delete on an asset. If asset is assocaited with owners, they will be deassociated and
	// all asset metadata about it's owners will be cleared.
	//
	// Returns a `NotFound` gRPC error if any of the records are not found.
	// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
	Delete(ctx context.Context, req *assetpb.DeleteRequest) (*assetpb.DeleteResponse, error)
	// DeletePermanent calls [AssetServiceClient.DeletePermanent] via gRPC client connection
	// to permanently deletes an asset and it's metadata.
	// It also deletes an asset from the Mux.
	// This action is irreversible.
	//
	// Returns a `NotFound` gRPC error if any of the records are not found.
	// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
	// Returns an `Unavailable` gRPC error if any of Mux API calls fails.
	DeletePermanent(ctx context.Context, req *assetpb.DeletePermanentRequest) (*assetpb.DeletePermanentResponse, error)
	// Restore calls [AssetServiceClient.List] via gRPC client connection
	// to restore a soft-deleted asset.
	//
	// Returns a `NotFound` gRPC error if any of the records are not found.
	// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
	Restore(ctx context.Context, req *assetpb.RestoreRequest) (*assetpb.RestoreResponse, error)
	// CreateUploadURL calls [AssetServiceClient.CreateUploadURL] via gRPC client connection
	// to create a new signed direct upload url for direct upload to the Mux.
	// It creates new asset instance and associates it with provided owner.
	// Asset information then will be populated via Mux webhooks.
	//
	// Returns a `NotFound` gRPC error if the owner is not found.
	// Returns an `InvalidArgument` gRPC rrror if the request payload is invalid or the owner already associated with another asset.
	// Returns an `Unavailable` gRPC error if any of Mux API calls fails.
	CreateUploadURL(ctx context.Context, req *assetpb.CreateUploadUrlRequest) (*assetpb.CreateUploadUrlResponse, error)
	// CreateUnownedUploadURL calls [AssetServiceClient.CreateUnownedUploadURL] via gRPC client connection
	// to create a new signed direct upload url for direct upload to the Mux.
	// It does not interact with asset-owner relationship, created asset will be unowned.
	// Asset information then will be populated via Mux webhooks.
	//
	// Returns a `InvalidArgument` gRPC error if the request payload is invalid.
	// Returns an `Unavailable` gRPC error if any of Mux API calls fails.
	CreateUnownedUploadURL(ctx context.Context, req *assetpb.CreateUnownedUploadURLRequest) (*assetpb.CreateUnownedUploadURLResponse, error)
	// Associate calls [AssetServiceClient.Associate] via gRPC client connection
	// to associate an asset with a single owner.
	// It updates local metadata and notifies another services via gRPC calls.
	//
	// Returns `NotFound` gRPC error if an asset/owner not found.
	// Returns `InvalidArgument` gRPC error if the request payload is invalid or owner aleady associated with another asset.
	Associate(ctx context.Context, req *assetpb.AssociateRequest) (*assetpb.AssociateResponse, error)
	// Deassociate calls [AssetServiceClient.Deassociate] via gRPC client connection
	// to deassociate an asset from a single owner.
	// It updates local metadata and notifies another services via gRPC calls.
	//
	// Returns `NotFound` gRPC error if an asset/owner not found.
	// Returns `InvalidArgument` gRPC error if the request payload is invalid.
	Deassociate(ctx context.Context, req *assetpb.DeassociateRequest) (*assetpb.DeassociateResponse, error)
	// UpdateOwners calls [AssetServiceClient.UpdateOwners] via gRPC client connection
	// to processe asset ownership relations changes. It recieves an updated list of asset owners, updates local metadata
	// for asset (about it's owners), processes the diff between old and new owners and notifies external services about this ownership
	// changes via gRPC connection.
	//
	// Returns `NotFound` gRPC error if an asset not found.
	// Returns `InvalidArgument` gRPC error if the request payload is invalid.
	UpdateOwners(ctx context.Context, req *assetpb.UpdateOwnersRequest) (*assetpb.UpdateOwnersResponse, error)
	// Close tears down connection to the client and all underlying connections.
	Close() error
}

// Client holds [grpc.ClientConn] to connect to the client and
// [assetpb.AssetServiceClient] client to call server-side methods.
// See more details about [underlying protobuf services].
//
// [underlying protobuf services]: https://github.com/mikhail5545/proto-go
type Client struct {
	conn   *grpc.ClientConn
	client assetpb.AssetServiceClient
}

// New creates a new [mux.Server] client.
func New(ctx context.Context, addr string, opt ...grpc.CallOption) (Service, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithDefaultCallOptions(opt...))
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection: %w", err)
	}
	log.Printf("Connection to mux asset service at %s established", addr)

	client := assetpb.NewAssetServiceClient(conn)
	return &Client{
		conn:   conn,
		client: client,
	}, nil
}

// Get calls [AssetServiceClient.Get] via gRPC client connection
// to retrieve a single not soft-deleted asset record and it's metadata.
//
// Returns a `NotFound` gRPC error if the record is not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (c *Client) Get(ctx context.Context, req *assetpb.GetRequest) (*assetpb.GetResponse, error) {
	return c.client.Get(ctx, req)
}

// GetWithDeleted calls [AssetServiceClient.GetWithDeleted] via gRPC client connection
// to retrieve a single asset record and it's metadata, including soft-deleted ones.
//
// Returns a `NotFound` gRPC error if the record is not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (c *Client) GetWithDeleted(ctx context.Context, req *assetpb.GetWithDeletedRequest) (*assetpb.GetWithDeletedResponse, error) {
	return c.client.GetWithDeleted(ctx, req)
}

// List calls [AssetServiceClient.List] via gRPC client connection
// to retrieve a paginated list of all not soft-deleted asset records with their metadata.
// The response also contains the total count of such records.
//
// Returns `InvalidArgument` gRPC error if the provided limit or offset are invalid.
func (c *Client) List(ctx context.Context, req *assetpb.ListRequest) (*assetpb.ListResponse, error) {
	return c.client.List(ctx, req)
}

// ListUnowned calls [AssetServiceClient.ListUnowned] via gRPC client connection
// to retrieve a paginated list of all unowned asset records with their metadata.
// The response also contains the total count of such records.
//
// Returns `InvalidArgument` gRPC error if the provided limit or offset are invalid.
func (c *Client) ListUnowned(ctx context.Context, req *assetpb.ListUnownedRequest) (*assetpb.ListUnownedResponse, error) {
	return c.client.ListUnowned(ctx, req)
}

// ListDeleted calls [AssetServiceClient.ListDeleted] via gRPC client connection
// to retrieve a paginated list of all soft-deleted asset records with their metadata.
// The response also contains the total count of such records.
//
// Returns `InvalidArgument` gRPC error if the provided limit or offset are invalid.
func (c *Client) ListDeleted(ctx context.Context, req *assetpb.ListDeletedRequest) (*assetpb.ListDeletedResponse, error) {
	return c.client.ListDeleted(ctx, req)
}

// Delete calls [AssetServiceClient.Delete] via gRPC client connection
// to perform a soft-delete on an asset. If asset is assocaited with owners, they will be deassociated and
// all asset metadata about it's owners will be cleared.
//
// Returns a `NotFound` gRPC error if any of the records are not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (c *Client) Delete(ctx context.Context, req *assetpb.DeleteRequest) (*assetpb.DeleteResponse, error) {
	return c.client.Delete(ctx, req)
}

// DeletePermanent calls [AssetServiceClient.DeletePermanent] via gRPC client connection
// to permanently deletes an asset and it's metadata.
// It also deletes an asset from the Mux.
// This action is irreversible.
//
// Returns a `NotFound` gRPC error if any of the records are not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
// Returns an `Unavailable` gRPC error if any of Mux API calls fails.
func (c *Client) DeletePermanent(ctx context.Context, req *assetpb.DeletePermanentRequest) (*assetpb.DeletePermanentResponse, error) {
	return c.client.DeletePermanent(ctx, req)
}

// Restore calls [AssetServiceClient.List] via gRPC client connection
// to restore a soft-deleted asset.
//
// Returns a `NotFound` gRPC error if any of the records are not found.
// Returns an `InvalidArgument` gRPC error if the provided ID is not a valid UUID.
func (c *Client) Restore(ctx context.Context, req *assetpb.RestoreRequest) (*assetpb.RestoreResponse, error) {
	return c.client.Restore(ctx, req)
}

// CreateUploadURL calls [AssetServiceClient.CreateUploadURL] via gRPC client connection
// to create a new signed direct upload url for direct upload to the Mux.
// It creates new asset instance and associates it with provided owner.
// Asset information then will be populated via Mux webhooks.
//
// Returns a `NotFound` gRPC error if the owner is not found.
// Returns an `InvalidArgument` gRPC rrror if the request payload is invalid or the owner already associated with another asset.
// Returns an `Unavailable` gRPC error if any of Mux API calls fails.
func (c *Client) CreateUploadURL(ctx context.Context, req *assetpb.CreateUploadUrlRequest) (*assetpb.CreateUploadUrlResponse, error) {
	return c.client.CreateUploadUrl(ctx, req)
}

// CreateUnownedUploadURL calls [AssetServiceClient.CreateUnownedUploadURL] via gRPC client connection
// to create a new signed direct upload url for direct upload to the Mux.
// It does not interact with asset-owner relationship, created asset will be unowned.
// Asset information then will be populated via Mux webhooks.
//
// Returns a `InvalidArgument` gRPC error if the request payload is invalid.
// Returns an `Unavailable` gRPC error if any of Mux API calls fails.
func (c *Client) CreateUnownedUploadURL(ctx context.Context, req *assetpb.CreateUnownedUploadURLRequest) (*assetpb.CreateUnownedUploadURLResponse, error) {
	return c.client.CreateUnownedUploadUrl(ctx, req)
}

// Associate calls [AssetServiceClient.Associate] via gRPC client connection
// to associate an asset with a single owner.
// It updates local metadata and notifies another services via gRPC calls.
//
// Returns `NotFound` gRPC error if an asset/owner not found.
// Returns `InvalidArgument` gRPC error if the request payload is invalid or owner aleady associated with another asset.
func (c *Client) Associate(ctx context.Context, req *assetpb.AssociateRequest) (*assetpb.AssociateResponse, error) {
	return c.client.Associate(ctx, req)
}

// Deassociate calls [AssetServiceClient.Deassociate] via gRPC client connection
// to deassociate an asset from a single owner.
// It updates local metadata and notifies another services via gRPC calls.
//
// Returns `NotFound` gRPC error if an asset/owner not found.
// Returns `InvalidArgument` gRPC error if the request payload is invalid.
func (c *Client) Deassociate(ctx context.Context, req *assetpb.DeassociateRequest) (*assetpb.DeassociateResponse, error) {
	return c.client.Deassociate(ctx, req)
}

// UpdateOwners calls [AssetServiceClient.UpdateOwners] via gRPC client connection
// to processe asset ownership relations changes. It recieves an updated list of asset owners, updates local metadata
// for asset (about it's owners), processes the diff between old and new owners and notifies external services about this ownership
// changes via gRPC connection.
//
// Returns `NotFound` gRPC error if an asset not found.
// Returns `InvalidArgument` gRPC error if the request payload is invalid.
func (c *Client) UpdateOwners(ctx context.Context, req *assetpb.UpdateOwnersRequest) (*assetpb.UpdateOwnersResponse, error) {
	return c.client.UpdateOwners(ctx, req)
}

// Close tears down connection to the client and all underlying connections.
func (c *Client) Close() error {
	return c.conn.Close()
}
