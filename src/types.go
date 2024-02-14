package main

import (
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

type APIServer struct {
	portAddress string
	storage     Storage
}

type CreateAccountRequest struct {
	Username string `json:"user"`
}

type DeleteAccountRequest struct {
	ID int32 `json:"id"`
}

type AddItemRequest struct {
	ItemID int32 `json:"item_id"`
}

type RemoveItemRequest struct {
	ItemID int32 `json:"item_id"`
}

type UpdateAccountRequest struct {
	Username string `json:"user"`
}

type CreateItemRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"desc"`
	Price       float64 `json:"price"`
}

type DeleteItemRequest struct {
	ID int32 `json:"id"`
}

type UpdateItemRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"desc"`
	Price       float64 `json:"price"`
}

type CreateOrderRequest struct {
	AccountID int32   `json:"account_id"`
	Items     []int32 `json:"items"`
	Total     float64 `json:"total"`
}

type DeleteOrderRequest struct {
	ID int32 `json:"id"`
}

type UpdateOrderRequest struct {
	Status string `json:"status"`
}

type Item struct {
	ID          uint32    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"desc"`
	Price       float64   `json:"price"`
	CreatedAt   time.Time `json:"created_at"`
}

func NewItem(name, description string, price float64) *Item {
	return &Item{
		Name:        name,
		Description: description,
		Price:       price,
		CreatedAt:   time.Now().UTC(),
	}
}

type AdminAccount struct {
	ID        uint32    `json:"id"`
	Username  string    `json:"username"`
	AuthToken string    `json:"auth_token"`
	CreatedAt time.Time `json:"created_at"`
}

func NewAdminAccount(username string) *AdminAccount {
	env := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
	})

	tokenString, err := token.SignedString([]byte(env))
	if err != nil {
		return nil
	}

	return &AdminAccount{
		Username:  username,
		AuthToken: tokenString,
		CreatedAt: time.Now().UTC(),
	}
}

type UserAccount struct {
	ID        uint32    `json:"id"`
	Username  string    `json:"username"`
	AuthToken string    `json:"auth_token"`
	Items     []int32   `json:"items"`
	Orders    []int32   `json:"orders"`
	CreatedAt time.Time `json:"created_at"`
}

func NewUserAccount(username string) *UserAccount {
	env := os.Getenv("JWT_SECRET")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": username,
	})

	tokenString, err := token.SignedString([]byte(env))
	if err != nil {
		return nil
	}

	return &UserAccount{
		Username:  username,
		AuthToken: tokenString,
		Items:     make([]int32, 0),
    Orders:    make([]int32, 0),
		CreatedAt: time.Now().UTC(),
	}
}

type Order struct {
	ID        uint32    `json:"id"`
	UserID    uint32    `json:"user_id"`
	Items     []int32   `json:"items"`
	Total     float64   `json:"total"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func NewOrder(userID uint32, items []int32, total float64) *Order {
	return &Order{
		UserID:    userID,
		Items:     items,
		Total:     total,
		Status:    "pending",
		CreatedAt: time.Now().UTC(),
	}
}
