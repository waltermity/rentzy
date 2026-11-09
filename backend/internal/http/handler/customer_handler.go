package handler

import (
	"github.com/gin-gonic/gin"

	"rentzy/backend/internal/http/middleware"
	"rentzy/backend/internal/http/respond"
	"rentzy/backend/internal/service"
)

type CustomerHandler struct{ svc *service.CustomerService }

func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

// List customers.
// @Summary   List customers
// @Tags      customers
// @Produce   json
// @Security  BearerAuth
// @Param     q    query     string  false  "Search in name, email, phone"
// @Success   200  {object}  object{data=[]domain.Customer}
// @Router    /api/customers [get]
func (h *CustomerHandler) List(c *gin.Context) {
	out, err := h.svc.List(c.Request.Context(), c.Query("q"))
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// Get a customer. Customers can read only their own record.
// @Summary   Get customer
// @Tags      customers
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      int  true  "Customer id"
// @Success   200  {object}  object{data=domain.Customer}
// @Failure   404  {object}  respond.ErrorEnvelope
// @Router    /api/customers/{id} [get]
func (h *CustomerHandler) Get(c *gin.Context) {
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

// Create a customer (without a login).
// @Summary   Create customer
// @Tags      customers
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      customerRequest  true  "Customer"
// @Success   201   {object}  object{data=domain.Customer}
// @Failure   400   {object}  respond.ErrorEnvelope
// @Failure   409   {object}  respond.ErrorEnvelope  "email taken"
// @Router    /api/customers [post]
func (h *CustomerHandler) Create(c *gin.Context) {
	var req customerRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.Create(c.Request.Context(), req.input())
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, out)
}

// Update a customer.
// @Summary   Update customer
// @Tags      customers
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     id    path      int              true  "Customer id"
// @Param     body  body      customerRequest  true  "Customer"
// @Success   200   {object}  object{data=domain.Customer}
// @Failure   400   {object}  respond.ErrorEnvelope
// @Failure   404   {object}  respond.ErrorEnvelope
// @Router    /api/customers/{id} [put]
func (h *CustomerHandler) Update(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req customerRequest
	if !bindJSON(c, &req) {
		return
	}
	out, err := h.svc.Update(c.Request.Context(), id, req.input())
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}
