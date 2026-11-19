package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"rentzy/backend/internal/http/respond"
	"rentzy/backend/internal/service"
)

type ReportHandler struct{ svc *service.ReportService }

func NewReportHandler(svc *service.ReportService) *ReportHandler { return &ReportHandler{svc: svc} }

func (h *ReportHandler) window(c *gin.Context) (service.Window, bool) {
	from, ok := queryTime(c, "from")
	if !ok {
		return service.Window{}, false
	}
	to, ok := queryTime(c, "to")
	if !ok {
		return service.Window{}, false
	}
	w, err := h.svc.ResolveWindow(from, to)
	if err != nil {
		respond.Err(c, err)
		return w, false
	}
	return w, true
}

// Popular lists the most rented tool types.
// @Summary   Most rented tools
// @Tags      reports
// @Produce   json
// @Security  BearerAuth
// @Param     from   query     string  false  "Default: 90 days before to"
// @Param     to     query     string  false  "Default: now"
// @Param     limit  query     int     false  "Default: 10"
// @Success   200    {object}  object{data=service.PopularReport}
// @Router    /api/reports/popular [get]
func (h *ReportHandler) Popular(c *gin.Context) {
	w, ok := h.window(c)
	if !ok {
		return
	}
	limit := 10
	if v := c.Query("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			badRequest(c, "limit must be a positive integer")
			return
		}
		limit = n
	}
	out, err := h.svc.Popular(c.Request.Context(), w, limit)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// Utilization gives utilization and revenue per tool type and item.
// @Summary   Utilization and revenue
// @Tags      reports
// @Produce   json
// @Security  BearerAuth
// @Param     from  query     string  false  "Default: 90 days before to"
// @Param     to    query     string  false  "Default: now"
// @Success   200   {object}  object{data=service.UtilizationReport}
// @Router    /api/reports/utilization [get]
func (h *ReportHandler) Utilization(c *gin.Context) {
	w, ok := h.window(c)
	if !ok {
		return
	}
	out, err := h.svc.Utilization(c.Request.Context(), w)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}

// Procurement suggests tool types to buy more items of.
// @Summary      Procurement suggestions
// @Description  Suggests more items for every type whose non-retired items' utilization is at or above the threshold, with the number needed to get below it.
// @Tags         reports
// @Produce      json
// @Security     BearerAuth
// @Param        from       query     string  false  "Default: 90 days before to"
// @Param        to         query     string  false  "Default: now"
// @Param        threshold  query     number  false  "Utilization threshold in (0, 1], default 0.7"
// @Success      200        {object}  object{data=service.ProcurementReport}
// @Router       /api/reports/procurement [get]
func (h *ReportHandler) Procurement(c *gin.Context) {
	w, ok := h.window(c)
	if !ok {
		return
	}
	th, ok := queryFloat(c, "threshold", service.DefaultProcurementThreshold)
	if !ok {
		return
	}
	out, err := h.svc.Procurement(c.Request.Context(), w, th)
	if err != nil {
		respond.Err(c, err)
		return
	}
	respond.OK(c, out)
}
