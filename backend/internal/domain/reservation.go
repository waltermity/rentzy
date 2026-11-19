package domain

import "time"

type ReservationStatus string

const (
	ReservationPending   ReservationStatus = "PENDING"
	ReservationConfirmed ReservationStatus = "CONFIRMED"
	ReservationCancelled ReservationStatus = "CANCELLED"
	ReservationConverted ReservationStatus = "CONVERTED" // became a rental
	ReservationExpired   ReservationStatus = "EXPIRED"
)

// Open reports whether the reservation still holds its period. New reservations
// are created CONFIRMED; PENDING only appears in older data and is treated the same.
func (s ReservationStatus) Open() bool {
	return s == ReservationPending || s == ReservationConfirmed
}

type Reservation struct {
	ID           int64             `json:"id"`
	CustomerID   int64             `json:"customer_id"`
	CustomerName string            `json:"customer_name,omitempty"`
	ToolItemID   int64             `json:"tool_item_id"`
	ToolTypeID   int64             `json:"tool_type_id"`
	ToolTypeName string            `json:"tool_type_name,omitempty"`
	StartDate    time.Time         `json:"start_date"`
	EndDate      time.Time         `json:"end_date"`
	Status       ReservationStatus `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
}

// BusyInterval is a period in which a tool item is taken by a reservation or rental.
type BusyInterval struct {
	ToolItemID int64     `json:"tool_item_id"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	// "reservation", "rental", or "rental_overdue": a late rental; its End is
	// only the end of the queried window, the item is busy until it is returned.
	Source string `json:"source"`
}
