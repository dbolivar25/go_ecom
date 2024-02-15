package main

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v4"
	_ "github.com/joho/godotenv/autoload"
	"github.com/lib/pq"
)

type Storage interface {
	// AdminAccount
	CreateAdminAccount(*AdminAccount) error
	LoginAdminAccount(string, string) (string, error)
	UpdateAdminAccount(*AdminAccount) error
	GetAdminAccount(int32) (*AdminAccount, error)
	DeleteAdminAccount(int32) error
	GetAdminAccounts() ([]*AdminAccount, error)

	// UserAccount
	CreateUserAccount(*UserAccount) error
	LoginUserAccount(string, string) (string, error)
	UpdateUserAccount(*UserAccount) error
	GetUserAccount(int32) (*UserAccount, error)
	DeleteUserAccount(int32) error
	AddItemToUserAccount(int32, int32) error
	RemoveItemFromUserAccount(int32, int32) error
	ClearUserItems(int32) error
	GetUserAccounts() ([]*UserAccount, error)

	// Item
	CreateItem(*Item) error
	UpdateItem(*Item) error
	GetItem(int32) (*Item, error)
	DeleteItem(int32) error
	GetItems() ([]*Item, error)
	GetItemsById([]int32) ([]*Item, float64, error)

	// Order
	CreateOrder(*Order) error
	UpdateOrder(*Order) error
	GetOrder(int32) (*Order, error)
	DeleteOrder(int32) error
	GetOrders() ([]*Order, error)
	GetOrdersById([]int32) ([]*Order, error)

	Init() error
	Close()
}

type PostgresStorage struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStorage, error) {
	user := os.Getenv("POSTGRES_USER")
	dbName := os.Getenv("POSTGRES_NAME")
	password := os.Getenv("POSTGRES_PASS")

	connStr := fmt.Sprintf("user=%s dbname=%s password=%s sslmode=disable", user, dbName, password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &PostgresStorage{db: db}, nil
}

func (self *PostgresStorage) Init() error {
	if err := self.createAdminAccountTable(); err != nil {
		return err
	}

	if err := self.createUserAccountTable(); err != nil {
		return err
	}

	if err := self.createItemTable(); err != nil {
		return err
	}

	if err := self.createOrderTable(); err != nil {
		return err
	}

	return nil
}

func (self *PostgresStorage) createAdminAccountTable() error {
	_, err := self.db.Exec(`
      CREATE TABLE IF NOT EXISTS admins (
        id SERIAL PRIMARY KEY,
        username TEXT NOT NULL,
        hashed_password TEXT NOT NULL,
        auth_token TEXT,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )
    `)
	if err != nil {
		return err
	}

	rootUser := os.Getenv("ROOT_USER")
	rootPass := os.Getenv("ROOT_PASS")
	if rootUser == "" || rootPass == "" {
		return fmt.Errorf("ROOT_USER and ROOT_PASS must be set")
	}

	rootAccount, err := NewAdminAccount(rootUser, rootPass)
	if err != nil {
		return err
	}

	_, err = self.db.Exec(`
      INSERT INTO admins (id, username, hashed_password)
      VALUES (1 ,$1, $2)
      ON CONFLICT DO NOTHING
    `, rootAccount.Username, rootAccount.HashedPassword)
	return err
}

func (self *PostgresStorage) createUserAccountTable() error {
	_, err := self.db.Exec(`
      CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        username TEXT NOT NULL,
        hashed_password TEXT NOT NULL,
        auth_token TEXT,
        items INT[],
        orders INT[],
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
      )
    `)

	return err
}

func (self *PostgresStorage) createItemTable() error {
	_, err := self.db.Exec(`
    CREATE TABLE IF NOT EXISTS items (
      id SERIAL PRIMARY KEY,
      name TEXT NOT NULL,
      description TEXT,
      price FLOAT,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
  `)

	return err
}

func (self *PostgresStorage) createOrderTable() error {
	_, err := self.db.Exec(`
    CREATE TABLE IF NOT EXISTS orders (
      id SERIAL PRIMARY KEY,
      user_id INT,
      items INT[],
      total FLOAT,
      status TEXT,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )
  `)

	return err
}

func (self *PostgresStorage) CreateAdminAccount(account *AdminAccount) error {
	var id int
	err := self.db.QueryRow(`
    INSERT INTO admins (username, hashed_password, created_at)
    VALUES ($1, $2, $3)
    RETURNING id
  `, account.Username, account.HashedPassword, account.CreatedAt).Scan(&id)
	if err != nil {
		return err
	}

	account.ID = uint32(id)

	return nil
}

func (self *PostgresStorage) CreateUserAccount(account *UserAccount) error {
	var id int
	err := self.db.QueryRow(`
    INSERT INTO users (username, hashed_password, items, orders, created_at)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id
  `, account.Username, account.HashedPassword, pq.Array(account.Items), pq.Array(account.Orders), account.CreatedAt).Scan(&id)
	if err != nil {
		return err
	}

	account.ID = uint32(id)

	return nil
}

func (self *PostgresStorage) UpdateAdminAccount(account *AdminAccount) error {
	res, err := self.db.Exec(`
    UPDATE admins
    SET username = $1
    WHERE id = $2
  `, account.Username, account.ID)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("Account %d not found", account.ID)
	}

	return nil
}

