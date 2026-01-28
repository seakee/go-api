// Package system provides data models and database operations for operation records.
// This package is primarily used for recording and managing system operation logs,
// including creating, querying, updating, and deleting operation records.
package system

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/qiniu/qmgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Custom error definitions
var (
	// ErrNoValidFields indicates no valid fields for building query
	ErrNoValidFields = errors.New("no valid fields for building query")
	// ErrInvalidID indicates the provided ID is invalid
	ErrInvalidID = errors.New("invalid ID")
)

// OperationRecord defines the fields for an operation record.
type OperationRecord struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	CreateAt     time.Time          `json:"create_at"  bson:"create_at"`
	IP           string             `json:"ip" form:"ip" bson:"ip"`
	Method       string             `json:"method" form:"method" bson:"method"`
	Path         string             `json:"path" form:"path" bson:"path"`
	Status       int                `json:"status" form:"status" bson:"status"`
	Latency      float64            `json:"latency" form:"latency" bson:"latency"`
	Agent        string             `json:"agent" form:"agent" bson:"agent"`
	ErrorMessage string             `json:"error_message" form:"error_message" bson:"error_message"`
	UserID       uint               `json:"user_id" form:"user_id" bson:"user_id"`
	UserName     string             `json:"user_name" form:"user_name" bson:"user_name"`
	Params       string             `json:"params" form:"params" bson:"params"`
	Resp         string             `json:"resp" form:"resp" bson:"resp"`
	TraceID      string             `json:"trace_id" form:"trace_id" bson:"trace_id"`
}

// SetID sets the ID of the OperationRecord.
//
// Parameters:
//   - id: string format ID
//
// Returns:
//   - error: returns error if ID is invalid
//
// Example:
//
//	record := &OperationRecord{}
//	err := record.SetID("5f5e7e9b9b9b9b9b9b9b9b9b")
//	if err != nil {
//	    log.Fatal(err)
//	}
func (o *OperationRecord) SetID(id string) error {
	// Convert string ID to ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidID, err)
	}
	o.ID = objectID
	return nil
}

// GetID retrieves the ID string of the OperationRecord.
//
// Returns:
//   - string: string representation of the ID
//
// Example:
//
//	record := &OperationRecord{ID: primitive.NewObjectID()}
//	idStr := record.GetID()
//	fmt.Println(idStr)
func (o *OperationRecord) GetID() string {
	return o.ID.Hex()
}

// CollectionName returns the collection name of OperationRecord in MongoDB.
//
// Returns:
//   - string: collection name
func (o *OperationRecord) CollectionName() string {
	return "operation_record"
}

// buildQuery builds a BSON document for querying.
//
// Returns:
//   - bson.M: BSON document containing non-zero value fields
//
// Note:
// This method uses reflection to iterate through all struct fields,
// adding non-zero value fields to the query document.
func (o *OperationRecord) buildQuery() bson.M {
	query := bson.M{}

	v := reflect.ValueOf(o).Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := v.Type().Field(i)

		bsonTag := fieldType.Tag.Get("bson")
		if bsonTag == "" || bsonTag == "-" {
			continue
		}

		if field.IsValid() && !field.IsZero() {
			// Special handling for ID field
			if bsonTag == "_id" {
				query[bsonTag] = field.Interface().(primitive.ObjectID)
			} else {
				query[bsonTag] = field.Interface()
			}
		}
	}

	return query
}

// FindByID queries an operation record by ID.
//
// Parameters:
//   - ctx: context
//   - db: database connection
//   - id: operation record ID
//
// Returns:
//   - *OperationRecord: the found operation record
//   - error: returns error if query fails
func (o *OperationRecord) FindByID(ctx context.Context, db *qmgo.Database, id string) (*OperationRecord, error) {
	var result OperationRecord
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID format: %w", err)
	}
	err = db.Collection(o.CollectionName()).Find(ctx, bson.M{"_id": objectID}).One(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to find operation record: %w", err)
	}
	return &result, nil
}

