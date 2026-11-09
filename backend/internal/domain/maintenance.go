package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type MaintenanceType string

const (
	MaintenanceScheduled  MaintenanceType = "SCHEDULED"
	MaintenanceRepair     MaintenanceType = "REPAIR"
	MaintenanceInspection MaintenanceType = "INSPECTION"
)

func (t MaintenanceType) Valid() bool {
	return t == MaintenanceScheduled || t == MaintenanceRepair || t == MaintenanceInspection
}

// Maintenance is one maintenance or repair event. It is open while CompletedAt is nil.
type Maintenance struct {
	ID           int64           `json:"id"`
	ToolItemID   int64           `json:"tool_item_id"`
	ToolTypeName string          `json:"tool_type_name,omitempty"`
	Type         MaintenanceType `json:"type"`
	Description  string          `json:"description"`
	StartedAt    time.Time       `json:"started_at"`
	CompletedAt  *time.Time      `json:"completed_at"`
	Cost         decimal.Decimal `json:"cost"`
	PerformedBy  string          `json:"performed_by"`
}

// MaintenanceUsage is how much a tool item was used since its last completed
// maintenance (or since acquisition), together with its type's thresholds.
type MaintenanceUsage struct {
	ToolItemID        int64      `json:"tool_item_id"`
	ToolTypeID        int64      `json:"tool_type_id"`
	ToolTypeName      string     `json:"tool_type_name"`
	Status            ToolStatus `json:"status"`
	LastMaintenanceAt *time.Time `json:"last_maintenance_at"`
	Since             time.Time  `json:"since"` // last maintenance, or acquisition
	RentalsSince      int        `json:"rentals_since"`
	HoursSince        float64    `json:"operating_hours_since"`

	EveryRentals *int     `json:"maintenance_every_rentals"`
	EveryDays    *int     `json:"maintenance_every_days"`
	EveryHours   *float64 `json:"maintenance_every_hours"`
}

// MaintenanceDue is the due-check result for one tool item.
type MaintenanceDue struct {
	MaintenanceUsage
	DaysSince int      `json:"days_since"`
	Due       bool     `json:"due"`
	Reasons   []string `json:"reasons"`
}

// CheckDue applies the thresholds to the usage: any reached threshold makes it due.
func CheckDue(u MaintenanceUsage, now time.Time) MaintenanceDue {
	d := MaintenanceDue{MaintenanceUsage: u, Reasons: []string{}}
	d.DaysSince = int(now.Sub(u.Since).Hours() / 24)
	if u.EveryRentals != nil && u.RentalsSince >= *u.EveryRentals {
		d.Reasons = append(d.Reasons, "RENTAL_COUNT")
	}
	if u.EveryDays != nil && d.DaysSince >= *u.EveryDays {
		d.Reasons = append(d.Reasons, "ELAPSED_DAYS")
	}
	if u.EveryHours != nil && u.HoursSince >= *u.EveryHours {
		d.Reasons = append(d.Reasons, "OPERATING_HOURS")
	}
	d.Due = len(d.Reasons) > 0
	return d
}
