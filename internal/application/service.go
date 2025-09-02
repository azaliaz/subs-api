package application

import (
	"context"
	"github.com/azaliaz/subs-api/internal/storage"
	"github.com/google/uuid"
	"log/slog"
)
//go:generate mockgen -source=service.go -destination=./mocks/service_mock.go -package=mocks

type SubscriptionsService interface {
	Create(ctx context.Context, request *CreateRequest) (*CreateResponse, error)
	GetInfo(ctx context.Context, request *GetInfoRequest) (*GetInfoResponse, error)
	List(ctx context.Context, request *ListRequest) (*ListResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateRequest) (*UpdateResponse, error)
	Delete(ctx context.Context, request *DeleteRequest) (*DeleteResponse, error)
	GetTotalSubscriptionsPrice(ctx context.Context, request *TotalRequest) (*TotalResponse, error)
}

type CreateRequest struct {
	UserID      uuid.UUID `json:"user_id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date"`
}
type CreateResponse struct {
	ID uuid.UUID `json:"id"`
}

type GetInfoRequest struct {
	ID uuid.UUID `json:"id"`
}
type GetInfoResponse struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ServiceName string    `json:"service_name"`
	Price       int       `json:"price"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date"`
}

type ListRequest struct {
	UserID      *uuid.UUID `json:"user_id"`
	ServiceName *string    `json:"service_name"`
	From        *string    `json:"from"`
	To          *string    `json:"to"`
	Limit       *int       `json:"limit"`
	Offset      *int       `json:"offset"`
}
type ListResponse struct {
	Subscriptions []GetInfoResponse
}
type UpdateRequest struct {
	ServiceName *string `json:"service_name"`
	Price       *int    `json:"price"`
	StartDate   *string `json:"start_date"`
	EndDate     *string `json:"end_date"`
}

type UpdateResponse struct {
	Updated bool `json:"updated"`
}
type DeleteRequest struct {
	ID uuid.UUID `json:"id"`
}
type DeleteResponse struct {
	Deleted bool `json:"deleted"`
}
type TotalRequest struct {
	UserID      *uuid.UUID `json:"user_id"`
	ServiceName *string    `json:"service_name"`
	From        string     `json:"from"`
	To          string     `json:"to"`
}
type TotalResponse struct {
	Total int `json:"total"`
}

type Service struct {
	log    *slog.Logger
	config *Config
	db     storage.SubscriptionsStorage
}

func NewService(
	logger *slog.Logger,
	config *Config,
	db storage.SubscriptionsStorage,
) *Service {
	return &Service{
		log:    logger,
		config: config,
		db:     db,
	}
}

func (s *Service) Init() error {
	return nil
}

func (s *Service) Run(ctx context.Context) {

}

func (s *Service) Stop() {

}
