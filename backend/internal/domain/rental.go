package domain

import (
	"math"
	"time"

	"github.com/shopspring/decimal"
)

type RentalStatus string

const (
	RentalActive    RentalStatus = "ACTIVE"
	RentalReturned  RentalStatus = "RETURNED"
	RentalOverdue   RentalStatus = "OVERDUE"
	RentalCancelled RentalStatus = "CANCELLED"
)

// Open reports whether the tool is still out with the customer.
func (s RentalStatus) Open() bool {
	return s == RentalActive || s == RentalOverdue
}

type Rental struct {
	ID                  int64            `json:"id"`
	CustomerID          int64            `json:"customer_id"`
	CustomerName        string           `json:"customer_name,omitempty"`
	ToolItemID          int64            `json:"tool_item_id"`
	ToolTypeID          int64            `json:"tool_type_id"`
	ToolTypeName        string           `json:"tool_type_name,omitempty"`
	ReservationID       int64            `json:"reservation_id"`
	StaffID             int64            `json:"staff_id"`
	StartDate           time.Time        `json:"start_date"`
	ExpectedEndDate     time.Time        `json:"expected_end_date"`
	ActualEndDate       *time.Time       `json:"actual_end_date"`
	DepositPaid         decimal.Decimal  `json:"deposit_paid"`
	ConditionAtHandover Condition        `json:"condition_at_handover"`
	ConditionAtReturn   *Condition       `json:"condition_at_return"`
	BaseFee             *decimal.Decimal `json:"base_fee"`
	LateFee             *decimal.Decimal `json:"late_fee"`
	TotalFee            *decimal.Decimal `json:"total_fee"`
	OperatingHours      *float64         `json:"operating_hours"`
	Status              RentalStatus     `json:"status"`
}

// Fees is the result of a fee calculation.
type Fees struct {
	BaseFee  decimal.Decimal `json:"base_fee"`
	LateFee  decimal.Decimal `json:"late_fee"`
	TotalFee decimal.Decimal `json:"total_fee"`
	LateDays int             `json:"late_days"`
}

// CalculateFees prices a rental.
//
//   - Base fee: daily rate x planned days, where every started 24 hours counts as a
//     day (minimum one). It is charged in full even if the tool comes back early.
//   - Late fee: daily rate x every started 24 hours after the expected end.
func CalculateFees(dailyRate decimal.Decimal, start, expectedEnd, actualEnd time.Time) Fees {
	plannedDays := startedDays(expectedEnd.Sub(start))
	if plannedDays < 1 {
		plannedDays = 1
	}
	lateDays := 0
	if actualEnd.After(expectedEnd) {
		lateDays = startedDays(actualEnd.Sub(expectedEnd))
	}
	base := dailyRate.Mul(decimal.NewFromInt(int64(plannedDays)))
	late := dailyRate.Mul(decimal.NewFromInt(int64(lateDays)))
	return Fees{BaseFee: base, LateFee: late, TotalFee: base.Add(late), LateDays: lateDays}
}

// startedDays counts every started 24-hour period of d.
func startedDays(d time.Duration) int {
	if d <= 0 {
		return 0
	}
	return int(math.Ceil(d.Hours() / 24))
}
