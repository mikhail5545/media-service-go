# Get operation Flowchart

This document illustrates the low-level logic of the all `Get` methods in the Mux service:

- `Get`
- `GetWithDeleted`

See [api documentation](../api.md) for more details.

## Flowchart

```mermaid
graph TD
    A((Start)) --> B["Validate AssetID"]
    B --> C{"Is AssetID a valid UUID?"}
    C --> |No| D[Return ErrInvalidArgument]
    C --> |Yes| E[Retrieve asset from the database]
    E --> F{Asset Found?}
    F --> |No| G[Return ErrNotFound]
    F --> |Yes| H[Return Asset]
    H --> I[End]
    D --> I
    G --> I
```
