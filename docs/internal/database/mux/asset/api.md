# Mux Asset Repository API

This document describes the API provided by the Mux Asset repository, which handles Mux asset data operations in PostgreSQL. The repository defines core data access functionality for storing and retrieving asset records in PostgreSQL using GORM.

For an overview of the repository, see [README.md](./README.md). For high-level processes, see [Architecture Documentation](../../../architecture/).

## Overview

The Mux Asset repository is defined by the `Repository` interface and implemented by the `gormRepository` struct. It provides methods for interacting with Mux asset records stored in PostgreSQL, including operations for creating, reading, updating, and deleting asset records with both soft-delete and permanent delete capabilities, as well as owner association management.

Key methods include:

- `Get`: Retrieves a single asset record from the database.
- `GetByUploadID`: Retrieves a single asset record from the database by it's MuxUploadID.
- `GetByAssetID`: Retrieves a single asset record from the database by it's MuxAssetID.
- `List`: Retrieves all asset records from the database.
- `ListByIDs`: Retrieves a paginated list of asset records from the database by their IDs.
- `Count`: Counts the total number of asset records in the database.
- `GetWithDeleted`: Retrieves a single asset record from the database including soft-deleted ones.
- `ListDeleted`: Retrieves all soft-deleted asset records from the database.
- `CountDeleted`: Counts the total number of soft-deleted asset records in the database.
- `Create`: Creates a new asset record in the database.
- `RemoveOwner`: Removes local asset association with the owner by setting it's `owner_id` and `owner_type` to nil.
- `SetOwner`: Sets local asset association with the owner by setting it's `owner_id` and `owner_type`.
- `Update`: Performs partial update of asset record in the database using updates.
- `Delete`: Performs soft-delete of asset record.
- `DeletePermanent`: Performs permanent delete of asset record.
- `Restore`: Restores soft-deleted asset record.
- `DB`: Returns the underlying gorm.DB instance.
- `WithTx`: Returns a new repository instance with the given transaction.

## Get

The `Get` method retrieves a single asset record from the database.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| id        | string          | Required | The UUID of the asset to get          |

### Output

| Type                    | Description                              |
|-------------------------|------------------------------------------|
| *assetmodel.Asset      | Asset record if found, nil otherwise     |
| error                   | Error if operation failed, nil otherwise |

## GetByUploadID

The `GetByUploadID` method retrieves a single asset record from the database by it's MuxUploadID.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| uploadID  | string          | Required | The Mux Upload ID to look up          |

### Output

| Type                    | Description                              |
|-------------------------|------------------------------------------|
| *assetmodel.Asset      | Asset record if found, nil otherwise     |
| error                   | Error if operation failed, nil otherwise |

## GetByAssetID

The `GetByAssetID` method retrieves a single asset record from the database by it's MuxAssetID.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| assetID   | string          | Required | The Mux Asset ID to look up           |

### Output

| Type                    | Description                              |
|-------------------------|------------------------------------------|
| *assetmodel.Asset      | Asset record if found, nil otherwise     |
| error                   | Error if operation failed, nil otherwise |

## List

The `List` method retrieves all asset records from the database.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| limit     | int             | Required | Maximum number of records to return    |
| offset    | int             | Required | Number of records to skip              |

### Output

| Type                   | Description                              |
|------------------------|------------------------------------------|
| []assetmodel.Asset     | List of asset records                    |
| error                  | Error if operation failed, nil otherwise |

## ListByIDs

The `ListByIDs` method retrieves a paginated list of asset records from the database by their IDs.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| limit     | int             | Required | Maximum number of records to return    |
| offset    | int             | Required | Number of records to skip              |
| ids       | ...string       | Required | List of asset IDs to retrieve         |

### Output

| Type                   | Description                              |
|------------------------|------------------------------------------|
| []assetmodel.Asset     | List of asset records                    |
| error                  | Error if operation failed, nil otherwise |

## Count

The `Count` method counts the total number of asset records in the database.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |

### Output

