package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/repository"
)

// ---------- categories ----------

type CategoryRepo struct{ db *DB }

func NewCategoryRepo(db *DB) *CategoryRepo { return &CategoryRepo{db: db} }

var _ repository.CategoryRepository = (*CategoryRepo)(nil)

func (r *CategoryRepo) List(ctx context.Context) ([]domain.Category, error) {
	rows, err := r.db.q(ctx).Query(ctx, `SELECT id, name FROM categories ORDER BY name`)
	if err != nil {
		return nil, mapErr(err, "list categories")
	}
	out, err := collect(rows, func(row pgx.Rows) (domain.Category, error) {
		var c domain.Category
		return c, row.Scan(&c.ID, &c.Name)
	})
	return out, mapErr(err, "list categories")
}

func (r *CategoryRepo) GetByID(ctx context.Context, id int64) (domain.Category, error) {
	var c domain.Category
	err := r.db.q(ctx).QueryRow(ctx, `SELECT id, name FROM categories WHERE id = $1`, id).Scan(&c.ID, &c.Name)
	return c, mapErr(err, "get category")
}

func (r *CategoryRepo) Create(ctx context.Context, c *domain.Category) error {
	err := r.db.q(ctx).QueryRow(ctx, `INSERT INTO categories (name) VALUES ($1) RETURNING id`, c.Name).Scan(&c.ID)
	return mapErr(err, "create category")
}

func (r *CategoryRepo) Update(ctx context.Context, c *domain.Category) error {
	tag, err := r.db.q(ctx).Exec(ctx, `UPDATE categories SET name = $2 WHERE id = $1`, c.ID, c.Name)
	if err != nil {
		return mapErr(err, "update category")
	}
	return expectOne(tag, "update category")
}

// ---------- tool types ----------

type ToolTypeRepo struct{ db *DB }

func NewToolTypeRepo(db *DB) *ToolTypeRepo { return &ToolTypeRepo{db: db} }

var _ repository.ToolTypeRepository = (*ToolTypeRepo)(nil)

const toolTypeSelect = `
SELECT tt.id, tt.name, tt.category_id, c.name,
       COALESCE(tt.manufacturer, ''), COALESCE(tt.model, ''),
       tt.purchase_price, tt.daily_rate, tt.deposit,
       tt.maintenance_every_rentals, tt.maintenance_every_days, tt.maintenance_every_hours::float8,
       COALESCE((SELECT jsonb_object_agg(s.status, s.n)
                 FROM (SELECT status, COUNT(*) AS n FROM tool_items
                       WHERE tool_type_id = tt.id GROUP BY status) s), '{}'::jsonb)
FROM tool_types tt
JOIN categories c ON c.id = tt.category_id`

func scanToolType(row pgx.Row) (domain.ToolType, error) {
	var t domain.ToolType
	err := row.Scan(&t.ID, &t.Name, &t.CategoryID, &t.CategoryName, &t.Manufacturer, &t.Model,
		&t.PurchasePrice, &t.DailyRate, &t.Deposit,
		&t.MaintenanceEveryRentals, &t.MaintenanceEveryDays, &t.MaintenanceEveryHours, &t.ItemCounts)
	return t, err
}

func (r *ToolTypeRepo) List(ctx context.Context, f repository.ToolTypeFilter) ([]domain.ToolType, error) {
	var where []string
	var args []any
	if f.CategoryID != nil {
		args = append(args, *f.CategoryID)
		where = append(where, fmt.Sprintf("tt.category_id = $%d", len(args)))
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		args = append(args, "%"+s+"%")
		n := len(args)
		where = append(where, fmt.Sprintf("(tt.name ILIKE $%d OR tt.manufacturer ILIKE $%d OR tt.model ILIKE $%d)", n, n, n))
	}
	q := toolTypeSelect
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY tt.name, tt.id"
	rows, err := r.db.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err, "list tool types")
	}
	out, err := collect(rows, func(row pgx.Rows) (domain.ToolType, error) { return scanToolType(row) })
	return out, mapErr(err, "list tool types")
}

func (r *ToolTypeRepo) GetByID(ctx context.Context, id int64) (domain.ToolType, error) {
	t, err := scanToolType(r.db.q(ctx).QueryRow(ctx, toolTypeSelect+" WHERE tt.id = $1", id))
	return t, mapErr(err, "get tool type")
}

func (r *ToolTypeRepo) Create(ctx context.Context, t *domain.ToolType) error {
	err := r.db.q(ctx).QueryRow(ctx, `
		INSERT INTO tool_types (name, category_id, manufacturer, model, purchase_price, daily_rate, deposit,
		                        maintenance_every_rentals, maintenance_every_days, maintenance_every_hours)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id`,
		t.Name, t.CategoryID, nullIfEmpty(t.Manufacturer), nullIfEmpty(t.Model),
		t.PurchasePrice, t.DailyRate, t.Deposit,
		t.MaintenanceEveryRentals, t.MaintenanceEveryDays, t.MaintenanceEveryHours,
	).Scan(&t.ID)
	return mapErr(err, "create tool type")
}

