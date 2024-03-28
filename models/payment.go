package models

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	PaymentId     uint `json:"payment_id" gorm:"primaryKey"`
	User          User `gorm:"ForeignKey:Userid"`
	Userid        uint
	PaymentMethod string `json:"payment_method" gorm:"not null"`
	TotalAmount   uint   `json:"total_amount" gorm:"not null"`
	Status        string `json:"Status" gorm:"not null"`
	Date          time.Time
}

type OrderDetails struct {
	Orderid     uint `json:"orderid" gorm:"primaryKey"`
	Userid      uint
	User        User    `gorm:"ForeignKey:Userid"`
	Address     Address `gorm:"ForeignKey:Addressid"`
	Addressid   uint    //*
	Payment     Payment `gorm:"ForeignKey:Paymentid"`
	Paymentid   uint
	OrderItem   OrderItem `gorm:"ForeignKey:OrderItemId"` //
	OrderItemId int
	Productid   uint
	Product     Product `gorm:"ForeignKey:Productid"`
	Quantity    uint
	Status      string `json:"status" gorm:"not null"`
	Couponid    uint    `gorm:"default:0"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type OrderItem struct {
	//	OrderId     int     `grom:"primaryKey"`
	gorm.Model
	User        User    `gorm:"ForeignKey:UserIdNo"`
	UserIdNo    uint    `json:"useridno" gorm:"not null"`
	TotalAmount uint    `jsont:"totalAmount" gorm:"not null"`
	Payment     Payment `gorm:"ForeignKey:Paymentid"`
	Paymentid   uint    `json:"PaymentId"`
	OrderStatus string  `json:"orderStatus"`
	Address     Address `gorm:"ForeignKey:AddId"`
	AddId       uint    `json:"addid"`
	// CreatedAt   time.Time
	// UpdatedAt   time.Time
}

type RazorPay struct {
	User_id          uint   `json:"userid"`
	RazorPayment_id  string `json:"razorpaymentid" gorm:"primaryKey"`
	RazorPayOrder_id string `json:"razorpayorderid"`
	Signature        string `json:"signature"`
	AmountPaid       string `json:"amountpaid"`
}

type Wallet struct {
	Id     uint
	User   User `gorm:"ForeignKey:Userid"`
	Userid uint
	Amount float64
}

type WalletHistory struct {
	Id              uint `json:"id" gorm:"primaryKey"`
	User            User `gorm:"ForeignKey:User_id"`
	User_id         uint
	Amount          float64
	TransactionType string
	Data            time.Time
}

type Coupon struct {
	ID            int
	CouponCode    string `gorm:"unique"`
	DiscountPrice float64
	CreatedAt     time.Time
	Expired       time.Time
}
