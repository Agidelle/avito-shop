package urls_test

import (
	urls "avito-shop/internal/http-server/handlers/url"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"avito-shop/internal/service/shop/storage"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"log/slog"
)

// MockService — мок для service.IService
type MockService struct {
	mock.Mock
}

func (m *MockService) CollectAllInfo(username string) (*storage.InfoResponse, error) {
	args := m.Called(username)
	return args.Get(0).(*storage.InfoResponse), args.Error(1)
}

func (m *MockService) Send(fromUsername string, scr *storage.SendCoinRequest) error {
	args := m.Called(fromUsername, scr)
	return args.Error(0)
}

func (m *MockService) Purchase(username, item string) error {
	args := m.Called(username, item)
	return args.Error(0)
}

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) CheckAuth(username string) (string, error) {
	args := m.Called(username)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) AddNewUser(username, password string) error {
	args := m.Called(username, password)
	return args.Error(0)
}

func (m *MockStorage) GetInfo(ir *storage.InfoResponse, username string) (int, error) {
	args := m.Called(ir, username)
	return args.Int(0), args.Error(1)
}

func (m *MockStorage) GetInventory(ir *storage.InfoResponse, id int) error {
	args := m.Called(ir, id)
	return args.Error(0)
}

func (m *MockStorage) GetReceivedHistory(ir *storage.InfoResponse, id int) error {
	args := m.Called(ir, id)
	return args.Error(0)
}

func (m *MockStorage) GetSendHistory(ir *storage.InfoResponse, id int) error {
	args := m.Called(ir, id)
	return args.Error(0)
}

func (m *MockStorage) BuyItem(name, item string, amount int) error {
	args := m.Called(name, item, amount)
	return args.Error(0)
}

func (m *MockStorage) SendCoins(username string, fromUserID int, toUserID int, scr *storage.SendCoinRequest) error {
	args := m.Called(username, fromUserID, toUserID, scr)
	return args.Error(0)
}

func TestInfoHandler_E2E(t *testing.T) {
	// Создаем мок сервиса
	mockService := new(MockService)

	// Определяем ожидаемый ответ от сервиса
	expectedResponse := &storage.InfoResponse{
		Coins: 1000,
		Inventory: []storage.Inventory{
			{Type: "t-shirt", Quantity: 1},
		},
		CoinHistory: storage.CoinHistory{
			Received: []storage.TransactionIn{
				{FromUser: "user2", Amount: 50},
			},
			Sent: []storage.TransactionOut{
				{ToUser: "user3", Amount: 30},
			},
		},
	}

	// Настройка мока для метода CollectAllInfo
	mockService.On("CollectAllInfo", "testuser").Return(expectedResponse, nil)

	// Создаем логгер
	logger := slog.New(slog.NewJSONHandler(nil, nil))

	// Создаем обработчик
	handlers := urls.NewHandlers(nil, mockService, logger)

	// Создаем HTTP-запрос
	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, "username", "testuser") // Добавляем username в контекст
	req = req.WithContext(ctx)

	// Создаем ResponseRecorder для записи ответа
	rr := httptest.NewRecorder()

	// Вызываем обработчик
	handler := handlers.Info()
	handler.ServeHTTP(rr, req)

	// Проверяем статус-код
	require.Equal(t, http.StatusOK, rr.Code)

	// Проверяем тело ответа
	var responseBody map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &responseBody)
	require.NoError(t, err)

	// Преобразуем ожидаемый ответ в JSON для сравнения
	expectedJSON, err := json.Marshal(expectedResponse)
	require.NoError(t, err)

	var expectedResponseBody map[string]interface{}
	err = json.Unmarshal(expectedJSON, &expectedResponseBody)
	require.NoError(t, err)

	// Сравниваем тела ответов
	require.Equal(t, expectedResponseBody, responseBody)

	// Проверяем, что метод CollectAllInfo был вызван
	mockService.AssertCalled(t, "CollectAllInfo", "testuser")
}
