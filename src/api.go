package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

type apiFunc func(http.ResponseWriter, *http.Request) error

type ApiError struct {
	Error string `json:"error"`
}

func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Status", strconv.Itoa(status))
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func makeHTTPHandlerFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func(start time.Time) {
			log.Printf("REQUEST: %s %s STATUS: %s DURATION: %s\n", r.Method, r.URL.Path, w.Header().Get("Status"), time.Since(start))
		}(time.Now())

		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
		}
	}
}

func validateJWT(token string) (*jwt.Token, error) {
	envSecret := os.Getenv("JWT_SECRET")

	return jwt.Parse(token, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(envSecret), nil
	})
}

func withJWTAdminAuth(handler http.HandlerFunc, storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("auth_token")

		// TODO: remove this
		if authHeader == "root_token" {
			handler(w, r)
			return
		}

		token, err := validateJWT(authHeader)
		if err != nil || !token.Valid {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		log.Println(claims)

		// verify that the user id in the url path matches the user id in the Token
		id, err := getID(r)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			return
		}

		account, err := storage.GetAdminAccount(int32(id))
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			return
		}

		if claims["username"] != account.Username {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		handler(w, r)
	}
}

func withJWTUserAuth(handler http.HandlerFunc, storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("auth_token")

		// TODO: remove this
		if authHeader == "root_token" {
			handler(w, r)
			return
		}

		token, err := validateJWT(authHeader)
		if err != nil || !token.Valid {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		log.Println(claims)

		// verify that the user id in the url path matches the user id in the Token
		id, err := getID(r)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			return
		}

		account, err := storage.GetUserAccount(int32(id))
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			return
		}

		if claims["username"] != account.Username {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		handler(w, r)
	}
}

type APIServer struct {
	portAddress string
	storage     Storage
}

func NewAPIServer(portAddress string, storage Storage) *APIServer {
	return &APIServer{
		portAddress: portAddress,
		storage:     storage,
	}
}

func (self *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/admin/{id}/dashboard", withJWTAdminAuth(makeHTTPHandlerFunc(self.handleAdminAccessDashboard), self.storage))
	router.HandleFunc("/admin/{id}/admins", withJWTAdminAuth(makeHTTPHandlerFunc(self.handleAdminAccessAdmins), self.storage))
	router.HandleFunc("/admin/{id}/users", withJWTAdminAuth(makeHTTPHandlerFunc(self.handleAdminAccessUsers), self.storage))
	router.HandleFunc("/admin/{id}/items", withJWTAdminAuth(makeHTTPHandlerFunc(self.handleAdminAccessItems), self.storage))
	router.HandleFunc("/user", makeHTTPHandlerFunc(self.handleNewUser))
	router.HandleFunc("/user/{id}", withJWTUserAuth(makeHTTPHandlerFunc(self.handleAccessUser), self.storage))
	router.HandleFunc("/user/{id}/items", withJWTUserAuth(makeHTTPHandlerFunc(self.handleAccessUserItems), self.storage))
	router.HandleFunc("/user/{id}/checkout", withJWTUserAuth(makeHTTPHandlerFunc(self.handleAccessUserCheckout), self.storage))
	router.HandleFunc("/items", makeHTTPHandlerFunc(self.handleAccessItems))
	router.HandleFunc("/items/{id}", makeHTTPHandlerFunc(self.handleAccessItem))

	log.Println("Running on port", self.portAddress)

	http.ListenAndServe(self.portAddress, router)
}

func (self *APIServer) handleAdminAccessDashboard(w http.ResponseWriter, r *http.Request) error {
	return WriteJSON(w, http.StatusOK, map[string]string{"message": "Welcome to the admin dashboard"})
}

