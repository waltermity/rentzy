package handler

import (
	"time"

	"github.com/gin-gonic/gin"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/http/middleware"
	"rentzy/backend/internal/http/respond"
	"rentzy/backend/internal/service"
)

type ReservationHandler struct{ svc *service.ReservationService }

func NewReservationHandler(svc *service.ReservationService) *ReservationHandler {
	return &ReservationHandler{svc: svc}
}

type createReservationRequest struct {
	ToolTypeID *int64    `json:"tool_type_id" example:"1"`
	ToolItemID *int64    `json:"tool_item_id"` // staff/admin only
	CustomerID *int64    `json:"customer_id"`  // staff/admin only (required for them)
	StartDate  time.Time `json:"start_date" binding:"required" example:"2026-11-20T08:00:00+01:00"`
	EndDate    time.Time `json:"end_date" binding:"required" example:"2026-11-22T08:00:00+01:00"`
}

// Create books a tool for a period.
// @Summary      Create reservation
// @Description  Customers send tool_type_id and get any free item of that type. Staff/admin must send customer_id and may send tool_item_id for an exact item. Overlapping bookings are refused with 409.
// @Tags         reservations
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        body  body      createReservationRequest  true  "Reservation"
// @Success      201   {object}  object{data=domain.Reservation}
// @Failure      400   {object}  respond.ErrorEnvelope
// @Failure      409   {object}  respond.ErrorEnvelope  "no free item"
// @Router       /api/reservations [post]
func (h *ReservationHandler) Create(c *gin.Context) {
	var req createReservationRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.Create(c.Request.Context(), middleware.PrincipalFrom(c), service.CreateReservationInput{
		ToolTypeID: req.ToolTypeID, ToolItemID: req.ToolItemID, CustomerID: req.CustomerID,
		StartDate: req.StartDate, EndDate: req.EndDate})
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, out)
}

// List reservations. Customers see only their own.
// @Summary   List reservations
// @Tags      reservations
// @Produce   json
// @Security  BearerAuth
// @Param     customer_id   query     int     false  "Filter (staff/admin)"
// @Param     tool_item_id  query     int     false  "Filter"
// @Param     status        query     string  false  "Comma-separated: PENDING,CONFIRMED,CANCELLED,CONVERTED,EXPIRED"
// @Param     from          query     string  false  "Overlapping period start"
// @Param     to            query     string  false  "Overlapping period end"
// @Success   200           {object}  object{data=[]domain.Reservation}
// @Router    /api/reservations [get]
func (h *ReservationHandler) List(c *gin.Context) {
	f := service.ReservationListFilter{}
	var ok bool
	if f.CustomerID, ok = queryInt64(c, "customer_id"); !ok {
		return
	}
	if f.ToolItemID, ok = queryInt64(c, "tool_item_id"); !ok {
		return
	}
	if f.From, ok = queryTime(c, "from"); !ok {
		return
	}
	if f.To, ok = queryTime(c, "to"); !ok {
		return
	}
	for _, s := range queryList(c, "status") {
		f.Statuses = append(f.Statuses, domain.ReservationStatus(s))
	}
	out, err := h.svc.List(c.Request.Context(), middleware.PrincipalFrom(c), f)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// Get a reservation.
// @Summary   Get reservation
// @Tags      reservations
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      int  true  "Reservation id"
// @Success   200  {object}  object{data=domain.Reservation}
// @Failure   404  {object}  respond.ErrorEnvelope
// @Router    /api/reservations/{id} [get]
func (h *ReservationHandler) Get(c *gin.Context) {
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

// Cancel a reservation that is still open.
// @Summary   Cancel reservation
// @Tags      reservations
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      int  true  "Reservation id"
// @Success   200  {object}  object{data=domain.Reservation}
// @Failure   404  {object}  respond.ErrorEnvelope
// @Failure   409  {object}  respond.ErrorEnvelope  "not open any more"
// @Router    /api/reservations/{id}/cancel [post]
func (h *ReservationHandler) Cancel(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Cancel(c.Request.Context(), middleware.PrincipalFrom(c), id)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}
