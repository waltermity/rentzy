package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

type ReservationRepo struct{ db *DB }

func NewReservationRepo(db *DB) *ReservationRepo { return &ReservationRepo{db: db} }

var _ repository.ReservationRepository = (*ReservationRepo)(nil)

// Statuses that hold a period of a tool item.
const (
	openReservationStatuses = `('PENDING','CONFIRMED')`
	openRentalStatuses      = `('ACTIVE','OVERDUE')`
)

const reservationSelect = `
SELECT r.id, r.customer_id, c.name, r.tool_item_id, ti.tool_type_id, tt.name,
       r.start_date, r.end_date, r.status, r.created_at
FROM reservations r
JOIN customers  c  ON c.id  = r.customer_id
JOIN tool_items ti ON ti.id = r.tool_item_id
JOIN tool_types tt ON tt.id = ti.tool_type_id`

func scanReservation(row pgx.Row) (domain.Reservation, error) {
	var r domain.Reservation
	err := row.Scan(&r.ID, &r.CustomerID, &r.CustomerName, &r.ToolItemID, &r.ToolTypeID, &r.ToolTypeName,
		&r.StartDate, &r.EndDate, &r.Status, &r.CreatedAt)
	return r, err
}

func (r *ReservationRepo) List(ctx context.Context, f repository.ReservationFilter) ([]domain.Reservation, error) {
	var where []string
	var args []any
	add := func(cond string, v any) {
		args = append(args, v)
		where = append(where, fmt.Sprintf(cond, len(args)))
	}
	if f.CustomerID != nil {
		add("r.customer_id = $%d", *f.CustomerID)
	}
	if f.ToolItemID != nil {
		add("r.tool_item_id = $%d", *f.ToolItemID)
	}
	if len(f.Statuses) > 0 {
		ss := make([]string, len(f.Statuses))
		for i, s := range f.Statuses {
			ss[i] = string(s)
		}
		add("r.status = ANY($%d)", ss)
	}
	if f.From != nil {
		add("r.end_date > $%d", *f.From)
	}
	if f.To != nil {
		add("r.start_date < $%d", *f.To)
	}
	q := reservationSelect
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY r.start_date DESC, r.id DESC"
	rows, err := r.db.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err, "list reservations")
	}
	out, err := collect(rows, func(row pgx.Rows) (domain.Reservation, error) { return scanReservation(row) })
	return out, mapErr(err, "list reservations")
}

func (r *ReservationRepo) GetByID(ctx context.Context, id int64) (domain.Reservation, error) {
	res, err := scanReservation(r.db.q(ctx).QueryRow(ctx, reservationSelect+" WHERE r.id = $1", id))
	return res, mapErr(err, "get reservation")
}

func (r *ReservationRepo) GetByIDForUpdate(ctx context.Context, id int64) (domain.Reservation, error) {
	res, err := scanReservation(r.db.q(ctx).QueryRow(ctx, reservationSelect+" WHERE r.id = $1 FOR UPDATE OF r", id))
	return res, mapErr(err, "lock reservation")
}

func (r *ReservationRepo) Create(ctx context.Context, res *domain.Reservation) error {
	err := r.db.q(ctx).QueryRow(ctx, `
		INSERT INTO reservations (customer_id, tool_item_id, start_date, end_date, status)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at`,
		res.CustomerID, res.ToolItemID, res.StartDate, res.EndDate, res.Status).Scan(&res.ID, &res.CreatedAt)
	return mapErr(err, "create reservation")
}

func (r *ReservationRepo) UpdateStatus(ctx context.Context, id int64, status domain.ReservationStatus) error {
	tag, err := r.db.q(ctx).Exec(ctx, `UPDATE reservations SET status = $2 WHERE id = $1`, id, status)
	if err != nil {
		return mapErr(err, "update reservation")
	}
	return expectOne(tag, "update reservation")
}

// An open rental holds its item until its expected end. A late rental (past its
// expected end, not returned) holds it open-endedly: nobody knows when it comes
// back, so no new period of that item can be promised until it is returned.
func (r *ReservationRepo) HasConflict(ctx context.Context, itemID int64, start, end time.Time, excludeID int64) (bool, error) {
	var conflict bool
	err := r.db.q(ctx).QueryRow(ctx, `
		SELECT EXISTS (
		    SELECT 1 FROM reservations
		    WHERE tool_item_id = $1 AND id <> $4 AND status IN `+openReservationStatuses+`
		      AND start_date < $3 AND end_date > $2
		) OR EXISTS (
		    SELECT 1 FROM rentals
		    WHERE tool_item_id = $1 AND status IN `+openRentalStatuses+`
		      AND start_date < $3 AND (expected_end_date > $2 OR expected_end_date < now())
		)`, itemID, start, end, excludeID).Scan(&conflict)
	return conflict, mapErr(err, "check conflict")
}

func (r *ReservationRepo) BusyIntervals(ctx context.Context, typeID int64, from, to time.Time) ([]domain.BusyInterval, error) {
	rows, err := r.db.q(ctx).Query(ctx, `
		SELECT res.tool_item_id, res.start_date, res.end_date, 'reservation'
		FROM reservations res JOIN tool_items ti ON ti.id = res.tool_item_id
		WHERE ti.tool_type_id = $1 AND res.status IN `+openReservationStatuses+`
		  AND res.start_date < $3 AND res.end_date > $2
		UNION ALL
		SELECT ren.tool_item_id, ren.start_date,
		       CASE WHEN ren.expected_end_date < now() THEN GREATEST($3, now()) ELSE ren.expected_end_date END,
		       CASE WHEN ren.expected_end_date < now() THEN 'rental_overdue' ELSE 'rental' END
		FROM rentals ren JOIN tool_items ti ON ti.id = ren.tool_item_id
		WHERE ti.tool_type_id = $1 AND ren.status IN `+openRentalStatuses+`
		  AND ren.start_date < $3 AND (ren.expected_end_date > $2 OR ren.expected_end_date < now())
		ORDER BY 1, 2`, typeID, from, to)
	if err != nil {
		return nil, mapErr(err, "busy intervals")
	}
	out, err := collect(rows, func(row pgx.Rows) (domain.BusyInterval, error) {
		var b domain.BusyInterval
		return b, row.Scan(&b.ToolItemID, &b.Start, &b.End, &b.Source)
	})
	return out, mapErr(err, "busy intervals")
}

func (r *ReservationRepo) ExpireStale(ctx context.Context, cutoff time.Time) (int64, error) {
	tag, err := r.db.q(ctx).Exec(ctx, `
		UPDATE reservations SET status = 'EXPIRED'
		WHERE status IN `+openReservationStatuses+` AND start_date < $1`, cutoff)
	if err != nil {
		return 0, mapErr(err, "expire reservations")
	}
	return tag.RowsAffected(), nil
}
