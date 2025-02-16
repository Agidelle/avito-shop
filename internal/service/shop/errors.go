package shop

import "errors"

var (
	ErrItemNotFound      = errors.New("предмет не найден в магазине")
	ErrInsufficientFunds = errors.New("недостаточно средств")
	ErrInternalServer    = errors.New("внутренняя ошибка сервера")
	ErrUserNotFound      = errors.New("пользователь не найден")
)