func generateToken(id uint32, username string, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":       id,
		"username": username,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	return token.SignedString([]byte(secret))
}

func (self *PostgresStorage) LoginAdminAccount(username, password string) (string, error) {
	rows, err := self.db.Query(`
    SELECT * FROM admins WHERE username = $1
  `, username)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		account, err := scanAdminAccount(rows)
		if err != nil {
			return "", err
		}

		if match, err := argon2id.ComparePasswordAndHash(password, account.HashedPassword); err != nil {
			return "", err
		} else if !match {
			return "", fmt.Errorf("Invalid password")
		} else {
			// do NOTHING
		}

		// update auth token in db then return it
		secret := os.Getenv("JWT_SECRET")
		token, err := generateToken(account.ID, account.Username, secret)
		if err != nil {
			return "", err
		}

		_, err = self.db.Exec(`
      UPDATE admins
      SET auth_token = $1
      WHERE id = $2 
    `, token, account.ID)
		if err != nil {
			return "", err
		}

		return token, nil
	}

	return "", fmt.Errorf("Account %s not found", username)
}

func (self *PostgresStorage) LoginUserAccount(username, password string) (string, error) {
	rows, err := self.db.Query(`
    SELECT * FROM users WHERE username = $1
  `, username)
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		account, err := scanUserAccount(rows)
		if err != nil {
			return "", err
		}

		if match, err := argon2id.ComparePasswordAndHash(password, account.HashedPassword); err != nil {
			return "", err
		} else if !match {
			return "", fmt.Errorf("Invalid password")
		} else {
			// do nothing
		}

		// update auth token in db then return it
		secret := os.Getenv("JWT_SECRET")
		token, err := generateToken(account.ID, account.Username, secret)
		if err != nil {
			return "", err
		}

		_, err = self.db.Exec(`
      UPDATE users
      SET auth_token = $1
      WHERE id = $2 
    `, token, account.ID)
		if err != nil {
			return "", err
		}

		return token, nil
	}

	return "", fmt.Errorf("Account %s not found", username)
}

func (self *PostgresStorage) UpdateUserAccount(account *UserAccount) error {
	res, err := self.db.Exec(`
    UPDATE users
    SET username = $1
    WHERE id = $2
  `, account.Username, account.ID)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("Account %d not found", account.ID)
	}

	return nil
}

func (self *PostgresStorage) GetAdminAccount(id int32) (*AdminAccount, error) {
	rows, err := self.db.Query(`
    SELECT * FROM admins WHERE id = $1
  `, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		return scanAdminAccount(rows)
	}

	return nil, fmt.Errorf("Account %d not found", id)
}

func (self *PostgresStorage) GetUserAccount(id int32) (*UserAccount, error) {
	rows, err := self.db.Query(`
    SELECT * FROM users WHERE id = $1
  `, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		return scanUserAccount(rows)
	}

	return nil, fmt.Errorf("Account %d not found", id)
}

