# go_ecom

Go_Ecom is a comprehensive eCommerce API developed in Go, designed to manage
online stores efficiently. It provides a secure and scalable way to handle user
and admin operations, including account management, item cataloging, and order
processing.

## Installation and Setup

1. **Clone the Repository:** Clone Go_Ecom to your machine and navigate into the
   project directory.
2. **Environment Configuration:** Create a .env file at the project root to
   store environment variables
   `POSTGRES_USER, POSTGRES_NAME, POSTGRES_PASS, PORT, ROOT_USER, ROOT_PASS, and JWT_SECRET`.
3. **Dependencies:** Use go mod tidy to install the required Go packages.
4. **Running the Service:** Execute go run . to start the Go_Ecom service.## API
   Endpoints

## API Overview

### Admin Authentication

- `/admin/login`: Login to admin account and obtain JWT.

### Admin Management

- `/admin/{id}`: View and update admin account details.
- `/admin/{id}/dash`: View dashboard data.
- `/admin/{id}/admins`: Manage admin accounts.
- `/admin/{id}/users`: Manage user accounts.
- `/admin/{id}/items`: View and manage item catalog.
- `/admin/{id}/orders`: View and manage orders.
- `/admin/{id}/items/{item_id}`: View and update specific item details.
- `/admin/{id}/orders/{order_id}`: View and update specific order details.

### User Authentication

- `/user/login`: Login to user account and obtain JWT.
- `/user/signup`: Create user account.

### User Management

- `/user/{id}`: View and update user account details.
- `/user/{id}/items`: View and manage items in the user's account.
- `/user/{id}/checkout`: Process checkout.
- `/user/{id}/orders`: View user's orders.

### General Item Management

- `/items`: View the item catalog.
- `/items/{id}`: View details of a specific item.

## Documentation

### Authentication Endpoints

#### Admin Login

- **POST** `/admin/login`
  - **Payload**:
    ```json
    {
      "user": "adminUsername",
      "password": "adminPassword"
    }
    ```
  - **Response**: Returns a JWT token for authentication.

#### User Login

- **POST** `/user/login`
  - **Payload**:
    ```json
    {
      "user": "userUsername",
      "password": "userPassword"
    }
    ```
  - **Response**: Returns a JWT token for authentication.

#### User Signup

- **POST** `/user/signup`
  - **Payload**:
    ```json
    {
      "user": "newUserUsername",
      "password": "newUserPassword"
    }
    ```
  - **Response**: Returns a newly created user account object.

### Admin Operations

#### Admin Account Access and Modification

- **GET, PUT** `/admin/{id}`
  - **PUT Payload**:
    ```json
    {
      "user": "updatedAdminUsername"
    }
    ```
  - **Response**: For `PUT`, returns the updated admin account details.

#### Dashboard Access

- **GET** `/admin/{id}/dash`
  - **Response**: Returns dashboard data including admin and user metrics.

#### Admin Account Management

- **GET, POST, DELETE** `/admin/{id}/admins`
  - **POST Payload**:
    ```json
    {
      "user": "newAdminUsername",
      "password": "newAdminPassword"
    }
    ```
  - **DELETE Payload**:
    ```json
    {
      "id": 123
    }
    ```
  - **Response**: For `POST`, returns the newly created admin account. For
    `DELETE`, confirms deletion.

#### User Account Management

- **GET, POST, DELETE** `/admin/{id}/users`
  - **POST Payload**:
    ```json
    {
      "user": "newUserUsername",
      "password": "newUserPassword"
    }
    ```
  - **DELETE Payload**:
    ```json
    {
      "id": 456
    }
    ```
  - **Response**: For `POST`, returns the newly created user account. For
    `DELETE`, confirms deletion.

#### Item Catalog Management

- **GET, POST, DELETE** `/admin/{id}/items`
  - **POST Payload**:
    ```json
    {
      "name": "NewItemName",
      "desc": "NewItemDescription",
      "price": 99.99
    }
    ```
  - **DELETE Payload**:
    ```json
    {
      "id": 789
    }
    ```
  - **Response**: For `POST`, returns the newly added item. For `DELETE`,
    confirms deletion.

#### Order Management

- **GET, POST, DELETE** `/admin/{id}/orders`
  - **POST Payload**:
    ```json
    {
      "account_id": 456,
      "items": [789, 1011],
      "total": 199.98
    }
    ```
  - **DELETE Payload**:
    ```json
    {
      "id": 11213
    }
    ```
  - **Response**: For `POST`, returns the newly created order. For `DELETE`,
    confirms deletion.

### User Operations

#### User Account and Item Management

- **GET, PUT** `/user/{id}`
  - **PUT Payload**:
    ```json
    {
      "user": "updatedUserUsername"
    }
    ```
  - **Response**: Returns the updated user account details.

#### Item Management in User Account

- **POST, DELETE** `/user/{id}/items`
  - **POST Payload**:
    ```json
    {
      "item_id": 789
    }
    ```
  - **DELETE Payload**:
    ```json
    {
      "item_id": 789
    }
    ```
  - **Response**: For `POST`, confirms item addition. For `DELETE`, confirms
    item removal.

#### Checkout

- **POST** `/user/{id}/checkout`
  - **Response**: Processes the checkout and returns the created order object.

#### User Orders

- **GET** `/user/{id}/orders`
  - **Response**: Returns a list of orders associated with the user's account.

### General Item Access

#### Item Catalog

- **GET** `/items`
  - **Response**: Returns a list of all items in the catalog.

#### Specific Item Details

- **GET** `/items/{id}`
  - **Response**: Returns details of a specific item.
