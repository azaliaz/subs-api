package rest

import (
	"github.com/azaliaz/subs-api/internal/application"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"strconv"
	"time"
)

func (api *Service) Create(c *fiber.Ctx) error {
	var req application.CreateRequest
	if err := c.BodyParser(&req); err != nil {
		api.log.Info("failed to parse body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   "invalid request body",
			"details": err.Error(),
		})
	}

	if req.UserID == uuid.Nil {
		api.log.Warn("User ID is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "user_id is required"})
	}
	if req.ServiceName == "" {
		api.log.Warn("Service name is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "service_name is required"})
	}
	if req.Price <= 0 {
		api.log.Warn("Price is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "price must be greater than 0"})
	}
	if _, err := time.Parse("01-2006", req.StartDate); err != nil {
		api.log.Warn("invalid start_date format", "start_date", req.StartDate, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid start_date format"})
	}

	if req.EndDate != nil && *req.EndDate != "" {
		if _, err := time.Parse("01-2006", *req.EndDate); err != nil {
			api.log.Warn("invalid end_date format", "end_date", *req.EndDate, "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid end_date format"})
		}
	}

	resp, err := api.app.Create(c.Context(), &req)
	if err != nil {
		api.log.Info("failed to create", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (api *Service) GetInfo(c *fiber.Ctx) error {
	idParam := c.Params("id")
	if idParam == "" {
		api.log.Warn("ID is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}
	subsID, err := uuid.Parse(idParam)
	if err != nil {
		api.log.Warn("ID is invalid", "id", idParam, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id format",
		})
	}
	resp, err := api.app.GetInfo(c.Context(), &application.GetInfoRequest{ID: subsID})
	if err != nil {
		api.log.Info("failed to get info", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if resp == nil {
		api.log.Info("subscription not found", "id", subsID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "subscription not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (api *Service) GetList(c *fiber.Ctx) error {
	var req application.ListRequest
	if userID := c.Query("user_id"); userID != "" {
		uid, err := uuid.Parse(userID)
		if err != nil {
			api.log.Warn("invalid user id format", "user_id", userID, "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid user_id"})
		}
		req.UserID = &uid
	}
	if service := c.Query("service_name"); service != "" {
		req.ServiceName = &service
	}
	if from := c.Query("from"); from != "" {
		if _, err := time.Parse("01-2006", from); err != nil {
			api.log.Warn("invalid from format", "from", from, "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid From date format"})
		}
		req.From = &from
	}
	if to := c.Query("to"); to != "" {
		if _, err := time.Parse("01-2006", to); err != nil {
			api.log.Warn("invalid to format", "to", to, "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid To date format"})
		}
		req.To = &to
	}
	if limit := c.Query("limit"); limit != "" {
		l, err := strconv.Atoi(limit)
		if err != nil || l <= 0 {
			api.log.Warn("invalid limit format", "limit", limit, "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid limit"})
		}
		req.Limit = &l
	}
	if offset := c.Query("offset"); offset != "" {
		o, err := strconv.Atoi(offset)
		if err != nil || o < 0 {
			api.log.Warn("invalid offset format", "offset", offset, "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid offset"})
		}
		req.Offset = &o
	}

	resp, err := api.app.List(c.Context(), &req)
	if err != nil {
		api.log.Info("failed to list", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (api *Service) Update(c *fiber.Ctx) error {
	idParam := c.Params("id")
	if idParam == "" {
		api.log.Warn("ID parameter is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}
	id, err := uuid.Parse(idParam)
	if err != nil {
		api.log.Warn("invalid id format", "id", idParam, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid id format"})
	}
	var req application.UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		api.log.Info("failed to parse body", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body",
		})
	}
	if req.StartDate != nil {
		if _, err := time.Parse("01-2006", *req.StartDate); err != nil {
			api.log.Warn("invalid start_date format", "start_date", *req.StartDate, "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid start date format",
			})
		}
	}
	if req.EndDate != nil {
		if _, err := time.Parse("01-2006", *req.EndDate); err != nil {
			api.log.Warn("invalid end_date format", "end_date", *req.EndDate, "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid end date format",
			})
		}
	}
	resp, err := api.app.Update(c.Context(), id, &req)
	if err != nil {
		api.log.Info("failed to update", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (api *Service) Delete(c *fiber.Ctx) error {
	idParam := c.Params("id")
	if idParam == "" {
		api.log.Warn("ID parameter is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "id parameter is required",
		})
	}
	subsID, err := uuid.Parse(idParam)
	if err != nil {
		api.log.Warn("ID is invalid", "id", idParam, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid id format",
		})
	}
	resp, err := api.app.Delete(c.Context(), &application.DeleteRequest{ID: subsID})
	if err != nil {
		api.log.Info("failed to delete", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if resp == nil {
		api.log.Warn("subscription not found", "id", subsID)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "subscription not found",
		})
	}
	return c.Status(fiber.StatusOK).JSON(resp)
}

func (api *Service) GetTotalSubscriptionsPrice(c *fiber.Ctx) error {
	var req application.TotalRequest

	if userID := c.Query("user_id"); userID != "" {
		uid, err := uuid.Parse(userID)
		if err != nil {
			api.log.Warn("invalid user id format", "user_id", userID, "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid user_id",
			})
		}
		req.UserID = &uid
	}

	if service := c.Query("service_name"); service != "" {
		req.ServiceName = &service
	}

	req.From = c.Query("from")
	if req.From == "" {
		api.log.Warn("from parameter is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "from is required",
		})
	}
	if _, err := time.Parse("01-2006", req.From); err != nil {
		api.log.Warn("invalid from format", "from", req.From, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid From date format (expected MM-YYYY)",
		})
	}

	req.To = c.Query("to")
	if req.To == "" {
		api.log.Warn("to parameter is required")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "to is required",
		})
	}
	if _, err := time.Parse("01-2006", req.To); err != nil {
		api.log.Warn("invalid to format", "to", req.To, "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid To date format (expected MM-YYYY)",
		})
	}

	resp, err := api.app.GetTotalSubscriptionsPrice(c.Context(), &req)
	if err != nil {
		api.log.Info("failed to get total subscriptions price", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(resp)
}
