# MUX Service API

This document describes the API provided by the MUX service, which handles interactions with MUX assets in the microservice. The service defines core business logic for creating, managing, and deleting assets stored in MUX.

For an overview of the service, see [README.md](./README.md). For high-level processes, see [Architecture Documentation](../../architecture/).

## Overview

The MUX service is defined by the `Service` interface and implemented by the `service` struct. It provides methods for interacting with MUX's API, such as creating signed upload URLs, deleting assets, and managing metadata. The service integrates with the relational database and ArangoDB for asset storage and metadata.

Key methods include:

- `Get`: Retrieves a single published and not soft-deleted mux upload record from the database along with it's metadata.
- `GetWithDeleted`: Retrieves a single mux upload record from the database along with it's metadata, including soft-deleted ones.
- `List`: Retrieves a paginated list of all published and not soft-deleted mux upload records along with their metadata.
- `ListDeleted`: Retrieves a paginated list of all soft-deleted mux upload records and their metadata.
- `ListUnowned`: Retrieves a paginated list of all unowned mux upload records and their metadata.
- `Delete`: Performs a soft delete of an asset. It should be called only for assets that don't have any owner or association.
- `DeletePermanent`: Performs a complete delete of a mux upload. It also deletes mux asset via MUX Direct Upload API if `upload.MuxAssetId` is populated.
- `Restore`: Performs a restore of a mux upload record. Mux upload record is not being published. This should be done manually.
- `CreateUploadURL`: Creates upload URL for the direct upload using mux direct upload api. If owner already has an association with the asset, both owner and asset will be deassociated and the new asset instance will be created.
- `CreateUnownedUploadURL`: Creates an upload URL for a new asset without an initial owner.
- `Associate`: Links an existing asset to an owner. It also updates asset metadata.
- `Deassociate`: Removes the link between an asset and an owner. It also deletes owner from asset metadata.
- `UpdateOwners`: Processes asset ownership relations changes. It receives an updated list of asset owners, updates local DB metadata for asset (about it's owners), processes the diff between old and new owners and notifies external services about this ownership changes via gRPC connection.
- `HandleAssetCreatedWebhook`: Processes an incoming Mux webhook with "video.asset.created" event type, finds the corresponding asset, and updates it in a patch-like manner.
- `HandleAssetReadyWebhook`: Processes an incoming Mux webhook with "video.asset.ready" event type, finds the corresponding asset, and updates it in a patch-like manner.

## Get

The `Get` method retrieves a single published and not soft-deleted mux upload record from the database along with it's metadata.

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
4. Retrieves the asset's metadata (title, owners, etc.).
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
fmt.Printf("Asset title: %s\n", asset.Title)
```

## GetWithDeleted

The `GetWithDeleted` method retrieves a single mux upload record from the database along with it's metadata, including soft-deleted ones.

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
4. Retrieves the asset's metadata (title, owners, etc.).
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
fmt.Printf("Asset mux upload id: %s\n", asset.MuxUploadID)
if asset.DeletedAt != nil{
    fmt.Printf("Asset was deleted at: %s\n", asset.DeletedAt.Time.String())
}
```

## List

The `List` method returns a paginated list of all published and not soft-deleted mux upload records along with their metadata.

### Input parameters

|Parameter|Type|Required|Description|
|---------|----|--------|-----------|
|limit    |int | required| Limits the maximum length of the resulting list.|
|offset   |int | required| Sets an offset for the list starting from 0.|

For example, if method was called with `limit = 10` and `offset = 5`, method will return **first 10** asset records **starting from 5th** record **ordered by creation time descending**.

### Output

Returns an `error` if operation fails. On success returns a slice of the assets `[]assetmodel.AssetResponse` and the total number of all published and not soft-deleted assets `int64` in the database.

### Description

The method:

1. Retrieves the list of assets from the database applying `limit` and `offset`.
2. If retrieve operation was unsuccessful, returns an `error`.
3. Counts the total number of published and not soft-deleted assets in the database.
4. If count operation was unsuccessful, returns an `error`.
5. For each asset, retrieves the corresponding metadata.
6. Returns assets and count.

### Errors

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

The `ListUnowned` method returns a paginated list of all unowned mux upload records and their metadata.

### Input parameters

|Parameter|Type|Required|Description|
|---------|----|--------|-----------|
|limit    |int | required| Limits the maximum length of the resulting list.|
|offset   |int | required| Sets an offset for the list starting from 0.|

For example, if method was called with `limit = 10` and `offset = 5`, method will return **first 10** asset records **starting from 5th** record **ordered by creation time descending**.

### Output

Returns an `error` if operation fails. On success returns a slice of the assets `[]assetmodel.AssetResponse` and the total number of all unowned assets `int64` in the database.

### Description

The method:

1. Retrieves the list of unowned asset IDs from the metadata repository.
2. If retrieve operation was unsuccessful, returns an `error`.
3. Fetches assets by the unowned IDs applying `limit` and `offset`.
4. Retrieves the metadata for each asset.
5. Returns assets and total count of unowned assets.

### Errors

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

The `ListDeleted` method returns a paginated list of all soft-deleted mux upload records and their metadata.

### Input parameters

|Parameter|Type|Required|Description|
|---------|----|--------|-----------|
|limit    |int | required| Limits the maximum length of the resulting list.|
|offset   |int | required| Sets an offset for the list starting from 0.|

For example, if method was called with `limit = 10` and `offset = 5`, method will return **first 10** asset records **starting from 5th** record **ordered by creation time descending**.

### Output

Returns an `error` if operation fails. On success returns a slice of the assets `[]assetmodel.AssetResponse` and the total number of all soft-deleted assets `int64` in the database.

### Description

The method:

1. Retrieves the list of soft-deleted assets from the database applying `limit` and `offset`.
2. If retrieve operation was unsuccessful, returns an `error`.
3. Counts the total number of soft-deleted assets in the database.
4. If count operation was unsuccessful, returns an `error`.
5. For each asset, retrieves the corresponding metadata.
6. Returns assets and count.

### Errors

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

## Delete

The `Delete` performs a soft delete of an asset. It should be called only for assets that don't have any owner or association. If asset has any owners, they will be deassociated and local asset metadata about ownership will be deleted.

### Input parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| id        | string | Required | The UUID of the asset to delete. |

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates that the `id` is a valid UUID.
2. Starts a database transaction.
3. Retrieves the asset.
4. If the asset has owners, it notifies external services to remove the associations and then deletes the ownership metadata from ArangoDB.
5. Soft-deletes the asset record in the relational database (e.g., by setting a `deleted_at` timestamp).
6. Commits the transaction.

### Errors

- `ErrInvalidArgument`: `id` is not a valid UUID (HTTP 400).
- `ErrNotFound`: Asset not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
if err := svc.Delete(ctx, id); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset soft-deleted successfully.")
```

## DeletePermanent

The `DeletePermanent` performs a complete delete of a mux upload. It also deletes mux asset via MUX Direct Upload API if `upload.MuxAssetId` is populated.

### Input parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| id        | string | Required | The UUID of the asset to delete. |

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates that the `id` is a valid UUID.
2. Starts a database transaction.
3. Retrieves the asset.
4. If asset has `MuxAssetID` field populated, it deletes the mux asset via MUX API.
5. Completely removes asset metadata from ArangoDB.
6. Deletes the asset permanently from the database.
7. Commits the transaction.

### Errors

- `ErrInvalidArgument`: `id` is not a valid UUID (HTTP 400).
- `ErrNotFound`: Asset not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
if err := svc.DeletePermanent(ctx, id); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset deleted permanently successfully.")
```

## Restore

The `Restore` performs a restore of a mux upload record. Mux upload record is not being published. This should be done manually.

### Input parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| id        | string | Required | The UUID of the asset to restore.|

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates that the `id` is a valid UUID.
2. Starts a database transaction.
3. Restores the asset record.
4. If asset not found, returns `ErrNotFound`.
5. Commits the transaction.

### Errors

- `ErrInvalidArgument`: `id` is not a valid UUID (HTTP 400).
- `ErrNotFound`: Asset not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
id := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
if err := svc.Restore(ctx, id); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset restored successfully.")
```

## CreateUploadURL

The `CreateUploadURL` creates upload URL for the direct upload using mux direct upload api. It uses [muxclient.Client.CreateUploadURL] method to access MUX direct upload API. If owner already has an association with the asset, both owner and asset will be deassociated and the new asset instance will be created.

### Input parameters

The method accepts a `CreateUploadURLRequest` struct with the following fields:

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| OwnerID   | string | Required | The UUID of the asset owner.     |
| OwnerType | string | Required | Type of the asset owner (e.g., "course_part"). |
| Title     | string | Required | Title for the asset.             |
| CreatorID | string | Required | The UUID of the asset creator (user). |

### Output

Returns an `error` if the operation fails. On success returns `*muxgo.UploadResponse` struct with resulting payload.

### Description

The method:

1. Validates that the request payload is valid.
2. Starts a database transaction.
3. Retrieves owner via gRPC call.
4. Validates that owner isn't associated with another asset already.
5. Calls Mux API to generate upload URL.
6. Creates new asset record.
7. Associates newly created asset with owner.
8. Commits the transaction.

### Errors

- `ErrInvalidArgument`: Request payload is invalid (HTTP 400).
- `ErrOwnerHasAsset`: Owner already associated with some asset (HTTP 400).
- `ErrNotFound`: Any record not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
req := &assetmodel.CreateUploadURLRequest{
    OwnerID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerType: "course_part",
    Title: "My new asset",
    CreatorID: "f22accvb-33cc-4372-a567-0e02b2c33sdf9",
}

res, err := svc.CreateUploadURL(ctx, req)
if err != nil{
    log.Fatal(err)
}
fmt.Printf("Upload URL: %s\n", res.Data.Url)
```

## CreateUnownedUploadURL

The `CreateUnownedUploadURL` creates an upload URL for a new asset without an initial owner.

### Input parameters

The method accepts a `CreateUnownedUploadURLRequest` struct with the following fields:

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| Title     | string | Required | Title for the asset.             |
| CreatorID | string | Required | The UUID of the asset creator (user). |

### Output

Returns an `error` if the operation fails. On success returns `*muxgo.UploadResponse` struct with resulting payload.

### Description

The method:

1. Validates that the request payload is valid.
2. Starts a database transaction.
3. Calls Mux API to generate upload URL.
4. Creates new asset record.
5. Creates asset metadata with no owners.
6. Commits the transaction.

### Errors

- `ErrInvalidArgument`: Request payload is invalid (HTTP 400).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
req := &assetmodel.CreateUnownedUploadURLRequest{
    Title: "My new asset",
    CreatorID: "f22accvb-33cc-4372-a567-0e02b2c33sdf9",
}

res, err := svc.CreateUnownedUploadURL(ctx, req)
if err != nil{
    log.Fatal(err)
}
fmt.Printf("Upload URL: %s\n", res.Data.Url)
```

## Associate

The `Associate` links an existing asset to an owner. It also updates asset metadata.

### Input parameters

The method accepts a `AssociateRequest` struct with the following fields:

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| ID        | string | Required | The UUID of the asset to associate.|
| OwnerID   | string | Required | The UUID of the owner to associate asset with.|
| OwnerType | string | Required | Type of the owner (e.g., "course_part").|

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates that the request payload is valid.
2. Starts a database transaction.
3. Retrieves owner via gRPC call.
4. Validates that owner isn't associated with another asset already.
5. Retrieves asset and it's metadata from the databases.
6. Updates asset's metadata with new owner.
7. Saves asset's metadata to the ArangoDB.
8. Notifies another service to associate owner with asset via gRPC.
9. Commits transaction.

### Errors

- `ErrInvalidArgument`: Request payload is invalid (HTTP 400).
- `ErrOwnerHasAsset`: Owner already associated with some asset (HTTP 400).
- `ErrNotFound`: Any record not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
req := &assetmodel.AssociateRequest{
    ID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerID: "f22accvb-33cc-4372-a567-0e02b2c33sdf9",
    OwnerType: "course_part",
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
| OwnerType | string | Required | Type of the owner (e.g., "course_part").|

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates that the request payload is valid.
2. Ensures asset exists in the database.
3. Retrieves asset metadata from ArangoDB.
4. Removes the specified owner from the list of owners.
5. Updates asset's metadata in ArangoDB to remove the owner.
6. Notifies external services to remove the association via gRPC.

### Errors

- `ErrInvalidArgument`: Request payload is invalid (HTTP 400).
- `ErrNotFound`: Any record not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
req := &assetmodel.DeassociateRequest{
    ID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerID: "f22accvb-33cc-4372-a567-0e02b2c33sdf9",
    OwnerType: "course_part",
}

if err := svc.Deassociate(ctx, req); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset deassociated from owner successfully")
```

## UpdateOwners

The `UpdateOwners` processes asset ownership relations changes. It receives an updated list of asset owners, updates local DB metadata for asset (about it's owners), processes the diff between old and new owners and notifies external services about this ownership changes via gRPC connection.

### Input parameters

The method accepts a `UpdateOwnersRequest` struct with the following fields:

| Parameter | Type              | Required | Description                      |
|-----------|-------------------|----------|----------------------------------|
| ID        | string            | Required | The UUID of the asset to update. |
| Owners    | []metamodel.Owner | Required | The new list of owners for the asset. |

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Validates that the request payload is valid.
2. Ensures asset exists in Postgres before updating metadata in ArangoDB.
3. Retrieves current asset metadata from ArangoDB.
4. Calculates the difference between current and new owners (what to add and what to delete).
5. Updates asset's metadata in ArangoDB with the new list of owners.
6. Notifies external services about ownership changes via gRPC.

### Errors

- `ErrInvalidArgument`: Request payload is invalid (HTTP 400).
- `ErrNotFound`: Asset not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
owners := []metamodel.Owner{
    {OwnerID: "f47ac10b-58cc-4372-a567-0e02b2c3d479", OwnerType: "course_part"},
    {OwnerID: "f22accvb-33cc-4372-a567-0e02b2c33sdf9", OwnerType: "lesson"},
}

req := &assetmodel.UpdateOwnersRequest{
    ID: "e47ac10b-58cc-4372-a567-0e02b2c3d480",
    Owners: owners,
}

if err := svc.UpdateOwners(ctx, req); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset owners updated successfully")
```

## HandleAssetCreatedWebhook

The `HandleAssetCreatedWebhook` processes an incoming Mux webhook with "video.asset.created" event type, finds the corresponding asset, and updates it in a patch-like manner.

### Input parameters

| Parameter | Type                      | Required | Description                      |
|-----------|---------------------------|----------|----------------------------------|
| payload   | *assetmodel.MuxWebhook   | Required | The MUX webhook payload to process. |

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Starts a database transaction.
2. Looks for the asset by UploadID if present in the payload, otherwise by AssetID.
3. If no asset found, returns `ErrNotFound`.
4. Builds updates map by comparing the existing asset with the webhook data.
5. Updates the asset in the database with the changes.
6. Commits the transaction.

### Errors

- `ErrNotFound`: Asset not found for the given upload_id or asset_id (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
payload := &assetmodel.MuxWebhook{
    // Webhook payload from MUX
}

if err := svc.HandleAssetCreatedWebhook(ctx, payload); err != nil{
    log.Fatal(err)
}
fmt.Println("Webhook processed successfully")
```

## HandleAssetReadyWebhook

The `HandleAssetReadyWebhook` processes an incoming Mux webhook with "video.asset.ready" event type, finds the corresponding asset, and updates it in a patch-like manner.

### Input parameters

| Parameter | Type                      | Required | Description                      |
|-----------|---------------------------|----------|----------------------------------|
| payload   | *assetmodel.MuxWebhook   | Required | The MUX webhook payload to process. |

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

### Description

The method:

1. Starts a database transaction.
2. Looks for the asset by UploadID if present in the payload, otherwise by AssetID.
3. If no asset found, returns `ErrNotFound`.
4. Builds updates map by comparing the existing asset with the webhook data.
5. Updates the asset in the database with the changes.
6. Commits the transaction.

### Errors

- `ErrNotFound`: Asset not found for the given upload_id or asset_id (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
payload := &assetmodel.MuxWebhook{
    // Webhook payload from MUX
}

if err := svc.HandleAssetReadyWebhook(ctx, payload); err != nil{
    log.Fatal(err)
}
fmt.Println("Webhook processed successfully")
```