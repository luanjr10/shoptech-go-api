package router

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	"github.com/luanjr10/shoptech-go-api/internal/config"
	"github.com/luanjr10/shoptech-go-api/internal/handler"
	"github.com/luanjr10/shoptech-go-api/internal/middleware"
	"github.com/luanjr10/shoptech-go-api/internal/model"
	"github.com/luanjr10/shoptech-go-api/internal/response"
	"github.com/luanjr10/shoptech-go-api/internal/service"
)

func New(db *gorm.DB, cfg config.Config) *gin.Engine {
	useJSONFieldNames()

	authH := handler.NewAuthHandler(service.NewAuthService(db, cfg))
	catalogH := handler.NewCatalogHandler(service.NewCatalogService(db))
	orderH := handler.NewOrderHandler(service.NewOrderService(db))

	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "ShopTech Go API đang chạy", gin.H{
			"health": "/health",
			"docs":   "Xem README.md và api.http để biết đầy đủ các endpoint",
			"endpoints": []string{
				"POST /api/v1/auth/register",
				"POST /api/v1/auth/login",
				"GET  /api/v1/auth/me",
				"GET  /api/v1/categories",
				"GET  /api/v1/products?q=&category_id=&page=&limit=",
				"GET  /api/v1/products/:id",
				"POST /api/v1/orders",
				"GET  /api/v1/orders",
				"POST /api/v1/orders/:id/cancel",
			},
		})
	})

	r.GET("/health", func(c *gin.Context) {
		response.Success(c, http.StatusOK, "OK", gin.H{"status": "up"})
	})

	auth := middleware.Auth(cfg.JWTSecret)
	adminOnly := middleware.RequireRole(model.RoleAdmin)

	v1 := r.Group("/api/v1")
	{
		v1.POST("/auth/register", authH.Register)
		v1.POST("/auth/login", authH.Login)
		v1.GET("/auth/me", auth, authH.Me)

		v1.GET("/categories", catalogH.ListCategories)
		v1.GET("/products", catalogH.ListProducts)
		v1.GET("/products/:id", catalogH.GetProduct)

		admin := v1.Group("", auth, adminOnly)
		admin.POST("/categories", catalogH.CreateCategory)
		admin.PUT("/categories/:id", catalogH.UpdateCategory)
		admin.DELETE("/categories/:id", catalogH.DeleteCategory)
		admin.POST("/products", catalogH.CreateProduct)
		admin.PUT("/products/:id", catalogH.UpdateProduct)
		admin.DELETE("/products/:id", catalogH.DeleteProduct)

		orders := v1.Group("/orders", auth)
		orders.POST("", orderH.Create)
		orders.GET("", orderH.ListMine)
		orders.GET("/:id", orderH.Get)
		orders.POST("/:id/cancel", orderH.Cancel)
	}

	r.NoRoute(func(c *gin.Context) {
		response.Error(c, http.StatusNotFound, "Không tìm thấy endpoint", nil)
	})
	return r
}

func useJSONFieldNames() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}
}
