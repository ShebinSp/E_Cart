package controllers

import (
	"net/http"

	"github.com/ShebinSp/e-cart/initializers"
	"github.com/ShebinSp/e-cart/models"
	"github.com/gin-gonic/gin"
)

// @Summary Add User Address
// @Description Add a new address for the user
// @Tags Add Address, Users
// @Security BearerToken
// @Accept json
// @Produce json
// @Param address body models.Address true "User address data"
// @Success 200 {object} map[string]interface{} "Address added successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/addaddress [post]
func AddAddress(c *gin.Context) {

	email := c.GetString("user")
	var db = initializers.DB

	var userAddress models.Address

	if c.Bind(&userAddress) != nil {
		err := map[string]interface{}{
			"Error": "Error binding JSON Data",
		}
		c.JSON(http.StatusBadRequest,err)
		return
	}
	id := getId(email, db)
	userAddress.Userid = id

	db.Model(&models.Address{}).Where("userid = ?", id).Update("default_add", false)
	result := db.Create(&userAddress)
	if result.Error != nil {
		err := map[string]interface{}{
			"Error": result.Error.Error(),
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	db.Model(&userAddress).Where("address_id = ?", userAddress.AddressId).Update("default_add", true)

	Ok := map[string]interface{}{
		"Message": "Address Added Successfully",
	}
	c.JSON(http.StatusOK, Ok)
}
