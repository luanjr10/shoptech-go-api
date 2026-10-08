package router_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/luanjr10/shoptech-go-api/internal/config"
	"github.com/luanjr10/shoptech-go-api/internal/database"
	"github.com/luanjr10/shoptech-go-api/internal/router"
	"github.com/luanjr10/shoptech-go-api/internal/service"
)

const (
	adminEmail    = "admin@test.local"
	adminPassword = "admin123"
)

type apiResponse struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    json.RawMessage   `json:"data"`
	Errors  map[string]string `json:"errors"`
}

type testApp struct {
	t      *testing.T
	engine *gin.Engine
}

func newTestApp(t *testing.T) *testApp {
	t.Helper()
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })

	if err := database.Migrate(db); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{JWTSecret: "test-secret", JWTTTL: time.Hour}
	if err := service.NewAuthService(db, cfg).EnsureAdmin(adminEmail, adminPassword); err != nil {
		t.Fatal(err)
	}
	return &testApp{t: t, engine: router.New(db, cfg)}
}

func (a *testApp) call(method, path, tok string, body any, wantStatus int) apiResponse {
	a.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	rec := httptest.NewRecorder()
	a.engine.ServeHTTP(rec, req)

	if rec.Code != wantStatus {
		a.t.Fatalf("%s %s: status = %d, muốn %d. Body: %s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	var res apiResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		a.t.Fatalf("response không phải JSON: %s", rec.Body.String())
	}
	return res
}

func decode[T any](t *testing.T, raw json.RawMessage) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

func (a *testApp) login(email, password string) string {
	res := a.call("POST", "/api/v1/auth/login", "", map[string]string{"email": email, "password": password}, http.StatusOK)
	return decode[struct{ Token string }](a.t, res.Data).Token
}

func (a *testApp) seedProduct(adminTok string, stock int) uint {
	cat := a.call("POST", "/api/v1/categories", adminTok, map[string]any{"name": "Laptop"}, http.StatusCreated)
	catID := decode[struct{ ID uint }](a.t, cat.Data).ID

	p := a.call("POST", "/api/v1/products", adminTok, map[string]any{
		"category_id": catID, "name": "MacBook Air M3", "price": 25_000_000, "stock": stock,
	}, http.StatusCreated)
	return decode[struct{ ID uint }](a.t, p.Data).ID
}

func (a *testApp) stockOf(productID uint) int {
	res := a.call("GET", fmt.Sprintf("/api/v1/products/%d", productID), "", nil, http.StatusOK)
	return decode[struct{ Stock int }](a.t, res.Data).Stock
}

func TestRegisterAndLogin(t *testing.T) {
	app := newTestApp(t)

	body := map[string]string{"name": "Luân", "email": "luan@test.local", "password": "secret123"}
	res := app.call("POST", "/api/v1/auth/register", "", body, http.StatusCreated)
	if !res.Success {
		t.Fatal("register phải trả success=true")
	}
	app.call("POST", "/api/v1/auth/register", "", body, http.StatusConflict)

	tok := app.login("luan@test.local", "secret123")
	me := app.call("GET", "/api/v1/auth/me", tok, nil, http.StatusOK)
	user := decode[map[string]any](t, me.Data)
	if user["role"] != "customer" {
		t.Errorf("role = %v, muốn customer", user["role"])
	}
	if _, leaked := user["password"]; leaked {
		t.Error("API không được trả về mật khẩu")
	}

	app.call("POST", "/api/v1/auth/login", "", map[string]string{"email": "luan@test.local", "password": "sai"}, http.StatusUnauthorized)
}

func TestValidationErrors(t *testing.T) {
	app := newTestApp(t)
	res := app.call("POST", "/api/v1/auth/register", "", map[string]string{"email": "khong-phai-email"}, http.StatusUnprocessableEntity)
	for _, field := range []string{"name", "email", "password"} {
		if _, ok := res.Errors[field]; !ok {
			t.Errorf("thiếu lỗi cho field %q: %v", field, res.Errors)
		}
	}
}

func TestAuthorization(t *testing.T) {
	app := newTestApp(t)
	app.call("POST", "/api/v1/auth/register", "", map[string]string{"name": "A", "email": "a@test.local", "password": "secret123"}, http.StatusCreated)
	customerTok := app.login("a@test.local", "secret123")

	app.call("POST", "/api/v1/categories", "", map[string]any{"name": "X"}, http.StatusUnauthorized)
	app.call("POST", "/api/v1/categories", "token-gia", map[string]any{"name": "X"}, http.StatusUnauthorized)
	app.call("POST", "/api/v1/categories", customerTok, map[string]any{"name": "X"}, http.StatusForbidden)
	app.call("POST", "/api/v1/categories", app.login(adminEmail, adminPassword), map[string]any{"name": "X"}, http.StatusCreated)
}

func TestProductCRUDAndSearch(t *testing.T) {
	app := newTestApp(t)
	adminTok := app.login(adminEmail, adminPassword)
	id := app.seedProduct(adminTok, 5)

	list := app.call("GET", "/api/v1/products?q=macbook&page=1&limit=10", "", nil, http.StatusOK)
	page := decode[struct {
		Items []struct{ Name string }
		Total int64
	}](t, list.Data)
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("tìm kiếm phải ra 1 sản phẩm, được %+v", page)
	}

	app.call("PUT", fmt.Sprintf("/api/v1/products/%d", id), adminTok, map[string]any{
		"category_id": 1, "name": "MacBook Air M3 2024", "price": 24_000_000, "stock": 0,
	}, http.StatusOK)
	if got := app.stockOf(id); got != 0 {
		t.Errorf("stock = %d, muốn 0", got)
	}

	app.call("DELETE", "/api/v1/categories/1", adminTok, nil, http.StatusConflict)
	app.call("DELETE", fmt.Sprintf("/api/v1/products/%d", id), adminTok, nil, http.StatusOK)
	app.call("GET", fmt.Sprintf("/api/v1/products/%d", id), "", nil, http.StatusNotFound)
}

