# Little Einstein Backend

## Installing Go and Pre-Commit on Mac and Windows

## Installing Homebrew (Mac Only)

If you haven't installed Homebrew yet, you can find more details [here](https://brew.sh/).

1. Open Terminal.
2. Run the following command:

   ```sh
   /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
   ```

3. Add Homebrew to your PATH:

   ```sh
   echo 'eval "$(/opt/homebrew/bin/brew shellenv)"' >> ~/.zprofile
   eval "$(/opt/homebrew/bin/brew shellenv)"
   ```

4. Verify the installation:

   ```sh
   brew --version
   ```

---

## Installing Go

### macOS

#### Using Homebrew (Recommended)

1. Open the terminal and run:

   ```sh
   brew install go
   ```

2. Verify the installation:

   ```sh
   go version
   ```

3. Set up Go environment variables (optional but recommended):

   ```sh
   echo 'export GOPATH="$HOME/go"' >> ~/.zshrc
   echo 'export PATH="$GOPATH/bin:$PATH"' >> ~/.zshrc
   source ~/.zshrc
   ```

#### Manual Installation

1. Download the latest Go package from the official site: [Go Downloads](https://go.dev/dl/)
2. Open the downloaded `.pkg` file and follow the installation steps.
3. Verify the installation:

   ```sh
   go version
   ```

### Windows

#### Using Chocolatey (Recommended)

1. Open **PowerShell as Administrator** and run:

   ```powershell
   choco install golang -y
   ```

2. Restart your terminal and verify installation:

   ```powershell
   go version
   ```

#### Manual Installation

1. Download the Windows installer from [Go Downloads](https://go.dev/dl/).
2. Run the installer and follow the installation steps.
3. Ensure Go is added to your system `PATH` (usually done automatically).
4. Verify installation:

   ```powershell
   go version
   ```

## Installing Pre-Commit

### macOS

#### Using Homebrew (Recommended)

1. Install **pre-commit**:

   ```sh
   brew install pre-commit
   ```

2. Verify installation:

   ```sh
   pre-commit --version
   ```

#### Using Pip

1. Install **pre-commit** using Python's package manager:

   ```sh
   pip install pre-commit
   ```

2. Verify installation:

   ```sh
   pre-commit --version
   ```

### Windows

#### Using Chocolatey

1. Open **PowerShell as Administrator** and run:

   ```powershell
   choco install pre-commit -y
   ```

2. Verify installation:

   ```powershell
   pre-commit --version
   ```

#### Using Pip

1. Install **pre-commit** using Python's package manager:

   ```powershell
   pip install pre-commit
   ```

2. Verify installation:

   ```powershell
   pre-commit --version
   ```

## Setting Up Pre-Commit in a Repository
#### Make sure  you have Cloned the repo before the next steps 
SSH: 
```sh
git clone git@github.com:littleeinsteinchildcare/littleeinstein-backend.git
```

HTTPS
```sh
git clone https://github.com/littleeinsteinchildcare/littleeinstein-backend.git
```

1. Navigate to your repository:

   ```sh
   cd $HOME/littleeinstein-backend
   ```

2. Installing node

   ```sh
   brew install node
   ```

3. Verify node and npm

   ```sh
   npm -v
   ```

   ```sh
   node -v
   ```
4. Install some go packages

   ```sh
   go install golang.org/x/tools/cmd/goimports@latest
   ```

   ```sh
   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
   ```
4. Installing commitlint/cli commitlint/config-conventional

   ```sh
   npm install -g @commitlint/cli @commitlint/config-conventional
   ```

5. Install pre-commit hooks:

   ```sh
   pre-commit install
   ```

6. Run pre-commit manually (optional, to test it):

   ```sh
   pre-commit run --all-files
   ```

Now your repository is set up with **pre-commit**, and hooks will run automatically on `git commit -m "Message"`!


# Go API Project Structure

This repository follows a clean, layered architecture for a Go REST API application. This structure is designed to promote separation of concerns, testability, and maintainability.

- To run the project do:  ```  go run cmd/api/main.go ```


## Project Structure
```
├── cmd/                    (Command line applications)
│   └── api/                (API server executable)
│       └── main.go         (Application entry point - initializes and starts the server)
├── internal/               (Private application code - not importable by external packages)
│   ├── api/                (API-specific code)
│   │   ├── routes/         (HTTP route definitions)
│   │   │   ├── user_routes.go     (User endpoints: GET /users, POST /users, etc.)
│   │   │   ├── event_routes.go    (Event endpoints: GET /events, POST /events, etc.)
│   │   │   └── router.go          (Central router configuration - combines all routes)
│   │   └── middleware/     (HTTP middleware functions) 
│   │       ├── auth.go            (Authentication/Authorization checks)
│   │       ├── cors.go            (Cross-Origin settings for frontend access)
│   │       └── logging.go         (Request logging and tracing)[NOT IMPLEMENTED YET]
│   ├── config/             (Application configuration processing)
│   │   ├── aztables_config.go     (Azure Table Storage configuration)
│   │   ├── server_config.go       (Server configuration settings)
│   │   └── firebase_config.go     [NEW] (Firebase initialization and configuration)
│   ├── handlers/           (HTTP request handlers - processes HTTP requests)
│   │   ├── event_handler.go       (Functions handling event-related requests)
│   │   ├── generic_handler.go     (Common handler utilities)
│   │   └── user_handler.go        (User email change, deletion, group check endpoints) [UPDATED]
│   ├── models/             (Data structures representing domain objects)
│   │   ├── event.go               (Event entity definition with fields)
│   │   └── user.go                (User entity definition with fields and validation)
│   ├── repositories/       (Database access layer)
│   │   ├── event_repo.go          (Event CRUD operations in the database)
│   │   ├── user_repo.go           (User CRUD operations in the database)
│   │   └── firebase_repo.go       (Firebase Auth operations: email change, user deletion, group checks)
│   └── services/           (Business logic layer)
│       ├── event_service.go       (Event-related business rules and operations)
│       └── user_service.go        (User business logic - coordinates Firebase and DB operations) [UPDATED]
├── pkg/                    (Reusable packages that could be used by external applications) [NOT IMPLEMENTED YET]
├── configs/                (Configuration files) [NOT IMPLEMENTED YET]
│   ├── app.env                    (Environment-specific variables)
│   └── app.yaml                   (Application settings in YAML format)
├── docs/                   (API documentation) [NOT IMPLEMENTED YET]
│   └── swagger.yaml               (OpenAPI/Swagger API specification)
└── .github/                (GitHub specific files)
    └── PULL_REQUEST_TEMPLATE.md   (Template for pull requests)
```


# Understanding the Layers

### Routes vs Handlers

- **Routes** (`internal/api/routes/`): Define URL patterns and HTTP methods your API responds to. They map each endpoint to a specific handler function. Think of routes as the "address" of your API endpoints.

- **Handlers** (`internal/handlers/`): Contain the implementation for processing HTTP requests. They extract data from requests, validate it, call the appropriate services, and format the HTTP response. Handlers bridge the HTTP world with your application's business logic.

Example comparison:
```go
// In routes/user_routes.go
router.GET("/users/:id", userHandler.GetUserByID)  // Defines the URL pattern

// In handlers/user_handler.go
func (h *UserHandler) GetUserByID(c *gin.Context) {
    id := c.Param("id")
    user, err := h.userService.GetUserByID(id) // Invoking the service layer
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
        return
    }
    c.JSON(http.StatusOK, user)
}
```

### Services

- **Services** (`internal/services/`): Form the heart of your application containing core business logic. They:
    - Implement domain-specific rules and workflows
    - Are independent of HTTP or other delivery mechanisms
    - Coordinate between different repositories when needed
    - Handle complex operations that span multiple data sources
    - Enforce business constraints and validation rules

Example:
```go
// In services/user_service.go
func (s *UserService) GetUserByID(id string) (*User, error) {
    user, err := s.userRepo.FindByID(id)
    if err != nil {
        return nil, err
    }
    return user, nil
}
```

### Repositories

- **Repositories** (`interal/repositories/`): Data Access Objects (DAO)
    - Connects to Azure Tables DB on instantiation
    - Handles CRUD operations on request from Services for a specific table


# Setting Up Local Development Database

- ### Azurite
    - **Azurite** is an open-source Azure Storage emulator, useful for testing locally (and for free) before transitioning to cost-accruing Azure Storage options like Azure Tables
    - **Installation**
        - **[Azurite Installation Instructions](https://learn.microsoft.com/en-us/azure/storage/common/storage-use-azurite?tabs=visual-studio%2Cblob-storage)** 
        - Easiest installation is via npm: `npm install -g azurite`
    - **Running Azurite**
        - To run Azurite in your terminal: ```azurite```  
        - Azurite will store config and JSON files in whichever directory you run the command in. You should be running the command inside `tmp/`
            - These files will maintain persistence locally on your computer, so subsequent uses of Azurite will reference all the previously created entities

- ### Azure Storage Explorer
    - **Azure Storage Explorer** Will let you explore all the different storage options locally, and also listen automatically for an emulator like **Azurite**
    - **Installation**
        - **[Azure Storage Explorer Installation Instructions](https://azure.microsoft.com/en-us/products/storage/storage-explorer)**
        - Download the Installer for your OS
        - Run the Installer
    - **Running Azure Storage Explorer**
        - Open Azure Storage Explorer
        - Make sure that Azurite is running
        - In the Explorer, select `Storage Accounts` --> `Emulator - Default Ports`
            - **If you're running this for the first time, make sure to add a .env file to the project folder**
                - Select `Emulator - Default Ports`
                    - In the Properties window (default is bottom left corner) you will need:
                        - `Account Name`
                        - `Primary Key`
                - **.env**
                    - ```
                        AZURE_STORAGE_ACCOUNT_NAME=<Account Name>
                        AZURE_STORAGE_ACCOUNT_KEY=<Primary Key>
                        AZURE_STORAGE_SERVICE_URL="http://127.0.0.1:10002/devstoreaccount1"
                    ```
        - Any updates to the table will be reflected in the `UserTable` table

### Running the Project with Air

To use Air for live reloading during development:

#### Install Air 
```
go install github.com/cosmtrek/air@latest
```

#### Running the project with Air:
```cd cmd/api && air```

This will watch for file changes and automatically restart the application for a smoother development experience.