package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"strings"
	"time"
)

func (r *Service) Create(ctx context.Context, request *CreateRequest) (*CreateResponse, error) {
	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		r.log.Error("failed to acquire DB connection", "error", err)
		return nil, err
	}
	defer conn.Release()

	parts := strings.Split(request.StartDate, "-")
	if len(parts) != 2 {
		r.log.Error("invalid start_date format in storage layer", "start_date", request.StartDate)
		return nil, fmt.Errorf("invalid start_date format, expected MM-YYYY")
	}
	startISO := fmt.Sprintf("%s-%s-01", parts[1], parts[0])

	var endVal interface{} = nil
	if request.EndDate != nil {
		ep := strings.Split(*request.EndDate, "-")
		if len(ep) != 2 {
			r.log.Error("invalid end_date format in storage layer", "end_date", *request.EndDate)
			return nil, fmt.Errorf("invalid end_date format, expected MM-YYYY")
		}
		endVal = fmt.Sprintf("%s-%s-01", ep[1], ep[0])
	}

	var id uuid.UUID
	err = conn.QueryRow(ctx,
		`INSERT INTO subscriptions (user_id, service_name, price, start_date, end_date)
         VALUES ($1, $2, $3, $4, $5)
         RETURNING id`,
		request.UserID,
		request.ServiceName,
		request.Price,
		startISO,
		endVal,
	).Scan(&id)
	if err != nil {
		r.log.Error("failed to insert subscription in storage layer",
			"error", err,
			"user_id", request.UserID,
			"service_name", request.ServiceName,
		)
		return nil, err
	}

	return &CreateResponse{ID: id}, nil
}

func (r *Service) GetInfo(ctx context.Context, id uuid.UUID) (*GetInfoResponse, error) {
	if id == uuid.Nil {
		r.log.Error("invalid subscription id in storage layer")
		return nil, errors.New("subscription with id not found")
	}

	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		r.log.Error("failed to acquire DB connection", "error", err)
		return nil, err
	}
	defer conn.Release()

	row := conn.QueryRow(ctx,
		`SELECT user_id, service_name, price, start_date, end_date
         FROM subscriptions
         WHERE id = $1`,
		id,
	)

	var userID uuid.UUID
	var serviceName string
	var startDate time.Time
	var endDate *time.Time
	var price int
	err = row.Scan(&userID, &serviceName, &price, &startDate, &endDate)
	if err != nil {
		if err == pgx.ErrNoRows || errors.Is(err, pgx.ErrNoRows) || strings.Contains(err.Error(), "no rows") {
			r.log.Warn("subscription not found in DB", "id", id)
			return nil, nil
		}
		r.log.Error("failed to scan subscription row in storage layer", "error", err, "id", id)
		return nil, err
	}

	startStr := startDate.Format("01-2006")
	var endStr *string
	if endDate != nil {
		s := endDate.Format("01-2006")
		endStr = &s
	}

	resp := &GetInfoResponse{
		ID:          id,
		UserID:      userID,
		ServiceName: serviceName,
		Price:       price,
		StartDate:   startStr,
		EndDate:     endStr,
	}

	r.log.Info("subscription info retrieved successfully in storage layer",
		"id", id,
		"user_id", userID,
		"service_name", serviceName,
		"price", price,
		"start_date", startStr,
		"end_date", endStr,
	)

	return resp, nil
}

