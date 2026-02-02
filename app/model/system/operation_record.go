// Package system provides SQL data models and database operations for operation records.
package system

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

// Custom error definitions.
var (
	// ErrNoValidFields indicates no valid fields for building query.
	ErrNoValidFields = errors.New("no valid fields for building query")
	// ErrInvalidID indicates the provided ID is invalid.
	ErrInvalidID = errors.New("invalid ID")
)

// OperationRecord defines the fields for an operation record.
type OperationRecord struct {
	gorm.Model

	IP           string  `gorm:"column:ip;type:varchar(50)" json:"ip"`
	Method       string  `gorm:"column:method;type:varchar(10)" json:"method"`
	Path         string  `gorm:"column:path;type:varchar(500)" json:"path"`
	Status       int     `gorm:"column:status;type:smallint" json:"status"`
	Latency      float64 `gorm:"column:latency;type:double precision" json:"latency"`
	Agent        string  `gorm:"column:agent;type:varchar(512)" json:"agent"`
	ErrorMessage string  `gorm:"column:error_message;type:text" json:"error_message"`
	UserID       uint    `gorm:"column:user_id" json:"user_id"`
	UserName     string  `gorm:"column:user_name;type:varchar(50)" json:"user_name"`
	Params       string  `gorm:"column:params;type:text" json:"params"`
	Resp         string  `gorm:"column:resp;type:text" json:"resp"`
	TraceID      string  `gorm:"column:trace_id;type:varchar(64)" json:"trace_id"`

	queryCondition string        `gorm:"-" json:"-"`
	queryArgs      []interface{} `gorm:"-" json:"-"`
}

// TableName specifies the table name for the OperationRecord model.
func (o *OperationRecord) TableName() string {
	return "sys_operation_record"
}

// Where sets query conditions for chaining with other methods.
func (o *OperationRecord) Where(query string, args ...interface{}) *OperationRecord {
	o.queryCondition = query
	o.queryArgs = args
	return o
}

// SetID sets the ID of the OperationRecord.
func (o *OperationRecord) SetID(id string) error {
	uintID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidID, err)
	}
	o.ID = uint(uintID)
	return nil
}

// GetID retrieves the ID string of the OperationRecord.
func (o *OperationRecord) GetID() string {
	return strconv.FormatUint(uint64(o.ID), 10)
}

// FindByID queries an operation record by ID.
func (o *OperationRecord) FindByID(ctx context.Context, db *gorm.DB, id string) (*OperationRecord, error) {
	uintID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrInvalidID, err)
	}

	var record OperationRecord
	if err := db.WithContext(ctx).First(&record, uint(uintID)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find by id failed: %w", err)
	}

	return &record, nil
}

// First queries the first operation record matching the criteria.
func (o *OperationRecord) First(ctx context.Context, db *gorm.DB) (*OperationRecord, error) {
	var record OperationRecord

	query := db.WithContext(ctx)
	if o.queryCondition != "" {
		query = query.Where(o.queryCondition, o.queryArgs...)
	} else {
		query = query.Where(o)
	}

	if err := query.First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find first failed: %w", err)
	}

	return &record, nil
}

// Create creates a new operation record.
func (o *OperationRecord) Create(ctx context.Context, db *gorm.DB) error {
	if err := db.WithContext(ctx).Create(o).Error; err != nil {
		return fmt.Errorf("create failed: %w", err)
	}
	return nil
}

// Delete deletes operation records matching the criteria.
func (o *OperationRecord) Delete(ctx context.Context, db *gorm.DB) error {
	if !o.hasCondition() {
		return ErrNoValidFields
	}

	query := db.WithContext(ctx)
	if o.queryCondition != "" {
		query = query.Where(o.queryCondition, o.queryArgs...)
	} else {
		query = query.Where(o)
	}

	if err := query.Delete(&OperationRecord{}).Error; err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	return nil
}

// Updates updates operation records matching the criteria.
func (o *OperationRecord) Updates(ctx context.Context, db *gorm.DB, updates map[string]interface{}) error {
	if !o.hasCondition() {
		return ErrNoValidFields
	}

	query := db.WithContext(ctx).Model(&OperationRecord{})
	if o.queryCondition != "" {
		query = query.Where(o.queryCondition, o.queryArgs...)
	} else if o.ID > 0 {
		query = query.Where("id = ?", o.ID)
	} else {
		query = query.Where(o)
	}

	if err := query.Updates(updates).Error; err != nil {
		return fmt.Errorf("updates failed: %w", err)
	}

	return nil
}

// Pagination queries operation records with pagination.
func (o *OperationRecord) Pagination(ctx context.Context, db *gorm.DB, page, size int) ([]OperationRecord, int64, error) {
	var records []OperationRecord

	query := db.WithContext(ctx)
	if o.queryCondition != "" {
		query = query.Where(o.queryCondition, o.queryArgs...)
	} else {
		query = query.Where(o)
	}

	var total int64
	if err := query.Model(&OperationRecord{}).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count failed: %w", err)
	}

	if err := query.Order("created_at desc").Offset((page-1)*size).Limit(size).Find(&records).Error; err != nil {
		return nil, 0, fmt.Errorf("paginate failed: %w", err)
	}

	return records, total, nil
}

func (o *OperationRecord) hasCondition() bool {
	if o.queryCondition != "" {
		return true
	}
	if o.ID > 0 {
		return true
	}

	switch {
	case o.IP != "":
		return true
	case o.Method != "":
		return true
	case o.Path != "":
		return true
	case o.Status != 0:
		return true
	case o.Latency != 0:
		return true
	case o.Agent != "":
		return true
	case o.ErrorMessage != "":
		return true
	case o.UserID != 0:
		return true
	case o.UserName != "":
		return true
	case o.Params != "":
		return true
	case o.Resp != "":
		return true
	case o.TraceID != "":
		return true
	default:
		return false
	}
}
