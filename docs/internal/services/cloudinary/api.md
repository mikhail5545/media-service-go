# Cloudinary Service API

This document describes the API provided by the Cloudinary service, which handles interactions with Cloudinary assets in the microservice. The service defines core business logic for creating, managing, and deleting assets stored in Cloudinary.

For an overview of the service, see [README.md](./README.md). For high-level processes, see [Architecture Documentation](../../architecture/).

## Overview

The Cloudinary service is defined by the `Service` interface and implemented by the `service` struct. It provides methods for interacting with Cloudinary’s API, such as creating signed upload URLs, deleting assets, and managing metadata. The service integrates with the relational database and ArangoDB for asset storage and metadata.

Key methods include:

- `CreateSignedUploadURL`: Generates parameters for a signed upload URL (see [Signed Upload URL Creation Process](./flow/create_signed_upload_url_flow.md)).
- `UpdateOwners`: Updates the ownership of an asset. (see [Owners Update Process](./flow/update_owners_flow.md))
- `SuccessfulUpload`: Creates an asset record after a successful upload to Cloudinary (see [Successful Upload Process](./flow/successful_upload_flow.md)).
- `CleanupOrphanAssets`: Removes assets from Cloudinary that don't have a corresponding record in the local database (see [Orphan Assets Cleanup Process](./flow/cleanup_orphan_assets_flow.md)).
- `Delete`: Soft-deletes an asset (see [Asset Deletion Process](./flow/delete_flow.md)).
- `DeletePermanent`: Permanently deletes an asset from the database and Cloudinary (see [Asset Permanent Deletion Process](./flow/delete_permanent_flow.md)).
- `Restore`: Restores a soft-deleted asset (see [Asset Restore Process](./flow/restore_flow.md)).

## CreateSignedUploadURL

The `CreateSignedUploadURL` method generates parameters required to construct a signed upload URL for Cloudinary.

### Input Parameters

The method accepts a `CreateSignedUploadURLRequest` struct with the following fields:

| Field     | Type   | Required | Description                                      |
|-----------|--------|----------|--------------------------------------------------|
| Eager     | *string | Optional | Transformations to apply (e.g., "c_thumb,w_200"). |
| PublicID  | string | Required | Unique identifier for the asset in Cloudinary.   |
| File      | string | Required | Path or identifier for the file to upload.       |

Example:

```go
e := "c_thumb,w_200"

req := &assetmodel.CreateSignedUploadURLRequest{
    Eager:    &e,
    PublicID: "asset_123",
    File:     "/path/to/file.jpg",
}
```

### Output

Returns a `map[string]any` containing parameters for constructing a signed upload URL. The URL must be assembled manually on the client side (e.g., frontend).

|Key|Type|Description|
|---|----|-----------|
|signature|string|Generated signature for the URL|
|public_id|string|Cloudinary public ID.|
|timestamp|int64|Unix timestamp for the request.|
|api_key|string|Cloudinary API key.|

Example:

```json
{
    "signature": "dbad3qa8fh3gbf",
    "public_id": "asset/233",
    "timestamp": 1634567890,
    "api_key": "cloudinary_api_key",
}
```

### Description

The method:

1. Validates the `CreateSignedUploadURLRequest` payload.
2. Calls the Cloudinary API client's `SignUploadParams` method with the provided `Eager`, `PublicID` and `File` values.
3. Constructs th eresponse map with the signature, public ID, timestamp, and API key.
4. Returns the map for use in constructing the upload URL.

The client must combine these parameters with Clousinary's base URL to form the final upload URL. See [Cloudinary documentation](https://cloudinary.com/documentation/upload_images#signed_uploads) for more details.

### Errors

The method returns the following errors:

- `ErrInvalidArgument`: Invalid or missing request fields (HTTP 400).
- `ErrCloudinaryAPI`: Cloudinary API call failed (HTTP 503).
- Other error (treat as internal): Database or internal error (HTTP 500).

### Example usage

```go
req := &CreateSignedUploadURLRequest{
    PublicID: "asset/111",
    File: "path/to/file.jpg",
}
resp, err := svc.CreateSignedUploadURL(ctx, req)
if err != nil{
    log.Fatal(err)
}
fmt.Printf("Signature: %w\n", resp["signature"])
```

## UpdateOwners

The `UpdateOwners` method processes changes in asset  ownership. It receives an updated list of owners, calculates the difference, updates the metadata database, and notifies external services of the changes.

### Input Parameters

The method accepts an `UpdateOwnesRequest` struct with the following fields:

