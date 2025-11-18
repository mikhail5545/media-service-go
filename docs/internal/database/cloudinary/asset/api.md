# Cloudinary Asset Repository API

This document describes the API provided by the Cloudinary Asset repository, which handles Cloudinary asset data operations in PostgreSQL. The repository defines core data access functionality for storing and retrieving asset records in PostgreSQL using GORM.

For an overview of the repository, see [README.md](./README.md). For high-level processes, see [Architecture Documentation](../../../architecture/).

## Overview

The Cloudinary Asset repository is defined by the `Repository` interface and implemented by the `gormRepository` struct. It provides methods for interacting with Cloudinary asset records stored in PostgreSQL, including operations for creating, reading, updating, and deleting asset records with both soft-delete and permanent delete capabilities.

Key methods include:

- `Get`: Retrieves a single asset record from the database.
- `List`: Retrieves all asset records from the database.
- `ListByIDs`: Retrieves a paginated list of asset records from the database by their IDs.
- `ListAllCloudinaryAssetIDs`: Returns all asset record's cloudinary asset id field value.
- `Count`: Counts the total number of asset records in the database.
- `GetWithDeleted`: Retrieves a single asset record from the database including soft-deleted ones.
- `GetWithDeletedByAssetID`: Retrieves a single asset record from the database by it's external CloudinaryAssetID including soft-deleted ones.
- `ListSelect`: Returns a list of all assets with specified fields populated.
- `ListDeleted`: Retrieves all soft-deleted asset records from the database.
- `CountDeleted`: Counts the total number of soft-deleted asset records in the database.
- `Create`: Creates a new asset record in the database.
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

| Type             | Description                              |
|------------------|------------------------------------------|
| *asset.Asset    | Asset record if found, nil otherwise     |
| error            | Error if operation failed, nil otherwise |

## List

The `List` method retrieves all asset records from the database.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| limit     | int             | Required | Maximum number of records to return    |
| offset    | int             | Required | Number of records to skip              |

### Output

| Type           | Description                              |
|----------------|------------------------------------------|
| []asset.Asset  | List of asset records                    |
| error          | Error if operation failed, nil otherwise |

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

| Type           | Description                              |
|----------------|------------------------------------------|
| []asset.Asset  | List of asset records                    |
| error          | Error if operation failed, nil otherwise |

## ListAllCloudinaryAssetIDs

The `ListAllCloudinaryAssetIDs` method returns all asset record's cloudinary asset id field value.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |

### Output

| Type                      | Description                              |
|---------------------------|------------------------------------------|
| map[string]struct{}       | Map of cloudinary asset IDs for O(1) lookup |
| error                     | Error if operation failed, nil otherwise |

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

| Type             | Description                              |
|------------------|------------------------------------------|
| *asset.Asset    | Asset record if found, nil otherwise     |
| error            | Error if operation failed, nil otherwise |

## GetWithDeletedByAssetID

The `GetWithDeletedByAssetID` method retrieves a single asset record from the database by it's external CloudinaryAssetID including soft-deleted ones.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| assetID   | string          | Required | The Cloudinary asset ID to get        |

### Output

| Type             | Description                              |
|------------------|------------------------------------------|
| *asset.Asset    | Asset record if found, nil otherwise     |
| error            | Error if operation failed, nil otherwise |

## ListSelect

The `ListSelect` method returns a list of all assets with specified fields populated.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| fields    | ...string       | Required | Fields to select in the query         |

### Output

| Type           | Description                              |
|----------------|------------------------------------------|
| []asset.Asset  | List of asset records with selected fields |
| error          | Error if operation failed, nil otherwise |

## ListDeleted

The `ListDeleted` method retrieves all soft-deleted asset records from the database.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| limit     | int             | Required | Maximum number of records to return    |
| offset    | int             | Required | Number of records to skip              |

### Output

| Type           | Description                              |
|----------------|------------------------------------------|
| []asset.Asset  | List of soft-deleted asset records       |
| error          | Error if operation failed, nil otherwise |

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
| Asset     | *asset.Asset   | Required | The asset record to create            |

### Output

| Type  | Description                             |
|-------|-----------------------------------------|
| error | Error if operation failed, nil otherwise |

## Update

The `Update` method performs partial update of asset record in the database using updates.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| Asset     | *asset.Asset   | Required | The asset record to update            |
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