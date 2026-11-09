package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// ToolStatus is the lifecycle status of a physical tool item.
type ToolStatus string

const (
	ToolAvailable   ToolStatus = "AVAILABLE"
	ToolRented      ToolStatus = "RENTED"
	ToolMaintenance ToolStatus = "MAINTENANCE"
	ToolBroken      ToolStatus = "BROKEN"
	ToolRetired     ToolStatus = "RETIRED"
)

func (s ToolStatus) Valid() bool {
	switch s {
	case ToolAvailable, ToolRented, ToolMaintenance, ToolBroken, ToolRetired:
		return true
	}
	return false
}

// Bookable reports whether an item in this status may get new reservations.
// Items in maintenance, broken or retired cannot be reserved or rented.
// RENTED items can be reserved for a later, non-overlapping period.
func (s ToolStatus) Bookable() bool {
	return s == ToolAvailable || s == ToolRented
}

// Condition is the physical condition of a tool item.
type Condition string

const (
	ConditionNew     Condition = "NEW"
	ConditionGood    Condition = "GOOD"
	ConditionFair    Condition = "FAIR"
	ConditionPoor    Condition = "POOR"
	ConditionDamaged Condition = "DAMAGED"
)

func (c Condition) Valid() bool {
	switch c {
	case ConditionNew, ConditionGood, ConditionFair, ConditionPoor, ConditionDamaged:
		return true
	}
	return false
}

type Category struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ToolType is a kind of tool (name, manufacturer, model, prices). Each physical
// copy is a ToolItem with its own id.
type ToolType struct {
	ID            int64           `json:"id"`
	Name          string          `json:"name"`
	CategoryID    int64           `json:"category_id"`
	CategoryName  string          `json:"category_name,omitempty"`
	Manufacturer  string          `json:"manufacturer"`
	Model         string          `json:"model"`
	PurchasePrice decimal.Decimal `json:"purchase_price"`
	DailyRate     decimal.Decimal `json:"daily_rate"`
	Deposit       decimal.Decimal `json:"deposit"`

	// Maintenance thresholds; nil means the rule is not used.
	MaintenanceEveryRentals *int     `json:"maintenance_every_rentals"`
	MaintenanceEveryDays    *int     `json:"maintenance_every_days"`
	MaintenanceEveryHours   *float64 `json:"maintenance_every_hours"`

	// Number of items per status; filled by list/get queries.
	ItemCounts map[ToolStatus]int `json:"item_counts,omitempty"`
}

// ToolItem is one physical, individually identified copy of a ToolType.
type ToolItem struct {
	ID               int64      `json:"id"`
	ToolTypeID       int64      `json:"tool_type_id"`
	ToolTypeName     string     `json:"tool_type_name,omitempty"`
	Status           ToolStatus `json:"status"`
	CurrentCondition Condition  `json:"current_condition"`
	AcquiredAt       time.Time  `json:"acquired_at"`
}
