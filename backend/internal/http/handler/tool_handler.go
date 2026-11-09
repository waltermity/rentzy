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

type ToolHandler struct{ svc *service.ToolService }

func NewToolHandler(svc *service.ToolService) *ToolHandler { return &ToolHandler{svc: svc} }

// ---------- categories ----------

type categoryRequest struct {
	Name string `json:"name" binding:"required" example:"Drilling & Demolition"`
}

// ListCategories lists tool categories.
// @Summary   List categories
// @Tags      tools
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  object{data=[]domain.Category}
// @Router    /api/categories [get]
func (h *ToolHandler) ListCategories(c *gin.Context) {
	out, err := h.svc.ListCategories(c.Request.Context())
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// CreateCategory adds a category.
// @Summary   Create category
// @Tags      tools
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      categoryRequest  true  "Category"
// @Success   201   {object}  object{data=domain.Category}
// @Failure   409   {object}  respond.ErrorEnvelope  "name taken"
// @Router    /api/categories [post]
func (h *ToolHandler) CreateCategory(c *gin.Context) {
	var req categoryRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.CreateCategory(c.Request.Context(), req.Name)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, out)
}

// UpdateCategory renames a category.
// @Summary   Update category
// @Tags      tools
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     id    path      int              true  "Category id"
// @Param     body  body      categoryRequest  true  "Category"
// @Success   200   {object}  object{data=domain.Category}
// @Failure   404   {object}  respond.ErrorEnvelope
// @Router    /api/categories/{id} [put]
func (h *ToolHandler) UpdateCategory(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req categoryRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.UpdateCategory(c.Request.Context(), id, req.Name)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// ---------- tool types ----------

type toolTypeRequest struct {
	Name                    string          `json:"name" binding:"required" example:"Rotary hammer (SDS-plus)"`
	CategoryID              int64           `json:"category_id" binding:"required" example:"1"`
	Manufacturer            string          `json:"manufacturer" example:"Bosch"`
	Model                   string          `json:"model" example:"GBH 2-28 F"`
	PurchasePrice           decimal.Decimal `json:"purchase_price" swaggertype:"number" example:"120000"`
	DailyRate               decimal.Decimal `json:"daily_rate" swaggertype:"number" example:"4500"`
	Deposit                 decimal.Decimal `json:"deposit" swaggertype:"number" example:"20000"`
	MaintenanceEveryRentals *int            `json:"maintenance_every_rentals" example:"20"`
	MaintenanceEveryDays    *int            `json:"maintenance_every_days" example:"180"`
	MaintenanceEveryHours   *float64        `json:"maintenance_every_hours" example:"50"`
}

func (r toolTypeRequest) input() service.ToolTypeInput {
	return service.ToolTypeInput{
		Name: r.Name, CategoryID: r.CategoryID, Manufacturer: r.Manufacturer, Model: r.Model,
		PurchasePrice: r.PurchasePrice, DailyRate: r.DailyRate, Deposit: r.Deposit,
		MaintenanceEveryRentals: r.MaintenanceEveryRentals,
		MaintenanceEveryDays:    r.MaintenanceEveryDays,
		MaintenanceEveryHours:   r.MaintenanceEveryHours,
	}
}

// ListTypes lists tool types (the catalog) with item counts per status.
// @Summary   List tool types
// @Tags      tools
// @Produce   json
// @Security  BearerAuth
// @Param     category_id  query     int     false  "Filter by category"
// @Param     q            query     string  false  "Search in name, manufacturer, model"
// @Success   200          {object}  object{data=[]domain.ToolType}
// @Router    /api/tool-types [get]
func (h *ToolHandler) ListTypes(c *gin.Context) {
	cat, ok := queryInt64(c, "category_id")
	if !ok {
		return
	}
	out, err := h.svc.ListTypes(c.Request.Context(), repository.ToolTypeFilter{CategoryID: cat, Search: c.Query("q")})
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// GetType returns one tool type.
// @Summary   Get tool type
// @Tags      tools
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      int  true  "Tool type id"
// @Success   200  {object}  object{data=domain.ToolType}
// @Failure   404  {object}  respond.ErrorEnvelope
// @Router    /api/tool-types/{id} [get]
func (h *ToolHandler) GetType(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.GetType(c.Request.Context(), id)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// CreateType adds a tool type.
// @Summary   Create tool type
// @Tags      tools
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      toolTypeRequest  true  "Tool type"
// @Success   201   {object}  object{data=domain.ToolType}
// @Failure   400   {object}  respond.ErrorEnvelope
// @Router    /api/tool-types [post]
func (h *ToolHandler) CreateType(c *gin.Context) {
	var req toolTypeRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.CreateType(c.Request.Context(), req.input())
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, out)
}

// UpdateType replaces a tool type's data.
// @Summary   Update tool type
// @Tags      tools
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     id    path      int              true  "Tool type id"
// @Param     body  body      toolTypeRequest  true  "Tool type"
// @Success   200   {object}  object{data=domain.ToolType}
// @Failure   400   {object}  respond.ErrorEnvelope
// @Failure   404   {object}  respond.ErrorEnvelope
// @Router    /api/tool-types/{id} [put]
func (h *ToolHandler) UpdateType(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req toolTypeRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.UpdateType(c.Request.Context(), id, req.input())
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// Availability shows which items of a type are busy in a period.
// @Summary      Tool type availability
// @Description  Busy intervals (open reservations and rentals) of the type's items in [from, to), the items that can be booked at all, and those free for the whole period.
// @Tags         tools
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int     true  "Tool type id"
// @Param        from  query     string  true  "RFC 3339 or YYYY-MM-DD"
// @Param        to    query     string  true  "RFC 3339 or YYYY-MM-DD"
// @Success      200   {object}  object{data=service.Availability}
// @Failure      400   {object}  respond.ErrorEnvelope
// @Router       /api/tool-types/{id}/availability [get]
func (h *ToolHandler) Availability(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	from, ok := queryTime(c, "from")
	if !ok {
		return
	}
	to, ok := queryTime(c, "to")
	if !ok {
		return
	}
	if from == nil || to == nil {
		badRequest(c, "from and to are required")
		return
	}
	out, err := h.svc.Availability(c.Request.Context(), id, *from, *to)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// ---------- tool items ----------

type addItemsRequest struct {
	Count      int              `json:"count" example:"1"`
	Condition  domain.Condition `json:"condition" enums:"NEW,GOOD,FAIR,POOR" example:"NEW"`
	AcquiredAt *time.Time       `json:"acquired_at"`
}

// AddItems adds new physical copies of a tool type, each with its own id.
// @Summary   Add tool items
// @Tags      tools
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     id    path      int              true  "Tool type id"
// @Param     body  body      addItemsRequest  true  "count defaults to 1, condition to NEW, acquired_at to now"
// @Success   201   {object}  object{data=[]domain.ToolItem}
// @Failure   400   {object}  respond.ErrorEnvelope
// @Failure   404   {object}  respond.ErrorEnvelope
// @Router    /api/tool-types/{id}/items [post]
func (h *ToolHandler) AddItems(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req addItemsRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.AddItems(c.Request.Context(), id, service.AddItemsInput{
		Count: req.Count, Condition: req.Condition, AcquiredAt: req.AcquiredAt})
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, out)
}

// ListItems lists tool items.
// @Summary   List tool items
// @Tags      tools
// @Produce   json
// @Security  BearerAuth
// @Param     tool_type_id  query     int     false  "Filter by tool type"
// @Param     status        query     string  false  "AVAILABLE, RENTED, MAINTENANCE, BROKEN or RETIRED"
// @Success   200           {object}  object{data=[]domain.ToolItem}
// @Router    /api/tool-items [get]
func (h *ToolHandler) ListItems(c *gin.Context) {
	typeID, ok := queryInt64(c, "tool_type_id")
	if !ok {
		return
	}
	f := repository.ToolItemFilter{ToolTypeID: typeID}
	if v := c.Query("status"); v != "" {
		st := domain.ToolStatus(v)
		f.Status = &st
	}
	out, err := h.svc.ListItems(c.Request.Context(), f)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// GetItem returns an item with its maintenance due status.
// @Summary   Get tool item
// @Tags      tools
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      int  true  "Tool item id"
// @Success   200  {object}  object{data=service.ToolItemDetail}
// @Failure   404  {object}  respond.ErrorEnvelope
// @Router    /api/tool-items/{id} [get]
func (h *ToolHandler) GetItem(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	out, err := h.svc.GetItem(c.Request.Context(), id)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

type updateItemRequest struct {
	Status    *domain.ToolStatus `json:"status" enums:"AVAILABLE,BROKEN,RETIRED"`
	Condition *domain.Condition  `json:"condition" enums:"NEW,GOOD,FAIR,POOR,DAMAGED"`
}

// UpdateItem sets an item's status (AVAILABLE, BROKEN, RETIRED) or condition.
// @Summary      Update tool item
// @Description  RENTED is set by handover, MAINTENANCE by starting maintenance. Rented items cannot be changed; items in maintenance can only be retired.
// @Tags         tools
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id    path      int                true  "Tool item id"
// @Param        body  body      updateItemRequest  true  "Fields to change"
// @Success      200   {object}  object{data=domain.ToolItem}
// @Failure      400   {object}  respond.ErrorEnvelope
// @Failure      409   {object}  respond.ErrorEnvelope
// @Router       /api/tool-items/{id} [patch]
func (h *ToolHandler) UpdateItem(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req updateItemRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.UpdateItem(c.Request.Context(), id, service.UpdateItemInput{Status: req.Status, Condition: req.Condition})
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}
