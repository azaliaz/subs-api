package tests

import (
	"context"
	"errors"
	"fmt"
	"github.com/azaliaz/subs-api/internal/application"
	"github.com/azaliaz/subs-api/internal/storage"
	"github.com/azaliaz/subs-api/internal/storage/mocks"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"testing"
)

func TestCreateSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name string
		req  *application.CreateRequest
		want func(mockStorage *mocks.MockSubscriptionsStorage) (*application.CreateResponse, error)
	}{
		{
			name: "success",
			req: &application.CreateRequest{
				UserID:      uuid.New(),
				ServiceName: "Netflix",
				Price:       10,
				StartDate:   "09-2025",
				EndDate:     func() *string { s := "12-2025"; return &s }(),
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.CreateResponse, error) {
				id := uuid.New()
				mockStorage.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(&storage.CreateResponse{ID: id}, nil)
				return &application.CreateResponse{ID: id}, nil
			},
		},
		{
			name: "nil request",
			req:  nil,
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.CreateResponse, error) {
				return nil, errors.New("request cannot be nil")
			},
		},
		{
			name: "nil UserID",
			req: &application.CreateRequest{
				UserID:      uuid.Nil,
				ServiceName: "Netflix",
				Price:       10,
				StartDate:   "09-2025",
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.CreateResponse, error) {
				return nil, errors.New("user_id is required")
			},
		},
		{
			name: "empty ServiceName",
			req: &application.CreateRequest{
				UserID:      uuid.New(),
				ServiceName: "",
				Price:       10,
				StartDate:   "09-2025",
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.CreateResponse, error) {
				return nil, errors.New("service_name is required")
			},
		},
		{
			name: "price <= 0",
			req: &application.CreateRequest{
				UserID:      uuid.New(),
				ServiceName: "Netflix",
				Price:       0,
				StartDate:   "09-2025",
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.CreateResponse, error) {
				return nil, errors.New("price must be greater than 0")
			},
		},
		{
			name: "invalid StartDate",
			req: &application.CreateRequest{
				UserID:      uuid.New(),
				ServiceName: "Netflix",
				Price:       10,
				StartDate:   "2025-09",
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.CreateResponse, error) {
				return nil, fmt.Errorf("invalid start_date format, expected MM-YYYY")
			},
		},
		{
			name: "invalid EndDate",
			req: &application.CreateRequest{
				UserID:      uuid.New(),
				ServiceName: "Netflix",
				Price:       10,
				StartDate:   "09-2025",
				EndDate:     func() *string { s := "2025-12"; return &s }(),
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.CreateResponse, error) {
				return nil, fmt.Errorf("invalid end_date format, expected MM-YYYY")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockSubscriptionsStorage(ctrl)
			wantResp, wantErr := tt.want(mockStorage)

			svc := application.NewService(
				slog.Default(),
				&application.Config{Secret: "test"},
				mockStorage,
			)

			got, err := svc.Create(context.Background(), tt.req)

			if wantErr != nil {
				assert.Error(t, err)

				switch tt.name {
				case "invalid StartDate":
					assert.ErrorContains(t, err, "invalid start_date format, expected MM-YYYY")
				case "invalid EndDate":
					assert.ErrorContains(t, err, "invalid end_date format, expected MM-YYYY")
				}

				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, wantResp, got)
			}

		})
	}
}

func TestGetInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name string
		req  *application.GetInfoRequest
		want func(mockStorage *mocks.MockSubscriptionsStorage) (*application.GetInfoResponse, error)
	}{
		{
			name: "success",
			req:  &application.GetInfoRequest{ID: validID},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.GetInfoResponse, error) {
				mockStorage.EXPECT().
					GetInfo(gomock.Any(), validID).
					Return(&storage.GetInfoResponse{
						ID:          validID,
						UserID:      userID,
						ServiceName: "Netflix",
						Price:       10,
						StartDate:   "09-2025",
						EndDate:     func() *string { s := "12-2025"; return &s }(),
					}, nil)
				return &application.GetInfoResponse{
					ID:          validID,
					UserID:      userID,
					ServiceName: "Netflix",
					Price:       10,
					StartDate:   "09-2025",
					EndDate:     func() *string { s := "12-2025"; return &s }(),
				}, nil
			},
		},
		{
			name: "nil request",
			req:  nil,
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.GetInfoResponse, error) {
				return nil, errors.New("request cannot be nil")
			},
		},
		{
			name: "invalid ID",
			req:  &application.GetInfoRequest{ID: uuid.Nil},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.GetInfoResponse, error) {
				return nil, errors.New("id is required")
			},
		},
		{
			name: "not found",
			req:  &application.GetInfoRequest{ID: validID},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.GetInfoResponse, error) {
				mockStorage.EXPECT().
					GetInfo(gomock.Any(), validID).
					Return(nil, nil)
				return nil, nil
			},
		},
		{
			name: "storage error",
			req:  &application.GetInfoRequest{ID: validID},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.GetInfoResponse, error) {
				mockStorage.EXPECT().
					GetInfo(gomock.Any(), validID).
					Return(nil, fmt.Errorf("db error"))
				return nil, fmt.Errorf("failed to get subscription info: %w", fmt.Errorf("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockSubscriptionsStorage(ctrl)
			wantResp, wantErr := tt.want(mockStorage)

			svc := application.NewService(
				slog.Default(),
				&application.Config{Secret: "test"},
				mockStorage,
			)

			got, err := svc.GetInfo(context.Background(), tt.req)

			if wantErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, wantErr.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, wantResp, got)
			}
		})
	}
}

