// Copyright 2024 Seakee.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

// Package auth provides authentication and authorization functionalities.
package auth

import (
	"context"
	"fmt"
	"reflect"

	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MgoApp represents an application in the authentication system.
type MgoApp struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	AppName     string             `bson:"app_name" json:"app_name"`
	AppID       string             `bson:"app_id" json:"app_id"`
	AppSecret   string             `bson:"app_secret" json:"app_secret"`
	RedirectUri string             `bson:"redirect_uri" json:"redirect_uri"`
	Description string             `bson:"description" json:"description"`
	Status      uint8              `bson:"status" json:"status"`
}

// CollectionName returns the name of the MongoDB collection for MgoApp.
func (a *MgoApp) CollectionName() string {
	return "auth_app"
}

// buildQuery constructs a BSON query based on non-zero fields of the MgoApp struct.
//
// This method uses reflection to iterate through the struct fields and builds
// a query using the bson tags and non-zero values.
//
// Returns:
//   - bson.M: A BSON map representing the query.
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

// First retrieves the first MgoApp document that matches the query.
//
// Parameters:
//   - ctx: A context.Context for the database operation.
//   - db: A pointer to the qmgo.Database to perform the operation on.
//
// Returns:
//   - *MgoApp: A pointer to the retrieved MgoApp, or nil if not found.
//   - error: An error if the operation fails, or nil on success.
//
// Example:
//
//	app := &MgoApp{AppID: "example_id"}
//	result, err := app.First(ctx, db)
//	if err != nil {
//	    log.Printf("Error finding app: %v", err)
//	    return
//	}
//	fmt.Printf("Found app: %+v\n", result)
func (a *MgoApp) First(ctx context.Context, db *qmgo.Database) (*MgoApp, error) {
	var app MgoApp

	err := db.Collection(a.CollectionName()).Find(ctx, a.buildQuery()).One(&app)
	if err != nil {
		return nil, fmt.Errorf("find failed: %w", err)
	}

	return &app, nil
}

// Last retrieves the last MgoApp document that matches the query, sorted by _id in descending order.
//
// Parameters:
//   - ctx: A context.Context for the database operation.
//   - db: A pointer to the qmgo.Database to perform the operation on.
//
// Returns:
//   - *MgoApp: A pointer to the retrieved MgoApp, or nil if not found.
//   - error: An error if the operation fails, or nil on success.
//
// Example:
//
//	app := &MgoApp{Status: 1}
//	result, err := app.Last(ctx, db)
//	if err != nil {
//	    log.Printf("Error finding last app: %v", err)
//	    return
//	}
//	fmt.Printf("Found last app: %+v\n", result)
func (a *MgoApp) Last(ctx context.Context, db *qmgo.Database) (*MgoApp, error) {
	var app MgoApp

	err := db.Collection(a.CollectionName()).Find(ctx, a.buildQuery()).Sort("-_id").One(&app)
	if err != nil {
		return nil, fmt.Errorf("find last failed: %w", err)
	}

	return &app, nil
}

// Create inserts a new MgoApp document into the database.
//
// Parameters:
//   - ctx: A context.Context for the database operation.
//   - db: A pointer to the qmgo.Database to perform the operation on.
//
// Returns:
//   - string: The hexadecimal representation of the inserted document's ObjectID.
//   - error: An error if the operation fails, or nil on success.
//
// Example:
//
//	newApp := &MgoApp{
//	    AppName: "New App",
//	    AppID: "new_app_id",
//	    AppSecret: "secret",
//	    Status: 1,
//	}
//	id, err := newApp.Create(ctx, db)
//	if err != nil {
//	    log.Printf("Error creating app: %v", err)
//	    return
//	}
//	fmt.Printf("Created app with ID: %s\n", id)
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

// Delete removes the MgoApp document that matches the query from the database.
//
// Parameters:
//   - ctx: A context.Context for the database operation.
//   - db: A pointer to the qmgo.Database to perform the operation on.
//
// Returns:
//   - error: An error if the operation fails, or nil on success.
//
// Example:
//
//	app := &MgoApp{AppID: "app_to_delete"}
//	err := app.Delete(ctx, db)
//	if err != nil {
//	    log.Printf("Error deleting app: %v", err)
//	    return
//	}
//	fmt.Println("App deleted successfully")
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

// Updates modifies the MgoApp document that matches the query with the provided updates.
//
// Parameters:
//   - ctx: A context.Context for the database operation.
//   - db: A pointer to the qmgo.Database to perform the operation on.
//   - updates: A bson.M containing the fields to update and their new values.
//
// Returns:
//   - error: An error if the operation fails, or nil on success.
//
// Example:
//
//	app := &MgoApp{AppID: "app_to_update"}
//	updates := bson.M{"status": 2, "description": "Updated description"}
//	err := app.Updates(ctx, db, updates)
//	if err != nil {
//	    log.Printf("Error updating app: %v", err)
//	    return
//	}
//	fmt.Println("App updated successfully")
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

