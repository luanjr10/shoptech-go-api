package dto

type RegisterRequest struct {
	Name     string `json:"name" binding:"required,max=100"`
	Email    string `json:"email" binding:"required,email,max=191"`
	Password string `json:"password" binding:"required,min=6,max=72"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type CategoryRequest struct {
	Name string `json:"name" binding:"required,max=100"`
}

type ProductRequest struct {
	CategoryID  uint   `json:"category_id" binding:"required"`
	Name        string `json:"name" binding:"required,max=200"`
	Description string `json:"description"`
	Price       int64  `json:"price" binding:"required,gt=0"`
	Stock       *int   `json:"stock" binding:"required,gte=0"`
}

type ProductFilter struct {
	Q          string `form:"q"`
	CategoryID uint   `form:"category_id"`
	Page       int    `form:"page"`
	Limit      int    `form:"limit"`
}

type OrderItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
}

type Paginated[T any] struct {
	Items []T   `json:"items"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}
