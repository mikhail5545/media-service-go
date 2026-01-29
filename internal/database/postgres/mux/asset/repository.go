package asset

import (
	"context"

	"github.com/google/uuid"
	"github.com/mikhail5545/media-service-go/internal/database/types"
	muxassetmodel "github.com/mikhail5545/media-service-go/internal/models/mux/asset"
	"gorm.io/gorm"
)

type GormRepository interface {
	DB() *gorm.DB
	WithTx(tx *gorm.DB) *Repository
	// Get retrieves a single mux asset based on the provided options and scopes.
	// If no scopes are provided, only active assets are considered.
	Get(ctx context.Context, opts GetOptions, scopes ...Scope) (*muxassetmodel.Asset, error)
	// List retrieves a paginated list of mux assets based on the provided options and scopes.
	// If no scopes are provided, only active assets are considered.
	List(ctx context.Context, opts ListOptions, scopes ...Scope) ([]*muxassetmodel.Asset, string, error)
	// ListAll retrieves all mux assets based on the provided options and scopes.
	// If no scopes are provided, only active assets are considered.
	// It does not support pagination, so it should be used with caution for large datasets.
	ListAll(ctx context.Context, opts ListAllOptions, scopes ...Scope) ([]*muxassetmodel.Asset, error)
	Create(ctx context.Context, asset *muxassetmodel.Asset) error
	// Update performs a partial update on mux assets matching the provided state operation options.
	// [muxassetmodel.Asset.Status] field cannot be updated using this method, use Restore, Archive instead.
	Update(ctx context.Context, updates map[string]any, opts StateOperationOptions) (int64, error)
	// Restore restores currently archived mux asset matching the provided state operation options.
	Restore(ctx context.Context, opts StateOperationOptions, auditOpts types.AuditTrailOptions) (int64, error)
	// Archive archives mux asset matching the provided state operation options.
	Archive(ctx context.Context, opts StateOperationOptions, auditOpts types.AuditTrailOptions) (int64, error)
	// Delete permanently deletes mux asset matching the provided state operation options.
	// Only currently soft-deleted (archived) assets can be permanently deleted.
	Delete(ctx context.Context, opts StateOperationOptions) (int64, error)
	MarkAsBroken(ctx context.Context, opts StateOperationOptions, auditOpts types.AuditTrailOptions) (int64, error)
}

type Repository struct {
	db *gorm.DB
}

var _ GormRepository = (*Repository)(nil)

func New(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) DB() *gorm.DB {
	return r.db
}

func (r *Repository) WithTx(tx *gorm.DB) *Repository {
	return &Repository{db: tx}
}

type Scope uint

const (
	ScopeAll                Scope = iota
	ScopeActive             Scope = iota
	ScopeUploadURLGenerated Scope = iota
	ScopeArchived           Scope = iota
	ScopeBroken             Scope = iota
)

type Filter struct {
	IDs          uuid.UUIDs
	MuxUploadIDs []string
	MuxAssetIDs  []string

	States          []muxassetmodel.State
	Statuses        []muxassetmodel.Status
	UploadStatuses  []muxassetmodel.UploadStatus
	AspectRatios    []string
	ResolutionTiers []string
	IngestTypes     []muxassetmodel.IngestType

	Fields []string

	OrderBy  muxassetmodel.OrderField
	OrderDir muxassetmodel.OrderDirection

	PageSize  int
	PageToken string
}

type GetOptions struct {
	ID          uuid.UUID
	MuxUploadID string
	MuxAssetID  string
	Fields      []string
}

type ListOptions struct {
	IDs          uuid.UUIDs
	MuxUploadIDs []string
	MuxAssetIDs  []string

	States          []muxassetmodel.State
	UploadStatuses  []muxassetmodel.UploadStatus
	AspectRatios    []string
	ResolutionTiers []string
	IngestTypes     []muxassetmodel.IngestType

	Fields []string

	OrderBy  muxassetmodel.OrderField
	OrderDir muxassetmodel.OrderDirection

	PageSize  int
	PageToken string
}

