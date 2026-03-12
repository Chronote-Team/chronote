package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"chronote/services"
)

var healthService = services.HealthService{}

// HealthCheck is a lightweight probe for Docker HEALTHCHECK.
func HealthCheck(ctx *gin.Context) {
	status := healthService.Check(ctx.Request.Context())
	if status.Healthy {
		ctx.JSON(http.StatusOK, gin.H{
			"code":    http.StatusOK,
			"message": "OK",
		})
		return
	}

	ctx.JSON(http.StatusServiceUnavailable, gin.H{
		"code":    http.StatusServiceUnavailable,
		"message": "Service Unavailable",
	})
}

// HealthDetails returns component-level diagnosis for operations and external tests.
func HealthDetails(ctx *gin.Context) {
	status := healthService.Check(ctx.Request.Context())

	httpStatus := http.StatusOK
	message := "All systems operational"

	if !status.Healthy {
		databaseOk := status.Components["database"].Status == "ok"
		redisOk := status.Components["redis"].Status == "ok"

		if databaseOk || redisOk {
			httpStatus = http.StatusMultiStatus
			message = "Some services degraded"
		} else {
			httpStatus = http.StatusServiceUnavailable
			message = "Service Unavailable"
		}
	}

	ctx.JSON(httpStatus, gin.H{
		"code":    httpStatus,
		"message": message,
		"data":    status,
	})
}
