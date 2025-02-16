package mysql

import (
	"avito-shop/internal/config"
	"avito-shop/internal/service/shop/storage"
	"database/sql"
	"errors"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type Storage struct {
	db *sql.DB
}

func (s *Storage) GetDB() *sql.DB {
	return s.db
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{db: db}
}

func New(cfg config.DB) (*Storage, error) {
	const op = "storage.mysql.New"

	connect := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", cfg.Username, cfg.Password, cfg.Host, cfg.Port, cfg.Database)
	db, err := sql.Open("mysql", connect)
	if err != nil {
		return nil, fmt.Errorf("%v: %w", op, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("%v: %w", op, err)
	}

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
            id INT AUTO_INCREMENT PRIMARY KEY,
            username VARCHAR(255) UNIQUE NOT NULL,
    		password_hash VARCHAR(255) NOT NULL,
            coins INT DEFAULT 1000
        );`,
		`CREATE TABLE IF NOT EXISTS transactions (
            id INT AUTO_INCREMENT PRIMARY KEY,
            from_user_id INT,
            to_user_id INT NOT NULL,
            amount INT NOT NULL,
            FOREIGN KEY (from_user_id) REFERENCES users(id),
            FOREIGN KEY (to_user_id) REFERENCES users(id)
        );`,
		`CREATE TABLE IF NOT EXISTS inventory (
            id INT AUTO_INCREMENT PRIMARY KEY,
            user_id INT NOT NULL,
            item_name VARCHAR(255) NOT NULL,
            quantity INT DEFAULT 0,
            FOREIGN KEY (user_id) REFERENCES users(id),
    		UNIQUE unique_user_item (user_id, item_name)
        );`,
	}

	for _, query := range queries {
		_, err = db.Exec(query)
		if err != nil {
			return nil, fmt.Errorf("%v: %w", op, err)
		}
	}

	return &Storage{db: db}, nil
}

func (s *Storage) AddNewUser(username, passwordHash string) error {
	stmt, err := s.db.Prepare("INSERT INTO users (username, password_hash) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(username, passwordHash)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) CheckAuth(username string) (string, error) {
	var storedPasswordHash string
	err := s.db.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&storedPasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrUserNotFound
		}
		return "", err
	}
	return storedPasswordHash, nil
}

func (s *Storage) GetInfo(ir *storage.InfoResponse, username string) (int, error) {
	var id int

	stmt, err := s.db.Prepare("SELECT id, coins FROM users WHERE username = ?;")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&id, &ir.Coins)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, storage.ErrUserNotFound
		}
		return 0, err
	}

	return id, nil
}

func (s *Storage) GetInventory(ir *storage.InfoResponse, id int) error {
	stmt, err := s.db.Prepare("SELECT item_name, quantity FROM inventory WHERE user_id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrUserNotFound
		}
		return err
	}

	for rows.Next() {
		var i storage.Inventory
		err = rows.Scan(&i.Type, &i.Quantity)
		if err != nil {
			return err
		}
		ir.Inventory = append(ir.Inventory, i)
	}
	defer rows.Close()
	return nil
}

func (s *Storage) GetReceivedHistory(ir *storage.InfoResponse, id int) error {
	stmt, err := s.db.Prepare("SELECT from_user_id, amount FROM transactions WHERE to_user_id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrUserNotFound
		}
		return err
	}

	for rows.Next() {
		var i storage.TransactionIn
		err = rows.Scan(&i.FromUser, &i.Amount)
		if err != nil {
			return err
		}
		ir.CoinHistory.Received = append(ir.CoinHistory.Received, i)
	}

	defer rows.Close()
	return nil
}

func (s *Storage) GetSendHistory(ir *storage.InfoResponse, id int) error {
	stmt, err := s.db.Prepare("SELECT to_user_id, amount FROM transactions WHERE from_user_id = ?;")
	if err != nil {
		return err
	}
	defer stmt.Close()

	rows, err := stmt.Query(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return storage.ErrUserNotFound
		}
		return err
	}

	for rows.Next() {
		var i storage.TransactionOut
		err = rows.Scan(&i.ToUser, &i.Amount)
		if err != nil {
			return err
		}
		ir.CoinHistory.Sent = append(ir.CoinHistory.Sent, i)
	}
	defer rows.Close()
	return nil
}

func (s *Storage) BuyItem(name, item string, amount int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			log.Println("Transaction rolled back")
		}
	}()
	query := `INSERT INTO inventory (user_id, item_name, quantity)
			VALUES ((SELECT id FROM users WHERE username = ?), ?, 1)
  			ON DUPLICATE KEY UPDATE quantity = quantity + 1;`
	_, err = tx.Exec(query, name, item)
	if err != nil {
		return err
	}
	_, err = tx.Exec("UPDATE users SET coins = coins - ? WHERE username = ?", amount, name)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}

func (s *Storage) SendCoins(username string, fromUserID int, toUserID int, scr *storage.SendCoinRequest) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			log.Println("Transaction rolled back")
		}
	}()
	query := `
	  UPDATE users
	  SET coins = CASE
	      WHEN username = ? THEN coins - ?
	      WHEN username = ? THEN coins + ?
	  END
	  WHERE username IN (?, ?);
	`
	_, err = tx.Exec(query, username, scr.Amount, scr.ToUser, scr.Amount, username, scr.ToUser)
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO transactions(from_user_id, to_user_id, amount) VALUES (?, ?, ?);", fromUserID, toUserID, scr.Amount)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	return nil
}
