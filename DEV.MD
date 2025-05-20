# LittleEinstein Backend - Docker Configuration for Local Development

This document explains how to run the LittleEinstein backend application in a local development environment using Docker and Docker Compose. It covers environment configuration, container setup, and testing approaches for developers.

## Containers vs Virtual Machines

Docker containers are lightweight compared to virtual machines because:
* VMs emulate complete hardware environments with their own operating systems
* Containers share the host operating system's kernel
* Containers isolate applications using namespaces and control groups
* This makes containers faster to start and more resource-efficient

## Prerequisites

- Docker and Docker Compose installed
- Git repository cloned

## Environment Configuration

The application requires specific environment variables for local development. Create a `.env` file in the project root with the following variables:

* **APP_ENV=development**
    * Must be set to `development` for local Docker setup

* **AZURE_STORAGE_ACCOUNT_NAME=devstoreaccount1**
    * Uses the standard Azurite emulator account name

* **AZURE_STORAGE_ACCOUNT_KEY=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==**
    * Standard development key for Azurite emulator

* **AZURE_TABLE_SERVICE_URL=http://127.0.0.1:10002/devstoreaccount1**
    * Points to the local Table service emulator

* **AZURE_BLOB_SERVICE_URL=http://127.0.0.1:10000/devstoreaccount1**
    * Points to the local Blob service emulator

* **AZURE_BLOB_CONTAINER_NAME=userblobs**
    * Name of the blob container to create/use

These values are configured to work with the local Azurite emulator (Azure Storage Emulator) which will run in Docker alongside your application.

## Docker Compose Setup

Our Docker Compose configuration sets up two services that work together:

1. **backend**: Your Go API application
    * Runs on port 8080
    * Connects to Azurite for Azure storage
    * Rebuilt automatically when you run `docker-compose up --build`
    * Port mapping: "8080:8080" connects host port to container port

2. **azurite**: Azure Storage Emulator
    * Runs three separate services on different ports:
        * Blob Storage: Port 10000
        * Queue Service: Port 10001
        * Table Service: Port 10002
    * Persists data between restarts using a Docker volume

The services run in separate containers but are connected through a Docker network, allowing them to communicate with each other. Your backend service can access Azurite using the service name in Docker's internal network.

### How Port Mapping Works

```
ports:
  - "8080:8080"  # HOST_PORT:CONTAINER_PORT
```

This creates a connection between your computer and the container:
- When a request comes to port 8080 on your computer (host)
- Docker forwards it to port 8080 inside the container
- Your Go API inside the container receives the request
- Think of it like mail forwarding between two addresses
- External apps/browsers use http://localhost:8080 to access your API
- The container has its own isolated network with its own port 8080

## Running the Application Locally

```bash
# Navigate to project directory
cd /path/to/littleeinstein-backend

# Start both services with rebuild
sudo docker-compose up --build

# To run in background (detached mode)
sudo docker-compose up -d --build

# To view logs while running in detached mode
sudo docker-compose logs -f
```

The API will be available at: `http://localhost:8080`

### What Happens When You Start the Application

1. Docker builds your Go API backend from the Dockerfile
2. Docker starts the Azurite emulator container
3. Docker starts your backend container once Azurite is ready
4. Your backend connects to Azurite using environment variables
5. The API becomes accessible at http://localhost:8080

### Stopping the Application

```bash
# If running in foreground, press Ctrl+C
# If running in background:
sudo docker-compose down

# To also remove volumes (deletes Azurite persisted data)
sudo docker-compose down -v
```

## Running Tests with Docker

You can run tests against the dockerized API without modifying your test scripts:

```bash
# Run test script from project directory
cd /path/to/littleeinstein-backend/test
./banner.sh
```

### Why This Works
- Your application inside Docker exposes port 8080 to your host machine
- Test scripts running on your host machine can access the API via `http://localhost:8080`
- No modification to test scripts is needed
- Tests run exactly as they would if you were running `go run cmd/api/main.go` directly

## Troubleshooting

### Viewing Container Logs

```bash
# View logs from all services
sudo docker-compose logs

# Follow logs in real-time
sudo docker-compose logs -f
```

## Next Steps

This configuration is for local development only. Production deployment configuration will be addressed separately in a future update.
