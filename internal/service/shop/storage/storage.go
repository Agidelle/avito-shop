package storage

type IStorage interface {
	CheckAuth(username string) (string, error)
	AddNewUser(username, password string) error
	GetInfo(ir *InfoResponse, username string) (int, error)
	GetInventory(ir *InfoResponse, id int) error
	GetReceivedHistory(ir *InfoResponse, id int) error
	GetSendHistory(ir *InfoResponse, id int) error
	BuyItem(name, item string, amount int) error
	SendCoins(username string, fromUserID int, toUserID int, scr *SendCoinRequest) error
}

type InfoResponse struct {
	Coins       int         `json:"coins"`
	Inventory   []Inventory `json:"inventory"`
	CoinHistory CoinHistory `json:"coinHistory"`
}

type Inventory struct {
	Type     string `json:"type"`
	Quantity int    `json:"quantity"`
}

type CoinHistory struct {
	Received []TransactionIn  `json:"received,omitempty"`
	Sent     []TransactionOut `json:"sent,omitempty"`
}

type TransactionIn struct {
	FromUser string `json:"fromUser,omitempty"`
	Amount   int    `json:"amount,omitempty"`
}

type TransactionOut struct {
	ToUser string `json:"toUser,omitempty"`
	Amount int    `json:"amount,omitempty"`
}

type SendCoinRequest struct {
	ToUser string `json:"toUser"`
	Amount int    `json:"amount"`
}

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

var MerchItems = map[string]int{
	"t-shirt":    80,
	"cup":        20,
	"book":       50,
	"pen":        10,
	"powerbank":  200,
	"hoody":      300,
	"umbrella":   200,
	"socks":      10,
	"wallet":     50,
	"pink-hoody": 500,
}