// List retrieves all MgoApp documents that match the query.
//
// Parameters:
//   - ctx: A context.Context for the database operation.
//   - db: A pointer to the qmgo.Database to perform the operation on.
//
// Returns:
//   - []MgoApp: A slice of MgoApp structs containing the matching documents.
//   - error: An error if the operation fails, or nil on success.
//
// Example:
//
//	app := &MgoApp{Status: 1}
//	results, err := app.List(ctx, db)
//	if err != nil {
//	    log.Printf("Error listing apps: %v", err)
//	    return
//	}
//	for _, result := range results {
//	    fmt.Printf("Found app: %+v\n", result)
//	}
func (a *MgoApp) List(ctx context.Context, db *qmgo.Database) ([]MgoApp, error) {
	var apps []MgoApp

	err := db.Collection(a.CollectionName()).Find(ctx, a.buildQuery()).All(&apps)
	if err != nil {
		return nil, fmt.Errorf("find list failed: %w", err)
	}

	return apps, nil
}

// Creates inserts multiple MgoApp documents into the database.
//
// Parameters:
//   - ctx: A context.Context for the database operation.
//   - db: A pointer to the qmgo.Database to perform the operation on.
//   - apps: A slice of MgoApp structs to be inserted.
//
// Returns:
//   - []primitive.ObjectID: A slice of ObjectIDs for the inserted documents.
//   - error: An error if the operation fails, or nil on success.
//
// Example:
//
//	newApps := []MgoApp{
//	    {AppName: "App 1", AppID: "app1", Status: 1},
//	    {AppName: "App 2", AppID: "app2", Status: 1},
//	}
//	ids, err := (&MgoApp{}).Creates(ctx, db, newApps)
//	if err != nil {
//	    log.Printf("Error creating apps: %v", err)
//	    return
//	}
//	for _, id := range ids {
//	    fmt.Printf("Created app with ID: %s\n", id.Hex())
//	}
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

// Pagination retrieves a paginated list of MgoApp documents that match the query.
//
// Parameters:
//   - ctx: A context.Context for the database operation.
//   - db: A pointer to the qmgo.Database to perform the operation on.
//   - page: The page number (1-based) to retrieve.
//   - size: The number of documents per page.
//
// Returns:
//   - []MgoApp: A slice of MgoApp structs containing the matching documents for the specified page.
//   - error: An error if the operation fails, or nil on success.
//
// Example:
//
//	app := &MgoApp{Status: 1}
//	results, err := app.Pagination(ctx, db, 2, 10) // Get the second page with 10 items per page
//	if err != nil {
//	    log.Printf("Error retrieving paginated apps: %v", err)
//	    return
//	}
//	for _, result := range results {
//	    fmt.Printf("Found app: %+v\n", result)
//	}
func (a *MgoApp) Pagination(ctx context.Context, db *qmgo.Database, page, size int) ([]MgoApp, error) {
	var apps []MgoApp

	err := db.Collection(a.CollectionName()).Find(ctx, a.buildQuery()).Skip(int64((page - 1) * size)).Limit(int64(size)).All(&apps)
	if err != nil {
		return nil, fmt.Errorf("find with pagination failed: %w", err)
	}

	return apps, nil
}

// FindWithSort retrieves all MgoApp documents that match the query, sorted according to the provided sort string.
//
// Parameters:
//   - ctx: A context.Context for the database operation.
//   - db: A pointer to the qmgo.Database to perform the operation on.
//   - sort: A string specifying the sort order (e.g., "-created_at" for descending order by created_at field).
//
// Returns:
//   - []MgoApp: A slice of MgoApp structs containing the matching documents, sorted as specified.
//   - error: An error if the operation fails, or nil on success.
//
// Example:
//
//	app := &MgoApp{Status: 1}
//	results, err := app.FindWithSort(ctx, db, "-app_name") // Sort by app_name in descending order
//	if err != nil {
//	    log.Printf("Error finding sorted apps: %v", err)
//	    return
//	}
//	for _, result := range results {
//	    fmt.Printf("Found app: %+v\n", result)
//	}
func (a *MgoApp) FindWithSort(ctx context.Context, db *qmgo.Database, sort string) ([]MgoApp, error) {
	var apps []MgoApp

	err := db.Collection(a.CollectionName()).Find(ctx, a.buildQuery()).Sort(sort).All(&apps)
	if err != nil {
		return nil, fmt.Errorf("find with sort failed: %w", err)
	}

	return apps, nil
}

// Count returns the number of MgoApp documents that match the query.
//
// Parameters:
//   - ctx: A context.Context for the database operation.
//   - db: A pointer to the qmgo.Database to perform the operation on.
//
// Returns:
//   - int64: The count of matching documents.
//   - error: An error if the operation fails, or nil on success.
//
// Example:
//
//	app := &MgoApp{Status: 1}
//	count, err := app.Count(ctx, db)
//	if err != nil {
//	    log.Printf("Error counting apps: %v", err)
//	    return
//	}
func (a *MgoApp) Count(ctx context.Context, db *qmgo.Database) (int64, error) {
	count, err := db.Collection(a.CollectionName()).Find(ctx, a.buildQuery()).Count()
	if err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}

	return count, nil
}