func (r *ToolTypeRepo) Update(ctx context.Context, t *domain.ToolType) error {
	tag, err := r.db.q(ctx).Exec(ctx, `
		UPDATE tool_types
		SET name = $2, category_id = $3, manufacturer = $4, model = $5,
		    purchase_price = $6, daily_rate = $7, deposit = $8,
		    maintenance_every_rentals = $9, maintenance_every_days = $10, maintenance_every_hours = $11
		WHERE id = $1`,
		t.ID, t.Name, t.CategoryID, nullIfEmpty(t.Manufacturer), nullIfEmpty(t.Model),
		t.PurchasePrice, t.DailyRate, t.Deposit,
		t.MaintenanceEveryRentals, t.MaintenanceEveryDays, t.MaintenanceEveryHours)
	if err != nil {
		return mapErr(err, "update tool type")
	}
	return expectOne(tag, "update tool type")
}

// ---------- tool items ----------

type ToolItemRepo struct{ db *DB }

func NewToolItemRepo(db *DB) *ToolItemRepo { return &ToolItemRepo{db: db} }

var _ repository.ToolItemRepository = (*ToolItemRepo)(nil)

const toolItemSelect = `
SELECT ti.id, ti.tool_type_id, tt.name, ti.status, ti.current_condition, ti.acquired_at
FROM tool_items ti
JOIN tool_types tt ON tt.id = ti.tool_type_id`

func scanToolItem(row pgx.Row) (domain.ToolItem, error) {
	var it domain.ToolItem
	err := row.Scan(&it.ID, &it.ToolTypeID, &it.ToolTypeName, &it.Status, &it.CurrentCondition, &it.AcquiredAt)
	return it, err
}

func (r *ToolItemRepo) List(ctx context.Context, f repository.ToolItemFilter) ([]domain.ToolItem, error) {
	var where []string
	var args []any
	if f.ToolTypeID != nil {
		args = append(args, *f.ToolTypeID)
		where = append(where, fmt.Sprintf("ti.tool_type_id = $%d", len(args)))
	}
	if f.Status != nil {
		args = append(args, string(*f.Status))
		where = append(where, fmt.Sprintf("ti.status = $%d", len(args)))
	}
	q := toolItemSelect
	if len(where) > 0 {
		q += " WHERE " + strings.Join(where, " AND ")
	}
	q += " ORDER BY ti.id"
	rows, err := r.db.q(ctx).Query(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err, "list tool items")
	}
	out, err := collect(rows, func(row pgx.Rows) (domain.ToolItem, error) { return scanToolItem(row) })
	return out, mapErr(err, "list tool items")
}

func (r *ToolItemRepo) GetByID(ctx context.Context, id int64) (domain.ToolItem, error) {
	it, err := scanToolItem(r.db.q(ctx).QueryRow(ctx, toolItemSelect+" WHERE ti.id = $1", id))
	return it, mapErr(err, "get tool item")
}

func (r *ToolItemRepo) GetByIDForUpdate(ctx context.Context, id int64) (domain.ToolItem, error) {
	it, err := scanToolItem(r.db.q(ctx).QueryRow(ctx, toolItemSelect+" WHERE ti.id = $1 FOR UPDATE OF ti", id))
	return it, mapErr(err, "lock tool item")
}

func (r *ToolItemRepo) ListByTypeForUpdate(ctx context.Context, toolTypeID int64) ([]domain.ToolItem, error) {
	rows, err := r.db.q(ctx).Query(ctx, toolItemSelect+" WHERE ti.tool_type_id = $1 ORDER BY ti.id FOR UPDATE OF ti", toolTypeID)
	if err != nil {
		return nil, mapErr(err, "lock tool items")
	}
	out, err := collect(rows, func(row pgx.Rows) (domain.ToolItem, error) { return scanToolItem(row) })
	return out, mapErr(err, "lock tool items")
}

func (r *ToolItemRepo) Create(ctx context.Context, it *domain.ToolItem) error {
	err := r.db.q(ctx).QueryRow(ctx, `
		INSERT INTO tool_items (tool_type_id, status, current_condition, acquired_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`,
		it.ToolTypeID, it.Status, it.CurrentCondition, it.AcquiredAt).Scan(&it.ID)
	return mapErr(err, "create tool item")
}

func (r *ToolItemRepo) UpdateState(ctx context.Context, id int64, status domain.ToolStatus, condition domain.Condition) error {
	tag, err := r.db.q(ctx).Exec(ctx,
		`UPDATE tool_items SET status = $2, current_condition = $3 WHERE id = $1`, id, status, condition)
	if err != nil {
		return mapErr(err, "update tool item")
	}
	return expectOne(tag, "update tool item")
}
