package service

import (
	"errors"
	"fmt"
	"sort"

	"gorm.io/gorm"

	"github.com/luanjr10/shoptech-go-api/internal/apperror"
	"github.com/luanjr10/shoptech-go-api/internal/dto"
	"github.com/luanjr10/shoptech-go-api/internal/model"
	"github.com/luanjr10/shoptech-go-api/internal/repository"
)

type OrderService struct {
	db *gorm.DB
}

func NewOrderService(db *gorm.DB) *OrderService { return &OrderService{db: db} }

func (s *OrderService) Create(userID uint, req dto.CreateOrderRequest) (*model.Order, error) {
	qty := map[uint]int{}
	for _, it := range req.Items {
		qty[it.ProductID] += it.Quantity
	}
	ids := make([]uint, 0, len(qty))
	for id := range qty {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	order := &model.Order{UserID: userID, Status: model.OrderPending}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		products := repository.NewProductRepository(tx)

		for _, id := range ids {
			p, err := products.FindForUpdate(id)
			if errors.Is(err, repository.ErrNotFound) {
				return apperror.NotFound(fmt.Sprintf("Sản phẩm #%d không tồn tại", id))
			}
			if err != nil {
				return err
			}
			if p.Stock < qty[id] {
				return apperror.Conflict(fmt.Sprintf("Sản phẩm \"%s\" chỉ còn %d trong kho", p.Name, p.Stock))
			}
			if err := products.AddStock(id, -qty[id]); err != nil {
				return err
			}
			order.Items = append(order.Items, model.OrderItem{
				ProductID:   p.ID,
				ProductName: p.Name,
				Price:       p.Price,
				Quantity:    qty[id],
			})
			order.Total += p.Price * int64(qty[id])
		}

		return repository.NewOrderRepository(tx).Create(order)
	})
	if err != nil {
		return nil, err
	}
	return order, nil
}

func (s *OrderService) ListMine(userID uint) ([]model.Order, error) {
	return repository.NewOrderRepository(s.db).ListByUser(userID)
}

func (s *OrderService) Get(id, userID uint, isAdmin bool) (*model.Order, error) {
	o, err := repository.NewOrderRepository(s.db).FindByID(id)
	if errors.Is(err, repository.ErrNotFound) || (err == nil && !isAdmin && o.UserID != userID) {
		return nil, apperror.NotFound("Không tìm thấy đơn hàng")
	}
	return o, err
}

func (s *OrderService) Cancel(id, userID uint) (*model.Order, error) {
	var order *model.Order
	err := s.db.Transaction(func(tx *gorm.DB) error {
		orders := repository.NewOrderRepository(tx)
		products := repository.NewProductRepository(tx)

		o, err := orders.FindForUpdate(id)
		if errors.Is(err, repository.ErrNotFound) || (err == nil && o.UserID != userID) {
			return apperror.NotFound("Không tìm thấy đơn hàng")
		}
		if err != nil {
			return err
		}
		if o.Status != model.OrderPending {
			return apperror.Conflict("Chỉ có thể huỷ đơn đang chờ xử lý")
		}
		for _, it := range o.Items {
			if err := products.AddStock(it.ProductID, it.Quantity); err != nil {
				return err
			}
		}
		if err := orders.UpdateStatus(o.ID, model.OrderCancelled); err != nil {
			return err
		}
		o.Status = model.OrderCancelled
		order = o
		return nil
	})
	return order, err
}
