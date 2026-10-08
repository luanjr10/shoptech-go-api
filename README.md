# ShopTech Go API

RESTful API cho hệ thống bán lẻ đồ công nghệ, xây dựng bằng **Go**, **Gin** và **MySQL**. Hệ thống gồm xác thực người dùng bằng JWT, phân quyền theo vai trò, quản lý danh mục và sản phẩm, đặt hàng với kiểm soát tồn kho trong transaction.

## Tính năng

- Đăng ký, đăng nhập, xác thực bằng JWT; mật khẩu được băm bằng bcrypt
- Phân quyền hai vai trò: `admin` và `customer`
- Quản lý danh mục, sản phẩm (CRUD), tìm kiếm theo tên, lọc theo danh mục, phân trang
- Đặt hàng: kiểm tra và trừ tồn kho trong một transaction, khoá dòng để chống bán vượt số lượng
- Huỷ đơn hàng ở trạng thái chờ xử lý và hoàn lại tồn kho
- Kiểm tra dữ liệu đầu vào, trả lỗi chi tiết theo từng trường
- Định dạng response thống nhất cho toàn bộ API
- Graceful shutdown, tự kết nối lại khi database chưa sẵn sàng
- Đóng gói bằng Docker, kiểm thử tự động bằng GitHub Actions

## Công nghệ sử dụng

| Thành phần | Công nghệ |
|---|---|
| Ngôn ngữ | Go 1.26 |
| Web framework | Gin |
| ORM | GORM |
| Cơ sở dữ liệu | MySQL 8 |
| Xác thực | JWT (HS256), bcrypt |
| Kiểm tra dữ liệu | go-playground/validator |
| Kiểm thử | testing, httptest, SQLite in-memory |
| Triển khai | Docker, Docker Compose, GitHub Actions |

## Kiến trúc

Ứng dụng được tổ chức theo mô hình phân lớp:

```
Request → Router → Middleware → Handler → Service → Repository → MySQL
```

| Lớp | Trách nhiệm |
|---|---|
| Router | Khai báo route, gắn middleware theo nhóm |
| Middleware | Xác thực JWT, kiểm tra quyền truy cập |
| Handler | Nhận request, kiểm tra dữ liệu, trả response JSON |
| Service | Xử lý nghiệp vụ, quản lý transaction |
| Repository | Truy vấn cơ sở dữ liệu |

## Cấu trúc thư mục

```
.
├── cmd/api/              # Entry point của ứng dụng
├── internal/
│   ├── apperror/         # Lỗi nghiệp vụ kèm HTTP status
│   ├── config/           # Đọc cấu hình từ biến môi trường
│   ├── database/         # Kết nối và migrate cơ sở dữ liệu
│   ├── dto/              # Request DTO và quy tắc validate
│   ├── handler/          # HTTP handler
│   ├── middleware/       # Xác thực và phân quyền
│   ├── model/            # Định nghĩa bảng dữ liệu
│   ├── repository/       # Truy vấn dữ liệu
│   ├── response/         # Chuẩn hoá response
│   ├── router/           # Định tuyến và integration test
│   ├── service/          # Nghiệp vụ
│   └── token/            # Tạo và xác thực JWT
├── api.http              # Request mẫu
├── docker-compose.yml
└── Dockerfile
```

## Cơ sở dữ liệu

| Bảng | Mô tả |
|---|---|
| `users` | Tài khoản, vai trò |
| `categories` | Danh mục sản phẩm |
| `products` | Sản phẩm, giá, tồn kho |
| `orders` | Đơn hàng, trạng thái, tổng tiền |
| `order_items` | Chi tiết đơn hàng, lưu tên và giá sản phẩm tại thời điểm đặt |

Các bảng được tạo tự động khi khởi động ứng dụng.

## Cài đặt và chạy

### Yêu cầu

- Go 1.26 trở lên
- MySQL 8

### Chạy trực tiếp

1. Tạo file cấu hình:

   ```bash
   cp .env.example .env
   ```

2. Tạo database `shoptech_go` và cập nhật thông tin kết nối trong `.env`.

3. Khởi động server:

   ```bash
   go run ./cmd/api
   ```

Server chạy tại `http://localhost:8080`. Tài khoản admin được tạo tự động ở lần chạy đầu tiên theo `ADMIN_EMAIL` và `ADMIN_PASSWORD`.

### Chạy bằng Docker

```bash
docker compose up --build
```

API chạy tại `http://localhost:8080`, MySQL được expose ra cổng `3307`.

### Biến môi trường

