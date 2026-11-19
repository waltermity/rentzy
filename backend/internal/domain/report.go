package domain

import "github.com/shopspring/decimal"

// ItemStats is the raw usage of one tool item in a reporting window.
type ItemStats struct {
	ToolItemID    int64           `json:"tool_item_id"`
	ToolTypeID    int64           `json:"tool_type_id"`
	ToolTypeName  string          `json:"tool_type_name"`
	CategoryName  string          `json:"category_name"`
	Status        ToolStatus      `json:"status"`
	PurchasePrice decimal.Decimal `json:"-"`
	RentalCount   int             `json:"rental_count"`
	RentedHours   float64         `json:"rented_hours"`
	CapacityHours float64         `json:"capacity_hours"` // hours the item existed in the window
	Revenue       decimal.Decimal `json:"revenue"`
}
