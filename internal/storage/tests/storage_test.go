package tests

import (
	"context"
	"fmt"
	"github.com/azaliaz/subs-api/internal/storage"
	"github.com/azaliaz/subs-api/migrations"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"log/slog"
	"strconv"
	"strings"
	"testing"
	"time"
)

func (s *RepositoryTestSuite) TestCreateSubscription() {
	ctx := context.Background()

	userID := uuid.New()
	serviceName := "Netflix"
	price := 10
	startDate := "09-2025"
	endDate := "12-2025"

	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM subscriptions WHERE user_id = $1`, userID)
		require.NoError(s.T(), err)
	}
	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM subscriptions WHERE user_id = $1`, userID)
		require.NoError(s.T(), err)
	}

	s.T().Run("Create subscription successfully", func(t *testing.T) {
		prepare()

		req := &storage.CreateRequest{
			UserID:      userID,
			ServiceName: serviceName,
			Price:       price,
			StartDate:   startDate,
			EndDate:     &endDate,
		}

		resp, err := s.repo.Create(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEqual(t, uuid.Nil, resp.ID)
	})

	s.T().Run("Create subscription with nil end date", func(t *testing.T) {
		prepare()

		req := &storage.CreateRequest{
			UserID:      userID,
			ServiceName: serviceName,
			Price:       price,
			StartDate:   startDate,
			EndDate:     nil,
		}

		resp, err := s.repo.Create(ctx, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEqual(t, uuid.Nil, resp.ID)
	})

	s.T().Run("Create subscription with invalid start date", func(t *testing.T) {
		prepare()

		req := &storage.CreateRequest{
			UserID:      userID,
			ServiceName: serviceName,
			Price:       price,
			StartDate:   "2025-09",
			EndDate:     &endDate,
		}

		resp, err := s.repo.Create(ctx, req)
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	s.T().Run("Create subscription with invalid end date", func(t *testing.T) {
		prepare()

		invalidEnd := "2025-12"
		req := &storage.CreateRequest{
			UserID:      userID,
			ServiceName: serviceName,
			Price:       price,
			StartDate:   startDate,
			EndDate:     &invalidEnd,
		}

		resp, err := s.repo.Create(ctx, req)
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	clear()
}
func (s *RepositoryTestSuite) TestGetInfo() {
	ctx := context.Background()

	userID := uuid.New()
	subscriptionID := uuid.New()
	serviceName := "Netflix"
	price := 15
	startDate := "09-2025"
	endDate := "12-2025"

	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, subscriptionID)
		require.NoError(s.T(), err)

		_, err = conn.Exec(ctx,
			`INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			subscriptionID,
			userID,
			serviceName,
			price,
			fmt.Sprintf("%s-%s-01", strings.Split(startDate, "-")[1], strings.Split(startDate, "-")[0]),
			fmt.Sprintf("%s-%s-01", strings.Split(endDate, "-")[1], strings.Split(endDate, "-")[0]),
		)
		require.NoError(s.T(), err)
	}
	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, subscriptionID)
		require.NoError(s.T(), err)
	}

	s.T().Run("Get subscription successfully", func(t *testing.T) {
		prepare()

		resp, err := s.repo.GetInfo(ctx, subscriptionID)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, subscriptionID, resp.ID)
		assert.Equal(t, userID, resp.UserID)
		assert.Equal(t, serviceName, resp.ServiceName)
		assert.Equal(t, price, resp.Price)
		assert.Equal(t, startDate, resp.StartDate)
		assert.Equal(t, &endDate, resp.EndDate)
	})

	s.T().Run("Get subscription with nil UUID", func(t *testing.T) {
		resp, err := s.repo.GetInfo(ctx, uuid.Nil)
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	s.T().Run("Get non-existing subscription", func(t *testing.T) {
		randomID := uuid.New()
		resp, err := s.repo.GetInfo(ctx, randomID)
		require.NoError(t, err)
		assert.Nil(t, resp)
	})

	clear()
}
func (s *RepositoryTestSuite) TestListSubscriptions() {
	ctx := context.Background()

	userID := uuid.New()
	otherUserID := uuid.New()
	sub1ID := uuid.New()
	sub2ID := uuid.New()
	service1 := "Netflix"
	service2 := "Spotify"
	start1 := "09-2025"
	start2 := "10-2025"
	end1 := "12-2025"
	end2 := "11-2025"

	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM subscriptions`)
		require.NoError(s.T(), err)

		subs := []struct {
			id          uuid.UUID
			userID      uuid.UUID
			serviceName string
			price       int
			start       string
			end         string
		}{
			{sub1ID, userID, service1, 10, start1, end1},
			{sub2ID, otherUserID, service2, 15, start2, end2},
		}

		for _, sub := range subs {
			_, err := conn.Exec(ctx,
				`INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				sub.id,
				sub.userID,
				sub.serviceName,
				sub.price,
				fmt.Sprintf("%s-%s-01", strings.Split(sub.start, "-")[1], strings.Split(sub.start, "-")[0]),
				fmt.Sprintf("%s-%s-01", strings.Split(sub.end, "-")[1], strings.Split(sub.end, "-")[0]),
			)
			require.NoError(s.T(), err)
		}
	}
	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()
		_, err = conn.Exec(ctx, `DELETE FROM subscriptions`)
		require.NoError(s.T(), err)
	}

	s.T().Run("List all subscriptions", func(t *testing.T) {
		prepare()
		req := &storage.ListRequest{}
		resp, err := s.repo.List(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Subscriptions, 2)
	})

	s.T().Run("Filter by UserID", func(t *testing.T) {
		req := &storage.ListRequest{UserID: &userID}
		resp, err := s.repo.List(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Subscriptions, 1)
		assert.Equal(t, userID, resp.Subscriptions[0].UserID)
	})

	s.T().Run("Filter by ServiceName", func(t *testing.T) {
		serviceFilter := "Spotify"
		req := &storage.ListRequest{ServiceName: &serviceFilter}
		resp, err := s.repo.List(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Subscriptions, 1)
		assert.Equal(t, "Spotify", resp.Subscriptions[0].ServiceName)
	})

	s.T().Run("Filter by From date", func(t *testing.T) {
		from := "10-2025"
		req := &storage.ListRequest{From: &from}
		resp, err := s.repo.List(ctx, req)
		require.NoError(t, err)
		require.Len(t, resp.Subscriptions, 1)
		assert.Equal(t, start2, resp.Subscriptions[0].StartDate)
	})

	s.T().Run("Invalid From date", func(t *testing.T) {
		invalidFrom := "2025-10"
		req := &storage.ListRequest{From: &invalidFrom}
		resp, err := s.repo.List(ctx, req)
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	s.T().Run("Nil request", func(t *testing.T) {
		resp, err := s.repo.List(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	clear()
}
func (s *RepositoryTestSuite) TestUpdateSubscription() {
	ctx := context.Background()

	userID := uuid.New()
	subID := uuid.New()
	service := "Netflix"
	price := 10
	start := "09-2025"
	end := "12-2025"

	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, subID)
		require.NoError(s.T(), err)

		_, err = conn.Exec(ctx,
			`INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			subID,
			userID,
			service,
			price,
			fmt.Sprintf("%s-%s-01", strings.Split(start, "-")[1], strings.Split(start, "-")[0]),
			fmt.Sprintf("%s-%s-01", strings.Split(end, "-")[1], strings.Split(end, "-")[0]),
		)
		require.NoError(s.T(), err)
	}
	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, subID)
		require.NoError(s.T(), err)
	}

	s.T().Run("Update all fields successfully", func(t *testing.T) {
		prepare()

		newService := "HBO Max"
		newPrice := 20
		newStart := "10-2025"
		newEnd := "01-2026"

		req := &storage.UpdateRequest{
			ServiceName: &newService,
			Price:       &newPrice,
			StartDate:   &newStart,
			EndDate:     &newEnd,
		}

		resp, err := s.repo.Update(ctx, subID, req)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Updated)

		info, err := s.repo.GetInfo(ctx, subID)
		require.NoError(t, err)
		assert.Equal(t, newService, info.ServiceName)
		assert.Equal(t, newPrice, info.Price)
		assert.Equal(t, newStart, info.StartDate)
		assert.Equal(t, &newEnd, info.EndDate)
	})

	s.T().Run("Update partial fields", func(t *testing.T) {
		newPrice := 25
		req := &storage.UpdateRequest{
			ServiceName: nil,
			Price:       &newPrice,
		}

		resp, err := s.repo.Update(ctx, subID, req)
		require.NoError(t, err)
		assert.True(t, resp.Updated)

		info, err := s.repo.GetInfo(ctx, subID)
		require.NoError(t, err)
		assert.Equal(t, newPrice, info.Price)
	})

	s.T().Run("Invalid StartDate format", func(t *testing.T) {
		invalid := "2025-10"
		req := &storage.UpdateRequest{StartDate: &invalid}

		resp, err := s.repo.Update(ctx, subID, req)
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	s.T().Run("Invalid EndDate format", func(t *testing.T) {
		invalid := "2025-13"
		req := &storage.UpdateRequest{EndDate: &invalid}

		resp, err := s.repo.Update(ctx, subID, req)
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	s.T().Run("Nil request", func(t *testing.T) {
		resp, err := s.repo.Update(ctx, subID, nil)
		require.Error(t, err)
		assert.Nil(t, resp)
	})

	s.T().Run("Update non-existing subscription", func(t *testing.T) {
		req := &storage.UpdateRequest{
			ServiceName: nil,
			Price:       nil,
		}
		resp, err := s.repo.Update(ctx, uuid.New(), req)
		require.NoError(t, err)
		assert.False(t, resp.Updated)
	})

	clear()
}

func (s *RepositoryTestSuite) TestDeleteSubscription() {
	ctx := context.Background()

	userID := uuid.New()
	subID := uuid.New()
	service := "Netflix"
	price := 10
	start := "09-2025"
	end := "12-2025"

	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, subID)
		require.NoError(s.T(), err)

		_, err = conn.Exec(ctx,
			`INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			subID,
			userID,
			service,
			price,
			fmt.Sprintf("%s-%s-01", strings.Split(start, "-")[1], strings.Split(start, "-")[0]),
			fmt.Sprintf("%s-%s-01", strings.Split(end, "-")[1], strings.Split(end, "-")[0]),
		)
		require.NoError(s.T(), err)
	}
	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, subID)
		require.NoError(s.T(), err)
	}

	s.T().Run("Delete existing subscription successfully", func(t *testing.T) {
		prepare()

		req := &storage.DeleteRequest{ID: subID}
		err := s.repo.Delete(ctx, req)
		require.NoError(t, err)

		info, err := s.repo.GetInfo(ctx, subID)
		require.NoError(t, err)
		assert.Nil(t, info)
	})

	s.T().Run("Delete non-existing subscription", func(t *testing.T) {
		req := &storage.DeleteRequest{ID: uuid.New()}
		err := s.repo.Delete(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not found")
	})

	s.T().Run("Nil request", func(t *testing.T) {
		err := s.repo.Delete(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "request object is nil")
	})

	s.T().Run("Nil UUID in request", func(t *testing.T) {
		req := &storage.DeleteRequest{ID: uuid.Nil}
		err := s.repo.Delete(ctx, req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "id is required")
	})

	clear()
}

func (s *RepositoryTestSuite) TestGetTotalSubscriptionsPrice() {
	ctx := context.Background()

	userID := uuid.New()
	otherUserID := uuid.New()
	service1 := "Netflix"
	service2 := "Spotify"
	start1 := "09-2025"
	start2 := "10-2025"
	end1 := "12-2025"
	end2 := "11-2025"
	price1 := 10
	price2 := 20

	prepare := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM subscriptions`)
		require.NoError(s.T(), err)

		subs := []struct {
			id          uuid.UUID
			userID      uuid.UUID
			serviceName string
			price       int
			start       string
			end         string
		}{
			{uuid.New(), userID, service1, price1, start1, end1},
			{uuid.New(), otherUserID, service2, price2, start2, end2},
		}

		for _, sub := range subs {
			_, err := conn.Exec(ctx,
				`INSERT INTO subscriptions (id, user_id, service_name, price, start_date, end_date)
				 VALUES ($1, $2, $3, $4, $5, $6)`,
				sub.id,
				sub.userID,
				sub.serviceName,
				sub.price,
				fmt.Sprintf("%s-%s-01", strings.Split(sub.start, "-")[1], strings.Split(sub.start, "-")[0]),
				fmt.Sprintf("%s-%s-01", strings.Split(sub.end, "-")[1], strings.Split(sub.end, "-")[0]),
			)
			require.NoError(s.T(), err)
		}
	}
	clear := func() {
		conn, err := s.db.Pool().Acquire(ctx)
		require.NoError(s.T(), err)
		defer conn.Release()

		_, err = conn.Exec(ctx, `DELETE FROM subscriptions`)
		require.NoError(s.T(), err)
	}

	s.T().Run("Get total price for all subscriptions", func(t *testing.T) {
		prepare()
		req := &storage.TotalRequest{
			From: "09-2025",
			To:   "12-2025",
		}
		total, err := s.repo.GetTotalSubscriptionsPrice(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, price1+price2, total)
	})

	s.T().Run("Filter by UserID", func(t *testing.T) {
		req := &storage.TotalRequest{
			UserID: &userID,
			From:   "09-2025",
			To:     "12-2025",
		}
		total, err := s.repo.GetTotalSubscriptionsPrice(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, price1, total)
	})

	s.T().Run("Filter by ServiceName", func(t *testing.T) {
		serviceFilter := "Spotify"
		req := &storage.TotalRequest{
			ServiceName: &serviceFilter,
			From:        "09-2025",
			To:          "12-2025",
		}
		total, err := s.repo.GetTotalSubscriptionsPrice(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, price2, total)
	})

	s.T().Run("Invalid From date", func(t *testing.T) {
		req := &storage.TotalRequest{
			From: "2025-09",
			To:   "12-2025",
		}
		total, err := s.repo.GetTotalSubscriptionsPrice(ctx, req)
		require.Error(t, err)
		assert.Equal(t, 0, total)
	})

	s.T().Run("Invalid To date", func(t *testing.T) {
		req := &storage.TotalRequest{
			From: "09-2025",
			To:   "2025-12",
		}
		total, err := s.repo.GetTotalSubscriptionsPrice(ctx, req)
		require.Error(t, err)
		assert.Equal(t, 0, total)
	})

	s.T().Run("Nil request", func(t *testing.T) {
		total, err := s.repo.GetTotalSubscriptionsPrice(ctx, nil)
		require.Error(t, err)
		assert.Equal(t, 0, total)
	})

	s.T().Run("No subscriptions in range", func(t *testing.T) {
		req := &storage.TotalRequest{
			From: "01-2020",
			To:   "02-2020",
		}
		total, err := s.repo.GetTotalSubscriptionsPrice(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, 0, total)
	})

	clear()
}

func TestRepositorySuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

type RepositoryTestSuite struct {
	container *postgres.PostgresContainer
	suite.Suite

	dbConfig storage.Config

	db   *storage.DB
	repo storage.SubscriptionsStorage
}

func (s *RepositoryTestSuite) SetupSuite() {
	ctx := context.Background()
	s.dbConfig = s.setupPostgres(ctx)

	logger := slog.Default()
	db := storage.NewDB(&s.dbConfig, logger)
	if err := db.Init(); err != nil {
		require.NoError(s.T(), err)
	}
	s.db = db
	s.repo = storage.NewService(db, logger)
}

func (s *RepositoryTestSuite) SetupTest() {
	ctx := context.Background()

	conn, err := s.db.Pool().Acquire(ctx)
	require.NoError(s.T(), err)
	defer conn.Release()

	_, err = conn.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	require.NoError(s.T(), err)

	require.NoError(s.T(), migrations.PostgresMigrate(s.dbConfig.UrlPostgres()))
}

func (s *RepositoryTestSuite) TearDownTest() {
	require.NoError(s.T(), migrations.PostgresMigrateDown(s.dbConfig.UrlPostgres()))
}
func (s *RepositoryTestSuite) setupPostgres(ctx context.Context) storage.Config {
	cfg := storage.Config{
		Host:             "",
		DbName:           "test-db",
		User:             "user",
		Password:         "1",
		MaxOpenConns:     10,
		ConnIdleLifetime: 60 * time.Second,
		ConnMaxLifetime:  60 * time.Minute,
	}
	// nolint
	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:14-alpine"),
		postgres.WithDatabase(cfg.DbName),
		postgres.WithUsername(cfg.User),
		postgres.WithPassword(cfg.Password),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	require.NoError(s.T(), err)
	s.container = pgContainer

	host, err := pgContainer.Host(ctx)
	require.NoError(s.T(), err)
	cfg.Host = host
	ports, err := pgContainer.MappedPort(ctx, "5432")
	require.NoError(s.T(), err)
	cfg.Host += ":" + strconv.Itoa(ports.Int())

	s.dbConfig = cfg
	return cfg
}

func (s *RepositoryTestSuite) TearDownSuite() {
	s.db.Stop()

	if err := s.container.Stop(context.Background(), nil); err != nil {
		s.T().Logf("failed to stop container: %v", err)
	}
}
