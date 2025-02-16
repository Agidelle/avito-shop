package mysql

import (
	"avito-shop/internal/service/shop/storage"
	"database/sql"
	"errors"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func NewTestDB(t *testing.T) (*Storage, func()) {
	dsn := "user:password@tcp(127.0.0.1:3306)/test_db"
	//connect := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.DatabaseTest)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	_, err = db.Exec("TRUNCATE TABLE users;")
	if err != nil {
		t.Fatalf("failed to clean up users table: %v", err)
	}
	_, err = db.Exec("TRUNCATE TABLE inventory;")
	if err != nil {
		t.Fatalf("failed to clean up users table: %v", err)
	}
	_, err = db.Exec("TRUNCATE TABLE transactions;")
	if err != nil {
		t.Fatalf("failed to clean up users table: %v", err)
	}

	return NewStorage(db), func() {
		db.Close()
	}
}

func TestAddNewUser_Success(t *testing.T) {
	store, cleanup := NewTestDB(t)
	defer cleanup()

	username := "test_user"
	passwordHash := "hashed_password"

	err := store.AddNewUser(username, passwordHash)
	assert.NoError(t, err)

	var count int
	err = store.GetDB().QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
	assert.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestAddNewUser_DuplicateUser(t *testing.T) {
	store, cleanup := NewTestDB(t)
	defer cleanup()

	username := "test_user"
	passwordHash := "hashed_password"

	err := store.AddNewUser(username, passwordHash)
	assert.NoError(t, err)

	err = store.AddNewUser(username, passwordHash)
	assert.Error(t, err)

}
func TestCheckAuth_Success(t *testing.T) {
	store, cleanup := NewTestDB(t)
	defer cleanup()

	username := "test_user"
	passwordHash := "hashed_password"

	_, err := store.GetDB().Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, passwordHash)
	assert.NoError(t, err)

	storedPasswordHash, err := store.CheckAuth(username)
	assert.NoError(t, err)
	assert.Equal(t, passwordHash, storedPasswordHash)
}

func TestCheckAuth_UserNotFound(t *testing.T) {
	store, cleanup := NewTestDB(t)
	defer cleanup()

	username := "non_existent_user"

	_, err := store.CheckAuth(username)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrUserNotFound))
}

func TestGetInfo_Success(t *testing.T) {
	store, cleanup := NewTestDB(t)
	defer cleanup()

	username := "test_user"
	passwordHash := "hashed_password"
	coins := 100

	_, err := store.GetDB().Exec("INSERT INTO users (username, password_hash, coins) VALUES (?, ?, ?)", username, passwordHash, coins)
	assert.NoError(t, err)

	var infoResponse storage.InfoResponse
	_, err = store.GetInfo(&infoResponse, username)
	assert.NoError(t, err)
	assert.Equal(t, coins, infoResponse.Coins)
}

func TestGetInfo_UserNotFound(t *testing.T) {
	store, cleanup := NewTestDB(t)
	defer cleanup()

	username := "non_existent_user"

	var infoResponse storage.InfoResponse
	_, err := store.GetInfo(&infoResponse, username)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, storage.ErrUserNotFound))
}

func TestGetInventory_Success(t *testing.T) {
	store, cleanup := NewTestDB(t)
	defer cleanup()

	username := "test_user"
	passwordHash := "hashed_password"
	inventory := []storage.Inventory{
		{Type: "sword", Quantity: 1},
	}

	row, err := store.GetDB().Exec("INSERT INTO users (username, password_hash, coins) VALUES (?, ?, ?)", username, passwordHash, 100)
	assert.NoError(t, err)
	userID, _ := row.LastInsertId()
	for _, item := range inventory {
		_, err := store.GetDB().Exec("INSERT INTO inventory (user_id, item_name, quantity) VALUES (?, ?, ?)", userID, item.Type, item.Quantity)
		assert.NoError(t, err)
	}

	var infoResponse storage.InfoResponse
	err = store.GetInventory(&infoResponse, int(userID))
	assert.NoError(t, err)
	assert.Equal(t, inventory, infoResponse.Inventory)
}

