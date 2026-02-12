package service

import (
	"context"
	"time"
)

// checkConsentExpiries triggers consent expiry checks for all tenants.
func (s *SchedulerService) checkConsentExpiries(ctx context.Context) {
	// Run once a day.
	// Ideally checking "last run" time from DB or similar.
	// For now, simpler: check if hour is 09:00 UTC?
	// Or just run it every hour (the service handles idempotency via renewal logs).
	// The service iterates ALL expiring sessions.

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	if err := s.expirySvc.CheckExpiries(ctx); err != nil {
		s.logger.Error("scheduled consent expiry check failed", "error", err)
	}
}