| Field  | Type           | Required | Description                               |
|--------|----------------|----------|-------------------------------------------|
| ID     | string         | Required | The UUID of the asset to update.          |
| Owners | []metamodel.Owner | Required | The new, complete list of asset owners.|

The `metamodel.Owner` struct has the following fields:

| Field     | Type   | Required | Description                               |
|-----------|--------|----------|-------------------------------------------|
| OwnerType | string | Required | The type of the owner (e.g., "product").  |
| OwnerID   | string | Required | The UUID of the owner.                    |

Example:

```go
req := &assetmodel.UpdateOwnersRequest{
    ID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    Owners: []metamodel.Owner{
        {OwnerType: "product", OwnerID: "c47ac10b-58cc-4372-a567-0e02b2c3d479"},
    },
}
```

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates the request payload.
2. Retrieves the asset from the relational database (Postgres) to ensure it exists.
3. Fetches the current list of owners from the metadata database (ArangoDB).
4. Compares the current and new owner lists to determine which owners to add and which to remove.
5. Updates the asset's owner list in ArangoDB with the new list.
6. Notifies relevant external services (via gRPC) about the owners that were added and removed.

### Errors

- `ErrInvalidArgument`: Invalid or missing request fields (HTTP 400).
- `ErrNotFound`: The specified asset ID does not exist (HTTP 404).
- Other error (treat as internal): Database or internal error (HTTP 500).

### Example usage

```go
req := &assetmodel.UpdateOwnersRequest{
    ID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    Owners: []metamodel.Owner{
        {OwnerType: "product", OwnerID: "c47ac10b-58cc-4372-a567-0e02b2c3d479"},
    },
}
err := svc.UpdateOwners(ctx, req)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Owners updated successfully.")
```

## SuccessfulUpload

The `SuccessfulUpload` method is called after a file has been successfully uploaded to Cloudinary. It creates a new asset record in the local database, establishes initial ownership relations, and notifies external services.

### Input Parameters

The method accepts a `SuccessfulUploadRequest` struct with the following fields:

| Field              | Type           | Required | Description                                       |
|--------------------|----------------|----------|---------------------------------------------------|
| CloudinaryAssetID  | string         | Required | The asset ID from Cloudinary.                     |
| CloudinaryPublicID | string         | Required | The public ID from Cloudinary.                    |
| ResourceType       | string         | Required | The type of resource (e.g., "image", "video").    |
| Format             | string         | Required | The format of the asset (e.g., "jpg", "png").     |
| Width              | int            | Required | The width of the asset in pixels.                 |
| Height             | int            | Required | The height of the asset in pixels.                |
| URL                | string         | Required | The insecure URL of the asset.                    |
| SecureURL          | string         | Required | The secure (HTTPS) URL of the asset.              |
| AssetFolder        | string         | Optional | The folder in Cloudinary where the asset is stored. |
| DisplayName        | string         | Optional | A user-friendly name for the asset.               |
| Owners             | []metamodel.Owner | Optional | A list of initial owners for the asset.           |

Example:

```go
req := &assetmodel.SuccessfulUploadRequest{
    CloudinaryAssetID:  "a1b2c3d4e5f6",
    CloudinaryPublicID: "folder/my-image",
    ResourceType:       "image",
    Format:             "png",
    Width:              1920,
    Height:             1080,
    URL:                "http://res.cloudinary.com/...",
    SecureURL:          "https://res.cloudinary.com/...",
    Owners: []metamodel.Owner{
        {OwnerType: "product", OwnerID: "c47ac10b-58cc-4372-a567-0e02b2c3d479"},
    },
}
```

### Output

Returns a pointer to the newly created `assetmodel.Asset` struct on success.

### Description

The method:

1. Validates the request payload.
2. Creates a new `Asset` record with a new UUID and the provided details.
3. Saves the new asset record to the relational database (Postgres).
4. If owners are provided, it creates the ownership metadata in the metadata database (ArangoDB).
5. Notifies relevant external services (via gRPC) about the newly created asset and its owners.

### Errors

- `ErrInvalidArgument`: Invalid or missing request fields (HTTP 400).
- Other error (treat as internal): Database or internal error (HTTP 500).

### Example usage

```go
req := &assetmodel.SuccessfulUploadRequest{
    // ... fill in fields
}
asset, err := svc.SuccessfulUpload(ctx, req)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Successfully created asset with ID: %s\n", asset.ID)
```

## CleanupOrphanAssets

The `CleanupOrphanAssets` method finds and deletes assets that exist in Cloudinary but do not have a corresponding record in the local database. This is useful for maintaining consistency and removing unused files.

### Input Parameters

