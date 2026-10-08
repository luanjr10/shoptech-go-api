package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/luanjr10/shoptech-go-api/internal/apperror"
	"github.com/luanjr10/shoptech-go-api/internal/response"
)

func bindJSON(c *gin.Context, dst any) bool {
	err := c.ShouldBindJSON(dst)
	if err == nil {
		return true
	}
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make(map[string]string, len(ve))
		for _, fe := range ve {
			_, field, _ := strings.Cut(fe.Namespace(), ".")
			fields[field] = validationMessage(fe)
		}
		response.Error(c, http.StatusUnprocessableEntity, "Dữ liệu không hợp lệ", fields)
		return false
	}
	response.Error(c, http.StatusBadRequest, "Body JSON không hợp lệ", nil)
	return false
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "không được để trống"
	case "email":
		return "email không hợp lệ"
	case "min":
		return "tối thiểu " + fe.Param()
	case "max":
		return "tối đa " + fe.Param()
	case "gt":
		return "phải lớn hơn " + fe.Param()
	case "gte":
		return "phải lớn hơn hoặc bằng " + fe.Param()
	default:
		return "không hợp lệ"
	}
}

func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Error(c, http.StatusBadRequest, "ID không hợp lệ", nil)
		return 0, false
	}
	return uint(id), true
}

func handleError(c *gin.Context, err error) {
	var appErr *apperror.Error
	if errors.As(err, &appErr) {
		response.Error(c, appErr.Status, appErr.Message, nil)
		return
	}
	log.Printf("lỗi không mong muốn: %v", err)
	response.Error(c, http.StatusInternalServerError, "Lỗi hệ thống, vui lòng thử lại sau", nil)
}
