package model

import "time"

const (
	RoleCustomer = "customer"
	RoleAdmin    = "admin"

	OrderPending   = "pending"
	OrderCancelled = "cancelled"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`
	Email     string    `gorm:"size:191;uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Role      string    `gorm:"size:20;not null;default:customer" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Category struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"size:100;uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	CategoryID  uint      `gorm:"index;not null" json:"category_id"`
	Category    *Category `json:"category,omitempty"`
	Name        string    `gorm:"size:200;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       int64     `gorm:"not null" json:"price"`
	Stock       int       `gorm:"not null;default:0" json:"stock"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Order struct {
	ID        uint        `gorm:"primaryKey" json:"id"`
	UserID    uint        `gorm:"index;not null" json:"user_id"`
	Status    string      `gorm:"size:20;not null;default:pending" json:"status"`
	Total     int64       `gorm:"not null" json:"total"`
	Items     []OrderItem `json:"items"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	OrderID     uint   `gorm:"index;not null" json:"order_id"`
	ProductID   uint   `gorm:"index;not null" json:"product_id"`
	ProductName string `gorm:"size:200;not null" json:"product_name"`
	Price       int64  `gorm:"not null" json:"price"`
	Quantity    int    `gorm:"not null" json:"quantity"`
}
