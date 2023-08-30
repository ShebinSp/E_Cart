package controllers

import (
	"net/http"

	"github.com/ShebinSp/e-cart/initializers"
	"github.com/ShebinSp/e-cart/models"
	"github.com/gin-gonic/gin"
)

type checkoutData struct {
	FullName  string
	Phone     uint
	HouseName string
	Area      string
	Landmark  string
	City      string
	District  string
	State     string
	Pincode   uint
}
// @Summary Checkout
// @Description Proceed to checkout for the items in the cart
// @Tags Cart, Users
// @Security BearerToken
// @Produce json
// @Success 200 {object} map[string]interface{} "Checkout details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 400 {object} map[string]interface{} "Cart is empty"
// @Failure 404 {object} map[string]interface{} "Address does not exist"
// @Router /user/cart/checkout [get]
func CheckOut(c *gin.Context) {
	
	email := c.GetString("user")
	var userAddressdata []checkoutData
	db := initializers.DB
	id := getId(email, db)

	var counter int64 // Geting items in cart
	_ = db.Find(&models.Cart{}).Where("userid = ?", id).Count(&counter)
	cartdata, error := CartItems(email)
	if error != nil {
		err := map[string]interface{}{
			"Error": error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if cartdata == nil {
		err := map[string]interface{}{
			"Message": "Cart is empty,Please add items into cart",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	result := db.Raw("SELECT full_name, phone, house_name, area, landmark, city, district, state, pincode FROM addresses WHERE userid = ? AND default_add = true", id).Scan(&userAddressdata)
	if result.Error != nil {
		err := map[string]interface{}{
			"Error":   result.Error.Error(),
			"Message": "Address does not exist",
		}
		c.JSON(http.StatusNotFound, err)
		return
	}
	if userAddressdata == nil {
		err := map[string]interface{}{
			"Message": "Please add your address",
		}
		c.JSON(http.StatusNotFound, err)
		return
	}

	var totalPrice float64

	r := db.Table("carts").Where("userid = ?", id).Select("SUM(total_price)").Scan(&totalPrice).Error
	if r != nil {
		err := map[string]interface{}{
			"Error": "Can not fetch total amount",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	oK := map[string]interface{}{
		"Your cart":                cartdata,
		"Default Address of User:": userAddressdata,
		"Total Price":              totalPrice,
	}
	c.JSON(http.StatusOK, oK)
}
