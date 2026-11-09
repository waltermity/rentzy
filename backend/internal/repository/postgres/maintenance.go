package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

type MaintenanceRepo struct{ db *DB }

func NewMaintenanceRepo(db *DB) *MaintenanceRepo { return &MaintenanceRepo{db: db} }

var _ repository.MaintenanceRepository = (*MaintenanceRepo)(nil)

const maintenanceSelect = `
SELECT m.id, m.tool_item_id, tt.name, m.type, COALESCE(m.description, ''), m.started_at, m.completed_at,
       COALESCE(m.cost, 0), COALESCE(m.performed_by, '')
FROM maintenance m
JOIN tool_items ti ON ti.id = m.tool_item_id
JOIN tool_types tt ON tt.id = ti.tool_type_id`

func scanMaintenance(row pgx.Row) (domain.Maintenance, error) {
	var m domain.Maintenance
	err := row.Scan(&m.ID, &m.ToolItemID, &m.ToolTypeName, &m.Type, &m.Description, &m.StartedAt,
		&m.CompletedAt, &m.Cost, &m.PerformedBy)
	return m, err
}

func (r *MaintenanceRepo) List(ctx context.Context, f repository.MaintenanceFilter) ([]domain.Maintenance, error) {
	var where []string
	var args []any
	if f.ToolItemID != nil {
		args = append(args, *f.ToolItemID)
		where = append(where, fmt.Sprintf("m.tool_item_id = $%d", len(args)))
	}
	if f.Open != nil {
		if *f.Open {
			where = append(where, "m.completed_at IS NULL")
		} else {
			where = append(where, "m.completed_at IS NOT NULL")
		}
	}
	q := maintenanceSelect
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY m.started_at DESC, m.id DESC"
	rows, err := r.db.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err, "list maintenance")
	}
	out, err := collect(rows, func(row pgx.Rows) (domain.Maintenance, error) { return scanMaintenance(row) })
	return out, mapErr(err, "list maintenance")
}

func (r *MaintenanceRepo) GetByID(ctx context.Context, id int64) (domain.Maintenance, error) {
	m, err := scanMaintenance(r.db.q(ctx).QueryRow(ctx, maintenanceSelect+" WHERE m.id = $1", id))
	return m, mapErr(err, "get maintenance")
}

func (r *MaintenanceRepo) GetByIDForUpdate(ctx context.Context, id int64) (domain.Maintenance, error) {
	m, err := scanMaintenance(r.db.q(ctx).QueryRow(ctx, maintenanceSelect+" WHERE m.id = $1 FOR UPDATE OF m", id))
	return m, mapErr(err, "lock maintenance")
}

func (r *MaintenanceRepo) Create(ctx context.Context, m *domain.Maintenance) error {
	err := r.db.q(ctx).QueryRow(ctx, `
		INSERT INTO maintenance (tool_item_id, type, description, started_at, completed_at, cost, performed_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id`,
		m.ToolItemID, m.Type, nullIfEmpty(m.Description), m.StartedAt, m.CompletedAt, m.Cost,
		nullIfEmpty(m.PerformedBy)).Scan(&m.ID)
	return mapErr(err, "create maintenance")
}

func (r *MaintenanceRepo) Complete(ctx context.Context, m *domain.Maintenance) error {
	tag, err := r.db.q(ctx).Exec(ctx, `
		UPDATE maintenance SET completed_at = $2, cost = $3, description = $4, performed_by = $5
		WHERE id = $1`,
		m.ID, m.CompletedAt, m.Cost, nullIfEmpty(m.Description), nullIfEmpty(m.PerformedBy))
	if err != nil {
		return mapErr(err, "complete maintenance")
	}
	return expectOne(tag, "complete maintenance")
}

func (r *MaintenanceRepo) CountOpen(ctx context.Context, itemID int64) (int, error) {
	var n int
	err := r.db.q(ctx).QueryRow(ctx,
		`SELECT COUNT(*) FROM maintenance WHERE tool_item_id = $1 AND completed_at IS NULL`, itemID).Scan(&n)
	return n, mapErr(err, "count open maintenance")
}

// Usage counts non-cancelled rentals started, and operating hours recorded on
// returned rentals, since the item's last completed maintenance.
func (r *MaintenanceRepo) Usage(ctx context.Context, itemID *int64) ([]domain.MaintenanceUsage, error) {
	q := `
		SELECT ti.id, ti.tool_type_id, tt.name, ti.status, lm.last_at, s.since,
		       (SELECT COUNT(*) FROM rentals r
		        WHERE r.tool_item_id = ti.id AND r.status <> 'CANCELLED' AND r.start_date >= s.since),
		       (SELECT COALESCE(SUM(r.operating_hours), 0)::float8 FROM rentals r
		        WHERE r.tool_item_id = ti.id AND r.status = 'RETURNED' AND r.start_date >= s.since),
		       tt.maintenance_every_rentals, tt.maintenance_every_days, tt.maintenance_every_hours::float8
		FROM tool_items ti
		JOIN tool_types tt ON tt.id = ti.tool_type_id
		LEFT JOIN LATERAL (SELECT MAX(m.completed_at) AS last_at FROM maintenance m
		                   WHERE m.tool_item_id = ti.id AND m.completed_at IS NOT NULL) lm ON true
		CROSS JOIN LATERAL (SELECT COALESCE(lm.last_at, ti.acquired_at) AS since) s
		WHERE ti.status <> 'RETIRED'`
	var args []any
	if itemID != nil {
		q += " AND ti.id = $1"
		args = append(args, *itemID)
	}
	q += " ORDER BY ti.id"
	rows, err := r.db.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err, "maintenance usage")
	}
	out, err := collect(rows, func(row pgx.Rows) (domain.MaintenanceUsage, error) {
		var u domain.MaintenanceUsage
		return u, row.Scan(&u.ToolItemID, &u.ToolTypeID, &u.ToolTypeName, &u.Status, &u.LastMaintenanceAt,
			&u.Since, &u.RentalsSince, &u.HoursSince, &u.EveryRentals, &u.EveryDays, &u.EveryHours)
	})
	return out, mapErr(err, "maintenance usage")
}
