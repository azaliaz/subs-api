package storage

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"log/slog"
)

//go:generate mockgen -source=service.go -destination=./mocks/service_mock.go -package=mocks

type SubscriptionsStorage interface {
	Create(ctx context.Context, request *CreateRequest) (*CreateResponse, error)
	GetInfo(ctx context.Context, id uuid.UUID) (*GetInfoResponse, error)
	List(ctx context.Context, request *ListRequest) (*ListResponse, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateRequest) (*UpdateResponse, error)
	Delete(ctx context.Context, request *DeleteRequest) error
	GetTotalSubscriptionsPrice(ctx context.Context, request *TotalRequest) (int, error)
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
type TotalRequest struct {
	UserID      *uuid.UUID `json:"user_id"`
	ServiceName *string    `json:"service_name"`
	From        string     `json:"from"`
	To          string     `json:"to"`
}

func NewService(db *DB, logger *slog.Logger) *Service {
	return &Service{DB: db, logger: logger}
}

type Service struct {
	*DB
	logger *slog.Logger
}

func NewDB(config *Config, logEntry *slog.Logger) *DB {
	return &DB{
		config: config,
		log:    logEntry,
	}
}

type DB struct {
	config *Config
	log    *slog.Logger
	pool   *pgxpool.Pool
	cancel func()
}

func (r *DB) Init() error {
	ctx, cancel := context.WithCancel(context.Background())
	r.cancel = cancel

	poolCfg, err := pgxpool.ParseConfig(r.config.dsnPostgres(r.log))
	if err != nil {
		return fmt.Errorf("error on parsing rw storage config: %w", err)
	}

	poolCfg.MaxConns = r.config.MaxOpenConns
	poolCfg.MaxConnIdleTime = r.config.ConnIdleLifetime
	poolCfg.MaxConnLifetime = r.config.ConnMaxLifetime

	pool, err := pgxpool.ConnectConfig(ctx, poolCfg)
	if err != nil {
		return fmt.Errorf("error on creating rw storage connection pool: %w", err)
	}

	r.pool = pool

	r.log.Info("connected to postgres")
	return nil
}

func (r *DB) Run(_ context.Context) {
}

func (r *DB) Stop() {
	r.log.Info("stopping storage service")
	if r.cancel != nil {
		r.cancel()
	}
	if r.pool != nil {
		r.pool.Close()
		r.log.Info("storage service has been stopped")
	} else {
		r.log.Warn("storage pool is nil, nothing to close")
	}
}

func (r *DB) Pool() *pgxpool.Pool {
	return r.pool
}
