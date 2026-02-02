package system

import (
	"context"
	"errors"
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
				IP:       "127.0.0.1",
				Method:   "GET",
				Path:     "/api/v1/users",
				Status:   200,
				Latency:  0.123,
				UserID:   1,
				UserName: "admin",
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
				optRedRepo: mockRepo,
			}

			err := svc.Create(context.Background(), tt.record)

			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
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

			svc := &operationRecordService{
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

// TestOperationRecordService_Interaction tests fetching operation record interaction detail.
func TestOperationRecordService_Interaction(t *testing.T) {
	tests := []struct {
		name       string
		recordID   string
		mockResult any
		mockErr    error
		wantNil    bool
		wantErr    bool
	}{
		{
			name:     "get interaction detail successfully",
			recordID: "1",
			mockResult: map[string]any{
				"params": `{"id": 1}`,
				"resp":   `{"code": 0, "message": "success"}`,
			},
			mockErr: nil,
			wantNil: false,
			wantErr: false,
		},
		{
			name:       "record not found",
			recordID:   "2",
			mockResult: nil,
			mockErr:    nil,
			wantNil:    true,
			wantErr:    false,
		},
		{
			name:       "query error",
			recordID:   "3",
			mockResult: nil,
			mockErr:    errors.New("database error"),
			wantNil:    true,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockOperationRecordRepo{
				InteractionFunc: func(ctx context.Context, id string) (any, error) {
					return tt.mockResult, tt.mockErr
				},
			}

			svc := &operationRecordService{
				optRedRepo: mockRepo,
			}

			result, err := svc.Interaction(context.Background(), tt.recordID)

			if (err != nil) != tt.wantErr {
				t.Errorf("Interaction() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (result == nil) != tt.wantNil {
				t.Errorf("Interaction() result nil = %v, wantNil %v", result == nil, tt.wantNil)
			}
		})
	}
}
