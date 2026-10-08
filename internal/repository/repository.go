package repository

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/luanjr10/shoptech-go-api/internal/dto"
	"github.com/luanjr10/shoptech-go-api/internal/model"
)

var ErrNotFound = errors.New("record not found")

func wrap(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

type UserRepository struct{ db *gorm.DB }

func NewUserRepository(db *gorm.DB) *UserRepository { return &UserRepository{db: db} }

func (r *UserRepository) Create(u *model.User) error { return r.db.Create(u).Error }

func (r *UserRepository) FindByEmail(email string) (*model.User, error) {
	var u model.User
	if err := r.db.Where("email = ?", email).First(&u).Error; err != nil {
		return nil, wrap(err)
	}
	return &u, nil
}

func (r *UserRepository) FindByID(id uint) (*model.User, error) {
	var u model.User
	if err := r.db.First(&u, id).Error; err != nil {
		return nil, wrap(err)
	}
	return &u, nil
}

type CategoryRepository struct{ db *gorm.DB }

func NewCategoryRepository(db *gorm.DB) *CategoryRepository { return &CategoryRepository{db: db} }

func (r *CategoryRepository) List() ([]model.Category, error) {
	var items []model.Category
	err := r.db.Order("name").Find(&items).Error
	return items, err
}

func (r *CategoryRepository) FindByID(id uint) (*model.Category, error) {
	var c model.Category
	if err := r.db.First(&c, id).Error; err != nil {
		return nil, wrap(err)
	}
	return &c, nil
}

func (r *CategoryRepository) ExistsByName(name string, exceptID uint) (bool, error) {
	var count int64
	err := r.db.Model(&model.Category{}).Where("name = ? AND id <> ?", name, exceptID).Count(&count).Error
	return count > 0, err
}

func (r *CategoryRepository) Save(c *model.Category) error { return r.db.Save(c).Error }

func (r *CategoryRepository) Delete(id uint) error { return r.db.Delete(&model.Category{}, id).Error }

type ProductRepository struct{ db *gorm.DB }

func NewProductRepository(db *gorm.DB) *ProductRepository { return &ProductRepository{db: db} }

func (r *ProductRepository) List(f dto.ProductFilter) ([]model.Product, int64, error) {
	filtered := func() *gorm.DB {
		q := r.db.Model(&model.Product{})
		if f.Q != "" {
			q = q.Where("name LIKE ?", "%"+f.Q+"%")
		}
		if f.CategoryID > 0 {
			q = q.Where("category_id = ?", f.CategoryID)
		}
		return q
	}

	var total int64
	if err := filtered().Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Product
	err := filtered().Preload("Category").
		Order("id DESC").
		Limit(f.Limit).Offset((f.Page - 1) * f.Limit).
		Find(&items).Error
	return items, total, err
}

func (r *ProductRepository) FindByID(id uint) (*model.Product, error) {
	var p model.Product
	if err := r.db.Preload("Category").First(&p, id).Error; err != nil {
		return nil, wrap(err)
	}
	return &p, nil
}

func (r *ProductRepository) FindForUpdate(id uint) (*model.Product, error) {
	var p model.Product
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&p, id).Error; err != nil {
		return nil, wrap(err)
	}
	return &p, nil
}

func (r *ProductRepository) AddStock(id uint, delta int) error {
	return r.db.Model(&model.Product{}).Where("id = ?", id).
		UpdateColumn("stock", gorm.Expr("stock + ?", delta)).Error
}

func (r *ProductRepository) CountByCategory(categoryID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.Product{}).Where("category_id = ?", categoryID).Count(&count).Error
	return count, err
}

func (r *ProductRepository) Save(p *model.Product) error { return r.db.Save(p).Error }

func (r *ProductRepository) Delete(id uint) error { return r.db.Delete(&model.Product{}, id).Error }

type OrderRepository struct{ db *gorm.DB }

func NewOrderRepository(db *gorm.DB) *OrderRepository { return &OrderRepository{db: db} }

func (r *OrderRepository) Create(o *model.Order) error { return r.db.Create(o).Error }

func (r *OrderRepository) ListByUser(userID uint) ([]model.Order, error) {
	var items []model.Order
	err := r.db.Preload("Items").Where("user_id = ?", userID).Order("id DESC").Find(&items).Error
	return items, err
}

func (r *OrderRepository) FindByID(id uint) (*model.Order, error) {
	var o model.Order
	if err := r.db.Preload("Items").First(&o, id).Error; err != nil {
		return nil, wrap(err)
	}
	return &o, nil
}

func (r *OrderRepository) FindForUpdate(id uint) (*model.Order, error) {
	var o model.Order
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(&o, id).Error; err != nil {
		return nil, wrap(err)
	}
	return &o, nil
}

func (r *OrderRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&model.Order{}).Where("id = ?", id).Update("status", status).Error
}
