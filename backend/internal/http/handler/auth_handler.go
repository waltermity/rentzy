package handler

import (
	"github.com/gin-gonic/gin"

	"rentzy/backend/internal/domain"
	"rentzy/backend/internal/http/middleware"
	"rentzy/backend/internal/http/respond"
	"rentzy/backend/internal/service"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

type loginRequest struct {
	Username string `json:"username" binding:"required" example:"kovacs.anna"`
	Password string `json:"password" binding:"required" example:"Rentzy123!"`
}

// Login exchanges username and password for a JWT.
// @Summary  Log in
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body  body      loginRequest  true  "Credentials"
// @Success  200   {object}  object{data=service.Session}
// @Failure  401   {object}  respond.ErrorEnvelope
// @Router   /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if !bindJSON(c, &req) {
		return
	}
	sess, err := h.svc.Login(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, sess)
}

type registerRequest struct {
	Username string `json:"username" binding:"required" example:"uj.ugyfel"`
	Email    string `json:"email" binding:"required" example:"uj.ugyfel@example.com"`
	Password string `json:"password" binding:"required" example:"Secret123!"`
	Name     string `json:"name" binding:"required" example:"Új Ügyfél"`
	Phone    string `json:"phone" example:"+3612345678"`
	Address  string `json:"address" example:"1111 Budapest, Fő utca 1."`
}

// Register creates a customer account (user + customer record) and logs in.
// @Summary  Register as a customer
// @Tags     auth
// @Accept   json
// @Produce  json
// @Param    body  body      registerRequest  true  "New customer"
// @Success  201   {object}  object{data=service.Session}
// @Failure  400   {object}  respond.ErrorEnvelope
// @Failure  409   {object}  respond.ErrorEnvelope  "username or email taken"
// @Router   /api/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if !bindJSON(c, &req) {
		return
	}
	sess, err := h.svc.Register(c.Request.Context(), service.RegisterInput{
		Username: req.Username, Email: req.Email, Password: req.Password,
		Name: req.Name, Phone: req.Phone, Address: req.Address,
	})
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.Created(c, sess)
}

type meResponse struct {
	User     domain.User      `json:"user"`
	Customer *domain.Customer `json:"customer,omitempty"`
}

// Me returns the logged-in user and, for customers, their customer record.
// @Summary   Current user
// @Tags      auth
// @Produce   json
// @Security  BearerAuth
// @Success   200  {object}  object{data=meResponse}
// @Failure   401  {object}  respond.ErrorEnvelope
// @Router    /api/auth/me [get]
func (h *AuthHandler) Me(c *gin.Context) {
	u, cust, err := h.svc.Me(c.Request.Context(), middleware.PrincipalFrom(c))
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, meResponse{User: u, Customer: cust})
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePassword changes the logged-in user's password.
// @Summary   Change own password
// @Tags      auth
// @Accept    json
// @Produce   json
// @Security  BearerAuth
// @Param     body  body      changePasswordRequest  true  "Old and new password"
// @Success   200   {object}  object{data=object{status=string}}
// @Failure   400   {object}  respond.ErrorEnvelope
// @Failure   401   {object}  respond.ErrorEnvelope
// @Router    /api/auth/password [post]
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if !bindJSON(c, &req) {
		return
	}
	if err := h.svc.ChangePassword(c.Request.Context(), middleware.PrincipalFrom(c), req.OldPassword, req.NewPassword); err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, gin.H{"status": "ok"})
}
