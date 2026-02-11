package system

import (
	"context"
	"encoding/json"
	"errors"
	repo "github.com/seakee/go-api/app/repository/system"
	"testing"
	"time"

	"github.com/seakee/go-api/app/model/system"
	"gorm.io/gorm"
)

// TestOperationRecordService_Create tests creating operation records.
func TestOperationRecordService_Create(t *testing.T) {
	tests := []struct {
		name    string
		record  *system.OperationRecord
		mockErr error
		wantErr bool
	}{
		{
			name: "create operation record successfully",
			record: &system.OperationRecord{
				IP:      "127.0.0.1",
				Method:  "GET",
				Path:    "/api/v1/users",
				Status:  200,
				Latency: 0.123,
				UserID:  1,
			},
			mockErr: nil,
			wantErr: false,
		},
		{
			name: "create failed",
			record: &system.OperationRecord{
				IP:     "127.0.0.1",
				Method: "POST",
				Path:   "/api/v1/users",
			},
			mockErr: errors.New("database error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOperationRecordRepo{
				CreateFunc: func(ctx context.Context, record *system.OperationRecord) error {
					return tt.mockErr
				},
			}

			svc := &operationRecordService{
				userRepo:   &mockUserRepo{},
				optRedRepo: mockRepo,
			}

			err := svc.Create(context.Background(), tt.record)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestOperationRecordService_Detail tests fetching operation record detail.
func TestOperationRecordService_Detail(t *testing.T) {
	tests := []struct {
		name       string
		recordID   string
		mockRecord *repo.OperationRecordDetail
		mockErr    error
		wantNil    bool
		wantErr    bool
		userName   string
	}{
		{
			name:     "get detail successfully",
			recordID: "1",
			mockRecord: &repo.OperationRecordDetail{
				Record: &system.OperationRecord{
					Model:  gorm.Model{ID: 1},
					Path:   "/go-api/internal/admin/system/user",
					UserID: 2,
				},
				Params: map[string]interface{}{"id": float64(1)},
				Resp:   map[string]interface{}{"code": float64(0)},
			},
			mockErr:  nil,
			wantNil:  false,
			wantErr:  false,
			userName: "tester",
		},
		{
			name:       "record not found",
			recordID:   "2",
			mockRecord: nil,
			mockErr:    nil,
			wantNil:    true,
			wantErr:    false,
		},
		{
			name:       "query error",
			recordID:   "3",
			mockRecord: nil,
			mockErr:    errors.New("database error"),
			wantNil:    true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOperationRecordRepo{
				DetailFunc: func(ctx context.Context, id string) (*repo.OperationRecordDetail, error) {
					return tt.mockRecord, tt.mockErr
				},
			}
			mockUser := &mockUserRepo{
				DetailByIDFunc: func(ctx context.Context, id uint) (*system.User, error) {
					if tt.userName == "" {
						return nil, nil
					}
					return &system.User{Model: gorm.Model{ID: id}, UserName: tt.userName}, nil
				},
			}

			svc := &operationRecordService{
				userRepo:   mockUser,
				optRedRepo: mockRepo,
			}

			detail, err := svc.Detail(context.Background(), tt.recordID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Detail() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (detail == nil) != tt.wantNil {
				t.Errorf("Detail() result nil = %v, wantNil %v", detail == nil, tt.wantNil)
			}
			if !tt.wantNil && detail.UserName != tt.userName {
				t.Errorf("Detail() UserName = %v, want %v", detail.UserName, tt.userName)
			}
		})
	}
}

// TestOperationRecordService_Paginate tests paginating operation records.
func TestOperationRecordService_Paginate(t *testing.T) {
	tests := []struct {
		name        string
		page        int
		size        int
		mockRecords []system.OperationRecord
		mockTotal   int64
		mockErr     error
		wantTotal   int64
		wantLen     int
		wantErr     bool
	}{
		{
			name: "get operation record list successfully",
			page: 1,
			size: 10,
			mockRecords: []system.OperationRecord{
				{
					Model: gorm.Model{
						ID:        1,
						CreatedAt: time.Now(),
					},
					IP:     "127.0.0.1",
					Method: "GET",
					Path:   "/api/v1/users",
					Status: 200,
				},
				{
					Model: gorm.Model{
						ID:        2,
						CreatedAt: time.Now(),
					},
					IP:     "127.0.0.1",
					Method: "POST",
					Path:   "/api/v1/users",
					Status: 201,
				},
			},
			mockTotal: 2,
			mockErr:   nil,
			wantTotal: 2,
			wantLen:   2,
			wantErr:   false,
		},
		{
			name:        "empty record list",
			page:        1,
			size:        10,
			mockRecords: []system.OperationRecord{},
			mockTotal:   0,
			mockErr:     nil,
			wantTotal:   0,
			wantLen:     0,
			wantErr:     false,
		},
		{
			name:        "query error",
			page:        1,
			size:        10,
			mockRecords: nil,
			mockTotal:   0,
			mockErr:     errors.New("database error"),
			wantTotal:   0,
			wantLen:     0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOperationRecordRepo{
				PaginationFunc: func(ctx context.Context, record *system.OperationRecord, page, pageSize int) ([]system.OperationRecord, int64, error) {
					return tt.mockRecords, tt.mockTotal, tt.mockErr
				},
			}
			mockUser := &mockUserRepo{
				ListByIDsFunc: func(ctx context.Context, ids []uint) ([]system.User, error) {
					users := make([]system.User, 0, len(ids))
					for _, id := range ids {
						users = append(users, system.User{Model: gorm.Model{ID: id}, UserName: "user"})
					}
					return users, nil
				},
			}

			svc := &operationRecordService{
				userRepo:   mockUser,
				optRedRepo: mockRepo,
			}

			list, total, err := svc.Paginate(context.Background(), &system.OperationRecord{}, tt.page, tt.size)

			if (err != nil) != tt.wantErr {
				t.Errorf("Paginate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if total != tt.wantTotal {
				t.Errorf("Paginate() total = %v, want %v", total, tt.wantTotal)
			}
			if len(list) != tt.wantLen {
				t.Errorf("Paginate() len = %v, want %v", len(list), tt.wantLen)
			}
		})
	}
}

func TestOperationRecordListItem_JSONTags(t *testing.T) {
	item := OperationRecordListItem{
		ID:       1,
		Method:   "GET",
		Path:     "/path",
		IP:       "127.0.0.1",
		Status:   0,
		UserName: "tester",
		TraceID:  "trace-1",
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	var m map[string]interface{}
	if err = json.Unmarshal(data, &m); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	wantKeys := []string{"id", "method", "path", "ip", "status", "user_name", "trace_id", "created_at"}
	for _, key := range wantKeys {
		if _, ok := m[key]; !ok {
			t.Errorf("missing key %q in json result: %s", key, string(data))
		}
	}

	notWantKeys := []string{"ID", "Method", "Path", "IP", "Status", "UserName", "TraceID", "CreatedAt"}
	for _, key := range notWantKeys {
		if _, ok := m[key]; ok {
			t.Errorf("unexpected key %q in json result: %s", key, string(data))
		}
	}
}
