# Secure Website Builder Backend

## Overview

This project provides the backend for the Secure Website Builder application, including a Go API and a PostgreSQL database. The setup supports easy configuration via environment variables and is ready for local development with Docker.

---

## Prerequisites

Before running the project, ensure your machine has:

- **Go 1.22** or higher installed:

  ```bash
  go version
  ```

- **Docker** installed and running:

  ```bash
  docker --version
  ```

- **Docker Compose** installed and running:

  ```bash
  docker compose version
  ```

- **Git** (to clone the repository)

---

## Clone Repository and Create `.env` File

1. Clone the repository:

```bash
git clone https://github.com/Secure-Website-Builder/Backend.git
cd Backend
```

2. Create a `.env` file in the **project root** with your configuration. Replace placeholders with your own credentials:

```env
# Application
APP_ENV=<development|production>
APP_PORT=<host-port-for-backend>

# Database
DB_USER=<your-db-user>
DB_PASSWORD=<your-db-password>
DB_NAME=<your-db-name>
DB_HOST=db
DB_PORT=<host-port-for-db>

# Auth
JWT_SECRET=<your-jwt-secret>
```

> This file stores secrets and host-specific configuration. **Do not commit it to version control.**

---

## Start Backend and Database

1. Launch services with Docker Compose:

```bash
docker compose up
```

2. Access the backend on your host machine:

```
http://localhost:<APP_PORT>
```

> **Port mapping explanation:**
>
> - The container listens internally on the port exposed in the Dockerfile (`EXPOSE 8080`).
> - The host port comes from `.env` (`APP_PORT`). Docker Compose maps host port → container port.
> - Example: `.env` has `APP_PORT=9090`, container exposes `8080`. Access via `http://localhost:9090`.

3. Stop the services:

```bash
docker compose down
```

---

## Notes

- The backend container mounts your local code for **live code updates**, so you don’t need to rebuild the image after code changes.
- Database port is configurable via `.env` (`DB_PORT`) and maps to container port `5432`.
- Ensure the `.env` file is in the project root before running `docker compose up`.
