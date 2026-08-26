package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"sterile-packaging-release-control/internal/repository"
	"sterile-packaging-release-control/internal/service"
	"sterile-packaging-release-control/internal/util"
)

type AuditHandler struct{ service service.AuditService }

func NewAuditHandler(auditService service.AuditService) *AuditHandler {
	return &AuditHandler{service: auditService}
}

func (h *AuditHandler) List(c *gin.Context) {
	var filter repository.AuditFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		util.RespondError(c, util.BadRequest(err.Error()))
		return
	}
	result, _ := h.service.List(c.Request.Context(), filter)
	util.Respond(c, http.StatusOK, result)
}
