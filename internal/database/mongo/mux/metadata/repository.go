package metadata

import (
	"context"

	"github.com/mikhail5545/media-service-go/internal/models/mux/metadata"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoRepository interface {
	Create(ctx context.Context, data *metadata.AssetMetadata) error
	Get(ctx context.Context, key string) (*metadata.AssetMetadata, error)
	GetByOwner(ctx context.Context, key string, owner *metadata.Owner) (*metadata.AssetMetadata, error)
	Update(ctx context.Context, key string, data *metadata.AssetMetadata) error
	Delete(ctx context.Context, key string) error
	ListUnownedIDs(ctx context.Context) ([]string, error)
	List(ctx context.Context) ([]*metadata.AssetMetadata, error)
	ListByKeys(ctx context.Context, keys []string) (map[string]*metadata.AssetMetadata, error)
}

type Repository struct {
	db             *mongo.Database
	collectionName string
}

var _ MongoRepository = (*Repository)(nil)

func New(db *mongo.Database, collectionName string) *Repository {
	return &Repository{db: db, collectionName: collectionName}
}

func (r *Repository) Create(ctx context.Context, data *metadata.AssetMetadata) error {
	collection := r.db.Collection(r.collectionName)
	_, err := collection.InsertOne(ctx, data)
	return err
}

func (r *Repository) Get(ctx context.Context, key string) (*metadata.AssetMetadata, error) {
	collection := r.db.Collection(r.collectionName)
	filter := bson.D{{Key: "_id", Value: key}}

	var result metadata.AssetMetadata
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *Repository) GetByOwner(ctx context.Context, key string, owner *metadata.Owner) (*metadata.AssetMetadata, error) {
	collection := r.db.Collection(r.collectionName)
	filter := bson.D{
		{Key: "_id", Value: key},
		{Key: "owners", Value: bson.D{
			{Key: "$elemMatch", Value: bson.D{
				{Key: "owner_id", Value: owner.OwnerID},
				{Key: "owner_type", Value: owner.OwnerType},
			}},
		}},
	}

	var result metadata.AssetMetadata
	err := collection.FindOne(ctx, filter).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *Repository) Update(ctx context.Context, key string, data *metadata.AssetMetadata) error {
	collection := r.db.Collection(r.collectionName)

	filter := bson.D{{Key: "_id", Value: key}}
	update := bson.D{{Key: "$set", Value: data}}
	opts := options.UpdateOne().SetUpsert(false)

	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return mongo.ErrNoDocuments
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, key string) error {
	collection := r.db.Collection(r.collectionName)
	filter := bson.D{{Key: "_id", Value: key}}
	_, err := collection.DeleteOne(ctx, filter)
	return err
}

func (r *Repository) ListUnownedIDs(ctx context.Context) ([]string, error) {
	collection := r.db.Collection(r.collectionName)

	filter := bson.D{{Key: "owners", Value: bson.D{{Key: "$size", Value: 0}}}}
	opts := options.Find().SetProjection(bson.D{{Key: "_id", Value: 1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		Key string `bson:"_id"`
	}

	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	ids := make([]string, len(results))
	for i, res := range results {
		ids[i] = res.Key
	}
	return ids, nil
}

func (r *Repository) List(ctx context.Context) ([]*metadata.AssetMetadata, error) {
	collection := r.db.Collection(r.collectionName)

	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var metadataList []*metadata.AssetMetadata
	if err := cursor.All(ctx, &metadataList); err != nil {
		return nil, err
	}
	return metadataList, nil
}

func (r *Repository) ListByKeys(ctx context.Context, keys []string) (map[string]*metadata.AssetMetadata, error) {
	collection := r.db.Collection(r.collectionName)

	filter := bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: keys}}}}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	metadataMap := make(map[string]*metadata.AssetMetadata)
	for cursor.Next(ctx) {
		var doc metadata.AssetMetadata
		if err := cursor.Decode(&doc); err != nil {
			return nil, err
		}
		metadataMap[doc.Key] = &doc
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return metadataMap, nil
}
