package service

import (
	"errors"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/luanjr10/shoptech-go-api/internal/apperror"
	"github.com/luanjr10/shoptech-go-api/internal/config"
	"github.com/luanjr10/shoptech-go-api/internal/dto"
	"github.com/luanjr10/shoptech-go-api/internal/model"
	"github.com/luanjr10/shoptech-go-api/internal/repository"
	"github.com/luanjr10/shoptech-go-api/internal/token"
)

type AuthService struct {
	users *repository.UserRepository
	cfg   config.Config
}

func NewAuthService(db *gorm.DB, cfg config.Config) *AuthService {
	return &AuthService{users: repository.NewUserRepository(db), cfg: cfg}
}

type AuthResult struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

func (s *AuthService) Register(req dto.RegisterRequest) (*AuthResult, error) {
	email := strings.ToLower(strings.TrimSpace(req.Email))

	_, err := s.users.FindByEmail(email)
	if err == nil {
		return nil, apperror.Conflict("Email đã được sử dụng")
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Name:     strings.TrimSpace(req.Name),
		Email:    email,
		Password: string(hash),
		Role:     model.RoleCustomer,
	}
	if err := s.users.Create(user); err != nil {
		return nil, err
	}
	return s.issue(user)
}

func (s *AuthService) Login(req dto.LoginRequest) (*AuthResult, error) {
	user, err := s.users.FindByEmail(strings.ToLower(strings.TrimSpace(req.Email)))
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.Unauthorized("Email hoặc mật khẩu không đúng")
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)) != nil {
		return nil, apperror.Unauthorized("Email hoặc mật khẩu không đúng")
	}
	return s.issue(user)
}

func (s *AuthService) Me(userID uint) (*model.User, error) {
	user, err := s.users.FindByID(userID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, apperror.NotFound("Không tìm thấy người dùng")
	}
	return user, err
}

func (s *AuthService) EnsureAdmin(email, password string) error {
	_, err := s.users.FindByEmail(email)
	if err == nil || !errors.Is(err, repository.ErrNotFound) {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.users.Create(&model.User{Name: "Admin", Email: email, Password: string(hash), Role: model.RoleAdmin})
}

func (s *AuthService) issue(user *model.User) (*AuthResult, error) {
	tok, err := token.Generate(s.cfg.JWTSecret, s.cfg.JWTTTL, user.ID, user.Role)
	if err != nil {
		return nil, err
	}
	return &AuthResult{Token: tok, User: user}, nil
}
