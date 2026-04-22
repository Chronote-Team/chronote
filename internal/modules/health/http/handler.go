package http

import (
	"net/http"

	healthapp "chronote-refactor/internal/modules/health/app"
	"chronote-refactor/internal/shared/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *healthapp.Service
}

func NewHandler(service *healthapp.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) HealthCheck(ctx *gin.Context) {
	status := h.service.Check(ctx.Request.Context())
	if status.Healthy {
		response.Write(ctx, http.StatusOK, "OK", nil)
		return
	}
	response.Write(ctx, http.StatusServiceUnavailable, "Service Unavailable", nil)
}

func (h *Handler) HealthDetails(ctx *gin.Context) {
	status := h.service.Check(ctx.Request.Context())
	httpStatus := http.StatusOK
	message := "All systems operational"

	if !status.Healthy {
		databaseOK := status.Components["database"].Status == "ok"
		redisOK := status.Components["redis"].Status == "ok"
		if databaseOK || redisOK {
			httpStatus = http.StatusMultiStatus
			message = "Some services degraded"
		} else {
			httpStatus = http.StatusServiceUnavailable
			message = "Service Unavailable"
		}
	}

	response.Write(ctx, httpStatus, message, status)
}
