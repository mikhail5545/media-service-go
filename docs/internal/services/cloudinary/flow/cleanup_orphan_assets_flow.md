# CleanupOrphanAssets Flowchart

This document illustrates the low-level logic of the `CleanupOrphanAssets` method in the Cloudinary service.

## Flowchart

```mermaid
graph TD
    A["Start"] --> B["Validate Request"]
    B --> C{"Request Valid?"}
    C -->|No| D["Return ErrInvalidArgument"]
    C -->|Yes| E["List Assets from Cloudinary by Folder"]
    E --> F{"API Call Successful?"}
    F -->|No| G["Return Cloudinary Error"]
    F -->|Yes| H["List All Cloudinary Asset IDs from Local DB"]
    H --> I{"DB Query Successful?"}
    I -->|No| J["Return DB Error"]
    I -->|Yes| K["Compare Cloudinary assets with local assets to find orphans"]
    K --> L{"Are there any orphans?"}
    L -->|No| M["Log 'No orphans found' and Return 0"]
    L -->|Yes| N["Log count of orphans to delete"]
    N --> O["Delete Orphan Assets from Cloudinary"]
    O --> P{"API Call Successful?"}
    P -->|No| Q["Return Cloudinary Error"]
    P -->|Yes| R["Return count of deleted orphans"]
    D --> S["End"]
    G --> S
    J --> S
    M --> S
    Q --> S
    R --> S
```
