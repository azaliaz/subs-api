package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/azaliaz/subs-api/internal/storage"
	"github.com/google/uuid"
	"time"
)

func validateDates(start, end *string) error {
	var startTime, endTime time.Time
	var err error

	if start != nil {
		startTime, err = time.Parse("01-2006", *start)
		if err != nil {
			return fmt.Errorf("invalid start_date format, expected MM-YYYY: %w", err)
		}
	}

	if end != nil {
		endTime, err = time.Parse("01-2006", *end)
		if err != nil {
			return fmt.Errorf("invalid end_date format, expected MM-YYYY: %w", err)
		}
	}

	if start != nil && end != nil && endTime.Before(startTime) {
		return errors.New("end_date cannot be before start_date")
	}

	return nil
}

func (s *Service) Create(ctx context.Context, request *CreateRequest) (*CreateResponse, error) {
	if request == nil {
		s.log.Warn("request is nil in application layer")
		return nil, errors.New("request cannot be nil")
	}

	if request.UserID == uuid.Nil {
		return nil, errors.New("user_id is required")
	}
	if request.ServiceName == "" {
		return nil, errors.New("service_name is required")
	}
	if request.Price <= 0 {
		return nil, errors.New("price must be greater than 0")
	}
	if err := validateDates(&request.StartDate, request.EndDate); err != nil {
		s.log.Warn("invalid start date format in application layer", "error", err)
		return nil, err
	}
	resp, err := s.db.Create(ctx, &storage.CreateRequest{
		UserID:      request.UserID,
		ServiceName: request.ServiceName,
		Price:       request.Price,
		StartDate:   request.StartDate,
		EndDate:     request.EndDate,
	})
	if err != nil {
		s.log.Error("failed to create subscription in storage layer", "error", err)
		return nil, fmt.Errorf("create request: %w", err)
	}
	return &CreateResponse{ID: resp.ID}, nil

}

func (s *Service) GetInfo(ctx context.Context, request *GetInfoRequest) (*GetInfoResponse, error) {
	if request == nil {
		s.log.Warn("request is nil in application layer")
		return nil, errors.New("request cannot be nil")
	}

	if request.ID == uuid.Nil {
		s.log.Warn("invalid ID in application layer")
		return nil, errors.New("id is required")
	}

	resp, err := s.db.GetInfo(ctx, request.ID)
	if err != nil {
		s.log.Error("failed to get info in storage layer", "error", err)
		return nil, fmt.Errorf("failed to get subscription info: %w", err)
	}
	if resp == nil {
		return nil, nil
	}

	return &GetInfoResponse{
		ID:          resp.ID,
		UserID:      resp.UserID,
		ServiceName: resp.ServiceName,
		Price:       resp.Price,
		StartDate:   resp.StartDate,
		EndDate:     resp.EndDate,
	}, nil
}

func (s *Service) List(ctx context.Context, request *ListRequest) (*ListResponse, error) {
	if request == nil {
		s.log.Warn("request is nil in application layer")
		return nil, errors.New("request cannot be nil")
	}

	if err := validateDates(request.From, request.To); err != nil {
		s.log.Warn("invalid date range in application layer", "error", err)
		return nil, err
	}
	storageResp, err := s.db.List(ctx, &storage.ListRequest{
		UserID:      request.UserID,
		ServiceName: request.ServiceName,
		From:        request.From,
		To:          request.To,
		Limit:       request.Limit,
		Offset:      request.Offset,
	})
	if err != nil {
		s.log.Error("failed to list subscriptions in storage layer", "error", err)
		return nil, fmt.Errorf("failed to list subscriptions: %w", err)
	}

	var appResp ListResponse
	for _, sub := range storageResp.Subscriptions {
		appResp.Subscriptions = append(appResp.Subscriptions, GetInfoResponse{
			ID:          sub.ID,
			UserID:      sub.UserID,
			ServiceName: sub.ServiceName,
			Price:       sub.Price,
			StartDate:   sub.StartDate,
			EndDate:     sub.EndDate,
		})
	}

	return &appResp, nil
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, request *UpdateRequest) (*UpdateResponse, error) {
	if request == nil {
		s.log.Warn("request is nil in application layer")
		return nil, errors.New("request cannot be nil")
	}

	if request.Price != nil && *request.Price <= 0 {
		s.log.Warn("price must be greater than 0 in application layer")
		return nil, errors.New("price must be greater than 0")
	}
	if err := validateDates(request.StartDate, request.EndDate); err != nil {
		s.log.Warn("invalid date range in application layer", "error", err)
		return nil, err
	}
	resp, err := s.db.Update(ctx, id, &storage.UpdateRequest{
		ServiceName: request.ServiceName,
		Price:       request.Price,
		StartDate:   request.StartDate,
		EndDate:     request.EndDate,
	})
	if err != nil {
		s.log.Error("failed to update subscription in storage layer", "error", err)
		return nil, fmt.Errorf("update request: %w", err)
	}

	return &UpdateResponse{Updated: resp.Updated}, nil
}

func (s *Service) Delete(ctx context.Context, request *DeleteRequest) (*DeleteResponse, error) {
	if request == nil {
		s.log.Warn("request is nil in application layer")
		return nil, errors.New("request cannot be nil")
	}

	if request.ID == uuid.Nil {
		s.log.Warn("invalid ID in application layer")
		return nil, errors.New("id is required")
	}
	err := s.db.Delete(ctx, &storage.DeleteRequest{
		ID: request.ID,
	})
	if err != nil {
		s.log.Error("failed to delete subscription in storage layer", "error", err)
		return nil, fmt.Errorf("delete request: %w", err)
	}

	return &DeleteResponse{Deleted: true}, nil
}

func (s *Service) GetTotalSubscriptionsPrice(ctx context.Context, request *TotalRequest) (*TotalResponse, error) {
	if request == nil {
		s.log.Warn("request is nil in application layer")
		return nil, errors.New("request cannot be nil")
	}

	if err := validateDates(&request.From, &request.To); err != nil {
		s.log.Warn("invalid date range in application layer", "error", err)
		return nil, err
	}

	total, err := s.db.GetTotalSubscriptionsPrice(ctx, &storage.TotalRequest{
		UserID:      request.UserID,
		ServiceName: request.ServiceName,
		From:        request.From,
		To:          request.To,
	})
	if err != nil {
		s.log.Error("failed to get total subscriptions price", "error", err)
		return nil, fmt.Errorf("get total subscriptions price: %w", err)
	}

	return &TotalResponse{Total: total}, nil
}