// First queries the first operation record matching the criteria.
//
// Parameters:
//   - ctx: context
//   - db: database connection
//
// Returns:
//   - *OperationRecord: the found operation record
//   - error: returns error if query fails
//
// Example:
//
//	record := &OperationRecord{UserID: 123}
//	result, err := record.First(ctx, db)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Found record: %+v\n", result)
func (o *OperationRecord) First(ctx context.Context, db *qmgo.Database) (*OperationRecord, error) {
	var result OperationRecord
	// Execute query
	err := db.Collection(o.CollectionName()).Find(ctx, o.buildQuery()).One(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to find first operation record: %w", err)
	}
	return &result, nil
}

// Create creates a new operation record.
//
// Parameters:
//   - ctx: context
//   - db: database connection
//
// Returns:
//   - error: returns error if creation fails
//
// Example:
//
//	record := &OperationRecord{
//	    Method: "GET",
//	    Path: "/api/users",
//	    Status: 200,
//	    UserID: 123,
//	}
//	err := record.Create(ctx, db)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (o *OperationRecord) Create(ctx context.Context, db *qmgo.Database) error {
	// If ID is zero value, generate a new ID
	if o.ID.IsZero() {
		o.ID = primitive.NewObjectID()
	}
	// If CreateAt is zero value, set to current time
	if o.CreateAt.IsZero() {
		o.CreateAt = time.Now()
	}

	// Insert record
	_, err := db.Collection(o.CollectionName()).InsertOne(ctx, o)
	if err != nil {
		return fmt.Errorf("failed to create operation record: %w", err)
	}

	return nil
}

// Delete deletes operation records matching the criteria.
//
// Parameters:
//   - ctx: context
//   - db: database connection
//
// Returns:
//   - error: returns error if deletion fails
//
// Example:
//
//	record := &OperationRecord{UserID: 123}
//	err := record.Delete(ctx, db)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (o *OperationRecord) Delete(ctx context.Context, db *qmgo.Database) error {
	query := o.buildQuery()
	if len(query) == 0 {
		return ErrNoValidFields
	}

	// Execute delete operation
	err := db.Collection(o.CollectionName()).Remove(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete operation record: %w", err)
	}

	return nil
}

// Updates updates operation records matching the criteria.
//
// Parameters:
//   - ctx: context
//   - db: database connection
//   - updates: map of fields and values to update
//
// Returns:
//   - error: returns error if update fails
//
// Example:
//
//	record := &OperationRecord{UserID: 123}
//	updates := bson.M{"status": 404, "error_message": "Not Found"}
//	err := record.Updates(ctx, db, updates)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (o *OperationRecord) Updates(ctx context.Context, db *qmgo.Database, updates bson.M) error {
	query := o.buildQuery()
	if len(query) == 0 {
		return ErrNoValidFields
	}

	// Execute update operation
	err := db.Collection(o.CollectionName()).UpdateOne(ctx, query, bson.M{"$set": updates})
	if err != nil {
		return fmt.Errorf("failed to update operation record: %w", err)
	}

	return nil
}

// Pagination queries operation records with pagination.
//
// Parameters:
//   - ctx: context
//   - db: database connection
//   - page: page number (starting from 1)
//   - size: number of records per page
//
// Returns:
//   - []OperationRecord: list of found operation records
//   - int64: total record count
//   - error: returns error if query fails
//
// Example:
//
//	record := &OperationRecord{Status: 200}
//	results, total, err := record.Pagination(ctx, db, 1, 10)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Printf("Found %d records\n", total)
//	for _, r := range results {
//	    fmt.Printf("%+v\n", r)
//	}
func (o *OperationRecord) Pagination(ctx context.Context, db *qmgo.Database, page, size int) ([]OperationRecord, int64, error) {
	var results []OperationRecord
	query := o.buildQuery()

	// Get total record count
	total, err := db.Collection(o.CollectionName()).Find(ctx, query).Count()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count operation records: %w", err)
	}

	// Execute pagination query
	err = db.Collection(o.CollectionName()).Find(ctx, query).
		Skip(int64((page - 1) * size)).
		Limit(int64(size)).
		Sort("-create_at").
		All(&results)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to paginate operation records: %w", err)
	}

	return results, total, nil
}
