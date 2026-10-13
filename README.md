# RSS Aggregator

A small REST API written in Go that lets users:

- create an account
- add RSS feeds
- follow feeds
- collect RSS posts in the background
- view posts from followed feeds

This project is useful for learning Go HTTP APIs, PostgreSQL, authentication middleware, SQLC-style database code, goroutines, and background jobs.

## How it works

```text
Client
  |
  v
Chi HTTP router
  |
  v
Handlers -> Authentication middleware -> PostgreSQL
  |
  +--> Background scraper fetches RSS feeds every minute
```

The scraper:

1. Finds feeds that need to be fetched.
2. Fetches several feeds concurrently.
3. Parses the RSS XML.
4. Saves posts in PostgreSQL.

## Requirements

- Go
- PostgreSQL
- An API client such as Thunder Client, Postman, or curl

## Setup

### 1. Create the database

Create a PostgreSQL database, for example:

```sql
CREATE DATABASE rssagg;
```

### 2. Configure environment variables

Create a `.env` file in the project root:

```env
PORT=8080
DB_URL=postgres://postgres:password@localhost:5432/rssagg?sslmode=disable
```

Replace the username, password, host, and database name with your local values.

### 3. Apply database migrations

Run the SQL migration files in this order:

```text
sql/schema/001_users.sql
sql/schema/002_users_apikey.sql
sql/schema/003_feeds.sql
sql/schema/004_feed_follows.sql
sql/schema/005_feeds_lastfetchedat.sql
sql/schema/006_posts.sql
```

Run each file using your preferred PostgreSQL migration tool or SQL client.

### 4. Start the server

```powershell
go run .
```

Or build and run the executable:

```powershell
go build
.\rss_Agregator.exe
```

The API starts at:

```text
http://localhost:8080
```

## Authentication

Protected endpoints require the API key returned when a user is created.

Use this exact header format:

```http
Authorization: Apikey YOUR_API_KEY
```

The prefix is `Apikey` with a capital `A` and lowercase remaining letters.

## API testing order

Use the following order when testing the API.

### 1. Health check

```http
GET http://localhost:8080/v1/healthz
```

Expected status:

```text
200 OK
```

### 2. Create a user

```http
POST http://localhost:8080/v1/users
Content-Type: application/json
```

Body:

```json
{
  "name": "Anuj"
}
```

Save the `api_key` from the response:

```json
{
  "id": "user-uuid",
  "name": "Anuj",
  "created_at": "timestamp",
  "updated_at": "timestamp",
  "api_key": "your-api-key"
}
```

### 3. Get available feeds

```http
GET http://localhost:8080/v1/feeds
```

Copy the `id` of one feed. Example:

```text
5dd2ee7a-9ceb-4b0c-9ce2-b9eaa929a134
```

### 4. Follow a feed

```http
POST http://localhost:8080/v1/feed_follows
Authorization: Apikey YOUR_API_KEY
Content-Type: application/json
```

Body:

```json
{
  "feed_id": "5dd2ee7a-9ceb-4b0c-9ce2-b9eaa929a134"
}
```

Expected status:

```text
201 Created
```

### 5. List followed feeds

```http
GET http://localhost:8080/v1/feed_follows
Authorization: Apikey YOUR_API_KEY
```

Do not send a request body with this `GET` request.

Expected status:

```text
200 OK
```

### 6. Get posts from followed feeds

```http
GET http://localhost:8080/v1/posts
Authorization: Apikey YOUR_API_KEY
```

Do not send a request body.

The scraper runs every minute, so wait for logs showing that posts were collected before testing this endpoint.

Expected response shape:

```json
[
  {
    "id": "post-uuid",
    "title": "Example title",
    "description": "Example description",
    "published_at": "timestamp",
    "url": "https://example.com/article",
    "feed_id": "feed-uuid"
  }
]
```

### 7. Delete a feed follow

Copy the follow record `id` from `GET /v1/feed_follows`, then send:

```http
DELETE http://localhost:8080/v1/feed_follows/FOLLOW_ID
Authorization: Apikey YOUR_API_KEY
```

Expected status:

```text
200 OK
```

### 8. Test the error endpoint

```http
GET http://localhost:8080/v1/err
```

The current implementation intentionally returns:

```text
400 Bad Request
```

with:

```json
{
  "error": "Something went wrong"
}
```

## Available endpoints

| Method | Endpoint | Auth | Purpose |
|---|---|---:|---|
| GET | `/v1/healthz` | No | Check whether the server is running |
| GET | `/v1/err` | No | Test an error response |
| POST | `/v1/users` | No | Create a user |
| GET | `/v1/users` | Yes | Get the authenticated user |
| GET | `/v1/feeds` | No | List feeds |
| POST | `/v1/feeds` | Yes | Create a feed |
| GET | `/v1/feed_follows` | Yes | List the user's followed feeds |
| POST | `/v1/feed_follows` | Yes | Follow a feed |
| DELETE | `/v1/feed_follows/{id}` | Yes | Unfollow a feed |
| GET | `/v1/posts` | Yes | Get posts from followed feeds |

## Create your own feed

```http
POST http://localhost:8080/v1/feeds
Authorization: Apikey YOUR_API_KEY
Content-Type: application/json
```

Body:

```json
{
  "name": "BBC News",
  "url": "https://feeds.bbci.co.uk/news/rss.xml"
}
```

The URL must point to a reachable RSS XML feed.

## Common problems

### `400` or `403` from a protected endpoint

Check the header:

```http
Authorization: Apikey YOUR_API_KEY
```

Make sure the key is valid and that there is exactly one space between `Apikey` and the key.

### Duplicate post URL

```text
pq: duplicate key value violates unique constraint "posts_url_key"
```

This means the RSS item was already saved. The database intentionally requires post URLs to be unique.

### RSS request timeout

```text
context deadline exceeded
```

The RSS server did not respond before the HTTP timeout. Try the feed again later or use another RSS URL.

### No posts are returned

- Confirm that the user follows a feed.
- Confirm that the scraper fetched the feed.
- Wait for the next scraper cycle.
- Check the application logs.

## Project structure

```text
.
├── main.go                    # Application startup and routes
├── handler_user.go            # User and post handlers
├── handler_feed.go            # Feed handlers
├── handler_feed_follows.go   # Follow and unfollow handlers
├── middleware_auth.go         # API-key authentication
├── rss.go                     # RSS HTTP/XML parsing
├── scraper.go                 # Background feed scraper
├── models.go                  # API response models
├── internal/auth/             # Authorization header parsing
├── internal/database/         # Database queries and models
└── sql/
    ├── queries/               # SQL queries
    └── schema/                # Database migrations
```

## Interview revision notes

### Why use a goroutine?

The scraper starts in a goroutine so the HTTP server can start immediately:

```go
go startScraping(apiCfg.DB, 10, time.Minute)
```

### Why use a `WaitGroup`?

Each feed is scraped concurrently. The `WaitGroup` keeps the scraper cycle from finishing until all feed goroutines finish.

### Why use API-key authentication?

The client sends an API key in the `Authorization` header. Middleware checks the key before calling protected handlers, keeping authentication logic separate from business logic.

### Why is the post URL unique?

RSS feeds are fetched repeatedly. A unique URL prevents the same article from being stored multiple times.

### Why use PostgreSQL?

Users, feeds, follows, and posts have relationships, so a relational database is a good fit. Foreign keys and cascading deletes help maintain data consistency.

## Useful commands

```powershell
go build
go run .
go test ./...
```

