# MUX Service API

This document describes the API provided by the MUX service, which handles interactions with MUX assets in the microservice. The service defines core business logic for creating, managing, and deleting assets stored in MUX.

For an overview of the service, see [README.md](./README.md). For high-level processes, see [Architecture Documentation](../../architecture/).

## Overview

The MUX service is defined by the `Service` interface and implemented by the `service` struct. It provides methods for interacting with MUX's API, such as creating signed upload URLs, deleting assets, and managing metadata. The service integrates with the relational database and ArangoDB for asset storage and metadata.

Key methods include:

- `Get`: Retrieves a single not soft-deleted asset and it's metadata from the database (see [Get Operation Flow](./flow/get_operation_flow.md)).
- `GetWithDeleted`: Retrieves a single asset and it's metadata from the database including soft-deleted ones (see [Get Operation Flow](./flow/get_operation_flow.md)).
- `List`: Retrieves a paginated list of all not soft-deleted assets from the database and their metadata (see [List Operation Flow](./flow/list_operation_flow.md)).
- `ListDeleted`: Retrieves a paginated list of all soft-deleted assets from the database and their metadata (see [List Operation Flow](./flow/list_operation_flow.md)).
- `ListUnowned`: Retrieves a paginated list of all unowned assets from the database and their metadata (see [List Operation Flow](./flow/list_operation_flow.md)).
- `Delete`: Performs soft-delete of an asset. (see [Delete Process Flow](./flow/delete_flow.md)).
- `DeletePermanent`: Performs a complete delete of a mux upload, also deletes mux asset via MUX Direct Upload API if upload.MuxAssetId is populated (see [Delete Permanent Process Flow](./flow/delete_permanent_flow.md)).
- `Restore`: Restores soft-deleted asset (see [Restore Process Flow](./flow/restore_flow.md)).
- `CreateUploadURL`: Creates signed upload url for direct upload to the MUX (see [Crate Upload URL Process Flow](./flow/create_upload_url_flow.md)).
- `Associate`: Associates an asset with a single owner.
- `CreateUnownedUploadURL`: Creates a MUX direct upload URL for an asset without an initial owner.
- `Deassociate`: Deassociates an asset from a single owner.
- `UpdateOwners`: Updates the list of owners for a given asset.
- `HandleAssetCreatedWebhook`: Processes an incoming Mux webhook with "video.asset.created" event type.
- `HandleAssetReadyWebhook`: Processes an incoming Mux webhook with "video.asset.ready" event type.

## Get

The `Get` method retrieves a single not soft-deleted asset from the database.

### Input parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| assetID   | string | Required | The UUID of the asset to get.    |

### Output

Returns an `error` if operation fails. Returns `*assetmodel.AssetResponse` struct on success, which contains the asset data and its metadata.

### Description

The method:

1. Validates that the `assetID` is a valid UUID.
2. Retrieves the asset from the database.
3. If no asset found, returns `ErrNotFound`.
4. Retrieves the asset's metadata (title, owners, etc.).
5. Returns a combined `AssetResponse`.

### Errors

- `ErrInvalidArgument`: `assetID` is not a valid UUID (HTTP 400).
- `ErrNotFound`: Asset not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
assetID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
asset, err := svc.Get(ctx, assetID)
if err != nil{
    log.Fatal(err)
}
fmt.Printf("Asset title: %s\n", asset.Title)
```

## GetWithDeleted

The `GetWithDeleted` method retrieves a single asset from the database including soft-deleted. It can be used when client doesn't care about potential deleted asset data fetch, or intentional fetch of soft-deleted asset data.

### Input parameters

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| assetID   | string | Required | The UUID of the asset to get.    |

### Output

Returns an `error` if operation fails. Returns `*Asset` struct on success.

### Description

The method:

1. Validates that the `assetID` is a valid UUID.
2. Retrieves the asset from the database.
3. If no asset found, returns `ErrNotFound`.
4. Returns asset.

### Errors

- `ErrInvalidArgument`: `assetID` is not a valid UUID (HTTP 400).
- `ErrNotFound`: Asset not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
assetID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
asset, err := svc.GetWithDeleted(ctx, assetID)
if err != nil{
    log.Fatal(err)
}
fmt.Printf("Asset mux uplaod id: %s\n", asset.MuxUploadID)
if asset.DeletedAt != nil{
    fmt.Printf("Asset was deleted at: %s\n", asset.DeletedAt.Time.String())
}
```

