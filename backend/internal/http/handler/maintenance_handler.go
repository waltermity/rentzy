package handler

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/http/respond"
	"rentzy/backend/internal/repository"
	"rentzy/backend/internal/service"
)

type MaintenanceHandler struct{ svc *service.MaintenanceService }

func NewMaintenanceHandler(svc *service.MaintenanceService) *MaintenanceHandler {
	return &MaintenanceHandler{svc: svc}
}

type startMaintenanceRequest struct {
	Type        domain.MaintenanceType `json:"type" binding:"required" enums:"SCHEDULED,REPAIR,INSPECTION"`
	Description string                 `json:"description" example:"Chain sharpening"`
	PerformedBy string                 `json:"performed_by" example:"toth.gabor"`
	StartedAt   *time.Time             `json:"started_at"` // defaults to now
}

// Start opens a maintenance record; the item goes into MAINTENANCE.
// @Summary   Start maintenance
// @Tags      maintenance
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     id    path      int                      true  "Tool item id"
// @Param     body  body      startMaintenanceRequest  true  "Maintenance"
// @Success   201   {object}  object{data=domain.Maintenance}
// @Failure   404   {object}  respond.ErrorEnvelope
// @Failure   409   {object}  respond.ErrorEnvelope  "item rented or retired"
// @Router    /api/tool-items/{id}/maintenance [post]
func (h *MaintenanceHandler) Start(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req startMaintenanceRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.Start(c.Request.Context(), id, service.StartMaintenanceInput{
		Type: req.Type, Description: req.Description, PerformedBy: req.PerformedBy, StartedAt: req.StartedAt})
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, out)
}

type completeMaintenanceRequest struct {
	Cost         decimal.Decimal   `json:"cost" swaggertype:"number" example:"6500"`
	Description  *string           `json:"description"`
	PerformedBy  *string           `json:"performed_by"`
	CompletedAt  *time.Time        `json:"completed_at"`                                   // defaults to now
	ResultStatus domain.ToolStatus `json:"result_status" enums:"AVAILABLE,BROKEN,RETIRED"` // defaults to AVAILABLE
	Condition    *domain.Condition `json:"condition" enums:"NEW,GOOD,FAIR,POOR,DAMAGED"`
}

// Complete closes a maintenance record and updates the item.
// @Summary   Complete maintenance
// @Tags      maintenance
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     id    path      int                         true  "Maintenance id"
// @Param     body  body      completeMaintenanceRequest  true  "Result"
// @Success   200   {object}  object{data=domain.Maintenance}
// @Failure   400   {object}  respond.ErrorEnvelope
// @Failure   404   {object}  respond.ErrorEnvelope
// @Failure   409   {object}  respond.ErrorEnvelope  "already completed"
// @Router    /api/maintenance/{id}/complete [post]
func (h *MaintenanceHandler) Complete(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req completeMaintenanceRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.Complete(c.Request.Context(), id, service.CompleteMaintenanceInput{
		Cost: req.Cost, Description: req.Description, PerformedBy: req.PerformedBy,
		CompletedAt: req.CompletedAt, ResultStatus: req.ResultStatus, Condition: req.Condition})
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// List maintenance records.
// @Summary   List maintenance records
// @Tags      maintenance
// @Produce   json
// @Security  BearerAuth
// @Param     tool_item_id  query     int   false  "Filter by item"
// @Param     open          query     bool  false  "true: in progress, false: completed"
// @Success   200           {object}  object{data=[]domain.Maintenance}
// @Router    /api/maintenance [get]
func (h *MaintenanceHandler) List(c *gin.Context) {
	itemID, ok := queryInt64(c, "tool_item_id")
	if !ok {
		return
	}
	h.list(c, itemID)
}

// History lists one item's maintenance and repair history.
// @Summary   Tool item maintenance history
// @Tags      maintenance
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      int  true  "Tool item id"
// @Success   200  {object}  object{data=[]domain.Maintenance}
// @Router    /api/tool-items/{id}/maintenance [get]
func (h *MaintenanceHandler) History(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	h.list(c, &id)
}

func (h *MaintenanceHandler) list(c *gin.Context, itemID *int64) {
	f := repository.MaintenanceFilter{ToolItemID: itemID}
	switch c.Query("open") {
	case "":
	case "true":
		t := true
		f.Open = &t
	case "false":
		f2 := false
		f.Open = &f2
	default:
		badRequest(c, "open must be true or false")
		return
	}
	out, err := h.svc.List(c.Request.Context(), f)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// Get a maintenance record.
// @Summary   Get maintenance record
// @Tags      maintenance
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      int  true  "Maintenance id"
// @Success   200  {object}  object{data=domain.Maintenance}
// @Failure   404  {object}  respond.ErrorEnvelope
// @Router    /api/maintenance/{id} [get]
func (h *MaintenanceHandler) Get(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// Due lists items whose maintenance is due by rental count, days or operating hours.
// @Summary   Maintenance due
// @Tags      maintenance
// @Produce   json
// @Security  BearerAuth
// @Param     all  query     bool  false  "true: list every item with its usage, not just the due ones"
// @Success   200  {object}  object{data=[]domain.MaintenanceDue}
// @Router    /api/maintenance/due [get]
func (h *MaintenanceHandler) Due(c *gin.Context) {
	out, err := h.svc.Due(c.Request.Context(), c.Query("all") != "true")
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}
