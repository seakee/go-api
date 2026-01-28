package system

import (
	"context"
	"github.com/qiniu/qmgo"
	"github.com/seakee/go-api/app/model/system"
	repo "github.com/seakee/go-api/app/repository/system"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
)

// OperationRecordService defines the operation record service interface.
type OperationRecordService interface {
	Create(ctx context.Context, operationRecord *system.OperationRecord) error
	Paginate(ctx context.Context, operationRecord *system.OperationRecord, page, size int) (list []system.OperationRecord, total int64, err error)
	Interaction(ctx context.Context, id string) (any, error)
}

type operationRecordService struct {
	redis      *redis.Manager
	logger     *logger.Manager
	optRedRepo repo.OperationRecordRepo
}

func (o operationRecordService) Create(ctx context.Context, operationRecord *system.OperationRecord) error {
	return o.optRedRepo.Create(ctx, operationRecord)
}

func (o operationRecordService) Paginate(ctx context.Context, operationRecord *system.OperationRecord, page, size int) (list []system.OperationRecord, total int64, err error) {
	return o.optRedRepo.Pagination(ctx, operationRecord, page, size)
}

func (o operationRecordService) Interaction(ctx context.Context, id string) (any, error) {
	return o.optRedRepo.Interaction(ctx, id)
}

// NewOperationRecordService creates a new OperationRecordService instance.
func NewOperationRecordService(redis *redis.Manager, logger *logger.Manager, db *qmgo.Database) OperationRecordService {
	return &operationRecordService{
		redis:      redis,
		logger:     logger,
		optRedRepo: repo.NewOperationRecordRepo(db, logger),
	}
}