func (self *PostgresStorage) DeleteAdminAccount(id int32) error {
	res, err := self.db.Exec(`
    DELETE FROM admins WHERE id = $1
  `, id)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("Account %d not found", id)
	}

	return nil
}

func (self *PostgresStorage) DeleteUserAccount(id int32) error {
	res, err := self.db.Exec(`
    DELETE FROM users WHERE id = $1
  `, id)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("Account %d not found", id)
	}

	return nil
}

func (self *PostgresStorage) AddItemToUserAccount(accountID, itemID int32) error {
	rows, err := self.db.Query(`
    SELECT * FROM items WHERE id = $1
  `, itemID)
	if err != nil {
		return fmt.Errorf("Item %d not found", itemID)
	}
	defer rows.Close()

	if !rows.Next() {
		return fmt.Errorf("Item %d not found", itemID)
	}

	res2, err := self.db.Exec(`
    UPDATE users 
    SET items = CASE
      WHEN $1 = ANY(items) THEN items
      ELSE array_append(items, $1)
    END
    WHERE id = $2
  `, itemID, accountID)
	if err != nil {
		return err
	}

	if count, _ := res2.RowsAffected(); count == 0 {
		return fmt.Errorf("Account %d not found", accountID)
	}

	return nil
}

func (self *PostgresStorage) RemoveItemFromUserAccount(accountID, itemID int32) error {
	rows, err := self.db.Query(`
    SELECT * FROM items WHERE id = $1
  `, itemID)
	if err != nil {
		return fmt.Errorf("Item %d not found", itemID)
	}
	defer rows.Close()

	if !rows.Next() {
		return fmt.Errorf("Item %d not found", itemID)
	}

	res, err := self.db.Exec(`
    UPDATE users 
    SET items = array_remove(items, $1)
    WHERE id = $2
  `, itemID, accountID)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("Item %d in account %d not found", itemID, accountID)
	}

	return nil
}

func (self *PostgresStorage) ClearUserItems(accountID int32) error {
	_, err := self.db.Exec(`
    UPDATE users
    SET items = '{}'
    WHERE id = $1
  `, accountID)

	return err
}

func (self *PostgresStorage) GetAdminAccounts() ([]*AdminAccount, error) {
	rows, err := self.db.Query(`
    SELECT * FROM admins
  `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]*AdminAccount, 0)
	for rows.Next() {
		account, err := scanAdminAccount(rows)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func (self *PostgresStorage) GetUserAccounts() ([]*UserAccount, error) {
	rows, err := self.db.Query(`
    SELECT * FROM users
  `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	accounts := make([]*UserAccount, 0)
	for rows.Next() {
		account, err := scanUserAccount(rows)
		if err != nil {
			return nil, err
		}

		accounts = append(accounts, account)
	}

	return accounts, nil
}

func scanAdminAccount(row *sql.Rows) (*AdminAccount, error) {
	account := new(AdminAccount)
	throwaway := new(string)

	err := row.Scan(
		&account.ID,
		&account.Username,
		&account.HashedPassword,
		&throwaway,
		&account.CreatedAt,
	)

	return account, err
}

func scanUserAccount(row *sql.Rows) (*UserAccount, error) {
	account := new(UserAccount)
	throwaway := new(string)

	err := row.Scan(
		&account.ID,
		&account.Username,
		&account.HashedPassword,
		&throwaway,
		pq.Array(&account.Items),
		pq.Array(&account.Orders),
		&account.CreatedAt,
	)

	return account, err
}

func (self *PostgresStorage) CreateItem(item *Item) error {
	var id int
	err := self.db.QueryRow(`
    INSERT INTO items (name, description, price, created_at)
    VALUES ($1, $2, $3, $4)
    RETURNING id
  `, item.Name, item.Description, item.Price, item.CreatedAt).Scan(&id)
	if err != nil {
		return err
	}

	item.ID = uint32(id)

	return nil
}

func (self *PostgresStorage) GetItem(id int32) (*Item, error) {
	rows, err := self.db.Query(`
    SELECT * FROM items WHERE id = $1
  `, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		return scanItem(rows)
	}

	return nil, fmt.Errorf("Item %d not found", id)
}

func (self *PostgresStorage) DeleteItem(id int32) error {
	res, err := self.db.Exec(`
    DELETE FROM items WHERE id = $1
  `, id)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("Item %d not found", id)
	}

	_, err2 := self.db.Exec(`
    UPDATE users 
    SET items = array_remove(items, $1) 
  `, id)

	return err2
}

func (self *PostgresStorage) UpdateItem(item *Item) error {
	res, err := self.db.Exec(`
    UPDATE items 
    SET name = $1, description = $2, price = $3 
    WHERE id = $4
  `, item.Name, item.Description, item.Price, item.ID)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("Item %d not found", item.ID)
	}
	return nil
}

