package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/luanjr10/shoptech-go-api/internal/dto"
	"github.com/luanjr10/shoptech-go-api/internal/middleware"
	"github.com/luanjr10/shoptech-go-api/internal/model"
	"github.com/luanjr10/shoptech-go-api/internal/response"
	"github.com/luanjr10/shoptech-go-api/internal/service"
)

type OrderHandler struct{ svc *service.OrderService }

func NewOrderHandler(svc *service.OrderService) *OrderHandler { return &OrderHandler{svc: svc} }

func (h *OrderHandler) Create(c *gin.Context) {
	var req dto.CreateOrderRequest
	if !bindJSON(c, &req) {
		return
	}
	order, err := h.svc.Create(middleware.UserID(c), req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Đặt hàng thành công", order)
}

func (h *OrderHandler) ListMine(c *gin.Context) {
	orders, err := h.svc.ListMine(middleware.UserID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", orders)
}

func (h *OrderHandler) Get(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	order, err := h.svc.Get(id, middleware.UserID(c), middleware.Role(c) == model.RoleAdmin)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", order)
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	order, err := h.svc.Cancel(id, middleware.UserID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Đã huỷ đơn hàng", order)
}