type ListAllOptions struct {
	IDs          uuid.UUIDs
	MuxUploadIDs []string
	MuxAssetIDs  []string

	States          []muxassetmodel.State
	UploadStatuses  []muxassetmodel.UploadStatus
	AspectRatios    []string
	ResolutionTiers []string
	IngestTypes     []muxassetmodel.IngestType

	Fields []string

	OrderBy  muxassetmodel.OrderField
	OrderDir muxassetmodel.OrderDirection
}

type StateOperationOptions struct {
	IDs          uuid.UUIDs
	MuxUploadIDs []string
	MuxAssetIDs  []string

	States          []muxassetmodel.State
	UploadStatuses  []muxassetmodel.UploadStatus
	AspectRatios    []string
	ResolutionTiers []string
	IngestTypes     []muxassetmodel.IngestType
}

// Get retrieves a single mux asset based on the provided options and scopes.
// If no scopes are provided, only active assets are considered.
func (r *Repository) Get(ctx context.Context, opts GetOptions, scopes ...Scope) (*muxassetmodel.Asset, error) {
	statuses := extractScopes(scopes)
	return r.get(ctx, &Filter{
		IDs:          uuid.UUIDs{opts.ID},
		MuxUploadIDs: []string{opts.MuxUploadID},
		MuxAssetIDs:  []string{opts.MuxAssetID},
		Statuses:     statuses,
		Fields:       opts.Fields,
	})
}

// List retrieves a paginated list of mux assets based on the provided options and scopes.
// If no scopes are provided, only active assets are considered.
func (r *Repository) List(ctx context.Context, opts ListOptions, scopes ...Scope) ([]*muxassetmodel.Asset, string, error) {
	return r.list(ctx, populateFromListOptions(opts, scopes))
}

// ListAll retrieves all mux assets based on the provided options and scopes.
// If no scopes are provided, only active assets are considered.
// It does not support pagination, so it should be used with caution for large datasets.
func (r *Repository) ListAll(ctx context.Context, opts ListAllOptions, scopes ...Scope) ([]*muxassetmodel.Asset, error) {
	statuses := extractScopes(scopes)
	return r.listAll(ctx, &Filter{
		IDs:             opts.IDs,
		MuxUploadIDs:    opts.MuxUploadIDs,
		MuxAssetIDs:     opts.MuxAssetIDs,
		Statuses:        statuses,
		UploadStatuses:  opts.UploadStatuses,
		AspectRatios:    opts.AspectRatios,
		ResolutionTiers: opts.ResolutionTiers,
		IngestTypes:     opts.IngestTypes,
		Fields:          opts.Fields,
		OrderBy:         opts.OrderBy,
		OrderDir:        opts.OrderDir,
	})
}

func (r *Repository) Create(ctx context.Context, asset *muxassetmodel.Asset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

// Update performs a partial update on mux assets matching the provided state operation options.
// [muxassetmodel.Asset.Status] field cannot be updated using this method, use Restore, Archive instead.
func (r *Repository) Update(ctx context.Context, updates map[string]any, opts StateOperationOptions) (int64, error) {
	return r.update(ctx, populateFromStateOperationOptions(opts), updates)
}

// Restore restores currently archived mux asset matching the provided state operation options.
func (r *Repository) Restore(ctx context.Context, opts StateOperationOptions, auditOpts types.AuditTrailOptions) (int64, error) {
	return r.restore(ctx, populateFromStateOperationOptions(opts), &auditOpts)
}

// Archive archives mux asset matching the provided state operation options.
func (r *Repository) Archive(ctx context.Context, opts StateOperationOptions, auditOpts types.AuditTrailOptions) (int64, error) {
	return r.archive(ctx, populateFromStateOperationOptions(opts), &auditOpts)
}

// Delete permanently deletes mux asset matching the provided state operation options.
// Only currently soft-deleted (archived) assets can be permanently deleted.
func (r *Repository) Delete(ctx context.Context, opts StateOperationOptions) (int64, error) {
	return r.delete(ctx, populateFromStateOperationOptions(opts))
}

func (r *Repository) MarkAsBroken(ctx context.Context, opts StateOperationOptions, auditOpts types.AuditTrailOptions) (int64, error) {
	return r.markAsBroken(ctx, populateFromStateOperationOptions(opts), &auditOpts)
}