func TestOrderDeductsStockAndCancelRestores(t *testing.T) {
	app := newTestApp(t)
	productID := app.seedProduct(app.login(adminEmail, adminPassword), 5)

	app.call("POST", "/api/v1/auth/register", "", map[string]string{"name": "B", "email": "b@test.local", "password": "secret123"}, http.StatusCreated)
	tok := app.login("b@test.local", "secret123")

	res := app.call("POST", "/api/v1/orders", tok, map[string]any{
		"items": []map[string]any{{"product_id": productID, "quantity": 2}, {"product_id": productID, "quantity": 1}},
	}, http.StatusCreated)
	order := decode[struct {
		ID     uint
		Total  int64
		Status string
	}](t, res.Data)
	if order.Total != 75_000_000 || order.Status != "pending" {
		t.Fatalf("đơn hàng sai: %+v", order)
	}
	if got := app.stockOf(productID); got != 2 {
		t.Fatalf("sau khi đặt 3, stock = %d, muốn 2", got)
	}

	app.call("POST", "/api/v1/orders", tok, map[string]any{
		"items": []map[string]any{{"product_id": productID, "quantity": 3}},
	}, http.StatusConflict)
	if got := app.stockOf(productID); got != 2 {
		t.Fatalf("đặt thất bại không được trừ kho, stock = %d", got)
	}

	app.call("POST", fmt.Sprintf("/api/v1/orders/%d/cancel", order.ID), tok, nil, http.StatusOK)
	if got := app.stockOf(productID); got != 5 {
		t.Fatalf("huỷ đơn phải hoàn kho, stock = %d, muốn 5", got)
	}
	app.call("POST", fmt.Sprintf("/api/v1/orders/%d/cancel", order.ID), tok, nil, http.StatusConflict)
}

func TestOrderOfAnotherUserIsHidden(t *testing.T) {
	app := newTestApp(t)
	productID := app.seedProduct(app.login(adminEmail, adminPassword), 5)

	for _, email := range []string{"c@test.local", "d@test.local"} {
		app.call("POST", "/api/v1/auth/register", "", map[string]string{"name": "U", "email": email, "password": "secret123"}, http.StatusCreated)
	}
	owner, other := app.login("c@test.local", "secret123"), app.login("d@test.local", "secret123")

	res := app.call("POST", "/api/v1/orders", owner, map[string]any{
		"items": []map[string]any{{"product_id": productID, "quantity": 1}},
	}, http.StatusCreated)
	id := decode[struct{ ID uint }](t, res.Data).ID

	app.call("GET", fmt.Sprintf("/api/v1/orders/%d", id), other, nil, http.StatusNotFound)
	app.call("POST", fmt.Sprintf("/api/v1/orders/%d/cancel", id), other, nil, http.StatusNotFound)
	app.call("GET", fmt.Sprintf("/api/v1/orders/%d", id), owner, nil, http.StatusOK)
}
