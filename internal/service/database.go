package service

import (
	"context"
	"fmt"
	"time"
)

// DatabasePing verifies database connectivity with a one-second timeout.
func (s *Service) DatabasePing(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	if err := s.database.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}

	return nil
}
