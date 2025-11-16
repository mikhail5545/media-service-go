# Restore Operation Flowchart

This document illustrates the low-level logic of the `Restore` method in the Mux service.

See [api documentation](../api.md) for more details.

## Flowchart

```mermaid
graph TD
    A((Start)) --> B[Validate AssetID]
    B --> C{Is AssetID a Valid UUID?}
    C -->|No| D[Return ErrInvalidArgument]
    C --> |Yes| E[Start DB Transaction]
    E --> F[Restore Asset in the DB]
    F --> G{Restore Successful?}
    G --> |No| H[Return Error]
    G --> I[Return RowsAffected]
    I --> J{Is RowsAffected == 0?}
    J --> |Yes| K[Return ErrNotFound]
    J --> |No| L[Commit DB Transaction]
    L --> M[End]
    D --> M
    H --> M
    K --> M
```
