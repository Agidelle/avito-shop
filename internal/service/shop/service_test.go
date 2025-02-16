package shop

import (
	"avito-shop/internal/service/shop/storage"

	"errors"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestGenerateJWT(t *testing.T) {
	type testCase struct {
		name           string
		secretKey      string
		username       string
		wantError      bool
		wantErrType    error
		wantTokenEmpty bool
		wantClaims     jwt.MapClaims
	}

	testCases := []testCase{
		{
			name:      "successful JWT generation",
			secretKey: "your-secret-key",
			username:  "test_user",
			wantError: false,
			wantClaims: jwt.MapClaims{
				"username": "test_user",
			},
		},
		{
			name:           "invalid signing method (empty secret key)",
			secretKey:      "",
			username:       "test_user",
			wantError:      true,
			wantErrType:    ErrInternalServer,
			wantTokenEmpty: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := GenerateJWT(tc.secretKey, tc.username)
			if tc.wantError {
				assert.Error(t, err)
				if tc.wantErrType != nil {
					assert.Equal(t, tc.wantErrType, err)
				}
				if tc.wantTokenEmpty {
					assert.Empty(t, token)
				}
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, token)

			parsedToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
				return []byte(tc.secretKey), nil
			})
			assert.NoError(t, err)
			assert.True(t, parsedToken.Valid)

			claims, ok := parsedToken.Claims.(jwt.MapClaims)
			assert.True(t, ok)
			assert.Equal(t, tc.wantClaims["username"], claims["username"])

			expectedExp := time.Now().Add(24 * time.Hour)
			actualExp := time.Unix(int64(claims["exp"].(float64)), 0)
			assert.WithinDuration(t, expectedExp, actualExp, time.Second)

			expectedIat := time.Now()
			actualIat := time.Unix(int64(claims["iat"].(float64)), 0)
			assert.WithinDuration(t, expectedIat, actualIat, time.Second)
		})
	}
}

func TestPurchase_TableDriven(t *testing.T) {
	mockStorage := &storage.IStorageMock{}
	service := NewService(mockStorage)

	tests := []struct {
		name          string
		item          string
		username      string
		setupMocks    func()
		expectedError error
	}{
		{
			name:     "Successful purchase",
			item:     "t-shirt",
			username: "test_user",
			setupMocks: func() {
				storage.MerchItems["t-shirt"] = 50 // Добавляем предмет в список товаров
				mockStorage.GetInfoFunc = func(res *storage.InfoResponse, username string) (int, error) {
					res.Coins = 100 // У пользователя достаточно средств
					return 1, nil
				}
				mockStorage.BuyItemFunc = func(name, item string, amount int) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:          "Item not found",
			item:          "unknown_item",
			username:      "test_user",
			setupMocks:    func() {},
			expectedError: ErrItemNotFound,
		},
		{
			name:     "Insufficient funds",
			item:     "t-shirt",
			username: "test_user",
			setupMocks: func() {
				storage.MerchItems["t-shirt"] = 50 // Добавляем предмет в список товаров
				mockStorage.GetInfoFunc = func(res *storage.InfoResponse, username string) (int, error) {
					res.Coins = 30 // У пользователя недостаточно средств
					return 1, nil
				}
			},
			expectedError: ErrInsufficientFunds,
		},
		{
			name:     "GetInfo error",
			item:     "t-shirt",
			username: "test_user",
			setupMocks: func() {
				storage.MerchItems["t-shirt"] = 50 // Добавляем предмет в список товаров
				mockStorage.GetInfoFunc = func(res *storage.InfoResponse, username string) (int, error) {
					return 0, errors.New("database error")
				}
			},
			expectedError: ErrInternalServer,
		},
		{
			name:     "BuyItem error",
			item:     "t-shirt",
			username: "test_user",
			setupMocks: func() {
				storage.MerchItems["t-shirt"] = 50 // Добавляем предмет в список товаров
				mockStorage.GetInfoFunc = func(res *storage.InfoResponse, username string) (int, error) {
					res.Coins = 100 // У пользователя достаточно средств
					return 1, nil
				}
				mockStorage.BuyItemFunc = func(name, item string, amount int) error {
					return errors.New("buy item error")
				}
			},
			expectedError: ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMocks != nil {
				tt.setupMocks()
			}

			err := service.Purchase(tt.username, tt.item)

			assert.Equal(t, tt.expectedError, err)
		})
	}
}

