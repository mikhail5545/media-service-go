# Mux Detail Repository API

This document describes the API provided by the Mux Detail repository, which handles Mux asset detail data operations in PostgreSQL. The repository defines core data access functionality for storing and retrieving detailed asset information in PostgreSQL using GORM.

For an overview of the repository, see [README.md](./README.md). For high-level processes, see [Architecture Documentation](../../../architecture/).

## Overview

The Mux Detail repository is defined by the `Repository` interface and implemented by the `gormRepository` struct. It provides methods for interacting with Mux asset detail records stored in PostgreSQL, including operations for creating, reading, updating, and bulk operations on asset detail records.

Key methods include:

- `Get`: Retrieves a single asset detail record.
- `ListByAssetIDs`: Retrieves multiple asset detail records by their asset IDs.
- `Upsert`: Creates or updates an asset detail record.
- `DB`: Returns the underlying gorm.DB instance.
- `WithTx`: Returns a new repository instance with the given transaction.

## Get

The `Get` method retrieves a single asset detail record.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| assetID   | string          | Required | The asset ID of the detail record to get|

### Output

| Type                              | Description                              |
|-----------------------------------|------------------------------------------|
| *detailmodel.AssetDetail         | Asset detail record if found, nil otherwise |
| error                             | Error if operation failed, nil otherwise |

## ListByAssetIDs

The `ListByAssetIDs` method retrieves multiple asset detail records by their asset IDs.

### Input parameters

| Parameter | Type            | Required | Description                           |
|-----------|-----------------|----------|---------------------------------------|
| ctx       | context.Context | Required | Context for managing request lifecycle |
| assetIDs  | ...string       | Required | List of asset IDs to retrieve details for |

### Output

| Type                                            | Description                              |
|-------------------------------------------------|------------------------------------------|
| map[string]*detailmodel.AssetDetail            | Map of asset IDs to detail records       |
| error                                           | Error if operation failed, nil otherwise |

## Upsert

The `Upsert` method creates or updates an asset detail record. It uses `clauses.OnConflict` to perform an "upsert" operation, updating the 'tracks' column if the asset_id already exists.

### Input parameters

| Parameter | Type                        | Required | Description                           |
|-----------|-----------------------------|----------|---------------------------------------|
| ctx       | context.Context             | Required | Context for managing request lifecycle |
| details   | *detailmodel.AssetDetail   | Required | The asset detail record to upsert     |

### Output

| Type  | Description                             |
|-------|-----------------------------------------|
| error | Error if operation failed, nil otherwise |

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