package handler

import (
	"github.com/gin-gonic/gin"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/http/respond"
	"rentzy/backend/internal/service"
)

type UserHandler struct{ svc *service.UserService }

func NewUserHandler(svc *service.UserService) *UserHandler { return &UserHandler{svc: svc} }

// List users, optionally by role.
// @Summary   List users
// @Tags      users
// @Produce   json
// @Security  BearerAuth
// @Param     role  query     string  false  "ADMIN, STAFF or CUSTOMER"
// @Success   200   {object}  object{data=[]domain.User}
// @Router    /api/users [get]
func (h *UserHandler) List(c *gin.Context) {
	var role *domain.Role
	if v := c.Query("role"); v != "" {
		r := domain.Role(v)
		if !r.Valid() {
			badRequest(c, "role must be ADMIN, STAFF or CUSTOMER")
			return
		}
		role = &r
	}
	users, err := h.svc.List(c.Request.Context(), role)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, users)
}

// Get a user.
// @Summary   Get user
// @Tags      users
// @Produce   json
// @Security  BearerAuth
// @Param     id   path      int  true  "User id"
// @Success   200  {object}  object{data=domain.User}
// @Failure   404  {object}  respond.ErrorEnvelope
// @Router    /api/users/{id} [get]
func (h *UserHandler) Get(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	u, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, u)
}

type customerRequest struct {
	Name    string `json:"name" binding:"required" example:"Kovács Anna"`
	Email   string `json:"email" binding:"required" example:"kovacs.anna@example.com"`
	Phone   string `json:"phone" example:"+3612345678"`
	Address string `json:"address" example:"1052 Budapest, Petőfi Sándor utca 8."`
}

func (r customerRequest) input() service.CustomerInput {
	return service.CustomerInput{Name: r.Name, Email: r.Email, Phone: r.Phone, Address: r.Address}
}

type createUserRequest struct {
	Username string      `json:"username" binding:"required"`
	Email    string      `json:"email" binding:"required"`
	Password string      `json:"password" binding:"required"`
	Role     domain.Role `json:"role" binding:"required" enums:"ADMIN,STAFF,CUSTOMER"`
	// CUSTOMER only: link an existing customer...
	CustomerID *int64 `json:"customer_id"`
	// ...or create a new one.
	Customer *customerRequest `json:"customer"`
}

// Create a user. CUSTOMER users need customer_id or customer.
// @Summary   Create user
// @Tags      users
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      createUserRequest  true  "New user"
// @Success   201   {object}  object{data=domain.User}
// @Failure   400   {object}  respond.ErrorEnvelope
// @Failure   409   {object}  respond.ErrorEnvelope
// @Router    /api/users [post]
func (h *UserHandler) Create(c *gin.Context) {
	var req createUserRequest
	if !bindJSON(c, &req) {
		return
	}
	in := service.CreateUserInput{Username: req.Username, Email: req.Email, Password: req.Password,
		Role: req.Role, CustomerID: req.CustomerID}
	if req.Customer != nil {
		ci := req.Customer.input()
		in.Customer = &ci
	}
	u, err := h.svc.Create(c.Request.Context(), in)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, u)
}

type updateUserRequest struct {
	Email    *string      `json:"email"`
	Role     *domain.Role `json:"role" enums:"ADMIN,STAFF"`
	Password *string      `json:"password"`
}

// Update a user's email, password, or role (ADMIN <-> STAFF only).
// @Summary   Update user
// @Tags      users
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     id    path      int                true  "User id"
// @Param     body  body      updateUserRequest  true  "Fields to change"
// @Success   200   {object}  object{data=domain.User}
// @Failure   400   {object}  respond.ErrorEnvelope
// @Failure   404   {object}  respond.ErrorEnvelope
// @Router    /api/users/{id} [patch]
func (h *UserHandler) Update(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req updateUserRequest
	if !bindJSON(c, &req) {
		return
	}
	u, err := h.svc.Update(c.Request.Context(), id, service.UpdateUserInput{
		Email: req.Email, Role: req.Role, Password: req.Password})
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, u)
}