| Type   | Description                              |
|--------|------------------------------------------|
| int64  | Total count of asset records             |
| error  | Error if operation failed, nil otherwise |

## GetWithDeleted

The `GetWithDeleted` method retrieves a single asset record from the database including soft-deleted ones.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| id        | string          | Required | The UUID of the asset to get          |

### Output

| Type                    | Description                              |
|-------------------------|------------------------------------------|
| *assetmodel.Asset      | Asset record if found, nil otherwise     |
| error                   | Error if operation failed, nil otherwise |

## ListDeleted

The `ListDeleted` method retrieves all soft-deleted asset records from the database.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| limit     | int             | Required | Maximum number of records to return    |
| offset    | int             | Required | Number of records to skip              |

### Output

| Type                   | Description                              |
|------------------------|------------------------------------------|
| []assetmodel.Asset     | List of soft-deleted asset records       |
| error                  | Error if operation failed, nil otherwise |

## CountDeleted

The `CountDeleted` method counts the total number of soft-deleted asset records in the database.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |

### Output

| Type   | Description                              |
|--------|------------------------------------------|
| int64  | Total count of soft-deleted asset records |
| error  | Error if operation failed, nil otherwise |

## Create

The `Create` method creates a new asset record in the database.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| asset     | *assetmodel.Asset | Required | The asset record to create          |

### Output

| Type  | Description                             |
|-------|-----------------------------------------|
| error | Error if operation failed, nil otherwise |

## RemoveOwner

The `RemoveOwner` method removes local asset association with the owner by setting it's `owner_id` and `owner_type` to nil.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| id        | string          | Required | The UUID of the asset to update       |

### Output

| Type   | Description                              |
|--------|------------------------------------------|
| int64  | Number of affected rows                  |
| error  | Error if operation failed, nil otherwise |

## SetOwner

The `SetOwner` method sets local asset association with the owner by setting it's `owner_id` and `owner_type`.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| id        | string          | Required | The UUID of the asset to update       |
| ownerID   | string          | Required | The ID of the owner to set            |
| ownerType | string          | Required | The type of the owner to set          |

### Output

| Type   | Description                              |
|--------|------------------------------------------|
| int64  | Number of affected rows                  |
| error  | Error if operation failed, nil otherwise |

## Update

The `Update` method performs partial update of asset record in the database using updates.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| asset     | *assetmodel.Asset | Required | The asset record to update          |
| updates   | any             | Required | Updates to apply to the record        |

### Output

| Type   | Description                              |
|--------|------------------------------------------|
| int64  | Number of affected rows                  |
| error  | Error if operation failed, nil otherwise |

## Delete

The `Delete` method performs soft-delete of asset record.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| id        | string          | Required | The UUID of the asset to delete       |

### Output

| Type   | Description                              |
|--------|------------------------------------------|
| int64  | Number of affected rows                  |
| error  | Error if operation failed, nil otherwise |

## DeletePermanent

The `DeletePermanent` method performs permanent delete of asset record.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| id        | string          | Required | The UUID of the asset to permanently delete |

### Output

| Type   | Description                              |
|--------|------------------------------------------|
| int64  | Number of affected rows                  |
| error  | Error if operation failed, nil otherwise |

## Restore

The `Restore` method restores soft-deleted asset record.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| id        | string          | Required | The UUID of the asset to restore      |

### Output

| Type   | Description                              |
|--------|------------------------------------------|
| int64  | Number of affected rows                  |
| error  | Error if operation failed, nil otherwise |

## DB

The `DB` method returns the underlying gorm.DB instance.

### Input parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| (none)    |      |          |             |

### Output

| Type      | Description                     |
|-----------|---------------------------------|
| *gorm.DB  | The underlying GORM database instance |

## WithTx

The `WithTx` method returns a new repository instance with the given transaction.

### Input parameters

| Parameter | Type      | Required | Description                |
|-----------|-----------|----------|----------------------------|
| tx        | *gorm.DB  | Required | GORM transaction instance  |

### Output

| Type            | Description                     |
|-----------------|---------------------------------|
| Repository      | New repository instance with the transaction |