func (self *PostgresStorage) GetItems() ([]*Item, error) {
	rows, err := self.db.Query(`
    SELECT * FROM items
  `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*Item, 0)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, nil
}

func (self *PostgresStorage) GetItemsById(ids []int32) ([]*Item, float64, error) {
	rows, err := self.db.Query(`
    SELECT * FROM items WHERE id = ANY($1)
  `, pq.Array(ids))
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var total float64

	items := make([]*Item, 0)
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, 0, err
		}

		items = append(items, item)
		total += item.Price
	}

	return items, total, nil
}

func scanItem(row *sql.Rows) (*Item, error) {
	item := new(Item)

	err := row.Scan(
		&item.ID,
		&item.Name,
		&item.Description,
		&item.Price,
		&item.CreatedAt,
	)

	return item, err
}

func (self *PostgresStorage) CreateOrder(order *Order) error {
	var id int
	err := self.db.QueryRow(`
    INSERT INTO orders (user_id, items, total, status, created_at)
    VALUES ($1, $2, $3, $4, $5)
    RETURNING id
  `, order.UserID, pq.Array(order.Items), order.Total, order.Status, order.CreatedAt).Scan(&id)
	if err != nil {
		return err
	}

	order.ID = uint32(id)

	err = self.db.QueryRow(`
    UPDATE users
    SET orders = array_append(orders, $1)
    WHERE id = $2 
    RETURNING id 
  `, order.ID, order.UserID).Scan(&id)
	if err != nil {
		return err
	}

	if id == 0 {
		return fmt.Errorf("User %d not found", order.UserID)
	}

	return nil
}

func (self *PostgresStorage) GetOrder(id int32) (*Order, error) {
	rows, err := self.db.Query(`
    SELECT * FROM orders WHERE id = $1
  `, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		return scanOrder(rows)
	}

	return nil, fmt.Errorf("Order %d not found", id)
}

func (self *PostgresStorage) DeleteOrder(id int32) error {
	res, err := self.db.Exec(`
    DELETE FROM orders WHERE id = $1
  `, id)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("Order %d not found", id)
	}

	return nil
}

func (self *PostgresStorage) UpdateOrder(order *Order) error {
	res, err := self.db.Exec(`
    UPDATE orders 
    SET status = $1
    WHERE id = $2
  `, order.Status, order.ID)
	if err != nil {
		return err
	}

	if count, _ := res.RowsAffected(); count == 0 {
		return fmt.Errorf("Order %d not found", order.ID)
	}
	return nil
}

func (self *PostgresStorage) GetOrders() ([]*Order, error) {
	rows, err := self.db.Query(`
    SELECT * FROM orders
  `)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*Order, 0)
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func (self *PostgresStorage) GetOrdersById(ids []int32) ([]*Order, error) {
	rows, err := self.db.Query(`
    SELECT * FROM orders WHERE id = ANY($1)
  `, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orders := make([]*Order, 0)
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}

		orders = append(orders, order)
	}

	return orders, nil
}

func scanOrder(row *sql.Rows) (*Order, error) {
	order := new(Order)

	err := row.Scan(
		&order.ID,
		&order.UserID,
		pq.Array(&order.Items),
		&order.Total,
		&order.Status,
		&order.CreatedAt,
	)

	return order, err
}

func (self *PostgresStorage) Close() {
	self.db.Close()
}
