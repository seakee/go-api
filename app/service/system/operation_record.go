package system

import (
	"context"
	"time"

	"github.com/seakee/go-api/app/model/system"
	repo "github.com/seakee/go-api/app/repository/system"
	"github.com/sk-pkg/logger"
	"github.com/sk-pkg/redis"
	"gorm.io/gorm"
)

// OperationRecordService defines the operation record service interface.
type OperationRecordService interface {
	Create(ctx context.Context, operationRecord *system.OperationRecord) error
	Paginate(ctx context.Context, operationRecord *system.OperationRecord, page, size int) (list []OperationRecordListItem, total int64, err error)
	Detail(ctx context.Context, id string) (*OperationRecordDetail, error)
}

type OperationRecordListItem struct {
	ID        uint      `json:"id"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	IP        string    `json:"ip"`
	Status    int       `json:"status"`
	UserName  string    `json:"user_name"`
	TraceID   string    `json:"trace_id"`
	CreatedAt time.Time `json:"created_at"`
}

type OperationRecordDetail struct {
	ID           uint                   `json:"id"`
	Method       string                 `json:"method"`
	Path         string                 `json:"path"`
	IP           string                 `json:"ip"`
	Status       int                    `json:"status"`
	UserID       uint                   `json:"user_id"`
	UserName     string                 `json:"user_name"`
	TraceID      string                 `json:"trace_id"`
	CreatedAt    time.Time              `json:"created_at"`
	Latency      float64                `json:"latency"`
	Agent        string                 `json:"agent"`
	ErrorMessage string                 `json:"error_message"`
	Params       map[string]interface{} `json:"params"`
	Resp         map[string]interface{} `json:"resp"`
}

type operationRecordService struct {
	redis      *redis.Manager
	logger     *logger.Manager
	userRepo   repo.UserRepo
	optRedRepo repo.OperationRecordRepo
}

func (o operationRecordService) Create(ctx context.Context, operationRecord *system.OperationRecord) error {
	return o.optRedRepo.Create(ctx, operationRecord)
}

func (o operationRecordService) Paginate(ctx context.Context, operationRecord *system.OperationRecord, page, size int) (list []OperationRecordListItem, total int64, err error) {
	records, total, err := o.optRedRepo.Pagination(ctx, operationRecord, page, size)
	if err != nil {
		return nil, 0, err
	}

	userNameMap, err := o.userNameMapByRecords(ctx, records)
	if err != nil {
		return nil, 0, err
	}

	list = make([]OperationRecordListItem, len(records))
	for i := range records {
		list[i] = OperationRecordListItem{
			ID:        records[i].ID,
			Method:    records[i].Method,
			Path:      records[i].Path,
			IP:        records[i].IP,
			Status:    records[i].Status,
			UserName:  userNameMap[records[i].UserID],
			TraceID:   records[i].TraceID,
			CreatedAt: records[i].CreatedAt,
		}
	}

	return list, total, nil
}

func (o operationRecordService) Detail(ctx context.Context, id string) (*OperationRecordDetail, error) {
	data, err := o.optRedRepo.Detail(ctx, id)
	if err != nil {
		return nil, err
	}
	if data == nil || data.Record == nil {
		return nil, nil
	}

	userName := ""
	if data.Record.UserID > 0 {
		user, userErr := o.userRepo.DetailByID(ctx, data.Record.UserID)
		if userErr != nil {
			return nil, userErr
		}
		if user != nil {
			userName = user.UserName
		}
	}

	return &OperationRecordDetail{
		ID:           data.Record.ID,
		Method:       data.Record.Method,
		Path:         data.Record.Path,
		IP:           data.Record.IP,
		Status:       data.Record.Status,
		UserID:       data.Record.UserID,
		UserName:     userName,
		TraceID:      data.Record.TraceID,
		CreatedAt:    data.Record.CreatedAt,
		Latency:      data.Record.Latency,
		Agent:        data.Record.Agent,
		ErrorMessage: data.Record.ErrorMessage,
		Params:       data.Params,
		Resp:         data.Resp,
	}, nil
}

// NewOperationRecordService creates a new OperationRecordService instance.
func NewOperationRecordService(redis *redis.Manager, logger *logger.Manager, db *gorm.DB) OperationRecordService {
	return &operationRecordService{
		redis:      redis,
		logger:     logger,
		userRepo:   repo.NewUserRepo(db, redis, logger),
		optRedRepo: repo.NewOperationRecordRepo(db, logger),
	}
}

func (o operationRecordService) userNameMapByRecords(ctx context.Context, records []system.OperationRecord) (map[uint]string, error) {
	ids := make([]uint, 0, len(records))
	idSet := make(map[uint]struct{}, len(records))
	for i := range records {
		if records[i].UserID == 0 {
			continue
		}
		if _, ok := idSet[records[i].UserID]; ok {
			continue
		}
		idSet[records[i].UserID] = struct{}{}
		ids = append(ids, records[i].UserID)
	}

	if len(ids) == 0 {
		return map[uint]string{}, nil
	}

	users, err := o.userRepo.ListByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	userNameMap := make(map[uint]string, len(users))
	for i := range users {
		userNameMap[users[i].ID] = users[i].UserName
	}

	return userNameMap, nil
}
