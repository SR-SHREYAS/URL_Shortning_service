# Mr. Compress - URL Shortening Service

## Quirk: URL Compression

This project implements a high-performance URL shortening service, akin to popular services like Bitly or TinyURL. It's built using Go with the Fiber web framework and leverages Redis for efficient data storage and retrieval, including URL mappings and rate limiting. The service provides a simple API to shorten long URLs and a redirection mechanism for accessing the original links.

## Table of Contents

- [Project Overview](#project-overview)
- [Key Technologies and Concepts](#key-technologies-and-concepts)
  - [Go (Golang)](#go-golang)
  - [Fiber Web Framework](#fiber-web-framework)
  - [Redis](#redis)
  - [go-redis/redis/v8](#go-redisredisv8)
  - [godotenv](#godotenv)
  - [govalidator](#govalidator)
  - [google/uuid](#googleuuid)
  - [Middleware (Logger, CORS)](#middleware-logger-cors)
  - [Rate Limiting](#rate-limiting)
  - [Frontend (HTML, Tailwind CSS, JavaScript)](#frontend-html-tailwind-css-javascript)
- [Flow of Execution](#flow-of-execution)
  - [1. Application Startup](#1-application-startup)
  - [2. URL Shortening Request](#2-url-shortening-request)
  - [3. URL Resolution Request](#3-url-resolution-request)
- [Future Goals](#future-goals)

---

## Project Overview

The service allows users to:

1.  **Shorten URLs**: Provide a long URL and receive a short, unique identifier.
2.  **Custom Aliases**: Optionally specify a custom short code for their URL.
3.  **Expiry**: Set an expiry duration for the shortened URL.
4.  **Redirection**: Access the original long URL by navigating to the shortened link.
5.  **Rate Limiting**: Control API usage to prevent abuse.

## Key Technologies and Concepts

### Go (Golang)

-   **Explanation**: Go is an open-source, statically typed, compiled programming language designed for building simple, reliable, and efficient software. It's known for its strong concurrency features and performance.
-   **Usage in Project**: Go serves as the primary language for developing the backend API. It handles all server-side logic, including request parsing, database interactions, business logic (like URL validation and short code generation), and response generation.

### Fiber Web Framework

-   **Explanation**: Fiber is an Express.js-inspired web framework built on top of Fasthttp, the fastest HTTP engine for Go. It's designed for speed and ease of use, making it ideal for building high-performance APIs.
-   **Usage in Project**: Fiber is used to:
    -   Define API endpoints (`/api/v1` for shortening, `/:url` for resolving).
    -   Handle incoming HTTP requests and route them to appropriate handlers.
    -   Parse request bodies (JSON).
    -   Serve static frontend files (`index.html`, CSS, JS).
    -   Apply middleware for logging and CORS.
    -   Construct and send JSON responses and HTTP redirects.

    ```go
    // api/main.go
    func setupRoutes(app *fiber.App) {
        app.Static("/", "./public", fiber.Static{
            Index: "index.html",
        })
        app.Post("/api/v1", routes.ShortenURL)
        app.Get("/:url", routes.ResolveURL)
    }
    ```

### Redis

-   **Explanation**: Redis (Remote Dictionary Server) is an open-source, in-memory data structure store, used as a database, cache, and message broker. It supports various data structures like strings, hashes, lists, sets, and sorted sets. Its in-memory nature makes it extremely fast.
-   **Usage in Project**: Redis is central to the service's functionality:
    -   **URL Mapping Storage**: Stores the mapping between short codes and original long URLs. Each short URL is a key, and the long URL is its value, with an optional expiry. (Using DB 0)
    -   **Rate Limiting**: Tracks the number of API requests made by each client IP address within a time window. (Using DB 1)
    -   **Visit Counter**: Increments a counter each time a shortened URL is resolved. (Using DB 1)

### go-redis/redis/v8

-   **Explanation**: This is a popular and robust Redis client library for Go, providing an idiomatic interface to interact with Redis.
-   **Usage in Project**: It's used to establish connections to Redis, and perform operations like `SET` (store URL), `GET` (retrieve URL), `DECR` (decrement rate limit), `TTL` (get time-to-live), and `INCR` (increment visit counter).

    ```go
    // api/database/database.go
    package database

    import (
    	"context"
    	"os"

    	"github.com/go-redis/redis/v8"
    )

    var Ctx = context.Background()

    func CreateClient(dbNo int) *redis.Client {
    	rdb := redis.NewClient(&redis.Options{
    		Addr:     os.Getenv("REDIS_ADDRESS"),
    		Password: os.Getenv("REDIS_PASSWORD"),
    		DB:       dbNo,
    	})
    	return rdb
    }
    ```

### godotenv

-   **Explanation**: A Go package that loads environment variables from a `.env` file into `os.Getenv()`. This is crucial for managing configuration settings without hardcoding them directly into the application.
-   **Usage in Project**: Used in `main.go` to load sensitive information and configuration parameters like `REDIS_ADDRESS`, `REDIS_PASSWORD`, `API_QUOTA`, and `DOMAIN` from the `.env` file.

### govalidator

-   **Explanation**: A Go package for validating strings, including common formats like URLs, emails, and UUIDs.
-   **Usage in Project**: In the `ShortenURL` handler, `govalidator.IsURL()` is used to ensure that the URL provided by the user is in a valid format before processing.

    ```go
    // api/routes/shorten.go (inside ShortenURL function)
    // ...
    if !govalidator.IsURL(body.URL) {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid url"})
    }
    // ...
    ```

### google/uuid

-   **Explanation**: A Go package for generating universally unique identifiers (UUIDs) based on RFC 4122.
-   **Usage in Project**: When a user does not provide a `custom_short` alias, a unique 6-character ID is generated using `uuid.New().String()[:6]` to serve as the short code for the URL.

### Middleware (Logger, CORS)

-   **Explanation**: Middleware functions are executed before or after the main request handler. They can perform tasks like logging, authentication, data compression, or handling cross-origin requests.
-   **Usage in Project**:
    -   **Logger**: `app.Use(logger.New())` logs details about incoming HTTP requests (e.g., method, path, status, latency) to the console, which is useful for debugging and monitoring.
    -   **CORS**: `app.Use(cors.New())` enables Cross-Origin Resource Sharing, allowing the frontend application (served from a potentially different origin) to make requests to the backend API.

### Rate Limiting

-   **Explanation**: Rate limiting is a strategy to control the number of requests a user or client can make to an API within a given time window. It prevents abuse, protects resources, and ensures fair usage.
-   **Usage in Project**: Implemented using Redis DB 1. Each client's IP address is used as a key, storing their remaining request quota. A `TTL` (Time To Live) is set on this key to reset the quota periodically. If a client exceeds their quota, further requests are blocked until the reset period.

    ```go
    // api/routes/shorten.go (inside ShortenURL function)
    // ...
    r2 := database.CreateClient(1) // connect to redis client db 1 where rate limits are stored
    defer r2.Close()

    value, err := r2.Get(database.Ctx, c.IP()).Result()
    if err == redis.Nil {
        _ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
    } else {
        valInt, _ := strconv.Atoi(value)
        if valInt <= 0 {
            limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
            return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "rate limit exceeded", "rate_limit_reset": limit})
        }
    }
    // ...
    r2.Decr(database.Ctx, c.IP()) // Decrement after successful processing
    // ...
    ```

### Frontend (HTML, Tailwind CSS, JavaScript)

-   **Explanation**: Standard web technologies used to create the user interface. HTML structures the content, Tailwind CSS provides utility-first styling, and JavaScript handles client-side interactivity and API communication.
-   **Usage in Project**: The `public/index.html` file, along with its linked CSS and JavaScript, provides a user-friendly interface for:
    -   Inputting a long URL.
    -   Optionally specifying a custom short alias and expiry.
    -   Submitting the request to the backend API.
    -   Displaying the shortened URL and allowing it to be copied.
    -   Showing error messages from the API.

## Flow of Execution

### 1. Application Startup

-   The `main.go` file loads environment variables from `.env`.
-   A new Fiber application instance is created.
-   Global middleware (logger for request logging, CORS for cross-origin requests) is applied.
-   Routes are configured:
    -   Static files from the `public` directory are served (e.g., `index.html`).
    -   `POST /api/v1` is mapped to the `ShortenURL` handler.
    -   `GET /:url` is mapped to the `ResolveURL` handler for redirection.
-   The Fiber app starts listening for incoming HTTP requests on the configured port.

### 2. URL Shortening Request

-   A user interacts with the frontend (`index.html`) and submits a long URL (and optional custom alias/expiry).
-   The frontend JavaScript sends a `POST` request to `/api/v1` with the URL data.
-   The `ShortenURL` handler in `api/routes/shorten.go` is invoked:
    1.  **Rate Limiting**: It connects to Redis DB 1. It checks the client's IP address against a rate limit. If the limit is exceeded, an error is returned. Otherwise, the quota is initialized or decremented later.
    2.  **Input Parsing & Validation**: The request body is parsed. The provided URL is validated using `govalidator.IsURL()`, and checks are performed to prevent shortening the service's own domain. `http://` or `https://` is enforced.
    3.  **Short Code Generation**: If a `custom_short` is provided, it's used. Otherwise, a 6-character UUID is generated.
    4.  **Redis Storage**: It connects to Redis DB 0. It checks if the generated/custom short code already exists. If so, a "forbidden" error is returned. If not, the short code (key) and original URL (value) are stored in Redis with the specified expiry (defaulting to 24 hours if not provided).
    5.  **Rate Limit Update**: The client's request count in Redis DB 1 is decremented.
    6.  **Response**: A JSON response containing the original URL, the generated short URL, expiry, and updated rate limit information is returned to the client.

    ```go
    // api/routes/shorten.go (core Redis interaction)
    // ...
    // Check if custom short is already in use
    value, _ = r.Get(database.Ctx, id).Result()
    if value != "" {
        return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "url custom short is already in use"})
    }

    // Set the URL in the database with expiry
    err = r.Set(database.Ctx, id, body.URL, body.Expiry*3600*time.Second).Err()
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not shorten url"})
    }
    // ...
    ```

### 3. URL Resolution Request

-   A user navigates to a shortened URL (e.g., `yourdomain.com/my-short-link`).
-   The `ResolveURL` handler in `api/routes/resolve.go` is invoked:
    1.  **Extract Short Code**: The short code (`my-short-link` in the example) is extracted from the URL parameters.
    2.  **Redis Lookup**: It connects to Redis DB 0. The short code is used as a key to retrieve the original long URL from Redis.
    3.  **Error Handling**: If the key is not found (`redis.Nil`), a 404 "Not Found" error is returned. If any other Redis error occurs, a 500 "Internal Server Error" is returned.
    4.  **Visit Counter**: It connects to Redis DB 1 and increments a global "counter" key to track total visits.
    5.  **Redirection**: The user's browser is redirected to the retrieved original long URL using an HTTP 301 (Moved Permanently) status code.

    ```go
    // api/routes/resolve.go (core Redis interaction and redirection)
    // ...
    url := c.Params("url")
    r := database.CreateClient(0)
    defer r.Close()

    value, err := r.Get(database.Ctx, url).Result()
    if err == redis.Nil {
        return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "short not found in the database"})
    } else if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "could not resolve short url"})
    }

    // Increment visit counter
    rInr := database.CreateClient(1)
    defer rInr.Close()
    _ = rInr.Incr(database.Ctx, "counter")

    // Redirect to original URL
    return c.Redirect(value, fiber.StatusMovedPermanently)
    ```

## Future Goals

1.  **Permanent URL Redirecting Service**: Implement an option for users to create shortened URLs that do not expire, or have a significantly longer expiry period, suitable for permanent links. This would involve a different storage strategy or a specific flag in the Redis entry.
```

## Flow of Execution

### Running the Project Locally with Docker

This project is containerized and can be easily run using Docker and Docker Compose.

**Prerequisites:**
*   Docker
*   Docker Compose

**1. Create Environment File**

Before running the application, you need to create a `.env` file in the `api/` directory. A `.env.example` file should be provided in the `api/` directory. You can copy it to create your own configuration:

```bash
cp api/.env.example api/.env
```

Your `api/.env` file should look like this:

```dotenv
# Port for the Go application
APP_PORT=4000

# The public-facing domain of the service
DOMAIN=http://localhost:4000

# Redis connection details (db is the service name in docker-compose.yml)
REDIS_ADDRESS=db:6379
REDIS_PASSWORD=

# Number of requests allowed per IP in a 30-minute window
API_QUOTA=10
```

**2. Build and Run the Containers**

From the root directory of the project (where `docker-compose.yml` is located), run the following command:

```bash
docker-compose up --build
```

This command builds the Docker images for the Go backend and Redis, then starts the containers. The API will be accessible at `http://localhost:4000`.

**3. Stopping the Application**

To stop and remove the containers, networks, and volumes created by `docker-compose up`, run:

```bash
docker-compose down
```

**4. Cleaning Up Docker Resources**

To remove all unused containers, networks, and images to free up disk space, you can use the `prune` command.

```bash
docker system prune -a
```
> **Warning:** This command is destructive and will remove all stopped containers, unused networks, and dangling images. Use with caution.

### 1. Application Startup

-   The `main.go` file loads environment variables from `.env`.
