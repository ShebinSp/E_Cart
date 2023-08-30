package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ProductId     uint     `json:"productId" gorm:"primarykey;unique"`
	ProductName   string   `json:"name" gorm:"not null"`
	SpecialPrice  uint     `json:"special_price"`
	ActualPrice   uint     `json:"actual_price" gorm:"not null"`
	Stock         int      `json:"quantity" gorm:"not null"`
	Description   string   `json:"description" gorm:"not null"`
	CategoryId    uint     `json:"CategoryId"`
	Category      Category `gorm:"ForeignKey:CategoryId"`
	BrandId       uint     `json:"BrandId"`
	Brand         Brand    `gorm:"ForeignKey:BrandId"`
	Offer_details string
	Offer_id      uint
}

type Category struct {
	Id            uint   `json:"id" gorm:"primaryKey"`
	Category_name string `json:"CategoryName"`
}

type Brand struct {
	ID         uint   `json:"id" gorm:"primaryKey"`
	Brand_name string `json:"BrandName" gorm:"not null"`
}

type Offers struct {
	ID             uint `gorm:"primaryKey;unique"`
	Offer_name     string
	Offer_discount uint
	Product_id     uint
	Category_id    uint
	Brand_id       uint
	Created_at     time.Time
	Expired_on     time.Time
}

type Cart struct {
	gorm.Model
	Product    Product `gorm:"ForeignKey:Productid"`
	Productid  uint
	Quantity   uint
	Price      uint
	TotalPrice uint
	Userid     uint
	User       User `gorm:"ForeignKey:Userid"`
	Coupon     uint `gorm:"default:0"`
}

type Image struct {
	ID        uint    `json:"id" gorm:"primaryKey"`
	Product   Product `gorm:"ForeignKey:Productid"`
	Productid uint    `json:"Product_id"`
	Image     string  `json:"image" gorm:"not null"`
}
