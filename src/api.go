package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/gorilla/mux"
)

func NewAPIServer(portAddress string, storage Storage) *APIServer {
	return &APIServer{
		portAddress: portAddress,
		storage:     storage,
	}
}

func (self *APIServer) Run() {
	router := mux.NewRouter()

	router.HandleFunc("/admin/login", makeHTTPHandlerFunc(self.handleAdminLogin))
	router.HandleFunc("/admin/{id}", withJWTAdminAuth(makeHTTPHandlerFunc(self.handleAdminAccessAdmin), self.storage))
	router.HandleFunc("/admin/{id}/dash", withJWTAdminAuth(makeHTTPHandlerFunc(self.handleAdminAccessDashboard), self.storage))
	router.HandleFunc("/admin/{id}/admins", withJWTAdminAuth(makeHTTPHandlerFunc(self.handleAdminAccessAdmins), self.storage))
	router.HandleFunc("/admin/{id}/users", withJWTAdminAuth(makeHTTPHandlerFunc(self.handleAdminAccessUsers), self.storage))
	router.HandleFunc("/admin/{id}/items", withJWTAdminAuth(makeHTTPHandlerFunc(self.handleAdminAccessItems), self.storage))
	router.HandleFunc("/admin/{id}/items/{item_id}", withJWTAdminAuth(makeHTTPHandlerFunc(self.handleAdminAccessItem), self.storage))
	router.HandleFunc("/admin/{id}/orders", withJWTAdminAuth(makeHTTPHandlerFunc(self.handleAdminAccessOrders), self.storage))
	router.HandleFunc("/admin/{id}/orders/{order_id}", withJWTAdminAuth(makeHTTPHandlerFunc(self.handleAdminAccessOrder), self.storage))

	router.HandleFunc("/user/login", makeHTTPHandlerFunc(self.handleUserLogin))
	router.HandleFunc("/user/signup", makeHTTPHandlerFunc(self.handleNewUser))
	router.HandleFunc("/user/{id}", withJWTUserAuth(makeHTTPHandlerFunc(self.handleAccessUser), self.storage))
	router.HandleFunc("/user/{id}/items", withJWTUserAuth(makeHTTPHandlerFunc(self.handleAccessUserItems), self.storage))
	router.HandleFunc("/user/{id}/checkout", withJWTUserAuth(makeHTTPHandlerFunc(self.handleAccessUserCheckout), self.storage))
	router.HandleFunc("/user/{id}/orders", withJWTUserAuth(makeHTTPHandlerFunc(self.handleAccessUserOrders), self.storage))
	router.HandleFunc("/items", makeHTTPHandlerFunc(self.handleAccessItems))
	router.HandleFunc("/items/{id}", makeHTTPHandlerFunc(self.handleAccessItem))

	log.Println("Running on port", self.portAddress)

	http.ListenAndServe(self.portAddress, router)
}