| Biến | Mặc định | Mô tả |
|---|---|---|
| `APP_PORT` | `8080` | Cổng HTTP |
| `DB_HOST` | `127.0.0.1` | Địa chỉ MySQL |
| `DB_PORT` | `3306` | Cổng MySQL |
| `DB_USER` | `root` | Tài khoản MySQL |
| `DB_PASSWORD` | | Mật khẩu MySQL |
| `DB_NAME` | `shoptech_go` | Tên database |
| `JWT_SECRET` | | Khoá ký JWT |
| `JWT_TTL` | `24h` | Thời hạn token |
| `ADMIN_EMAIL` | `admin@shoptech.local` | Email tài khoản admin |
| `ADMIN_PASSWORD` | `admin123` | Mật khẩu tài khoản admin |

## API

Base URL: `/api/v1`. Các endpoint cần xác thực sử dụng header:

```
Authorization: Bearer <token>
```

### Xác thực

| Method | Endpoint | Quyền | Mô tả |
|---|---|---|---|
| POST | `/auth/register` | Công khai | Đăng ký tài khoản |
| POST | `/auth/login` | Công khai | Đăng nhập |
| GET | `/auth/me` | Đã đăng nhập | Thông tin tài khoản hiện tại |

### Danh mục

| Method | Endpoint | Quyền | Mô tả |
|---|---|---|---|
| GET | `/categories` | Công khai | Danh sách danh mục |
| POST | `/categories` | Admin | Tạo danh mục |
| PUT | `/categories/:id` | Admin | Cập nhật danh mục |
| DELETE | `/categories/:id` | Admin | Xoá danh mục (không cho phép khi còn sản phẩm) |

### Sản phẩm

| Method | Endpoint | Quyền | Mô tả |
|---|---|---|---|
| GET | `/products` | Công khai | Danh sách sản phẩm |
| GET | `/products/:id` | Công khai | Chi tiết sản phẩm |
| POST | `/products` | Admin | Tạo sản phẩm |
| PUT | `/products/:id` | Admin | Cập nhật sản phẩm |
| DELETE | `/products/:id` | Admin | Xoá sản phẩm |

Tham số truy vấn của `GET /products`: `q` (tìm theo tên), `category_id`, `page`, `limit` (tối đa 100).

### Đơn hàng

| Method | Endpoint | Quyền | Mô tả |
|---|---|---|---|
| POST | `/orders` | Đã đăng nhập | Tạo đơn hàng |
| GET | `/orders` | Đã đăng nhập | Danh sách đơn hàng của người dùng |
| GET | `/orders/:id` | Đã đăng nhập | Chi tiết đơn hàng |
| POST | `/orders/:id/cancel` | Đã đăng nhập | Huỷ đơn hàng |

### Định dạng response

Thành công:

```json
{
  "success": true,
  "message": "Đặt hàng thành công",
  "data": {
    "id": 1,
    "status": "pending",
    "total": 44000000,
    "items": [
      { "product_id": 1, "product_name": "iPhone 16", "price": 22000000, "quantity": 2 }
    ]
  }
}
```

Lỗi validate:

```json
{
  "success": false,
  "message": "Dữ liệu không hợp lệ",
  "errors": {
    "email": "email không hợp lệ",
    "password": "tối thiểu 6"
  }
}
```

| HTTP status | Ý nghĩa |
|---|---|
| `400` | Request sai định dạng |
| `401` | Chưa đăng nhập hoặc token không hợp lệ |
| `403` | Không đủ quyền |
| `404` | Không tìm thấy tài nguyên |
| `409` | Xung đột dữ liệu (trùng email, hết hàng, đơn không thể huỷ) |
| `422` | Dữ liệu không hợp lệ |

## Xử lý đặt hàng

Toàn bộ quá trình đặt hàng được thực hiện trong một transaction:

1. Gộp các dòng trùng sản phẩm và sắp xếp theo ID để các transaction luôn khoá theo cùng thứ tự, tránh deadlock.
2. Khoá từng sản phẩm bằng `SELECT ... FOR UPDATE`.
3. Kiểm tra tồn kho, trả lỗi `409` nếu không đủ hàng.
4. Trừ tồn kho và tạo đơn hàng cùng các dòng chi tiết.

Nếu bất kỳ bước nào thất bại, transaction được rollback và tồn kho giữ nguyên. Huỷ đơn hàng cũng được thực hiện trong transaction: khoá đơn, kiểm tra trạng thái, hoàn tồn kho và cập nhật trạng thái.

## Kiểm thử

```bash
go test ./... -v
```

Integration test gọi API qua `httptest` trên SQLite in-memory, không cần MySQL. Các kịch bản được kiểm thử:

- Đăng ký, đăng nhập, trùng email, sai mật khẩu
- Lỗi validate theo từng trường
- Phân quyền: thiếu token, token không hợp lệ, không đủ quyền
- CRUD và tìm kiếm sản phẩm
- Đặt hàng trừ tồn kho, đặt vượt tồn kho bị từ chối và không làm thay đổi tồn kho
- Huỷ đơn hoàn tồn kho, không cho huỷ hai lần
- Người dùng không xem hoặc huỷ được đơn hàng của người khác

GitHub Actions tự động chạy `go vet`, test và build Docker image ở mỗi lần push.