func TestCollectAllInfo_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		setupMocks     func(mockStorage *storage.IStorageMock)
		expectedResult *storage.InfoResponse
		expectedError  error
	}{
		{
			name:     "Successful collection",
			username: "test_user",
			setupMocks: func(mockStorage *storage.IStorageMock) {
				mockStorage.GetInfoFunc = func(res *storage.InfoResponse, username string) (int, error) {
					res.Coins = 100
					return 1, nil
				}
				mockStorage.GetInventoryFunc = func(res *storage.InfoResponse, id int) error {
					res.Inventory = []storage.Inventory{
						{Type: "t-shirt", Quantity: 2},
					}
					return nil
				}
				mockStorage.GetSendHistoryFunc = func(res *storage.InfoResponse, id int) error {
					res.CoinHistory.Sent = []storage.TransactionOut{
						{ToUser: "jane_doe", Amount: 50},
					}
					return nil
				}
				mockStorage.GetReceivedHistoryFunc = func(res *storage.InfoResponse, id int) error {
					res.CoinHistory.Received = []storage.TransactionIn{
						{FromUser: "john_doe", Amount: 30},
					}
					return nil
				}
			},
			expectedResult: &storage.InfoResponse{
				Coins: 100,
				Inventory: []storage.Inventory{
					{Type: "t-shirt", Quantity: 2},
				},
				CoinHistory: storage.CoinHistory{
					Sent: []storage.TransactionOut{
						{ToUser: "jane_doe", Amount: 50},
					},
					Received: []storage.TransactionIn{
						{FromUser: "john_doe", Amount: 30},
					},
				},
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			mockStorage := &storage.IStorageMock{}
			service := NewService(mockStorage)

			if tt.setupMocks != nil {
				tt.setupMocks(mockStorage)
			}

			result, err := service.CollectAllInfo(tt.username)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestSend_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		fromUsername  string
		scr           *storage.SendCoinRequest
		setupMocks    func(mockStorage *storage.IStorageMock)
		expectedError error
	}{
		{
			name:         "Successful send",
			fromUsername: "from_user",
			scr: &storage.SendCoinRequest{
				ToUser: "to_user",
				Amount: 50,
			},
			setupMocks: func(mockStorage *storage.IStorageMock) {
				mockStorage.GetInfoFunc = func(res *storage.InfoResponse, username string) (int, error) {
					if username == "from_user" {
						res.Coins = 100
						return 1, nil
					}
					if username == "to_user" {
						res.Coins = 50
						return 2, nil
					}
					return 0, errors.New("unknown user")
				}
				mockStorage.SendCoinsFunc = func(username string, fromUserID int, toUserID int, scr *storage.SendCoinRequest) error {
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name:         "Insufficient funds",
			fromUsername: "from_user",
			scr: &storage.SendCoinRequest{
				ToUser: "to_user",
				Amount: 50,
			},
			setupMocks: func(mockStorage *storage.IStorageMock) {
				mockStorage.GetInfoFunc = func(res *storage.InfoResponse, username string) (int, error) {
					if username == "from_user" {
						res.Coins = 30
						return 1, nil
					}
					if username == "to_user" {
						res.Coins = 50
						return 2, nil
					}
					return 0, errors.New("unknown user")
				}
			},
			expectedError: ErrInsufficientFunds,
		},
		{
			name:         "GetInfo error for fromUser",
			fromUsername: "from_user",
			scr: &storage.SendCoinRequest{
				ToUser: "to_user",
				Amount: 50,
			},
			setupMocks: func(mockStorage *storage.IStorageMock) {
				mockStorage.GetInfoFunc = func(res *storage.InfoResponse, username string) (int, error) {
					if username == "from_user" {
						return 0, errors.New("database error")
					}
					if username == "to_user" {
						res.Coins = 50
						return 2, nil
					}
					return 0, errors.New("unknown user")
				}
			},
			expectedError: ErrInternalServer,
		},
		{
			name:         "GetInfo error for toUser",
			fromUsername: "from_user",
			scr: &storage.SendCoinRequest{
				ToUser: "to_user",
				Amount: 50,
			},
			setupMocks: func(mockStorage *storage.IStorageMock) {
				mockStorage.GetInfoFunc = func(res *storage.InfoResponse, username string) (int, error) {
					if username == "from_user" {
						res.Coins = 100
						return 1, nil
					}
					if username == "to_user" {
						return 0, errors.New("database error")
					}
					return 0, errors.New("unknown user")
				}
			},
			expectedError: ErrInternalServer,
		},
		{
			name:         "SendCoins error",
			fromUsername: "from_user",
			scr: &storage.SendCoinRequest{
				ToUser: "to_user",
				Amount: 50,
			},
			setupMocks: func(mockStorage *storage.IStorageMock) {
				mockStorage.GetInfoFunc = func(res *storage.InfoResponse, username string) (int, error) {
					if username == "from_user" {
						res.Coins = 100
						return 1, nil
					}
					if username == "to_user" {
						res.Coins = 50
						return 2, nil
					}
					return 0, errors.New("unknown user")
				}
				mockStorage.SendCoinsFunc = func(username string, fromUserID int, toUserID int, scr *storage.SendCoinRequest) error {
					return errors.New("send coins error")
				}
			},
			expectedError: ErrInternalServer,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := &storage.IStorageMock{}
			service := NewService(mockStorage)

			if tt.setupMocks != nil {
				tt.setupMocks(mockStorage)
			}

			err := service.Send(tt.fromUsername, tt.scr)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
