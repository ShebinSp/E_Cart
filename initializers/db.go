package initializers

import (
	"fmt"
	"os"

	"github.com/ShebinSp/e-cart/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectToDb() {
	var err error

	DSN := os.Getenv("dsn")
	DB, err = gorm.Open(postgres.Open(DSN), &gorm.Config{})
	if err != nil {
		fmt.Println("Failed to connect to database")
	}

	DB.AutoMigrate(&models.User{})
	DB.AutoMigrate(&models.Product{})
	DB.AutoMigrate(&models.Brand{})
	DB.AutoMigrate(&models.Category{})
	DB.AutoMigrate(&models.Cart{})
	DB.AutoMigrate(&models.Image{})
	DB.AutoMigrate(&models.Address{})
	DB.AutoMigrate(&models.OrderDetails{})
	DB.AutoMigrate(&models.OrderItem{})
	DB.AutoMigrate(&models.Payment{})
	DB.AutoMigrate(&models.RazorPay{})
	DB.AutoMigrate(&models.Wallet{})
	DB.AutoMigrate(&models.WalletHistory{})
	DB.AutoMigrate(&models.Coupon{})
	DB.AutoMigrate(&models.Offers{})
	DB.AutoMigrate(&models.Referal_info{})
}