## List

The `List` method returns a paginated list of all not soft-deleted assets in the database along with the total number of such records in the database.

### Input parameters

|Parameter|Type|Required|Description|
|---------|----|--------|-----------|
|limit    |int | required| Limits the maximum length of the resulting list.|
|offset|int|required| Sets an offset for the list starting from 0.|

For example, if method was called with `limit = 10` and `offst = 5`, method will return **first 10** asset records **starting from 5th** record **ordered by creation time descending**.

### Output

Returns an `error` if operation fails. On success returns a slice of the assets `[]assetmodel.Asset` and the total number of all not soft-deleted assets `int64` in the database.


### Description

The method:

1. Retrieves the list of assets from the database applying `limit` and `offset`.
3. If retrieve operation was unsuccessful, returns an `errror`.
3. Counts the total number of not soft-deleted assets in the database.
4. If count operation was unsuccessful, returns an `error`.
5. Returns assets and count.

### Errors

- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
limit, offset := 10, 5
assets, err := svc.List(ctx, limit, offset)
if err != nil{
    log.Fatal(err)
}
fmt.Printf("List returned %d assets\n", len(assets))
```

## ListDeleted

The `ListDeleted` method returns a paginated list of all soft-deleted assets in the database along with the total number of such records in the database.

### Input parameters

|Parameter|Type|Required|Description|
|---------|----|--------|-----------|
|limit    |int | required| Limits the maximum length of the resulting list.|
|offset|int|required| Sets an offset for the list starting from 0.|

For example, if method was called with `limit = 10` and `offst = 5`, method will return **first 10** asset records **starting from 5th** record **ordered by creation time descending**.

### Output

Returns an `error` if operation fails. On success returns a slice of the assets `[]assetmodel.Asset` and the total number of all soft-deleted assets `int64` in the database.


### Description

The method:

1. Retrieves the list of assets from the database applying `limit` and `offset`.
3. If retrieve operation was unsuccessful, returns an `errror`.
3. Counts the total number of soft-deleted assets in the database.
4. If count operation was unsuccessful, returns an `error`.
5. Returns assets and count.

### Errors

- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
limit, offset := 10, 5
assets, err := svc.ListDeleted(ctx, limit, offset)
if err != nil{
    log.Fatal(err)
}
fmt.Println("List returned %d assets\n", len(assets))
```

## ListUnowned

The `ListUnowned` method returns a paginated list of unowned assets in the database along with the total number of such records in the database. Asset treats as **Unowned** *if it doesn't have any owners*. Deleted assets cannot be included in the result. 

### Input parameters

|Parameter|Type|Required|Description|
|---------|----|--------|-----------|
|limit    |int | required| Limits the maximum length of the resulting list.|
|offset|int|required| Sets an offset for the list starting from 0.|

For example, if method was called with `limit = 10` and `offst = 5`, method will return **first 10** asset records **starting from 5th** record **ordered by creation time descending**.

### Output

Returns an `error` if operation fails. On success returns a slice of the assets `[]assetmodel.Asset` and the total number of unowned assets `int64` in the database.


### Description

The method:

1. Retrieves the list of assets from the database applying `limit` and `offset`.
3. If retrieve operation was unsuccessful, returns an `errror`.
3. Counts the total number of unowned assets in the database.
4. If count operation was unsuccessful, returns an `error`.
5. Returns assets and count.

### Errors

- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
limit, offset := 10, 5
assets, err := svc.ListUnowned(ctx, limit, offset)
if err != nil{
    log.Fatal(err)
}
fmt.Println("List returned %d assets\n", len(assets))
```

## Delete

The `Delete` method performs soft-delete of the asset. If asset has any owners, they will be deassociated and all asset metadata about owners will be cleared.

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
3. Retrieves the asset and its ownership metadata.
4. If the asset has owners, it notifies external services to remove the associations and then deletes the ownership metadata from ArangoDB.
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

The `DeletePermanent` permanently deletes an asset and completely clears it's metadata, also deleting asset from MUX. It should be called only for soft-deleted assets.

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
4. If asset has `MuxUploadID` field populated, it deletes the mux upload.
5. Completely removes asset metadata from ArangoDB.
6. Deletes the asset itself.
7. Commits the transaction.

### Errors

- `ErrInvalidArgument`: `assetID` is not a valid UUID (HTTP 400).
- `ErrNotFound`: Asset not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
assetID := "f47ac10b-58cc-4372-a567-0e02b2c3d479"
if err := svc.DeletePermanent(ctx, assetID); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset deleted successfully.")
```

## Restore

The `Restore` restores a soft-deleted aset.

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

## CreateUploadURL

The `CreateUploadURL` creates a signed upload url for direct asset upload to the MUX. It returns complete generated upload url.

### Input parameters

The method accepts a `CreateUploadURLRequest` struct with the following fields:

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| OwnerID   | string | Required | The UUID of the asset owner.     |
| OwnerType   | string | Required | Type of the asset owner (e.g., "course_part").     |
| Title   | string | Required | Title for the asset.     |
| CreatorID   | string | Required | The UUID of the asset creator (user).     |

### Output

Returns an `nil` and `error` if the operation fails. On success returns `*muxgo.UploadResponse` struct with resulting payload and `nil` as an error.

### Description

The method:

1. Validates that the request payload is valid.
2. Starts a database transaction.
3. Retrieves owner via gRPC call.
4. Validates that owner isn't associated with another asset already.
5. Calls Mux API to generate upload url.
6. Creates new asset record.
7. Associates newly created asset with owner.
8. Commits the transaction.

### Errors

- `ErrInvalidArgument`: Request payload is invalid (HTTP 400).
- `ErrOwnerHasAsset`: Owner already associated with some asset (HTTP 400).
- `ErrNotFound`: Any record not found in the database (HTTP 404).
- `ErrMuxAPI`: Mux API error (HTTP 503).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
req := &CreateUploadURLRequest{
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

## Associate

The `Associate` method associates an asset with a single owner.

### Input parameters

The method accepts a `AssociateRequest` struct with the following fields:

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| ID        | string | Required | The UUID of the asset to associate.|
| OwnerID   | string | Required | The UUID of the owner to associate asset with.|
| OwnerType | string | Required | Type of the owner (e.g., "course_part").|

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

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
req := &AssociateRequest{
    ID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerID: "f22accvb-33cc-4372-a567-0e02b2c33sdf9",
    OwnerType: "course_part",
}

if err := svc.Associate(ctx, req); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset associated with new owner successfuly")
```

## Deassociate

The `Deassociate` method deassociated an asset from a single owner.

### Input parameters

The method accepts a `DeassociateRequest` struct with the following fields:

| Parameter | Type   | Required | Description                      |
|-----------|--------|----------|----------------------------------|
| ID        | string | Required | The UUID of the asset to deassociate.|
| OwnerID   | string | Required | The UUID of the owner to deassociate asset from.|
| OwnerType | string | Required | Type of the owner (e.g., "course_part").|

### Output

Returns an `error` if the operation fails. A `nil` error indicates success.

1. Validates that the request payload is valid.
2. Starts a database transaction.
3. Retrieves asset and it's metadata from the databases.
4. Removes owner from asset metadata.
5. Saves updated asset's metadata to the ArangoDB.
6. Notifies another service to associate owner with asset via gRPC.
7. Commits transaction.

### Errors

- `ErrInvalidArgument`: Request payload is invalid (HTTP 400).
- `ErrNotFound`: Any record not found in the database (HTTP 404).
- Other error (treat as internal): Database error (HTTP 500).

### Example usage

```go
req := &DeassociateRequest{
    ID: "f47ac10b-58cc-4372-a567-0e02b2c3d479",
    OwnerID: "f22accvb-33cc-4372-a567-0e02b2c33sdf9",
    OwnerType: "course_part",
}

if err := svc.Deassociate(ctx, req); err != nil{
    log.Fatal(err)
}
fmt.Println("Asset deassociated from owner successfuly")
```
