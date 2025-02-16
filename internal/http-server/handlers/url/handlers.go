package urls

import (
	"avito-shop/internal/http-server/handlers/auth"
	"avito-shop/internal/service/shop"
	"avito-shop/internal/service/shop/storage"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type Handlers struct {
	service shop.IService
	storage storage.IStorage
	log     *slog.Logger
}
type IHandlers interface {
	Auth(authKey string) http.HandlerFunc
	Info() http.HandlerFunc
	SendCoin() http.HandlerFunc
	BuyItem() http.HandlerFunc
}

func NewHandlers(storage storage.IStorage, service shop.IService, log *slog.Logger) *Handlers {
	return &Handlers{storage: storage, service: service, log: log}
}

func (h *Handlers) writeErrorResponse(w http.ResponseWriter, errorMessage string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"errors": errorMessage})
}

func (h *Handlers) Auth(authKey string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input storage.AuthRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			h.log.Warn("Invalid request body", slog.String("error", err.Error()))
			h.writeErrorResponse(w, "Неверный запрос.", http.StatusBadRequest)
			return
		}
		if input.Username == "" || input.Password == "" {
			h.log.Warn("Invalid request body")
			h.writeErrorResponse(w, "Неверный запрос.", http.StatusBadRequest)
			return
		}
		token, err := auth.AuthenticateUser(h.storage, input.Username, input.Password, authKey)
		if err != nil {
			h.log.Warn("Authentication failed", slog.String("username", input.Username))
			h.writeErrorResponse(w, "Неавторизован.", http.StatusUnauthorized)
			return
		}
		h.log.Info("User authenticated successfully", slog.String("username", input.Username))
		w.Header().Add("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(map[string]string{"token": token})
		if err != nil {
			h.log.Error("Failed to encode response", slog.String("error", err.Error()))
			h.writeErrorResponse(w, fmt.Sprintf("Внутренняя ошибка сервера: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handlers) Info() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)
		w.Header().Add("Content-Type", "application/json")

		resp, err := h.service.CollectAllInfo(username)
		if err != nil {
			h.log.Error("Failed to collect user info", slog.String("error", err.Error()))
			h.writeErrorResponse(w, "Внутренняя ошибка сервера.", http.StatusInternalServerError)
			return
		}

		err = json.NewEncoder(w).Encode(&resp)
		if err != nil {
			h.log.Error("Failed to encode response", slog.String("error", err.Error()))
			h.writeErrorResponse(w, fmt.Sprintf("Внутренняя ошибка сервера: %s", err.Error()), http.StatusInternalServerError)
			return
		}
	}
}

func (h *Handlers) SendCoin() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)

		var input storage.SendCoinRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			h.log.Warn("Invalid request body")
			h.writeErrorResponse(w, "Неверный запрос.", http.StatusBadRequest)
			return
		}

		err := h.service.Send(username, &input)
		if err != nil {
			switch {
			case errors.Is(err, shop.ErrInsufficientFunds):
				h.log.Warn("Insufficient funds")
				h.writeErrorResponse(w, "Недостаточно средств.", http.StatusBadRequest)
			default:
				h.log.Error("Failed to process transaction", slog.String("error", err.Error()))
				h.writeErrorResponse(w, fmt.Sprintf("Внутренняя ошибка сервера: %s", err.Error()), http.StatusInternalServerError)
			}
			return
		}
		h.log.Info("Coins transferred successfully")
		w.WriteHeader(http.StatusOK)
	}
}

func (h *Handlers) BuyItem() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		username := r.Context().Value("username").(string)

		item := r.PathValue("item")
		if item == "" {
			h.log.Warn("Bad Request: item is empty")
			h.writeErrorResponse(w, "Неверный запрос. Предмет не указан.", http.StatusBadRequest)
			return
		}

		err := h.service.Purchase(username, item)
		if err != nil {
			switch {
			case errors.Is(err, shop.ErrItemNotFound):
				h.log.Warn("Item not found", slog.String("item", item))
				h.writeErrorResponse(w, fmt.Sprintf("Предмет '%s' не найден.", item), http.StatusBadRequest)
			case errors.Is(err, shop.ErrInsufficientFunds):
				h.log.Warn("Insufficient funds", slog.String("username", username))
				h.writeErrorResponse(w, "Недостаточно средств.", http.StatusBadRequest)
			default:
				h.log.Error("Failed to process purchase", slog.String("error", err.Error()))
				h.writeErrorResponse(w, fmt.Sprintf("Внутренняя ошибка сервера: %s", err.Error()), http.StatusInternalServerError)
			}
			return
		}

		h.log.Info("Item purchased successfully", slog.String("username", username), slog.String("item", item))
		w.WriteHeader(http.StatusOK)
	}
}
