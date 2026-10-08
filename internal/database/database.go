package database

import (
	"log"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/luanjr10/shoptech-go-api/internal/model"
)

func Connect(dsn string) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)
	for attempt := 1; attempt <= 10; attempt++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			sqlDB, _ := db.DB()
			if err = sqlDB.Ping(); err == nil {
				sqlDB.SetMaxOpenConns(20)
				sqlDB.SetMaxIdleConns(10)
				sqlDB.SetConnMaxLifetime(30 * time.Minute)
				return db, nil
			}
		}
		log.Printf("chưa kết nối được MySQL (lần %d/10): %v", attempt, err)
		time.Sleep(3 * time.Second)
	}
	return nil, err
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&model.User{},
		&model.Category{},
		&model.Product{},
		&model.Order{},
		&model.OrderItem{},
	)
}
