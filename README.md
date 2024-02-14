# go_ecom

go_ecom is a Go-based eCommerce API designed for managing online store
operations. It supports account management, item cataloging, and order
processing with secure authentication. The API provides structured endpoints for
both user and admin functions, enabling straightforward integration and backend
management for eCommerce platforms.

## Installation and Setup

To get go_ecom up and running, follow these steps:

1. **Clone the Repository**

   - Clone go_ecom to your local machine using Git:
     ```
     git clone git@github.com:dbolivar25/go_ecom.git
     ```
   - Navigate into the cloned repository:
     ```
     cd go_ecom
     ```

2. **Set Up the Environment**

   - Create a `.env` file in the root of your project directory. This file will
     store environment variables such as database connection details and the JWT
     secret key. Here's an example of what your `.env` file should contain:

     ```
     POSTGRES_USER=<user>
     POSTGRES_NAME=<name>
     POSTGRES_PASS=<pass>

     PORT=<port>

     JWT_SECRET=<secret>
     ```

3. **Install Dependencies**

   - go_ecom requires certain Go packages to function properly. Install these by
     running:
     ```
     go mod tidy
     ```
     This command downloads the necessary packages as defined in `go.mod`.

4. **Run the Application**
   - With the dependencies installed and the environment variables set, you can
     now run the API:
     ```
     go run .
     ```
     This starts the go_ecom service, making it listen for requests as per the
     configuration in your code.

By following these steps, you'll have a local instance of go_ecom running, ready
to handle eCommerce operations through its API endpoints.

## Endpoints

- `/admin/{id}`
- `/admin/{id}/dash`
- `/admin/{id}/admins`
- `/admin/{id}/users`
- `/admin/{id}/items`
- `/admin/{id}/items/{item_id}`
- `/admin/{id}/orders`
- `/admin/{id}/orders/{order_id}`

- `/user`
- `/user/{id}`
- `/user/{id}/items`
- `/user/{id}/checkout`
- `/user/{id}/orders`

- `/items`
- `/items/{id}`

## Documentation

### `/admin/{id}`

- **GET**: Retrieves the details of an admin account.
- **PUT**: Updates an existing admin account.

### `/admin/{id}/dash`

- **GET**: Fetches dashboard data for an admin.

### `/admin/{id}/admins`

- **GET**: Retrieves a list of all admin accounts.
- **POST**: Creates a new admin account.
- **DELETE**: Deletes an admin account.

### `/admin/{id}/users`

- **GET**: Fetches a list of all user accounts.
- **POST**: Creates a new user account.
- **DELETE**: Deletes a user account.

### `/admin/{id}/items`

- **GET**: Retrieves a list of all items in the catalog.
- **POST**: Adds a new item to the catalog.
- **DELETE**: Removes an item from the catalog.

### `/admin/{id}/items/{item_id}`

- **GET**: Fetches details of a specific item.
- **PUT**: Updates the details of a specific item.

### `/admin/{id}/orders`

- **GET**: Retrieves a list of all orders.
- **POST**: Creates a new order.
- **DELETE**: Deletes an order.

### `/admin/{id}/orders/{order_id}`

- **GET**: Fetches details of a specific order.
- **PUT**: Updates the details of a specific order.

### `/user`

- **POST**: Registers a new user account.

### `/user/{id}`

- **GET**: Retrieves the details of a user account.
- **PUT**: Updates an existing user account.

### `/user/{id}/items`

- **GET**: Lists the items associated with a user's account.
- **POST**: Adds an item to a user's account.
- **DELETE**: Removes an item from a user's account.

### `/user/{id}/checkout`

- **POST**: Processes a checkout operation for the items in a user's account.

### `/user/{id}/orders`

- **GET**: Lists all orders associated with a user's account.

### `/items`

- **GET**: Retrieves a list of all items available in the catalog.

### `/items/{id}`

- **GET**: Fetches details of a specific item by its ID.