func (self *APIServer) handleAdminLogin(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "POST":
		return self.handlePostAdminLogin(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleAdminAccessAdmin(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetAdminAccount(w, r)
	case "PUT":
		return self.handleUpdateAdminAccount(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleAdminAccessDashboard(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetDashboard(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
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

func (self *APIServer) handleAdminAccessItem(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetItem(w, r)
	case "PUT":
		return self.handleUpdateItem(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleAdminAccessOrders(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetOrders(w, r)
	case "POST":
		return self.handleCreateOrder(w, r)
	case "DELETE":
		return self.handleDeleteOrder(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleAdminAccessOrder(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetOrder(w, r)
	case "PUT":
		return self.handleUpdateOrder(w, r)
	}

	return fmt.Errorf("Invalid method: \"%s\"", r.Method)
}

func (self *APIServer) handleUserLogin(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "POST":
		return self.handlePostUserLogin(w, r)
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

func (self *APIServer) handleAccessUserOrders(w http.ResponseWriter, r *http.Request) error {
	switch r.Method {
	case "GET":
		return self.handleGetUserOrders(w, r)
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

func (self *APIServer) handlePostAdminLogin(w http.ResponseWriter, r *http.Request) error {
	loginRequest := new(LoginRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&loginRequest); err != nil {
		return err
	}

	auth_token, err := self.storage.LoginAdminAccount(loginRequest.Username, loginRequest.Password)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, struct {
		AuthToken string `json:"auth_token"`
	}{auth_token})
}

func (self *APIServer) handleGetAdminAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	account, err := self.storage.GetAdminAccount(int32(id))
	if err != nil {
		return err
	}

	account.HashedPassword = ""

	return WriteJSON(w, http.StatusOK, account)
}

func (self *APIServer) handleUpdateAdminAccount(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	updateAdminAccountRequest := new(UpdateAccountRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&updateAdminAccountRequest); err != nil {
		return err
	}

	account := AdminAccount{
		ID:       uint32(id),
		Username: updateAdminAccountRequest.Username,
	}

	if err := self.storage.UpdateAdminAccount(&account); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, struct {
		UpdatedAccount int32 `json:"updated_account"`
	}{int32(id)})
}

func (self *APIServer) handleGetDashboard(w http.ResponseWriter, r *http.Request) error {
	admins, err := self.storage.GetAdminAccounts()
	if err != nil {
		return err
	}

	for _, admin := range admins {
		admin.HashedPassword = ""
	}

	users, err := self.storage.GetUserAccounts()
	if err != nil {
		return err
	}

	for _, user := range users {
		user.HashedPassword = ""
	}

	items, err := self.storage.GetItems()
	if err != nil {
		return err
	}

	orders, err := self.storage.GetOrders()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, struct {
		Admins      []*AdminAccount `json:"admins"`
		TotalAdmins int             `json:"total_admins"`
		Users       []*UserAccount  `json:"users"`
		TotalUsers  int             `json:"total_users"`
		Items       []*Item         `json:"items"`
		TotalItems  int             `json:"total_items"`
		Orders      []*Order        `json:"orders"`
		TotalOrders int             `json:"total_orders"`
	}{admins, len(admins), users, len(users), items, len(items), orders, len(orders)})
}

func (self *APIServer) handleGetAdminAccounts(w http.ResponseWriter, r *http.Request) error {
	accounts, err := self.storage.GetAdminAccounts()
	if err != nil {
		return err
	}

	for _, account := range accounts {
		account.HashedPassword = ""
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

	account, err := NewAdminAccount(createAdminAccountRequest.Username, createAdminAccountRequest.Password)
	if err != nil {
		return err
	}

	if err := self.storage.CreateAdminAccount(account); err != nil {
		return err
	}

	account.HashedPassword = ""

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

	return WriteJSON(w, http.StatusOK, struct {
		DeletedAccount int32 `json:"deleted_account"`
	}{deleteAdminAccountRequest.ID})
}

func (self *APIServer) handleGetUserAccounts(w http.ResponseWriter, r *http.Request) error {
	accounts, err := self.storage.GetUserAccounts()
	if err != nil {
		return err
	}

	for _, account := range accounts {
		account.HashedPassword = ""
	}

	return WriteJSON(w, http.StatusOK, accounts)
}

func (self *APIServer) handlePostUserLogin(w http.ResponseWriter, r *http.Request) error {
	loginRequest := new(LoginRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&loginRequest); err != nil {
		return err
	}

	auth_token, err := self.storage.LoginUserAccount(loginRequest.Username, loginRequest.Password)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, struct {
		AuthToken string `json:"auth_token"`
	}{auth_token})
}

func (self *APIServer) handleCreateUserAccount(w http.ResponseWriter, r *http.Request) error {
	createUserAccountRequest := new(CreateAccountRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&createUserAccountRequest); err != nil {
		return err
	}

	account, err := NewUserAccount(createUserAccountRequest.Username, createUserAccountRequest.Password)
	if err != nil {
		return err
	}

	if err := self.storage.CreateUserAccount(account); err != nil {
		return err
	}

	account.HashedPassword = ""

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

	return WriteJSON(w, http.StatusOK, struct {
		DeletedAccount int32 `json:"deleted_account"`
	}{deleteUserAccountRequest.ID})
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

	account.HashedPassword = ""

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

	return WriteJSON(w, http.StatusOK, struct {
		UpdatedAccount int32 `json:"updated_account"`
	}{int32(id)})
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

	return WriteJSON(w, http.StatusOK, struct {
		AddedItem int32 `json:"added_item"`
		Account   int32 `json:"account"`
	}{addItemRequest.ItemID, id})
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

	return WriteJSON(w, http.StatusOK, struct {
		AddedItem int32 `json:"removed_item"`
		Account   int32 `json:"account"`
	}{removeItemRequest.ItemID, id})
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

	_, total, err := self.storage.GetItemsById(account.Items)
	if err != nil {
		return err
	}

	order := NewOrder(uint32(id), account.Items, total)
	if err := self.storage.CreateOrder(order); err != nil {
		return err
	}

	if err := self.storage.ClearUserItems(id); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, struct {
		Order  *Order `json:"order"`
		Status string `json:"status"`
	}{order, "pending"})
}

func (self *APIServer) handleGetUserOrders(w http.ResponseWriter, r *http.Request) error {
	id, err := getID(r)
	if err != nil {
		return err
	}

	account, err := self.storage.GetUserAccount(int32(id))
	if err != nil {
		return err
	}

	orders, err := self.storage.GetOrdersById(account.Orders)
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, orders)
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

	return WriteJSON(w, http.StatusOK, struct {
		DeletedItem int32 `json:"deleted_item"`
	}{deleteItemRequest.ID})
}

func (self *APIServer) handleGetItem(w http.ResponseWriter, r *http.Request) error {
	id, err := getItemID(r)
	if err != nil {
		return err
	}

	item, err := self.storage.GetItem(int32(id))
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, item)
}

func (self *APIServer) handleUpdateItem(w http.ResponseWriter, r *http.Request) error {
	id, err := getItemID(r)
	if err != nil {
		return err
	}

	updateItemRequest := new(UpdateItemRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&updateItemRequest); err != nil {
		return err
	}

	item := Item{
		ID:          uint32(id),
		Name:        updateItemRequest.Name,
		Description: updateItemRequest.Description,
		Price:       updateItemRequest.Price,
	}

	if err := self.storage.UpdateItem(&item); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, struct {
		UpdatedItem int32 `json:"updated_item"`
	}{int32(id)})
}

func (self *APIServer) handleGetOrders(w http.ResponseWriter, r *http.Request) error {
	orders, err := self.storage.GetOrders()
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, orders)
}

func (self *APIServer) handleCreateOrder(w http.ResponseWriter, r *http.Request) error {
	createOrderRequest := new(CreateOrderRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&createOrderRequest); err != nil {
		return err
	}

	order := NewOrder(uint32(createOrderRequest.AccountID), createOrderRequest.Items, createOrderRequest.Total)
	if err := self.storage.CreateOrder(order); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, order)
}

func (self *APIServer) handleDeleteOrder(w http.ResponseWriter, r *http.Request) error {
	deleteOrderRequest := new(DeleteOrderRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&deleteOrderRequest); err != nil {
		return err
	}

	if err := self.storage.DeleteOrder(deleteOrderRequest.ID); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, struct {
		DeletedOrder int32 `json:"deleted_order"`
	}{deleteOrderRequest.ID})
}

func (self *APIServer) handleGetOrder(w http.ResponseWriter, r *http.Request) error {
	id, err := getOrderID(r)
	if err != nil {
		return err
	}

	order, err := self.storage.GetOrder(int32(id))
	if err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, order)
}

