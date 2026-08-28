package handler

import "context"

// Database defines the methods required for database operations used by handlers
type Database interface {
	PingContext(ctx context.Context) error
}
