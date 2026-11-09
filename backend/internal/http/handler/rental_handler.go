package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/http/middleware"
	"rentzy/backend/internal/http/respond"
	"rentzy/backend/internal/service"
)

type RentalHandler struct{ svc *service.RentalService }

func NewRentalHandler(svc *service.RentalService) *RentalHandler { return &RentalHandler{svc: svc} }

type handoverRequest struct {
	ReservationID int64             `json:"reservation_id" binding:"required" example:"10"`
	Condition     *domain.Condition `json:"condition_at_handover" enums:"NEW,GOOD,FAIR,POOR"`
}

// Handover hands a reserved tool over and starts the rental.
// @Summary      Hand over (start rental)
// @Description  Starts now, expected end = reservation end, deposit = tool type deposit. Refused (409) if the item is in maintenance, broken, retired or still rented.
// @Tags         rentals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      handoverRequest  true  "Reservation to convert"
// @Success      201   {object}  object{data=domain.Rental}
// @Failure      404   {object}  respond.ErrorEnvelope
// @Failure      409   {object}  respond.ErrorEnvelope
// @Router       /api/rentals [post]
func (h *RentalHandler) Handover(c *gin.Context) {
	var req handoverRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.Handover(c.Request.Context(), middleware.PrincipalFrom(c), service.HandoverInput{
		ReservationID: req.ReservationID, Condition: req.Condition})
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, out)
}

type returnRequest struct {
	Condition      domain.Condition `json:"condition_at_return" binding:"required" enums:"NEW,GOOD,FAIR,POOR,DAMAGED"`
	OperatingHours *float64         `json:"operating_hours" example:"6.5"`
	ReturnedAt     *time.Time       `json:"returned_at"` // defaults to now
}

// Return takes a tool back and computes the fees.
// @Summary      Return (close rental)
// @Description  Base fee = daily rate x started planned days; late fee = daily rate x started days late. The item becomes AVAILABLE, or BROKEN if returned DAMAGED.
// @Tags         rentals
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int            true  "Rental id"
// @Param        body  body      returnRequest  true  "Return data"
// @Success      200   {object}  object{data=service.ReturnResult}
// @Failure      404   {object}  respond.ErrorEnvelope
// @Failure      409   {object}  respond.ErrorEnvelope
// @Router       /api/rentals/{id}/return [post]
func (h *RentalHandler) Return(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req returnRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.Return(c.Request.Context(), id, service.ReturnInput{
		Condition: req.Condition, OperatingHours: req.OperatingHours, ReturnedAt: req.ReturnedAt})
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

type cancelRentalRequest struct {
	Condition *domain.Condition `json:"condition_at_return" enums:"NEW,GOOD,FAIR,POOR,DAMAGED"`
}

// Cancel closes a rental free of charge (tool came back unused).
// @Summary   Cancel rental
// @Tags      rentals
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     id    path      int                  true   "Rental id"
// @Param     body  body      cancelRentalRequest  false  "Condition defaults to the handover condition"
// @Success   200   {object}  object{data=domain.Rental}
// @Failure   404   {object}  respond.ErrorEnvelope
// @Failure   409   {object}  respond.ErrorEnvelope
// @Router    /api/rentals/{id}/cancel [post]
func (h *RentalHandler) Cancel(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req cancelRentalRequest
	if c.Request.ContentLength > 0 && !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.Cancel(c.Request.Context(), id, service.CancelRentalInput{Condition: req.Condition})
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// List rentals. Customers see only their own.
// @Summary   List rentals
// @Tags      rentals
// @Produce   json
// @Security  BearerAuth
// @Param     customer_id   query     int     false  "Filter (staff/admin)"
// @Param     tool_item_id  query     int     false  "Filter"
// @Param     staff_id      query     int     false  "Filter"
// @Param     status        query     string  false  "Comma-separated: ACTIVE,OVERDUE,RETURNED,CANCELLED"
// @Success   200           {object}  object{data=[]domain.Rental}
// @Router    /api/rentals [get]
func (h *RentalHandler) List(c *gin.Context) {
	f := service.RentalListFilter{}
	var ok bool
	if f.CustomerID, ok = queryInt64(c, "customer_id"); !ok {
		return
	}
	if f.ToolItemID, ok = queryInt64(c, "tool_item_id"); !ok {
		return
	}
	if f.StaffID, ok = queryInt64(c, "staff_id"); !ok {
		return
	}
	for _, s := range queryList(c, "status") {
		f.Statuses = append(f.Statuses, domain.RentalStatus(s))
	}
	out, err := h.svc.List(c.Request.Context(), middleware.PrincipalFrom(c), f)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// Get a rental.
// @Summary   Get rental
// @Tags      rentals
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      int  true  "Rental id"
// @Success   200  {object}  object{data=domain.Rental}
// @Failure   404  {object}  respond.ErrorEnvelope
// @Router    /api/rentals/{id} [get]
func (h *RentalHandler) Get(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Get(c.Request.Context(), middleware.PrincipalFrom(c), id)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}
