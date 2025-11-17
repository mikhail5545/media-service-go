# Cloudinary Service API

This document describes the API provided by the Cloudinary service, which handles interactions with Cloudinary assets in the microservice. The service defines core business logic for creating, managing, and deleting assets stored in Cloudinary.

For an overview of the service, see [README.md](./README.md). For high-level processes, see [Architecture Documentation](../../architecture/).

## Overview

The Cloudinary service is defined by the `Service` interface and implemented by the `service` struct. It provides methods for interacting with Cloudinary's API, such as creating signed upload URLs, deleting assets, and managing metadata. The service integrates with the relational database and ArangoDB for asset storage and metadata.

Key methods include:

- `Get`: Retrieves a single not soft-deleted asset record from the database along with it's metadata.
- `GetWithDeleted`: Retrieves a single asset record from the database along with it's metadata, including soft-deleted ones.
- `List`: Retrieves a paginated list of all not soft-deleted asset records along with their metadata.
- `ListUnowned`: Retrieves a paginated list of all unowned asset records along with their metadata.
- `ListDeleted`: Retrieves a paginated list of all soft-deleted asset records along with their metadata.
- `CreateSignedUploadURL`: Creates a signature for a direct frontend upload. Direct upload url should be constructed using this params, this function only creates signature for signed upload.
- `UpdateOwners`: Processes asset ownership relations changes. It recieves an updated list of asset owners, updates local DB metadata for asset (about it's owners), processes the diff between old and new owners and notifies external services about this ownership changes via gRPC connection.
- `Associate`: Links an existing asset to an owner. It also updates asset medatada.
- `Deassociate`: Removes the link between an asset and an owner. It also deletes owner from asset metadata.
- `SuccessfulUpload`: Creates a new asset with provided information and creates owner relations for it. It saves asset metadata about owner relations in the local noSQL db and notifies external services about ownership changes via gRPC connection. This method should be called after successful cloudinary image upload.
- `CleanupOrphanAssets`: Finds and deletes assets that exist in Cloudinary but not in the local database.
- `Delete`: Performs a soft-delete of an asset. It does not delete Cloudinary asset. If assset has owners, it will be deassociated from them first.
- `DeletePermanent`: Performs a complete delete of an asset. It also deletes Cloudinary asset. By this time, asset shouldn't have any owners. They should be deleted when asset is being soft-deleted. This action is irreversable.
- `Restore`: Performs a restore of an asset.

## Get

The `Get` method retrieves a single not soft-deleted asset record from the database along with it's metadata.

### Input parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| id        | string | Required | The UUID of the asset to get.    |

### Output

Returns an `error` if operation fails. Returns `*assetmodel.AssetResponse` struct on success, which contains the asset data and its metadata.

### Description

The method:

1. Validates that the `id` is a valid UUID.
2. Retrieves the asset from the database.
3. If no asset found, returns `ErrNotFound`.
4. Retrieves the asset's metadata (owners, etc.).
5. Returns a combined `AssetResponse`.

### Errors

- `ErrInvalidArgument`: `id` is not a valid UUID (HTTP 400).
- `ErrNotFound`: Asset not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
asset, err := svc.Get(ctx, id)
if err != nil{
    log.Fatal(err)
}
fmt.Printf("Asset URL: %s\n", asset.URL)
```

## GetWithDeleted

The `GetWithDeleted` method retrieves a single asset record from the database along with it's metadata, including soft-deleted ones.

### Input parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| id        | string | Required | The UUID of the asset to get.    |

### Output

Returns an `error` if operation fails. Returns `*assetmodel.AssetResponse` struct on success, which contains the asset data and its metadata.

### Description

The method:

1. Validates that the `id` is a valid UUID.
2. Retrieves the asset from the database including soft-deleted ones.
3. If no asset found, returns `ErrNotFound`.
4. Retrieves the asset's metadata (owners, etc.).
5. Returns a combined `AssetResponse`.

### Errors

- `ErrInvalidArgument`: `id` is not a valid UUID (HTTP 400).
- `ErrNotFound`: Asset not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
asset, err := svc.GetWithDeleted(ctx, id)
if err != nil{
    log.Fatal(err)
}
fmt.Printf("Asset public ID: %s\n", asset.CloudinaryPublicID)
if asset.DeletedAt != nil{
    fmt.Printf("Asset was deleted at: %s\n", asset.DeletedAt.Time.String())
}
```

## List

The `List` method retrieves a paginated list of all not soft-deleted asset records along with their metadata.

### Input parameters

|Parameter|Type|Required|Description|
|---------|----|--------|-----------|
|limit    |int | required| Limits the maximum length of the resulting list. Use -1 for no limit.|
|offset   |int | required| Sets an offset for the list starting from 0.|

For example, if method was called with `limit = 10` and `offset = 5`, method will return **first 10** asset records **starting from 5th** record **ordered by creation time descending**.

### Output

Returns an `error` if operation fails. On success returns a slice of the assets `[]assetmodel.AssetResponse` and the total number of all not soft-deleted assets `int64` in the database.

### Description

The method:

1. Validates the limit and offset parameters.
2. Retrieves the list of assets from the database applying `limit` and `offset`.
3. If retrieve operation was unsuccessful, returns an `error`.
4. Counts the total number of not soft-deleted assets in the database.
5. If count operation was unsuccessful, returns an `error`.
6. For each asset, retrieves the corresponding metadata.
7. Returns assets and count.

### Errors

- `ErrInvalidArgument`: Invalid limit or offset values (HTTP 400).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
limit, offset := 10, 5
assets, total, err := svc.List(ctx, limit, offset)
if err != nil{
    log.Fatal(err)
}
fmt.Printf("List returned %d assets (total: %d)\n", len(assets), total)
```

## ListUnowned

The `ListUnowned` method retrieves a paginated list of all unowned asset records along with their metadata.

### Input parameters

|Parameter|Type|Required|Description|
|---------|----|--------|-----------|
|limit    |int | required| Limits the maximum length of the resulting list. Use -1 for no limit.|
|offset   |int | required| Sets an offset for the list starting from 0.|

For example, if method was called with `limit = 10` and `offset = 5`, method will return **first 10** asset records **starting from 5th** record **ordered by creation time descending**.

### Output

Returns an `error` if operation fails. On success returns a slice of the assets `[]assetmodel.AssetResponse` and the total number of all unowned assets `int64` in the database.

### Description

The method:

1. Validates the limit and offset parameters.
2. Retrieves the list of unowned asset IDs from the metadata repository.
3. If retrieve operation was unsuccessful, returns an `error`.
4. Fetches assets by the unowned IDs applying `limit` and `offset`.
5. Retrieves the metadata for each asset.
6. Returns assets and total count of unowned assets.

### Errors

- `ErrInvalidArgument`: Invalid limit or offset values (HTTP 400).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
limit, offset := 10, 5
assets, total, err := svc.ListUnowned(ctx, limit, offset)
if err != nil{
    log.Fatal(err)
}
fmt.Printf("List returned %d unowned assets (total: %d)\n", len(assets), total)
```

## ListDeleted

The `ListDeleted` method retrieves a paginated list of all soft-deleted asset records along with their metadata.

### Input parameters

|Parameter|Type|Required|Description|
|---------|----|--------|-----------|
|limit    |int | required| Limits the maximum length of the resulting list. Use -1 for no limit.|
|offset   |int | required| Sets an offset for the list starting from 0.|

For example, if method was called with `limit = 10` and `offset = 5`, method will return **first 10** asset records **starting from 5th** record **ordered by creation time descending**.

### Output

Returns an `error` if operation fails. On success returns a slice of the assets `[]assetmodel.AssetResponse` and the total number of all soft-deleted assets `int64` in the database.

### Description

The method:

1. Validates the limit and offset parameters.
2. Retrieves the list of soft-deleted assets from the database applying `limit` and `offset`.
3. If retrieve operation was unsuccessful, returns an `error`.
4. Counts the total number of soft-deleted assets in the database.
5. If count operation was unsuccessful, returns an `error`.
6. For each asset, retrieves the corresponding metadata.
7. Returns assets and count.

### Errors

- `ErrInvalidArgument`: Invalid limit or offset values (HTTP 400).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
limit, offset := 10, 5
assets, total, err := svc.ListDeleted(ctx, limit, offset)
if err != nil{
    log.Fatal(err)
}
fmt.Printf("List returned %d deleted assets (total: %d)\n", len(assets), total)
```

## CreateSignedUploadURL

The `CreateSignedUploadURL` creates a signature for a direct frontend upload. Direct upload url should be constructed using this params, this function only creates signature for signed upload.

### Input parameters

The method accepts a `CreateSignedUploadURLRequest` struct with the following fields:

| Field     | Type   | Required | Description                                      |
|-----------|--------|----------|--------------------------------------------------|
| Eager     | *string | Optional | Transformations to apply (e.g., "c_thumb,w_200"). |
| PublicID  | string | Required | Unique identifier for the asset in Cloudinary.   |

Example:

```go
e := "c_thumb,w_200"

req := &assetmodel.CreateSignedUploadURLRequest{
    Eager:    &e,
    PublicID: "asset_123",
}
```

### Output

Returns an `error` if operation fails. On success returns `map[string]string` containing parameters for constructing a signed upload URL. The URL must be assembled manually on the client side (e.g., frontend).

|Key|Type|Description|
|---|----|-----------|
|signature|string|Generated signature for the URL|
|public_id|string|Cloudinary public ID.|
|timestamp|string|Unix timestamp for the request.|
|api_key|string|Cloudinary API key.|

Example:

```json
{
    "signature": "dbad3qa8fh3gbf",
    "public_id": "asset/233",
    "timestamp": "1634567890",
    "api_key": "cloudinary_api_key",
}
```

### Description

The method:

1. Validates the `CreateSignedUploadURLRequest` payload.
2. Creates a timestamp and URL parameters for signing.
3. Calls the Cloudinary API client's `SignUploadParams` method to generate signature.
4. Constructs the response map with the signature, public ID, timestamp, and API key.
5. Returns the map for use in constructing the upload URL.

The client must combine these parameters with Cloudinary's base URL to form the final upload URL. See [Cloudinary documentation](https://cloudinary.com/documentation/upload_images#signed_uploads) for more details.

### Errors

The method returns the following errors:

- `ErrInvalidArgument`: Invalid or missing request fields (HTTP 400).
- Other error (treat as internal): Database or internal error (HTTP 500).

### Example usage

```go
req := &assetmodel.CreateSignedUploadURLRequest{
    PublicID: "asset/111",
}
resp, err := svc.CreateSignedUploadURL(ctx, req)
if err != nil{
    log.Fatal(err)
}
fmt.Printf("Signature: %s\n", resp["signature"])
```

## UpdateOwners

The `UpdateOwners` processes asset ownership relations changes. It recieves an updated list of asset owners, updates local DB metadata for asset (about it's owners), processes the diff between old and new owners and notifies external services about this ownership changes via gRPC connection.

### Input Parameters

The method accepts an `UpdateOwnersRequest` struct with the following fields:

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

## Associate

The `Associate` links an existing asset to an owner. It also updates asset medatada.

### Input parameters

The method accepts an `AssociateRequest` struct with the following fields:

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| ID        | string | Required | The UUID of the asset to associate.|
| OwnerID   | string | Required | The UUID of the owner to associate asset with.|
| OwnerType | string | Required | Type of the owner (e.g., "product").|

Example:

```go
req := &assetmodel.AssociateRequest{
    ID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerID: "c47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerType: "product",
}
```

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates that the request payload is valid.
2. Starts a database transaction.
3. Retrieves asset from the database.
4. Retrieves asset metadata from ArangoDB.
5. Updates the asset's metadata with the new owner by appending to the existing owners list.
6. Saves the updated metadata to ArangoDB.
7. Notifies external services to associate the owner with the asset via gRPC.
8. Commits the transaction.

### Errors

- `ErrInvalidArgument`: Request payload is invalid (HTTP 400).
- `ErrNotFound`: Any record not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
req := &assetmodel.AssociateRequest{
    ID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerID: "c47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerType: "product",
}

if err := svc.Associate(ctx, req); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset associated with new owner successfully")
```

## Deassociate

The `Deassociate` removes the link between an asset and an owner. It also deletes owner from asset metadata.

### Input parameters

The method accepts a `DeassociateRequest` struct with the following fields:

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| ID        | string | Required | The UUID of the asset to deassociate.|
| OwnerID   | string | Required | The UUID of the owner to deassociate asset from.|
| OwnerType | string | Required | Type of the owner (e.g., "product").|

Example:

```go
req := &assetmodel.DeassociateRequest{
    ID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerID: "c47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerType: "product",
}
```

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates that the request payload is valid.
2. Starts a database transaction.
3. Ensures asset exists in the database.
4. Retrieves asset metadata from ArangoDB.
5. Removes the specified owner from the list of owners.
6. Updates asset's metadata in ArangoDB to remove the owner.
7. Notifies external services to remove the association via gRPC.
8. Commits the transaction.

### Errors

- `ErrInvalidArgument`: Request payload is invalid (HTTP 400).
- `ErrNotFound`: Any record not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
req := &assetmodel.DeassociateRequest{
    ID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerID: "c47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerType: "product",
}

if err := svc.Deassociate(ctx, req); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset deassociated from owner successfully")
```

## SuccessfulUpload

The `SuccessfulUpload` creates a new asset with provided information and creates owner relations for it. It saves asset metadata about owner relations in the local noSQL db and notifies external services about ownership changes via gRPC connection. This method should be called after successful cloudinary image upload.

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
    URL:                "http://example.com/image.png",
    SecureURL:          "https://example.com/image.png",
    AssetFolder:        "assets",
    DisplayName:        "My Image",
    Owners: []metamodel.Owner{
        {OwnerType: "product", OwnerID: "c47ac10b-58cc-4372-a567-0e02b2c3d479"},
    },
}
```

### Output

Returns an `error` if the operation fails. On success returns `*assetmodel.Asset` with the newly created asset.

### Description

The method:

1. Validates the request payload.
2. Creates a new asset record in the database with the provided information.
3. If the asset has owners, updates the metadata in ArangoDB with the owner information.
4. Notifies external services about the ownership changes via gRPC.
5. Returns the newly created asset.

### Errors

- `ErrInvalidArgument`: Invalid or missing request fields (HTTP 400).
- Other error (treat as internal): Database or internal error (HTTP 500).

### Example usage

```go
req := &assetmodel.SuccessfulUploadRequest{
    CloudinaryAssetID:  "a1b2c3d4e5f6",
    CloudinaryPublicID: "folder/my-image",
    ResourceType:       "image",
    Format:             "png",
    Width:              1920,
    Height:             1080,
    URL:                "http://example.com/image.png",
    SecureURL:          "https://example.com/image.png",
}

asset, err := svc.SuccessfulUpload(ctx, req)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("New asset created with ID: %s\n", asset.ID)
```

## CleanupOrphanAssets

The `CleanupOrphanAssets` finds and deletes assets that exist in Cloudinary but not in the local database.

### Input Parameters

The method accepts a `CleanupOrphanAssetsRequest` struct with the following fields:

| Field  | Type   | Required | Description                               |
|--------|--------|----------|-------------------------------------------|
| Folder | string | Required | The Cloudinary folder to check for orphans.|

Example:

```go
req := &assetmodel.CleanupOrphanAssetsRequest{
    Folder: "my-assets",
}
```

### Output

Returns an `error` if the operation fails. On success returns the number of cleaned assets `int`.

### Description

The method:

1. Validates the request payload.
2. Lists all assets in Cloudinary for the specified folder.
3. Retrieves all asset IDs from the local database.
4. Compares Cloudinary assets to local database assets to identify orphans.
5. Deletes orphan assets from Cloudinary.
6. Returns the number of assets that were cleaned up.

### Errors

- `ErrInvalidArgument`: Invalid or missing request fields (HTTP 400).
- Other error (treat as internal): Database or internal error (HTTP 500).

### Example usage

```go
req := &assetmodel.CleanupOrphanAssetsRequest{
    Folder: "my-assets",
}

cleaned, err := svc.CleanupOrphanAssets(ctx, req)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Cleaned up %d orphan assets\n", cleaned)
```

## Delete

The `Delete` performs a soft-delete of an asset. It does not delete Cloudinary asset. If assset has owners, it will be deassociated from them first.

### Input parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| assetID   | string | Required | The UUID of the asset to delete. |

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates that the `assetID` is a valid UUID.
2. Starts a database transaction.
3. Retrieves the asset.
4. If the asset has owners, it notifies external services to remove the associations and then deletes the ownership metadata from the metadata database.
5. Soft-deletes the asset record in the relational database (e.g., by setting a `deleted_at` timestamp).
6. Commits the transaction.

### Errors

- `ErrInvalidArgument`: `assetID` is not a valid UUID (HTTP 400).
- `ErrNotFound`: Asset not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
assetID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
if err := svc.Delete(ctx, assetID); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset soft-deleted successfully.")
```

## DeletePermanent

The `DeletePermanent` performs a complete delete of an asset. It also deletes Cloudinary asset. By this time, asset shouldn't have any owners. They should be deleted when asset is being soft-deleted. This action is irreversable.

### Input parameters

The method accepts a `DestroyAssetRequest` struct with the following fields:

| Field        | Type   | Required | Description                      |
|--------------|--------|----------|----------------------------------|
| ID           | string | Required | The UUID of the asset to delete. |
| ResourceType | string | Required | The type of resource to delete from Cloudinary. |

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates the request payload.
2. Starts a database transaction.
3. Retrieves the asset.
4. Deletes the asset permanently from the database.
5. Deletes the asset from Cloudinary.
6. Commits the transaction.

### Errors

- `ErrInvalidArgument`: Request payload is invalid (HTTP 400).
- `ErrNotFound`: Asset not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
req := &assetmodel.DestroyAssetRequest{
    ID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    ResourceType: "image",
}

if err := svc.DeletePermanent(ctx, req); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset deleted permanently successfully.")
```

## Restore

The `Restore` performs a restore of an asset.

### Input parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| assetID   | string | Required | The UUID of the asset to restore.|

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates that the `assetID` is a valid UUID.
2. Starts a database transaction.
3. Restores the asset record.
4. If asset not found, returns `ErrNotFound`.
5. Commits the transaction.

### Errors

- `ErrInvalidArgument`: `assetID` is not a valid UUID (HTTP 400).
- `ErrNotFound`: Asset not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
assetID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
if err := svc.Restore(ctx, assetID); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset restored successfully.")
```