The method accepts a `CleanupOrphanAssetsRequest` struct with the following fields:

| Field  | Type   | Required | Description                                       |
|--------|--------|----------|---------------------------------------------------|
| Folder | string | Required | The Cloudinary folder to scan for orphan assets.  |

Example:

```go
req := &assetmodel.CleanupOrphanAssetsRequest{
    Folder: "products/images",
}
```

### Output

Returns an `int` representing the number of orphan assets that were deleted, and an `error` if the operation fails.

### Description

The method:

1. Validates the request payload.
2. Fetches a list of all assets from the specified folder in Cloudinary.
3. Fetches a list of all Cloudinary asset IDs from the local relational database.
4. Compares the two lists to identify assets that are in Cloudinary but not in the local DB (orphans).
5. If orphans are found, it calls the Cloudinary API to delete them in a batch.
6. Returns the count of deleted assets.

### Errors

- `ErrInvalidArgument`: Invalid or missing request fields (HTTP 400).
- Other error (treat as internal): Cloudinary API or database error (HTTP 500).

### Example usage

```go
req := &assetmodel.CleanupOrphanAssetsRequest{
    Folder: "products/images",
}
count, err := svc.CleanupOrphanAssets(ctx, req)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Cleaned up %d orphan assets.\n", count)
```

## Delete

The `Delete` method performs a soft-delete on an asset. It marks the asset as deleted in the local database but does not remove the actual file from Cloudinary. If the asset has any owners, they are disassociated first.

### Input Parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| assetID   | string | Required | The UUID of the asset to delete. |

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates that the `assetID` is a valid UUID.
2. Starts a database transaction.
3. Retrieves the asset and its ownership metadata.
4. If the asset has owners, it notifies external services to remove the associations and then deletes the ownership metadata from ArangoDB.
5. Soft-deletes the asset record in the relational database (e.g., by setting a `deleted_at` timestamp).
6. Commits the transaction.

+### Errors

- `ErrInvalidArgument`: The `assetID` is not a valid UUID (HTTP 400).
- `ErrNotFound`: The specified asset ID does not exist (HTTP 404).
- Other error (treat as internal): Database or internal error (HTTP 500).

### Example usage

```go
assetID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
err := svc.Delete(ctx, assetID)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Asset soft-deleted successfully.")
```

## DeletePermanent

The `DeletePermanent` method performs a hard, irreversible delete of an asset. It removes the asset record from the local database and deletes the corresponding file from Cloudinary. This should typically be called on assets that have already been soft-deleted and have no owners.

### Input Parameters

The method accepts a `DestroyAssetRequest` struct with the following fields:

| Field        | Type   | Required | Description                                    |
|--------------|--------|----------|------------------------------------------------|
| ID           | string | Required | The UUID of the asset to permanently delete.   |
| ResourceType | string | Required | The resource type in Cloudinary (e.g., "image"). |

Example:

```go
req := &assetmodel.DestroyAssetRequest{
    ID:           "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    ResourceType: "image",
}
```

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates the request payload.
2. Starts a database transaction.
3. Retrieves the asset from the database to get its Cloudinary Public ID.
4. Permanently deletes the asset record from the relational database.
5. Calls the Cloudinary API to delete the asset file from cloud storage.
6. Commits the transaction. If the Cloudinary deletion fails, the transaction is rolled back.

### Errors

- `ErrInvalidArgument`: Invalid or missing request fields (HTTP 400).
- `ErrNotFound`: The specified asset ID does not exist (HTTP 404).
- Other error (treat as internal): Database or Cloudinary API error (HTTP 500).

### Example usage

```go
req := &assetmodel.DestroyAssetRequest{
    ID:           "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    ResourceType: "image",
}
err := svc.DeletePermanent(ctx, req)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Asset permanently deleted successfully.")
```

## Restore

The `Restore` method restores a soft-deleted asset, making it active again.

### Input Parameters

| Parameter | Type   | Required | Description                       |
|-----------|--------|----------|-----------------------------------|
| assetID   | string | Required | The UUID of the asset to restore. |

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates that the `assetID` is a valid UUID.
2. Starts a database transaction.
3. Restores the soft-deleted asset record in the relational database (e.g., by setting `deleted_at` to `NULL`).
4. Commits the transaction.

### Errors

- `ErrInvalidArgument`: The `assetID` is not a valid UUID (HTTP 400).
- `ErrNotFound`: The asset was not found or was not in a soft-deleted state (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
assetID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
err := svc.Restore(ctx, assetID)
if err != nil {
    log.Fatal(err)
}
fmt.Println("Asset restored successfully.")
```
