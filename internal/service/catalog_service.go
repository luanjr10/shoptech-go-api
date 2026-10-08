package service

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/luanjr10/shoptech-go-api/internal/apperror"
	"github.com/luanjr10/shoptech-go-api/internal/dto"
	"github.com/luanjr10/shoptech-go-api/internal/model"
	"github.com/luanjr10/shoptech-go-api/internal/repository"
)

type CatalogService struct {
	categories *repository.CategoryRepository
	products   *repository.ProductRepository
}

func NewCatalogService(db *gorm.DB) *CatalogService {
	return &CatalogService{
		categories: repository.NewCategoryRepository(db),
		products:   repository.NewProductRepository(db),
	}
}

func (s *CatalogService) ListCategories() ([]model.Category, error) {
	return s.categories.List()
}

func (s *CatalogService) CreateCategory(req dto.CategoryRequest) (*model.Category, error) {
	c := &model.Category{}
	return c, s.saveCategory(c, req)
}

func (s *CatalogService) UpdateCategory(id uint, req dto.CategoryRequest) (*model.Category, error) {
	c, err := s.categories.FindByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("Không tìm thấy danh mục")
	}
	if err != nil {
		return nil, err
	}
	return c, s.saveCategory(c, req)
}

func (s *CatalogService) saveCategory(c *model.Category, req dto.CategoryRequest) error {
	name := strings.TrimSpace(req.Name)
	exists, err := s.categories.ExistsByName(name, c.ID)
	if err != nil {
		return err
	}
	if exists {
		return apperror.Conflict("Tên danh mục đã tồn tại")
	}
	c.Name = name
	return s.categories.Save(c)
}

func (s *CatalogService) DeleteCategory(id uint) error {
	if _, err := s.categories.FindByID(id); errors.Is(err, repository.ErrNotFound) {
		return apperror.NotFound("Không tìm thấy danh mục")
	} else if err != nil {
		return err
	}
	count, err := s.products.CountByCategory(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return apperror.Conflict("Danh mục vẫn còn sản phẩm, không thể xoá")
	}
	return s.categories.Delete(id)
}

func (s *CatalogService) ListProducts(f dto.ProductFilter) (*dto.Paginated[model.Product], error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.Limit < 1 || f.Limit > 100 {
		f.Limit = 10
	}
	items, total, err := s.products.List(f)
	if err != nil {
		return nil, err
	}
	return &dto.Paginated[model.Product]{Items: items, Page: f.Page, Limit: f.Limit, Total: total}, nil
}

func (s *CatalogService) GetProduct(id uint) (*model.Product, error) {
	p, err := s.products.FindByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("Không tìm thấy sản phẩm")
	}
	return p, err
}

func (s *CatalogService) CreateProduct(req dto.ProductRequest) (*model.Product, error) {
	p := &model.Product{}
	if err := s.saveProduct(p, req); err != nil {
		return nil, err
	}
	return s.products.FindByID(p.ID)
}

func (s *CatalogService) UpdateProduct(id uint, req dto.ProductRequest) (*model.Product, error) {
	p, err := s.GetProduct(id)
	if err != nil {
		return nil, err
	}
	p.Category = nil
	if err := s.saveProduct(p, req); err != nil {
		return nil, err
	}
	return s.products.FindByID(p.ID)
}

func (s *CatalogService) saveProduct(p *model.Product, req dto.ProductRequest) error {
	if _, err := s.categories.FindByID(req.CategoryID); errors.Is(err, repository.ErrNotFound) {
		return apperror.Unprocessable("Danh mục không tồn tại")
	} else if err != nil {
		return err
	}
	p.CategoryID = req.CategoryID
	p.Name = strings.TrimSpace(req.Name)
	p.Description = req.Description
	p.Price = req.Price
	p.Stock = *req.Stock
	return s.products.Save(p)
}

func (s *CatalogService) DeleteProduct(id uint) error {
	if _, err := s.GetProduct(id); err != nil {
		return err
	}
	return s.products.Delete(id)
}
