package shop

import (
	"errors"
	"sync"
	"time"

	"avito-shop/internal/service/shop/storage"

	"github.com/golang-jwt/jwt/v5"
)

type Service struct {
	Storage storage.IStorage
}

func NewService(storage storage.IStorage) *Service {
	return &Service{
		Storage: storage,
	}
}

type IService interface {
	CollectAllInfo(username string) (*storage.InfoResponse, error)
	Send(fromUsername string, scr *storage.SendCoinRequest) error
	Purchase(username, item string) error
}

func GenerateJWT(secretKey string, username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}
	if secretKey == "" {
		return "", ErrInternalServer
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenSign, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return "", ErrInternalServer
	}

	return tokenSign, nil
}

func (s *Service) CollectAllInfo(username string) (*storage.InfoResponse, error) {
	var res storage.InfoResponse
	var wg sync.WaitGroup
	var mu sync.Mutex
	var errs []error

	id, err := s.Storage.GetInfo(&res, username)
	if err != nil {
		return nil, ErrInternalServer
	}

	runCollect := func(fn func(*storage.InfoResponse, int) error) {
		defer wg.Done()
		err := fn(&res, id)
		if err != nil {
			mu.Lock()
			errs = append(errs, ErrInternalServer)
			mu.Unlock()
		}
	}
	wg.Add(3)
	go runCollect(s.Storage.GetInventory)
	go runCollect(s.Storage.GetSendHistory)
	go runCollect(s.Storage.GetReceivedHistory)

	wg.Wait()

	if len(errs) > 0 {
		return nil, errs[0]
	}

	return &res, nil
}

func (s *Service) Send(fromUsername string, scr *storage.SendCoinRequest) error {
	var (
		infoResponseFrom, infoResponseTo storage.InfoResponse
		fromUserID, toUserID             int
		wg                               sync.WaitGroup
		mu                               sync.Mutex
		errs                             []error
	)

	fetchUserInfo := func(username string, infoResponse *storage.InfoResponse, userID *int) {
		defer wg.Done()
		id, err := s.Storage.GetInfo(infoResponse, username)
		if err != nil {
			if errors.Is(err, storage.ErrUserNotFound) {
				mu.Lock()
				errs = append(errs, ErrUserNotFound)
				mu.Unlock()
				return
			}
			mu.Lock()
			errs = append(errs, ErrInternalServer)
			mu.Unlock()
			return
		}
		*userID = id
	}

	wg.Add(2)
	go fetchUserInfo(fromUsername, &infoResponseFrom, &fromUserID)
	go fetchUserInfo(scr.ToUser, &infoResponseTo, &toUserID)
	wg.Wait()

	if len(errs) > 0 {
		return errs[0]
	}

	if infoResponseFrom.Coins < scr.Amount {
		return ErrInsufficientFunds
	}

	err := s.Storage.SendCoins(fromUsername, fromUserID, toUserID, scr)
	if err != nil {
		return ErrInternalServer
	}

	return nil
}

func (s *Service) Purchase(username, item string) error {
	price, exists := storage.MerchItems[item]
	if !exists {
		return ErrItemNotFound
	}

	var infoResponse storage.InfoResponse
	_, err := s.Storage.GetInfo(&infoResponse, username)
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			return storage.ErrUserNotFound
		}
		return ErrInternalServer
	}

	if infoResponse.Coins < price {
		return ErrInsufficientFunds
	}

	err = s.Storage.BuyItem(username, item, price)
	if err != nil {
		return ErrInternalServer
	}

	return nil
}