func TestListSubscriptions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	serviceName := "Netflix"
	from := "09-2025"
	to := "12-2025"
	limit := 10
	offset := 0

	tests := []struct {
		name string
		req  *application.ListRequest
		want func(mockStorage *mocks.MockSubscriptionsStorage) (*application.ListResponse, error)
	}{
		{
			name: "success",
			req: &application.ListRequest{
				UserID:      &userID,
				ServiceName: &serviceName,
				From:        &from,
				To:          &to,
				Limit:       &limit,
				Offset:      &offset,
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.ListResponse, error) {
				mockStorage.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(&storage.ListResponse{
						Subscriptions: []storage.GetInfoResponse{
							{
								ID:          uuid.New(),
								UserID:      userID,
								ServiceName: serviceName,
								Price:       10,
								StartDate:   from,
								EndDate:     &to,
							},
						},
					}, nil)
				return &application.ListResponse{
					Subscriptions: []application.GetInfoResponse{
						{
							ID:          uuid.Nil,
							UserID:      userID,
							ServiceName: serviceName,
							Price:       10,
							StartDate:   from,
							EndDate:     &to,
						},
					},
				}, nil
			},
		},
		{
			name: "nil request",
			req:  nil,
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.ListResponse, error) {
				return nil, errors.New("request cannot be nil")
			},
		},
		{
			name: "invalid date range",
			req: &application.ListRequest{
				From: func() *string { s := "2025-09"; return &s }(),
				To:   func() *string { s := "2025-12"; return &s }(),
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.ListResponse, error) {
				return nil, fmt.Errorf("invalid start_date format, expected MM-YYYY: parsing time \"2025-09\": month out of range")
			},
		},
		{
			name: "storage error",
			req: &application.ListRequest{
				UserID:      &userID,
				ServiceName: &serviceName,
				From:        &from,
				To:          &to,
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.ListResponse, error) {
				mockStorage.EXPECT().
					List(gomock.Any(), gomock.Any()).
					Return(nil, fmt.Errorf("db error"))
				return nil, fmt.Errorf("failed to list subscriptions: %w", fmt.Errorf("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockSubscriptionsStorage(ctrl)
			wantResp, wantErr := tt.want(mockStorage)

			svc := application.NewService(
				slog.Default(),
				&application.Config{Secret: "test"},
				mockStorage,
			)

			got, err := svc.List(context.Background(), tt.req)

			if wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), wantErr.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(wantResp.Subscriptions), len(got.Subscriptions))
				if len(wantResp.Subscriptions) > 0 {
					assert.Equal(t, wantResp.Subscriptions[0].UserID, got.Subscriptions[0].UserID)
					assert.Equal(t, wantResp.Subscriptions[0].ServiceName, got.Subscriptions[0].ServiceName)
					assert.Equal(t, wantResp.Subscriptions[0].Price, got.Subscriptions[0].Price)
					assert.Equal(t, wantResp.Subscriptions[0].StartDate, got.Subscriptions[0].StartDate)
					assert.Equal(t, wantResp.Subscriptions[0].EndDate, got.Subscriptions[0].EndDate)
				}
			}
		})
	}
}

func TestUpdateSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validID := uuid.New()
	validPrice := 100
	validStart := "09-2025"
	validEnd := "12-2025"
	serviceName := "Netflix"

	tests := []struct {
		name string
		id   uuid.UUID
		req  *application.UpdateRequest
		want func(mockStorage *mocks.MockSubscriptionsStorage) (*application.UpdateResponse, error)
	}{
		{
			name: "success",
			id:   validID,
			req: &application.UpdateRequest{
				ServiceName: &serviceName,
				Price:       &validPrice,
				StartDate:   &validStart,
				EndDate:     &validEnd,
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.UpdateResponse, error) {
				mockStorage.EXPECT().
					Update(gomock.Any(), validID, gomock.Any()).
					Return(&storage.UpdateResponse{Updated: true}, nil)
				return &application.UpdateResponse{Updated: true}, nil
			},
		},
		{
			name: "nil request",
			id:   validID,
			req:  nil,
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.UpdateResponse, error) {
				return nil, errors.New("request cannot be nil")
			},
		},
		{
			name: "invalid price <= 0",
			id:   validID,
			req: &application.UpdateRequest{
				Price: func() *int { i := 0; return &i }(),
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.UpdateResponse, error) {
				return nil, errors.New("price must be greater than 0")
			},
		},
		{
			name: "invalid StartDate format",
			id:   validID,
			req: &application.UpdateRequest{
				StartDate: func() *string { s := "2025-09"; return &s }(),
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.UpdateResponse, error) {
				return nil, fmt.Errorf("invalid start_date format")
			},
		},
		{
			name: "storage error",
			id:   validID,
			req: &application.UpdateRequest{
				ServiceName: &serviceName,
				Price:       &validPrice,
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.UpdateResponse, error) {
				mockStorage.EXPECT().
					Update(gomock.Any(), validID, gomock.Any()).
					Return(nil, errors.New("db error"))
				return nil, fmt.Errorf("update request")
			},
		},
		{
			name: "update not found (updated=false)",
			id:   validID,
			req: &application.UpdateRequest{
				Price: &validPrice,
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.UpdateResponse, error) {
				mockStorage.EXPECT().
					Update(gomock.Any(), validID, gomock.Any()).
					Return(&storage.UpdateResponse{Updated: false}, nil)
				return &application.UpdateResponse{Updated: false}, nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockSubscriptionsStorage(ctrl)
			wantResp, wantErr := tt.want(mockStorage)

			svc := application.NewService(
				slog.Default(),
				&application.Config{Secret: "test"},
				mockStorage,
			)

			got, err := svc.Update(context.Background(), tt.id, tt.req)

			if wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), wantErr.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, wantResp, got)
			}
		})
	}
}

func TestDeleteSubscription(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	validID := uuid.New()

	tests := []struct {
		name string
		req  *application.DeleteRequest
		want func(mockStorage *mocks.MockSubscriptionsStorage) (*application.DeleteResponse, error)
	}{
		{
			name: "success",
			req:  &application.DeleteRequest{ID: validID},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.DeleteResponse, error) {
				mockStorage.EXPECT().
					Delete(gomock.Any(), &storage.DeleteRequest{ID: validID}).
					Return(nil)
				return &application.DeleteResponse{Deleted: true}, nil
			},
		},
		{
			name: "nil request",
			req:  nil,
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.DeleteResponse, error) {
				return nil, errors.New("request cannot be nil")
			},
		},
		{
			name: "invalid ID",
			req:  &application.DeleteRequest{ID: uuid.Nil},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.DeleteResponse, error) {
				return nil, errors.New("id is required")
			},
		},
		{
			name: "storage error",
			req:  &application.DeleteRequest{ID: validID},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.DeleteResponse, error) {
				mockStorage.EXPECT().
					Delete(gomock.Any(), &storage.DeleteRequest{ID: validID}).
					Return(errors.New("db error"))
				return nil, fmt.Errorf("delete request")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockSubscriptionsStorage(ctrl)
			wantResp, wantErr := tt.want(mockStorage)

			svc := application.NewService(
				slog.Default(),
				&application.Config{Secret: "test"},
				mockStorage,
			)

			got, err := svc.Delete(context.Background(), tt.req)

			if wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), wantErr.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, wantResp, got)
			}
		})
	}
}

func TestGetTotalSubscriptionsPrice(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userID := uuid.New()
	serviceName := "Netflix"
	from := "09-2025"
	to := "12-2025"

	tests := []struct {
		name string
		req  *application.TotalRequest
		want func(mockStorage *mocks.MockSubscriptionsStorage) (*application.TotalResponse, error)
	}{
		{
			name: "success",
			req: &application.TotalRequest{
				UserID:      &userID,
				ServiceName: &serviceName,
				From:        from,
				To:          to,
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.TotalResponse, error) {
				mockStorage.EXPECT().
					GetTotalSubscriptionsPrice(gomock.Any(), gomock.Any()).
					Return(50, nil)
				return &application.TotalResponse{Total: 50}, nil
			},
		},
		{
			name: "nil request",
			req:  nil,
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.TotalResponse, error) {
				return nil, errors.New("request cannot be nil")
			},
		},
		{
			name: "storage error",
			req: &application.TotalRequest{
				UserID:      &userID,
				ServiceName: &serviceName,
				From:        from,
				To:          to,
			},
			want: func(mockStorage *mocks.MockSubscriptionsStorage) (*application.TotalResponse, error) {
				mockStorage.EXPECT().
					GetTotalSubscriptionsPrice(gomock.Any(), gomock.Any()).
					Return(0, fmt.Errorf("db error"))
				return nil, fmt.Errorf("get total subscriptions price: %w", fmt.Errorf("db error"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := mocks.NewMockSubscriptionsStorage(ctrl)
			wantResp, wantErr := tt.want(mockStorage)

			svc := application.NewService(
				slog.Default(),
				&application.Config{Secret: "test"},
				mockStorage,
			)

			got, err := svc.GetTotalSubscriptionsPrice(context.Background(), tt.req)

			if wantErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), wantErr.Error())
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, wantResp.Total, got.Total)
			}
		})
	}
}
