package system

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/seakee/go-api/app/model/system"
	"github.com/sk-pkg/logger"
	"gorm.io/gorm"
)

const (
	defaultPageSize = 10
	maxPageSize     = 100
)

// OperationRecordRepo defines the operation record repository interface.
type OperationRecordRepo interface {
	Pagination(ctx context.Context, operationRecord *system.OperationRecord, page, pageSize int) ([]system.OperationRecord, int64, error)
	Interaction(ctx context.Context, id string) (interface{}, error)
	Create(ctx context.Context, operationRecord *system.OperationRecord) error
}

type operationRecordRepo struct {
	db     *gorm.DB
	logger *logger.Manager
}

func (opr *operationRecordRepo) Pagination(ctx context.Context, operationRecord *system.OperationRecord, page, pageSize int) ([]system.OperationRecord, int64, error) {
	page, pageSize = normalizePageParams(page, pageSize)
	return operationRecord.Pagination(ctx, opr.db, page, pageSize)
}

func (opr *operationRecordRepo) Interaction(ctx context.Context, id string) (interface{}, error) {
	operationRecord := &system.OperationRecord{}

	record, err := operationRecord.FindByID(ctx, opr.db, id)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}

	interaction := struct {
		Params map[string]interface{} `json:"params"`
		Resp   map[string]interface{} `json:"resp"`
	}{}

	interaction.Params = parseParams(record.Params)
	interaction.Resp = parseResp(record.Resp)

	return interaction, nil
}

func (opr *operationRecordRepo) Create(ctx context.Context, operationRecord *system.OperationRecord) error {
	return operationRecord.Create(ctx, opr.db)
}

// NewOperationRecordRepo creates a new OperationRecordRepo instance.
func NewOperationRecordRepo(db *gorm.DB, logger *logger.Manager) OperationRecordRepo {
	return &operationRecordRepo{db: db, logger: logger}
}

// normalizePageParams normalizes page and page size parameters.
func normalizePageParams(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}

	switch {
	case pageSize > maxPageSize:
		pageSize = maxPageSize
	case pageSize <= 0:
		pageSize = defaultPageSize
	}

	return page, pageSize
}

func parseParams(params string) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(params), &result); err == nil {
		return result
	}

	if vals, err := url.ParseQuery(params); err == nil {
		result = make(map[string]interface{})
		for k, v := range vals {
			result[k] = strings.Join(v, ", ")
		}
		return result
	}

	return map[string]interface{}{"params": params}
}

func parseResp(resp string) map[string]interface{} {
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(resp), &result); err == nil {
		return result
	}

	return map[string]interface{}{"resp": resp}
}
