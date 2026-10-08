package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/luanjr10/shoptech-go-api/internal/dto"
	"github.com/luanjr10/shoptech-go-api/internal/response"
	"github.com/luanjr10/shoptech-go-api/internal/service"
)

type CatalogHandler struct{ svc *service.CatalogService }

func NewCatalogHandler(svc *service.CatalogService) *CatalogHandler {
	return &CatalogHandler{svc: svc}
}

func (h *CatalogHandler) ListCategories(c *gin.Context) {
	items, err := h.svc.ListCategories()
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", items)
}

func (h *CatalogHandler) CreateCategory(c *gin.Context) {
	var req dto.CategoryRequest
	if !bindJSON(c, &req) {
		return
	}
	cat, err := h.svc.CreateCategory(req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Tạo danh mục thành công", cat)
}

func (h *CatalogHandler) UpdateCategory(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.CategoryRequest
	if !bindJSON(c, &req) {
		return
	}
	cat, err := h.svc.UpdateCategory(id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Cập nhật danh mục thành công", cat)
}

func (h *CatalogHandler) DeleteCategory(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteCategory(id); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Đã xoá danh mục", nil)
}

func (h *CatalogHandler) ListProducts(c *gin.Context) {
	var f dto.ProductFilter
	if err := c.ShouldBindQuery(&f); err != nil {
		response.Error(c, http.StatusBadRequest, "Tham số truy vấn không hợp lệ", nil)
		return
	}
	page, err := h.svc.ListProducts(f)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", page)
}

func (h *CatalogHandler) GetProduct(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	p, err := h.svc.GetProduct(id)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "OK", p)
}

func (h *CatalogHandler) CreateProduct(c *gin.Context) {
	var req dto.ProductRequest
	if !bindJSON(c, &req) {
		return
	}
	p, err := h.svc.CreateProduct(req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusCreated, "Tạo sản phẩm thành công", p)
}

func (h *CatalogHandler) UpdateProduct(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	var req dto.ProductRequest
	if !bindJSON(c, &req) {
		return
	}
	p, err := h.svc.UpdateProduct(id, req)
	if err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Cập nhật sản phẩm thành công", p)
}

func (h *CatalogHandler) DeleteProduct(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := h.svc.DeleteProduct(id); err != nil {
		handleError(c, err)
		return
	}
	response.Success(c, http.StatusOK, "Đã xoá sản phẩm", nil)
}
