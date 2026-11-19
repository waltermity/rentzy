package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

type ReportRepo struct{ db *DB }

func NewReportRepo(db *DB) *ReportRepo { return &ReportRepo{db: db} }

var _ repository.ReportRepository = (*ReportRepo)(nil)

// ItemStats: the window ends at `to` or now, whichever is earlier, for the hour
// figures. Rented hours are the part of each non-cancelled rental (until its
// return, or now if still out) that falls into the window. Capacity hours are the
// part of the window after the item was acquired. Revenue is the total fee of
// rentals returned in the window.
func (r *ReportRepo) ItemStats(ctx context.Context, from, to time.Time) ([]domain.ItemStats, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		WITH w AS (SELECT $1::timestamptz AS f, LEAST($2::timestamptz, now()) AS t, $2::timestamptz AS t_rev)
		SELECT ti.id, ti.tool_type_id, tt.name, c.name, ti.status, tt.purchase_price,
		       (SELECT COUNT(*) FROM rentals r
		        WHERE r.tool_item_id = ti.id AND r.status <> 'CANCELLED'
		          AND r.start_date >= w.f AND r.start_date < w.t_rev),
		       (SELECT COALESCE(SUM(EXTRACT(EPOCH FROM GREATEST(
		                   LEAST(COALESCE(r.actual_end_date, now()), w.t) - GREATEST(r.start_date, w.f),
		                   interval '0'))), 0) / 3600.0
		        FROM rentals r
		        WHERE r.tool_item_id = ti.id AND r.status <> 'CANCELLED')::float8,
		       (EXTRACT(EPOCH FROM GREATEST(w.t - GREATEST(w.f, ti.acquired_at), interval '0')) / 3600.0)::float8,
		       (SELECT COALESCE(SUM(r.total_fee), 0) FROM rentals r
		        WHERE r.tool_item_id = ti.id AND r.status = 'RETURNED'
		          AND r.actual_end_date >= w.f AND r.actual_end_date < w.t_rev)
		FROM tool_items ti
		JOIN tool_types tt ON tt.id = ti.tool_type_id
		JOIN categories c  ON c.id  = tt.category_id
		CROSS JOIN w
		ORDER BY ti.id`, from, to)
	if err != nil {
		return nil, mapErr(err, "item stats")
	}
	out, err := collect(rows, func(row pgx.Rows) (domain.ItemStats, error) {
		var s domain.ItemStats
		return s, row.Scan(&s.ToolItemID, &s.ToolTypeID, &s.ToolTypeName, &s.CategoryName, &s.Status,
			&s.PurchasePrice, &s.RentalCount, &s.RentedHours, &s.CapacityHours, &s.Revenue)
	})
	return out, mapErr(err, "item stats")
}
