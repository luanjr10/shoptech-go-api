package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/luanjr10/shoptech-go-api/internal/dto"
	"github.com/luanjr10/shoptech-go-api/internal/middleware"
	"github.com/luanjr10/shoptech-go-api/internal/response"
	"github.com/luanjr10/shoptech-go-api/internal/service"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.svc.Register(req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Đăng ký thành công", res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.svc.Login(req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Đăng nhập thành công", res)
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, err := h.svc.Me(middleware.UserID(c))
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", user)
}
