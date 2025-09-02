package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/azaliaz/subs-api/internal/application"
	"github.com/azaliaz/subs-api/internal/application/mocks"
	"github.com/azaliaz/subs-api/internal/facade/rest"
	"github.com/gofiber/fiber/v2"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)

	userID := uuid.New()
	serviceName := "Netflix"
	price := 10
	startDate := "09-2025"
	endDate := "12-2025"

	mockApp.EXPECT().
		Create(gomock.Any(), &application.CreateRequest{
			UserID:      userID,
			ServiceName: serviceName,
			Price:       price,
			StartDate:   startDate,
			EndDate:     &endDate,
		}).
		Return(&application.CreateResponse{
			ID: uuid.New(),
		}, nil)

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("POST", "/api/create", api.Create)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"user_id":      userID,
		"service_name": serviceName,
		"price":        price,
		"start_date":   startDate,
		"end_date":     endDate,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/create", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)
}

func TestCreate_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("POST", "/api/create", api.Create)

	req := httptest.NewRequest(http.MethodPost, "/api/create", bytes.NewReader([]byte("{invalid_json}")))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestCreate_InvalidFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("POST", "/api/create", api.Create)

	userID := uuid.Nil
	serviceName := ""
	price := 0
	startDate := "invalid-date"

	reqBody, _ := json.Marshal(map[string]interface{}{
		"user_id":      userID.String(),
		"service_name": serviceName,
		"price":        price,
		"start_date":   startDate,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/create", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestCreate_AppError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	userID := uuid.New()
	serviceName := "Netflix"
	price := 10
	startDate := "09-2025"

	mockApp.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("db error"))

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("POST", "/api/create", api.Create)

	requestBody, _ := json.Marshal(map[string]interface{}{
		"user_id":      userID.String(),
		"service_name": serviceName,
		"price":        price,
		"start_date":   startDate,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/create", bytes.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")

	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
func TestGetInfo_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	subID := uuid.New()

	mockApp.EXPECT().GetInfo(gomock.Any(), &application.GetInfoRequest{
		ID: subID,
	}).Return(&application.GetInfoResponse{
		ID:          subID,
		UserID:      uuid.New(),
		ServiceName: "Netflix",
		Price:       10,
		StartDate:   "09-2025",
		EndDate:     nil,
	}, nil)

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/info/:id", api.GetInfo)

	req := httptest.NewRequest(http.MethodGet, "/api/info/"+subID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetInfo_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Get("/api/info/*", api.GetInfo)

	req := httptest.NewRequest(http.MethodGet, "/api/info/", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetInfo_InvalidIDFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/info/:id", api.GetInfo)

	req := httptest.NewRequest(http.MethodGet, "/api/info/invalid-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetInfo_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	subID := uuid.New()

	mockApp.EXPECT().GetInfo(gomock.Any(), &application.GetInfoRequest{
		ID: subID,
	}).Return(nil, nil)

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/info/:id", api.GetInfo)

	req := httptest.NewRequest(http.MethodGet, "/api/info/"+subID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestGetInfo_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	subID := uuid.New()

	mockApp.EXPECT().GetInfo(gomock.Any(), &application.GetInfoRequest{
		ID: subID,
	}).Return(nil, fmt.Errorf("db error"))

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/info/:id", api.GetInfo)

	req := httptest.NewRequest(http.MethodGet, "/api/info/"+subID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestGetList_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	userID := uuid.New()
	serviceName := "Netflix"
	from := "09-2025"
	to := "12-2025"
	limit := 10
	offset := 0

	mockApp.EXPECT().List(gomock.Any(), &application.ListRequest{
		UserID:      &userID,
		ServiceName: &serviceName,
		From:        &from,
		To:          &to,
		Limit:       &limit,
		Offset:      &offset,
	}).Return(&application.ListResponse{
		Subscriptions: []application.GetInfoResponse{
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

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/list", api.GetList)

	req := httptest.NewRequest(http.MethodGet,
		fmt.Sprintf("/api/list?user_id=%s&service_name=%s&from=%s&to=%s&limit=%d&offset=%d",
			userID.String(), serviceName, from, to, limit, offset),
		nil,
	)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetList_InvalidUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/list", api.GetList)

	req := httptest.NewRequest(http.MethodGet, "/api/list?user_id=invalid-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetList_InvalidFromDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/list", api.GetList)

	req := httptest.NewRequest(http.MethodGet, "/api/list?from=2025-09", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetList_InvalidToDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/list", api.GetList)

	req := httptest.NewRequest(http.MethodGet, "/api/list?to=2025-12", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetList_InvalidLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/list", api.GetList)

	req := httptest.NewRequest(http.MethodGet, "/api/list?limit=-1", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetList_InvalidOffset(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/list", api.GetList)

	req := httptest.NewRequest(http.MethodGet, "/api/list?offset=-5", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetList_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	userID := uuid.New()

	mockApp.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, fmt.Errorf("db error"))

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/list", api.GetList)

	req := httptest.NewRequest(http.MethodGet, "/api/list?user_id="+userID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
func TestUpdate_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	subID := uuid.New()
	startDate := "09-2025"
	endDate := "12-2025"

	mockApp.EXPECT().Update(gomock.Any(), subID, &application.UpdateRequest{
		StartDate: &startDate,
		EndDate:   &endDate,
	}).Return(&application.UpdateResponse{
		Updated: true,
	}, nil)

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("PUT", "/api/update/:id", api.Update)

	body, _ := json.Marshal(map[string]string{
		"start_date": startDate,
		"end_date":   endDate,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/update/"+subID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)

	var respBody application.UpdateResponse
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.True(t, respBody.Updated)
}

func TestUpdate_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("PUT", "/api/update/*", api.Update)

	body, _ := json.Marshal(map[string]string{
		"start_date": "09-2025",
		"end_date":   "12-2025",
	})
	req := httptest.NewRequest(http.MethodPut, "/api/update/", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	var respBody map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&respBody)
	assert.Equal(t, "id parameter is required", respBody["error"])
}

func TestUpdate_InvalidIDFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("PUT", "/api/update/:id", api.Update)

	req := httptest.NewRequest(http.MethodPut, "/api/update/invalid-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUpdate_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	subID := uuid.New()

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("PUT", "/api/update/:id", api.Update)

	req := httptest.NewRequest(http.MethodPut, "/api/update/"+subID.String(), bytes.NewReader([]byte("{invalid_json}")))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUpdate_InvalidStartDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	subID := uuid.New()
	startDate := "2025-09"

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("PUT", "/api/update/:id", api.Update)

	body, _ := json.Marshal(map[string]string{
		"start_date": startDate,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/update/"+subID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUpdate_InvalidEndDate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	subID := uuid.New()
	endDate := "2025-13"

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("PUT", "/api/update/:id", api.Update)

	body, _ := json.Marshal(map[string]string{
		"end_date": endDate,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/update/"+subID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestUpdate_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	subID := uuid.New()
	startDate := "09-2025"

	mockApp.EXPECT().Update(gomock.Any(), subID, &application.UpdateRequest{
		StartDate: &startDate,
	}).Return(nil, fmt.Errorf("db error"))

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("PUT", "/api/update/:id", api.Update)

	body, _ := json.Marshal(map[string]string{
		"start_date": startDate,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/update/"+subID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
func TestDelete_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	subID := uuid.New()

	mockApp.EXPECT().Delete(gomock.Any(), &application.DeleteRequest{
		ID: subID,
	}).Return(&application.DeleteResponse{Deleted: true}, nil)

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("DELETE", "/api/delete/:id", api.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/delete/"+subID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestDelete_MissingID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("DELETE", "/api/delete/:id?", api.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/delete", nil)
	resp, _ := app.Test(req)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDelete_InvalidIDFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("DELETE", "/api/delete/:id", api.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/delete/invalid-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestDelete_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	subID := uuid.New()

	mockApp.EXPECT().Delete(gomock.Any(), &application.DeleteRequest{
		ID: subID,
	}).Return(nil, nil)

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("DELETE", "/api/delete/:id", api.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/delete/"+subID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestDelete_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)
	subID := uuid.New()

	mockApp.EXPECT().Delete(gomock.Any(), &application.DeleteRequest{
		ID: subID,
	}).Return(nil, fmt.Errorf("db error"))

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("DELETE", "/api/delete/:id", api.Delete)

	req := httptest.NewRequest(http.MethodDelete, "/api/delete/"+subID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}

func TestGetTotalSubscriptionsPrice_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)

	req := application.TotalRequest{
		From: "09-2025",
		To:   "12-2025",
	}
	mockApp.EXPECT().
		GetTotalSubscriptionsPrice(gomock.Any(), &req).
		Return(&application.TotalResponse{Total: 100}, nil)

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/total", api.GetTotalSubscriptionsPrice)

	url := "/api/total?from=09-2025&to=12-2025"
	reqHTTP := httptest.NewRequest(http.MethodGet, url, nil)
	resp, _ := app.Test(reqHTTP)

	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestGetTotalSubscriptionsPrice_InvalidUserID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/total", api.GetTotalSubscriptionsPrice)

	url := "/api/total?user_id=not-a-uuid&from=09-2025&to=12-2025"
	reqHTTP := httptest.NewRequest(http.MethodGet, url, nil)
	resp, _ := app.Test(reqHTTP)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetTotalSubscriptionsPrice_MissingFrom(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/total", api.GetTotalSubscriptionsPrice)

	url := "/api/total?to=12-2025"
	reqHTTP := httptest.NewRequest(http.MethodGet, url, nil)
	resp, _ := app.Test(reqHTTP)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetTotalSubscriptionsPrice_InvalidFromFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/total", api.GetTotalSubscriptionsPrice)

	url := "/api/total?from=2025-09&to=12-2025"
	reqHTTP := httptest.NewRequest(http.MethodGet, url, nil)
	resp, _ := app.Test(reqHTTP)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetTotalSubscriptionsPrice_MissingTo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/total", api.GetTotalSubscriptionsPrice)

	url := "/api/total?from=09-2025"
	reqHTTP := httptest.NewRequest(http.MethodGet, url, nil)
	resp, _ := app.Test(reqHTTP)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetTotalSubscriptionsPrice_InvalidToFormat(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/total", api.GetTotalSubscriptionsPrice)

	url := "/api/total?from=09-2025&to=2025-12"
	reqHTTP := httptest.NewRequest(http.MethodGet, url, nil)
	resp, _ := app.Test(reqHTTP)

	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestGetTotalSubscriptionsPrice_ServiceError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockApp := mocks.NewMockSubscriptionsService(ctrl)

	req := application.TotalRequest{
		From: "09-2025",
		To:   "12-2025",
	}
	mockApp.EXPECT().
		GetTotalSubscriptionsPrice(gomock.Any(), &req).
		Return(nil, fmt.Errorf("db error"))

	api := rest.NewAPI(slog.Default(), nil, mockApp)
	app := fiber.New()
	app.Add("GET", "/api/total", api.GetTotalSubscriptionsPrice)

	url := "/api/total?from=09-2025&to=12-2025"
	reqHTTP := httptest.NewRequest(http.MethodGet, url, nil)
	resp, _ := app.Test(reqHTTP)

	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
}
