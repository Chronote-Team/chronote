package response

import (
	"encoding/json"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func JSON(code int, message string, data interface{}) ([]byte, error) {
	return json.Marshal(Envelope{
		Code:    code,
		Message: message,
		Data:    data,
	})
}

func Write(ctx *gin.Context, code int, message string, data interface{}) {
	ctx.JSON(code, Envelope{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
