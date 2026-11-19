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

type RentalRepo struct{ db *DB }

func NewRentalRepo(db *DB) *RentalRepo { return &RentalRepo{db: db} }

var _ repository.RentalRepository = (*RentalRepo)(nil)

const rentalSelect = `
SELECT r.id, r.customer_id, c.name, r.tool_item_id, ti.tool_type_id, tt.name, r.reservation_id, r.staff_id,
       r.start_date, r.expected_end_date, r.actual_end_date, r.deposit_paid,
       r.condition_at_handover, r.condition_at_return, r.base_fee, r.late_fee, r.total_fee,
       r.operating_hours::float8, r.status
FROM rentals r
JOIN customers  c  ON c.id  = r.customer_id
JOIN tool_items ti ON ti.id = r.tool_item_id
JOIN tool_types tt ON tt.id = ti.tool_type_id`

func scanRental(row pgx.Row) (domain.Rental, error) {
	var r domain.Rental
	err := row.Scan(&r.ID, &r.CustomerID, &r.CustomerName, &r.ToolItemID, &r.ToolTypeID, &r.ToolTypeName,
		&r.ReservationID, &r.StaffID, &r.StartDate, &r.ExpectedEndDate, &r.ActualEndDate, &r.DepositPaid,
		&r.ConditionAtHandover, &r.ConditionAtReturn, &r.BaseFee, &r.LateFee, &r.TotalFee,
		&r.OperatingHours, &r.Status)
	return r, err
}

func (r *RentalRepo) List(ctx context.Context, f repository.RentalFilter) ([]domain.Rental, error) {
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
	if f.StaffID != nil {
		add("r.staff_id = $%d", *f.StaffID)
	}
	if len(f.Statuses) > 0 {
		ss := make([]string, len(f.Statuses))
		for i, s := range f.Statuses {
			ss[i] = string(s)
		}
		add("r.status = ANY($%d)", ss)
	}
	q := rentalSelect
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY r.start_date DESC, r.id DESC"
	rows, err := r.db.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err, "list rentals")
	}
	out, err := collect(rows, func(row pgx.Rows) (domain.Rental, error) { return scanRental(row) })
	return out, mapErr(err, "list rentals")
}

func (r *RentalRepo) GetByID(ctx context.Context, id int64) (domain.Rental, error) {
	ren, err := scanRental(r.db.q(ctx).QueryRow(ctx, rentalSelect+" WHERE r.id = $1", id))
	return ren, mapErr(err, "get rental")
}

func (r *RentalRepo) GetByIDForUpdate(ctx context.Context, id int64) (domain.Rental, error) {
	ren, err := scanRental(r.db.q(ctx).QueryRow(ctx, rentalSelect+" WHERE r.id = $1 FOR UPDATE OF r", id))
	return ren, mapErr(err, "lock rental")
}

func (r *RentalRepo) Create(ctx context.Context, ren *domain.Rental) error {
	err := r.db.q(ctx).QueryRow(ctx, `
		INSERT INTO rentals (customer_id, tool_item_id, reservation_id, staff_id, start_date, expected_end_date,
		                     deposit_paid, condition_at_handover, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id`,
		ren.CustomerID, ren.ToolItemID, ren.ReservationID, ren.StaffID, ren.StartDate, ren.ExpectedEndDate,
		ren.DepositPaid, ren.ConditionAtHandover, ren.Status).Scan(&ren.ID)
	return mapErr(err, "create rental")
}

func (r *RentalRepo) Close(ctx context.Context, ren *domain.Rental) error {
	tag, err := r.db.q(ctx).Exec(ctx, `
		UPDATE rentals
		SET actual_end_date = $2, condition_at_return = $3, base_fee = $4, late_fee = $5, total_fee = $6,
		    operating_hours = $7, status = $8
		WHERE id = $1`,
		ren.ID, ren.ActualEndDate, ren.ConditionAtReturn, ren.BaseFee, ren.LateFee, ren.TotalFee,
		ren.OperatingHours, ren.Status)
	if err != nil {
		return mapErr(err, "close rental")
	}
	return expectOne(tag, "close rental")
}

func (r *RentalRepo) MarkOverdue(ctx context.Context, now time.Time) (int64, error) {
	tag, err := r.db.q(ctx).Exec(ctx,
		`UPDATE rentals SET status = 'OVERDUE' WHERE status = 'ACTIVE' AND expected_end_date < $1`, now)
	if err != nil {
		return 0, mapErr(err, "mark overdue")
	}
	return tag.RowsAffected(), nil
}
