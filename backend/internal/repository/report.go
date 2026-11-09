package repository

import (
	"context"
	"time"

	"rentzy/backend/internal/domain"
)

type ReportRepository interface {
	// ItemStats returns usage per tool item in [from, to): rentals started,
	// rented hours, capacity hours and revenue of rentals returned in the window.
	ItemStats(ctx context.Context, from, to time.Time) ([]domain.ItemStats, error)
}