func TestGetReceivedHistory_Success(t *testing.T) {
	store, cleanup := NewTestDB(t)
	defer cleanup()

	username := "test_user"
	passwordHash := "hashed_password"
	usernameTo := "test_user_two"
	passwordHashTwo := "hashed_password"

	receivedHistory := []storage.TransactionIn{
		{FromUser: "test_user_two", Amount: 50},
	}

	row, err := store.GetDB().Exec("INSERT INTO users (username, password_hash, coins) VALUES (?, ?, ?)", username, passwordHash, 100)
	assert.NoError(t, err)
	userID, _ := row.LastInsertId()

	rowTwo, err := store.GetDB().Exec("INSERT INTO users (username, password_hash, coins) VALUES (?, ?, ?)", usernameTo, passwordHashTwo, 100)
	assert.NoError(t, err)
	userIDTwo, _ := rowTwo.LastInsertId()

	for _, transaction := range receivedHistory {
		_, err := store.GetDB().Exec("INSERT INTO transactions (from_user_id, to_user_id, amount) VALUES (?, ?, ?)", userIDTwo, userID, transaction.Amount)
		assert.NoError(t, err)
	}

	var infoResponse storage.InfoResponse
	err = store.GetReceivedHistory(&infoResponse, int(userID))
	receivedHistory[0].FromUser = strconv.Itoa(int(userIDTwo))
	assert.NoError(t, err)
	assert.Equal(t, receivedHistory, infoResponse.CoinHistory.Received)
}

func TestBuyItem_Success(t *testing.T) {
	db, cleanup := NewTestDB(t)
	defer cleanup()
	storagex := NewStorage(db.db)

	err := storagex.AddNewUser("testuser", "hashedpassword")
	require.NoError(t, err)

	var initialCoins int
	err = db.db.QueryRow("SELECT coins FROM users WHERE username = ?", "testuser").Scan(&initialCoins)
	require.NoError(t, err)
	require.Equal(t, 1000, initialCoins)

	err = storagex.BuyItem("testuser", "t-shirt", 50)
	require.NoError(t, err)

	var updatedCoins int
	err = db.db.QueryRow("SELECT coins FROM users WHERE username = ?", "testuser").Scan(&updatedCoins)
	require.NoError(t, err)
	require.Equal(t, 950, updatedCoins)

	var quantity int
	err = db.db.QueryRow(`
        SELECT quantity 
        FROM inventory 
        WHERE user_id = (SELECT id FROM users WHERE username = ?) AND item_name = ?
    `, "testuser", "t-shirt").Scan(&quantity)
	require.NoError(t, err)
	require.Equal(t, 1, quantity)
}

func TestSendCoins_Success(t *testing.T) {
	store, cleanup := NewTestDB(t)
	defer cleanup()

	usernameSender := "sender"
	usernameRecipient := "recipient"
	passwordHash := "hashed_password"
	rowSender, err := store.GetDB().Exec("INSERT INTO users (username, password_hash, coins) VALUES (?, ?, ?)", usernameSender, passwordHash, 1000)
	assert.NoError(t, err)
	senderID, _ := rowSender.LastInsertId()
	rowRecipient, err := store.GetDB().Exec("INSERT INTO users (username, password_hash, coins) VALUES (?, ?, ?)", usernameRecipient, passwordHash, 1000)
	assert.NoError(t, err)
	recipientID, _ := rowRecipient.LastInsertId()

	scr := &storage.SendCoinRequest{
		ToUser: usernameRecipient,
		Amount: 50,
	}
	err = store.SendCoins(usernameSender, int(senderID), int(recipientID), scr)
	assert.NoError(t, err)

	var senderUpdatedCoins int
	err = store.GetDB().QueryRow("SELECT coins FROM users WHERE id = ?", senderID).Scan(&senderUpdatedCoins)
	assert.NoError(t, err)
	assert.Equal(t, 950, senderUpdatedCoins)

	var recipientUpdatedCoins int
	err = store.GetDB().QueryRow("SELECT coins FROM users WHERE id = ?", recipientID).Scan(&recipientUpdatedCoins)
	assert.NoError(t, err)
	assert.Equal(t, 1050, recipientUpdatedCoins)

	var transactionAmount int
	err = store.GetDB().QueryRow(`
        SELECT amount 
        FROM transactions 
        WHERE from_user_id = ? AND to_user_id = ?
    `, senderID, recipientID).Scan(&transactionAmount)
	assert.NoError(t, err)
	assert.Equal(t, 50, transactionAmount)
}