func getID(r *http.Request) (int32, error) {
	idStr := mux.Vars(r)["id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("Invalid id: \"%s\"", idStr)
	}

	return int32(id), nil
}

func (self *APIServer) handleUpdateOrder(w http.ResponseWriter, r *http.Request) error {
	id, err := getOrderID(r)
	if err != nil {
		return err
	}

	updateOrderRequest := new(UpdateOrderRequest)
	jsonDecoderHandle := json.NewDecoder(r.Body)
	jsonDecoderHandle.DisallowUnknownFields()
	if err := jsonDecoderHandle.Decode(&updateOrderRequest); err != nil {
		return err
	}

	order := Order{
		ID:     uint32(id),
		Status: updateOrderRequest.Status,
	}

	if err := self.storage.UpdateOrder(&order); err != nil {
		return err
	}

	return WriteJSON(w, http.StatusOK, struct {
		UpdatedOrder int32 `json:"updated_order"`
	}{int32(id)})
}

func getItemID(r *http.Request) (int32, error) {
	idStr := mux.Vars(r)["item_id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("Invalid id: \"%s\"", idStr)
	}

	return int32(id), nil
}

func getOrderID(r *http.Request) (int32, error) {
	idStr := mux.Vars(r)["order_id"]

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, fmt.Errorf("Invalid id: \"%s\"", idStr)
	}

	return int32(id), nil
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
		authHeader := r.Header.Get("Authorization")
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) != 2 {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}
		tokenString := splitToken[1]

		token, err := validateJWT(tokenString)
		if err != nil || !token.Valid {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		id, err := getID(r)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			return
		}

		if claimId := int32(claims["id"].(float64)); claimId != id {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		account, err := storage.GetAdminAccount(id)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		if account.Username != claims["username"] {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		handler(w, r)
	}
}

func withJWTUserAuth(handler http.HandlerFunc, storage Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) != 2 {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}
		tokenString := splitToken[1]

		token, err := validateJWT(tokenString)
		if err != nil || !token.Valid {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		claims := token.Claims.(jwt.MapClaims)

		// verify that the user id in the url path matches the user id in the Token
		id, err := getID(r)
		if err != nil {
			WriteJSON(w, http.StatusBadRequest, ApiError{Error: err.Error()})
			return
		}

		if claimId := int32(claims["id"].(float64)); claimId != id {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		account, err := storage.GetUserAccount(id)
		if err != nil {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		if account.Username != claims["username"] {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		if int64(claims["exp"].(float64)) < time.Now().Unix() {
			WriteJSON(w, http.StatusUnauthorized, ApiError{Error: "Unauthorized"})
			return
		}

		handler(w, r)
	}
}
