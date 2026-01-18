package asset

import (
	"time"

	metamodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/metadata"
)

type OrderDirection string

const (
	OrderAscending  OrderDirection = "ASC"
	OrderDescending OrderDirection = "DESC"
)

type OrderField string

const (
	OrderCreatedAt    OrderField = "created_at"
	OrderUpdatedAt    OrderField = "updated_at"
	OrderResourceType OrderField = "resource_type"
	OrderFormat       OrderField = "format"
)

// Details is a DTO that combines the core Asset model with its metadata.
type Details struct {
	Asset    *Asset
	Metadata *metamodel.AssetMetadata
}

type GetFilter struct {
	ID string `param:"id" json:"-"`
}

type ListRequest struct {
	IDs                 []string `query:"ids" json:"-"`
	CloudinaryAssetIDs  []string `query:"cloudinary_asset_ids" json:"-"`
	CloudinaryPublicIDs []string `query:"cloudinary_public_ids" json:"-"`

	ResourceTypes []string `query:"resource_types" json:"-"`
	Formats       []string `query:"formats" json:"-"`

	OrderDir   OrderDirection `query:"order_dir" json:"-"`
	OrderField OrderField     `query:"order_field" json:"-"`

	PageSize  int    `query:"page_size" json:"-"`
	PageToken string `query:"page_token" json:"-"`
}

type ManageOwnerRequest struct {
	ID        string `param:"id" json:"-"`
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_type"`
}

type CreateSignedUploadURLRequest struct {
	Eager    *string `json:"eager"`
	PublicID string  `json:"public_id"`
	File     string  `json:"file"`
}

type GeneratedSignedParams struct {
	Signature    string  `json:"signature"`
	Timestamp    string  `json:"timestamp"`
	ApiKey       string  `json:"api_key"`
	Eager        *string `json:"eager,omitempty"`
	PublicID     string  `json:"public_id"`
	ResourceType string  `json:"resource_type,omitempty"`
}

type ChangeStateRequest struct {
	ID        string `param:"id" json:"-"`
	AdminID   string `json:"admin_id"`
	AdminName string `json:"admin_name"`
	Note      string `json:"note"`
}

type SuccessfulUploadRequest struct {
	CloudinaryAssetID  string `json:"cloudinary_asset_id"`
	CloudinaryPublicID string `json:"cloudinary_public_id"`
	ResourceType       string `json:"resource_type"`
	Format             string `json:"format"`
	Width              *int   `json:"width"`
	Height             *int   `json:"height"`
	URL                string `json:"url"`
	SecureURL          string `json:"secure_url"`
	AssetFolder        string `json:"asset_folder"`
	DisplayName        string `json:"display_name"`
}

// CloudinaryUploadWebhook represents Cloudinary API webhook triggered by an asset upload.
type CloudinaryUploadWebhook struct {
	NotificationType    string              `json:"notification_type"`
	Timestamp           time.Time           `json:"timestamp"`
	RequestID           string              `json:"request_id"`
	AssetID             string              `json:"asset_id"`
	PublicID            string              `json:"public_id"`
	Width               int                 `json:"width"`
	Height              int                 `json:"height"`
	Format              string              `json:"format"`
	ResourceType        string              `json:"resource_type"`
	CreatedAt           time.Time           `json:"created_at"`
	Tags                []string            `json:"tags"`
	Url                 string              `json:"url"`
	SecureUrl           string              `json:"secure_url"`
	AssetFolder         string              `json:"asset_folder"`
	DisplayName         string              `json:"display_name"`
	ApiKey              string              `json:"api_key"`
	Context             *Context            `json:"context,omitempty"`
	NotificationContext NotificationContext `json:"notification_context"`
	SignatureKey        string              `json:"signature_key"`
}

// Context represents the context object in a Cloudinary webhook.
type Context struct {
	Custom CustomContext `json:"custom"`
}

// CustomContext is a map that holds the custom key-value pairs
// sent during an upload. The keys are strings (e.g., "product_ids")
// and the values are also strings (e.g., "uuid1|uuid2").
type CustomContext map[string]string

// NotificationContext represents Cloudinary API webhook notification context.
type NotificationContext struct {
	TriggeredAt time.Time   `json:"triggered_at"`
	TriggeredBy TriggeredBy `json:"triggered_by"`
}

// TriggeredBy represents Cloudinary API webhook payload about source of a trigger.
type TriggeredBy struct {
	Source string `json:"source"`
	ID     string `json:"id"`
}

// CloudinaryContextChangeWebhook represents Cloudinary API webhook triggered by an asset/assets context change.
type CloudinaryContextChangeWebhook struct {
	NotificationType    string              `json:"notification_type"`
	Source              string              `json:"source"`
	Resources           map[string]Resource `json:"resources"`
	NotificationContext NotificationContext `json:"notification_context"`
	SignatureKey        string              `json:"signature_key"`
}

// Resource represents Cloudinary API webhook payload about context changes for particular asset.
type Resource struct {
	Added        []KeyVal    `json:"added"`
	Removed      []KeyVal    `json:"removed"`
	Updated      []UpdateVal `json:"updated"`
	AssetID      string      `json:"asset_id"`
	ResourceType string      `json:"resource_type"`
	Type         string      `json:"type"`
}

// KeyVal represents Cloudinary API webhook payload about context changes in key-value format.
type KeyVal struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// UpdateVal represents Cloudinary API webhook payload about context changes for updated keys.
type UpdateVal struct {
	Name     string `json:"name"`
	OldValue string `json:"old_value"`
	NewValue string `json:"new_value"`
}
