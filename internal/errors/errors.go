package errors

import "errors"

// Handler errors and messages
var (
	ErrInvalidID = errors.New("invalid id")
)

// Middleware errors
var (
	ErrClaimsNotFound    = errors.New("claims not found")
	ErrInvalidClaims     = errors.New("invalid claims")
	ErrAdminOnly         = errors.New("admin only")
	ErrAuthHeaderMissing = errors.New("authorization header missing or invalid")
	ErrUnauthorizedToken = errors.New("unauthorized or invalid token")
)

// Connection Errors (for external dependencies)
var (
	// Postgres connection errors
	ErrInvalidPostgresDSN       = errors.New("invalid Postgres DSN")
	ErrPostgresConnectionFailed = errors.New("failed to connect to Postgres")
)

// Connection to Mongo
var (
	ErrMongoConnectionFailed = errors.New("failed to connect to Mongo")
	ErrMongoPingConnection   = errors.New("failed to ping Mongo")
)
