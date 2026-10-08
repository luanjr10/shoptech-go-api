package response

import "github.com/gin-gonic/gin"

type Body struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, Body{Success: true, Message: message, Data: data})
}

func Error(c *gin.Context, status int, message string, errs any) {
	c.AbortWithStatusJSON(status, Body{Success: false, Message: message, Errors: errs})
}