func (self *APIServer) handleAdminAccessAdmins(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetAdminAccounts(w, r)
	case "POST":
		return self.handleCreateAdminAccount(w, r)
	case "DELETE":
		return self.handleDeleteAdminAccount(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleAdminAccessUsers(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetUserAccounts(w, r)
	case "POST":
		return self.handleCreateUserAccount(w, r)
	case "DELETE":
		return self.handleDeleteUserAccount(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleAdminAccessItems(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetItems(w, r)
	case "POST":
		return self.handleCreateItem(w, r)
	case "DELETE":
		return self.handleDeleteItem(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleNewUser(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "POST":
		return self.handleCreateUserAccount(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleAccessUser(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetUserAccount(w, r)
	case "PUT":
		return self.handleUpdateUserAccount(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleAccessUserItems(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetUserItems(w, r)
	case "POST":
		return self.handleAddItemToUserAccount(w, r)
	case "DELETE":
		return self.handleRemoveItemFromUserAccount(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleAccessUserCheckout(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "POST":
		return self.handleCheckoutUserAccount(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleAccessItems(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetItems(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleAccessItem(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetItem(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

// method specific handlers

func (self *APIServer) handleGetAdminAccounts(w http.ResponseWriter, r *http.Request) error {
	accounts, err := self.storage.GetAdminAccounts()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

func (self *APIServer) handleCreateAdminAccount(w http.ResponseWriter, r *http.Request) error {
	createAdminAccountRequest := new(CreateAccountRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&createAdminAccountRequest); err != nil {
		return err
	}

	account := NewAdminAccount(createAdminAccountRequest.Username)
	if err := self.storage.CreateAdminAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (self *APIServer) handleDeleteAdminAccount(w http.ResponseWriter, r *http.Request) error {
	deleteAdminAccountRequest := new(DeleteAccountRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&deleteAdminAccountRequest); err != nil {
		return err
	}

	if err := self.storage.DeleteAdminAccount(deleteAdminAccountRequest.ID); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int32{"deleted account": deleteAdminAccountRequest.ID})
}

func (self *APIServer) handleGetUserAccounts(w http.ResponseWriter, r *http.Request) error {
	accounts, err := self.storage.GetUserAccounts()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

func (self *APIServer) handleCreateUserAccount(w http.ResponseWriter, r *http.Request) error {
	createUserAccountRequest := new(CreateAccountRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&createUserAccountRequest); err != nil {
		return err
	}

	account := NewUserAccount(createUserAccountRequest.Username)
	if err := self.storage.CreateUserAccount(account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (self *APIServer) handleDeleteUserAccount(w http.ResponseWriter, r *http.Request) error {
	deleteUserAccountRequest := new(DeleteAccountRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&deleteUserAccountRequest); err != nil {
		return err
	}

	if err := self.storage.DeleteUserAccount(deleteUserAccountRequest.ID); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int32{"deleted account": deleteUserAccountRequest.ID})
}

func (self *APIServer) handleGetUserAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	account, err := self.storage.GetUserAccount(int32(id))
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, account)
}

func (self *APIServer) handleUpdateUserAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	updateUserAccountRequest := new(UpdateAccountRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&updateUserAccountRequest); err != nil {
		return err
	}

	account := UserAccount{
		ID:       uint32(id),
		Username: updateUserAccountRequest.Username,
	}

	if err := self.storage.UpdateUserAccount(&account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int32{"updated account": id})
}

func (self *APIServer) handleGetUserItems(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	account, err := self.storage.GetUserAccount(int32(id))
	if err != nil {
		return err
	}

	items, total, err := self.storage.GetItemsById(account.Items)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, struct {
		Items []*Item `json:"items"`
		Total float64 `json:"total"`
	}{items, total})
}

func (self *APIServer) handleAddItemToUserAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	addItemRequest := new(AddItemRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&addItemRequest); err != nil {
		return err
	}

	if err := self.storage.AddItemToUserAccount(id, addItemRequest.ItemID); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int32{"added item": addItemRequest.ItemID, "to account": id})
}

func (self *APIServer) handleRemoveItemFromUserAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	removeItemRequest := new(RemoveItemRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&removeItemRequest); err != nil {
		return err
	}

	if err := self.storage.RemoveItemFromUserAccount(id, removeItemRequest.ItemID); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int32{"removed item": removeItemRequest.ItemID, "from account": id})
}

func (self *APIServer) handleCheckoutUserAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	account, err := self.storage.GetUserAccount(int32(id))
	if err != nil {
		return err
	}

	items, total, err := self.storage.GetItemsById(account.Items)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, struct {
		Items []*Item `json:"items"`
		Total float64 `json:"total"`
	}{items, total})
}

func (self *APIServer) handleGetItems(w http.ResponseWriter, r *http.Request) error {
	items, err := self.storage.GetItems()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, items)
}

func (self *APIServer) handleCreateItem(w http.ResponseWriter, r *http.Request) error {
	createItemRequest := new(CreateItemRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&createItemRequest); err != nil {
		return err
	}

	item := NewItem(createItemRequest.Name, createItemRequest.Description, createItemRequest.Price)
	if err := self.storage.CreateItem(item); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, item)
}

func (self *APIServer) handleDeleteItem(w http.ResponseWriter, r *http.Request) error {
	deleteItemRequest := new(DeleteItemRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&deleteItemRequest); err != nil {
		return err
	}

	if err := self.storage.DeleteItem(deleteItemRequest.ID); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, map[string]int32{"deleted item": deleteItemRequest.ID})
}

func (self *APIServer) handleGetItem(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	item, err := self.storage.GetItem(int32(id))
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, item)
}

func getID(r *http.Request) (int32, error) {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("Invalid id: \"%s\"", idStr)
	}

	return int32(id), nil
}
