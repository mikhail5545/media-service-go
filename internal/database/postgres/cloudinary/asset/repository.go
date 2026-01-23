package asset

import (
	"context"

	"github.com/google/uuid"
	"github.com/mikhail5545/media-service-go/internal/database/types"
	cldassetmodel "github.com/mikhail5545/media-service-go/internal/models/cloudinary/asset"
	"gorm.io/gorm"
)

type GormRepository interface {
	DB() *gorm.DB
	WithTx(tx *gorm.DB) *Repository
	Get(ctx context.Context, opts GetOptions, scopes ...Scope) (*cldassetmodel.Asset, error)
	List(ctx context.Context, opts ListOptions, scopes ...Scope) ([]*cldassetmodel.Asset, string, error)
	ListAll(ctx context.Context, opts ListAllOptions, scopes ...Scope) ([]*cldassetmodel.Asset, error)
	Create(ctx context.Context, asset *cldassetmodel.Asset) error
	Update(ctx context.Context, updates map[string]any, opts StateOperationOptions) (int64, error)
	Archive(ctx context.Context, opts StateOperationOptions, auditOpts *types.AuditTrailOptions) (int64, error)
	Restore(ctx context.Context, opts StateOperationOptions, auditOpts *types.AuditTrailOptions) (int64, error)
	Delete(ctx context.Context, opts StateOperationOptions) (int64, error)
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
	ScopeUploadURLGenerated Scope = iota
	ScopeActive             Scope = iota
	ScopeArchived           Scope = iota
	ScopeBroken             Scope = iota
)

type Filter struct {
	IDs                 uuid.UUIDs
	CloudinaryAssetIDs  []string
	CloudinaryPublicIDs []string

	ResourceTypes []string
	Formats       []string

	Fields   []string
	Statuses []cldassetmodel.Status

	OrderDir   cldassetmodel.OrderDirection
	OrderField cldassetmodel.OrderField

	PageSize  int
	PageToken string
}

type GetOptions struct {
	ID                 uuid.UUID
	CloudinaryAssetID  string
	CloudinaryPublicID string
	Fields             []string
}

type ListOptions struct {
	IDs                 uuid.UUIDs
	CloudinaryAssetIDs  []string
	CloudinaryPublicIDs []string

	ResourceTypes []string
	Formats       []string

	Fields   []string
	Statuses []cldassetmodel.Status

	OrderDir   cldassetmodel.OrderDirection
	OrderField cldassetmodel.OrderField

	PageSize  int
	PageToken string
}

type ListAllOptions struct {
	IDs                 uuid.UUIDs
	CloudinaryAssetIDs  []string
	CloudinaryPublicIDs []string

	ResourceTypes []string
	Formats       []string

	Fields   []string
	Statuses []cldassetmodel.Status

	OrderDir   cldassetmodel.OrderDirection
	OrderField cldassetmodel.OrderField
}

type StateOperationOptions struct {
	IDs                 uuid.UUIDs
	CloudinaryAssetIDs  []string
	CloudinaryPublicIDs []string

	ResourceTypes []string
	Formats       []string
}

func (r *Repository) Get(ctx context.Context, opts GetOptions, scopes ...Scope) (*cldassetmodel.Asset, error) {
	return r.get(ctx, &Filter{
		IDs:                 uuid.UUIDs{opts.ID},
		CloudinaryAssetIDs:  []string{opts.CloudinaryAssetID},
		CloudinaryPublicIDs: []string{opts.CloudinaryPublicID},
		Fields:              opts.Fields,
		Statuses:            extractScopes(scopes),
	})
}

func (r *Repository) List(ctx context.Context, opts ListOptions, scopes ...Scope) ([]*cldassetmodel.Asset, string, error) {
	return r.list(ctx, populateFromListOptions(&opts, scopes))
}

func (r *Repository) ListAll(ctx context.Context, opts ListAllOptions, scopes ...Scope) ([]*cldassetmodel.Asset, error) {
	return r.listAll(ctx, &Filter{
		IDs:                 opts.IDs,
		CloudinaryAssetIDs:  opts.CloudinaryAssetIDs,
		CloudinaryPublicIDs: opts.CloudinaryPublicIDs,
		ResourceTypes:       opts.ResourceTypes,
		Formats:             opts.Formats,
		Fields:              opts.Fields,
		Statuses:            extractScopes(scopes),
		OrderDir:            opts.OrderDir,
		OrderField:          opts.OrderField,
	})
}

func (r *Repository) Create(ctx context.Context, asset *cldassetmodel.Asset) error {
	return r.db.WithContext(ctx).Create(asset).Error
}

func (r *Repository) Update(ctx context.Context, updates map[string]any, opts StateOperationOptions) (int64, error) {
	return r.update(ctx, populateFromStateOperationOptions(&opts), updates)
}

func (r *Repository) Archive(ctx context.Context, opts StateOperationOptions, auditOpts *types.AuditTrailOptions) (int64, error) {
	return r.archive(ctx, populateFromStateOperationOptions(&opts), auditOpts)
}

func (r *Repository) Restore(ctx context.Context, opts StateOperationOptions, auditOpts *types.AuditTrailOptions) (int64, error) {
	return r.restore(ctx, populateFromStateOperationOptions(&opts), auditOpts)
}

func (r *Repository) Delete(ctx context.Context, opts StateOperationOptions) (int64, error) {
	return r.delete(ctx, populateFromStateOperationOptions(&opts))
}