func (r *Service) List(ctx context.Context, request *ListRequest) (*ListResponse, error) {
	if request == nil {
		r.log.Error("request object is nil in storage layer")
		return nil, errors.New("request object is nil")
	}

	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		r.log.Error("failed to acquire DB connection", "error", err)
		return nil, err
	}
	defer conn.Release()

	var args []interface{}
	conds := []string{"1=1"}

	argIdx := 1
	if request.UserID != nil && *request.UserID != uuid.Nil {
		conds = append(conds, fmt.Sprintf("user_id = $%d", argIdx))
		args = append(args, *request.UserID)
		argIdx++
	}
	if request.ServiceName != nil && *request.ServiceName != "" {
		conds = append(conds, fmt.Sprintf("service_name ILIKE $%d", argIdx))
		args = append(args, "%"+*request.ServiceName+"%")
		argIdx++
	}
	if request.From != nil {
		fromDate, err := time.Parse("01-2006", *request.From)
		if err != nil {
			r.log.Error("invalid From date in storage layer", "from", *request.From, "error", err)
			return nil, fmt.Errorf("invalid From date: %w", err)
		}
		conds = append(conds, fmt.Sprintf("start_date >= $%d", argIdx))
		args = append(args, fromDate)
		argIdx++
	}
	if request.To != nil {
		toDate, err := time.Parse("01-2006", *request.To)
		if err != nil {
			r.log.Error("invalid To date in storage layer", "to", *request.To, "error", err)
			return nil, fmt.Errorf("invalid To date: %w", err)
		}
		toDate = toDate.AddDate(0, 1, -1)
		conds = append(conds, fmt.Sprintf("start_date <= $%d", argIdx))
		args = append(args, toDate)
		argIdx++
	}

	limit := 50
	if request.Limit != nil && *request.Limit > 0 {
		limit = *request.Limit
	}
	offset := 0
	if request.Offset != nil && *request.Offset >= 0 {
		offset = *request.Offset
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, service_name, price, start_date, end_date
		FROM subscriptions
		WHERE %s
		ORDER BY start_date
		LIMIT $%d OFFSET $%d`, strings.Join(conds, " AND "), argIdx, argIdx+1)

	args = append(args, limit, offset)

	r.log.Info("executing query in storage layer", "query", query, "args", args)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		r.log.Error("failed to execute query in storage layer", "error", err)
		return nil, err
	}
	defer rows.Close()

	var resp ListResponse
	for rows.Next() {
		var (
			id          uuid.UUID
			userID      uuid.UUID
			serviceName string
			price       int
			startDate   time.Time
			endDate     *time.Time
		)
		if err := rows.Scan(&id, &userID, &serviceName, &price, &startDate, &endDate); err != nil {
			r.log.Error("failed to scan row in storage layer", "error", err)
			return nil, err
		}

		startStr := startDate.Format("01-2006")
		var endStr *string
		if endDate != nil {
			s := endDate.Format("01-2006")
			endStr = &s
		}

		resp.Subscriptions = append(resp.Subscriptions, GetInfoResponse{
			ID:          id,
			UserID:      userID,
			ServiceName: serviceName,
			Price:       price,
			StartDate:   startStr,
			EndDate:     endStr,
		})
	}

	return &resp, nil
}

func (r *Service) Update(ctx context.Context, id uuid.UUID, request *UpdateRequest) (*UpdateResponse, error) {
	if request == nil {
		r.log.Error("request object is nil in storage layer")
		return nil, errors.New("request object is nil")
	}

	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		r.log.Error("failed to acquire DB connection", "error", err)
		return nil, err
	}
	defer conn.Release()

	var startDate, endDate interface{}
	if request.StartDate != nil {
		t, err := time.Parse("01-2006", *request.StartDate)
		if err != nil {
			r.log.Error("invalid start_date format in storage layer", "startDate", *request.StartDate, "error", err)
			return nil, fmt.Errorf("invalid start_date format: %w", err)
		}
		startDate = t
	}
	if request.EndDate != nil {
		t, err := time.Parse("01-2006", *request.EndDate)
		if err != nil {
			r.log.Error("invalid end_date format in storage layer", "endDate", *request.EndDate, "error", err)
			return nil, fmt.Errorf("invalid end_date format: %w", err)
		}
		endDate = t
	}

	cmdTag, err := conn.Exec(ctx, `
		UPDATE subscriptions
		SET
			service_name = COALESCE($2, service_name),
			price        = COALESCE($3, price),
			start_date   = COALESCE($4, start_date),
			end_date     = COALESCE($5, end_date)
		WHERE id = $1
	`, id, request.ServiceName, request.Price, startDate, endDate)
	if err != nil {
		r.log.Error("failed to update subscription in storage layer", "error", err)
		return nil, err
	}

	updated := cmdTag.RowsAffected() > 0
	return &UpdateResponse{Updated: updated}, nil
}

func (r *Service) Delete(ctx context.Context, request *DeleteRequest) error {
	if request == nil {
		r.log.Error("request object is nil in storage layer")
		return errors.New("request object is nil")
	}
	if request.ID == uuid.Nil {
		r.log.Error("id is required in storage layer")
		return errors.New("id is required")
	}

	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		r.log.Error("failed to acquire DB connection", "error", err)
		return err
	}
	defer conn.Release()

	cmdTag, err := conn.Exec(ctx, `DELETE FROM subscriptions WHERE id = $1`, request.ID)
	if err != nil {
		r.log.Error("failed to delete subscription in storage layer", "error", err)
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		r.log.Warn("subscription not found in storage layer", "id", request.ID)
		return fmt.Errorf("subscription with id %s not found", request.ID)
	}

	return nil
}

func (r *Service) GetTotalSubscriptionsPrice(ctx context.Context, request *TotalRequest) (int, error) {
	if request == nil {
		r.log.Error("request object is nil in storage layer")
		return 0, errors.New("request object is nil")
	}

	conn, err := r.Pool().Acquire(ctx)
	if err != nil {
		r.log.Error("failed to acquire DB connection", "error", err)
		return 0, err
	}
	defer conn.Release()

	fromDate, err := time.Parse("01-2006", request.From)
	if err != nil {
		r.log.Error("invalid From date format in storage layer", "from", request.From, "error", err)
		return 0, fmt.Errorf("invalid From date: %w", err)
	}

	toDate, err := time.Parse("01-2006", request.To)
	if err != nil {
		r.log.Error("invalid To date format in storage layer", "to", request.To, "error", err)
		return 0, fmt.Errorf("invalid To date: %w", err)
	}
	toDate = toDate.AddDate(0, 1, -1)

	var total int
	query := `
		SELECT COALESCE(SUM(price), 0)
		FROM subscriptions
		WHERE ($1::uuid IS NULL OR user_id = $1)
		  AND ($2::text IS NULL OR service_name ILIKE '%' || $2 || '%')
		  AND start_date >= $3
		  AND start_date <= $4
	`

	err = conn.QueryRow(ctx, query, request.UserID, request.ServiceName, fromDate, toDate).Scan(&total)
	if err != nil {
		r.log.Error("failed to get total subscriptions price in storage layer", "error", err)
		return 0, err
	}

	return total, nil
}
