package main

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

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
		CreatedAt: time.Now().UTC(),
	}
}
