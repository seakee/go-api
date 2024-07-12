package auth

import (
	"context"
	"fmt"
	"reflect"

	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MgoApp struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AppName     string             `bson:"app_name" json:"app_name"`
	AppID       string             `bson:"app_id" json:"app_id"`
	AppSecret   string             `bson:"app_secret" json:"app_secret"`
	RedirectUri string             `bson:"redirect_uri" json:"redirect_uri"`
	Description string             `bson:"description" json:"description"`
	Status      uint8              `bson:"status" json:"status"`
}

func (a *MgoApp) CollectionName() string {
	return "auth_app"
}

// buildQuery 构建查询条件
func (a *MgoApp) buildQuery() bson.M {
	query := bson.M{}

	v := reflect.ValueOf(a).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		bsonTag := fieldType.Tag.Get("bson")
		if bsonTag == "" || bsonTag == "-" {
			continue
		}

		if field.IsValid() && !field.IsZero() {
			query[bsonTag] = field.Interface()
		}
	}

	return query
}

func (a *MgoApp) First(ctx context.Context, db *qmgo.Database) (*MgoApp, error) {
	var app MgoApp

	err := db.Collection(a.CollectionName()).Find(ctx, a.buildQuery()).One(&app)
	if err != nil {
		return nil, fmt.Errorf("find failed: %w", err)
	}

	return &app, nil
}

func (a *MgoApp) Last(ctx context.Context, db *qmgo.Database) (*MgoApp, error) {
	var app MgoApp

	err := db.Collection(a.CollectionName()).Find(ctx, a.buildQuery()).Sort("-_id").One(&app)
	if err != nil {
		return nil, fmt.Errorf("find last failed: %w", err)
	}

	return &app, nil
}

func (a *MgoApp) Create(ctx context.Context, db *qmgo.Database) (string, error) {
	result, err := db.Collection(a.CollectionName()).InsertOne(ctx, a)
	if err != nil {
		return "", fmt.Errorf("create failed: %w", err)
	}

	objectID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", fmt.Errorf("create failed: unable to convert inserted ID to ObjectID")
	}

	return objectID.Hex(), nil
}

func (a *MgoApp) Delete(ctx context.Context, db *qmgo.Database) error {
	query := a.buildQuery()
	if len(query) == 0 {
		return fmt.Errorf("delete failed: no valid fields to build query")
	}

	err := db.Collection(a.CollectionName()).Remove(ctx, query)
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}

	return nil
}

func (a *MgoApp) Updates(ctx context.Context, db *qmgo.Database, updates bson.M) error {
	query := a.buildQuery()
	if len(query) == 0 {
		return fmt.Errorf("update failed: no valid fields to build query")
	}

	err := db.Collection(a.CollectionName()).UpdateOne(ctx, query, bson.M{"$set": updates})
	if err != nil {
		return fmt.Errorf("updates failed: %w", err)
	}

	return nil
}

func (a *MgoApp) List(ctx context.Context, db *qmgo.Database) ([]MgoApp, error) {
	var apps []MgoApp

	err := db.Collection(a.CollectionName()).Find(ctx, a.buildQuery()).All(&apps)
	if err != nil {
		return nil, fmt.Errorf("find list failed: %w", err)
	}

	return apps, nil
}

func (a *MgoApp) Creates(ctx context.Context, db *qmgo.Database, apps []MgoApp) ([]primitive.ObjectID, error) {
	docs := make([]interface{}, len(apps))
	for i, app := range apps {
		docs[i] = app
	}

	result, err := db.Collection(a.CollectionName()).InsertMany(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("batch insert failed: %w", err)
	}

	objectIDs := make([]primitive.ObjectID, len(result.InsertedIDs))
	for i, id := range result.InsertedIDs {
		objectID, ok := id.(primitive.ObjectID)
		if !ok {
			return nil, fmt.Errorf("batch insert failed: unable to convert inserted ID to ObjectID")
		}
		objectIDs[i] = objectID
	}

	return objectIDs, nil
}

func (a *MgoApp) Pagination(ctx context.Context, db *qmgo.Database, page, size int) ([]MgoApp, error) {
	var apps []MgoApp

	err := db.Collection(a.CollectionName()).Find(ctx, a.buildQuery()).Skip(int64((page - 1) * size)).Limit(int64(size)).All(&apps)
	if err != nil {
		return nil, fmt.Errorf("find with pagination failed: %w", err)
	}

	return apps, nil
}

func (a *MgoApp) FindWithSort(ctx context.Context, db *qmgo.Database, sort string) ([]MgoApp, error) {
	var apps []MgoApp

	err := db.Collection(a.CollectionName()).Find(ctx, a.buildQuery()).Sort(sort).All(&apps)
	if err != nil {
		return nil, fmt.Errorf("find with sort failed: %w", err)
	}

	return apps, nil
}

func (a *MgoApp) Count(ctx context.Context, db *qmgo.Database) (int64, error) {
	count, err := db.Collection(a.CollectionName()).Find(ctx, a.buildQuery()).Count()
	if err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}

	return count, nil
}
