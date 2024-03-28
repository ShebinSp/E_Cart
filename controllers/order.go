package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/ShebinSp/e-cart/initializers"
	"github.com/ShebinSp/e-cart/models"
	"github.com/gin-gonic/gin"
)

func OrderDetails(email string) error {
	db := initializers.DB
	id := getId(email, db)
	var UserAddress models.Address
	var userPayment models.Payment
	var Order_item models.OrderItem
	var UserCart []models.Cart

	r := db.Find(&UserAddress, "userid = ? AND default_add = true", id)
	if r.Error != nil {
		return r.Error
	}
	r = db.Find(&UserCart, "userid = ?", id)
	if r.Error != nil {
		return r.Error
	}
	db.Last(&userPayment, "userid = ?", id)

	// Fetch the OrderItem associated with the cartItem
	r = db.Last(&Order_item, "add_id = ?", UserAddress.AddressId)
	if r.Error != nil {
		return r.Error
	}

	for _, UserCart := range UserCart {
		orderDetails := models.OrderDetails{
			Userid:      id,
			Addressid:   UserAddress.AddressId,
			Paymentid:   userPayment.PaymentId,
			Productid:   UserCart.Productid,
			Status:      "Processing",
			Quantity:    UserCart.Quantity,
			OrderItemId: int(Order_item.ID),
			Couponid:    UserCart.Coupon,
		}

		r = db.Create(&orderDetails)
		if r.Error != nil {

			return r.Error
		}
		// paymentid := strconv.Itoa(int(orderDetails.Paymentid))
		// orderid := strconv.Itoa(orderDetails.OrderItemId)
		// message = append(message, "Order placed successfully", "Your order is "+orderDetails.Status, "Payment id = "+paymentid, "Order id = "+orderid)
	}
	return nil
}

// @Summary Show User Orders
// @Description Get a list of user orders
// @Tags Orders, Users
// @Accept json
// @Security ApiKeyAuth
// @Produce json
// @Success 200 {object} map[string]interface{} "User order details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "No orders found"
// @Router /user/showorders [get]
func ShowOrder(c *gin.Context) {

	email := c.GetString("user")
	db := initializers.DB
	id := getId(email, db)
	var userOrder []models.OrderDetails
	var products []models.Product
	var count int64

	r := db.Find(&userOrder, "userid = ?", id)
	if r.RowsAffected == 0 {
		msg := map[string]interface{}{
			"Message": "Order List is empty",
		}
		c.JSON(http.StatusNotFound, msg)
		return
	}
	count = r.RowsAffected

	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	// Fetch the product details for each order and populate the products slice
	for _, order := range userOrder {
		var product models.Product
		r := db.Find(&product, "product_id = ?", order.Productid)
		if r.Error != nil {
			err := map[string]interface{}{
				"Error": r.Error.Error(),
			}
			c.JSON(http.StatusBadRequest, err)
			return
		}
		products = append(products, product)
	}

	var orderDetails []gin.H
	for i, order := range userOrder {
		r := db.Find(&userOrder, "productid = ?", order.Productid)
		if r.Error != nil {
			err := map[string]interface{}{
				"Error": r.Error.Error(),
			}
			c.JSON(http.StatusBadRequest, err)
			return
		}
		orderDetail := gin.H{
			"Order ID":     order.Orderid,
			"Product Name": products[i].ProductName,
			"Price":        products[i].SpecialPrice,
			"Description":  products[i].Description,
			"Status":       order.Status,
			"Quantity":     order.Quantity,
		}
		orderDetails = append(orderDetails, orderDetail)
	}
	oK := map[string]interface{}{
		"Your orders":  orderDetails,
		"Total orders": count,
	}
	c.JSON(http.StatusOK, oK)
}

func ShowOrders(c *gin.Context) {

	type addressData struct {
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
	//	var address addressData//

	type orderDetails struct {
		User_id          uint
		Order_id         uint
		Status           string
		Payment_method   string
		Total_amount     uint
		CreatedAt        time.Time
		UpdatedAt        time.Time
		Shipping_Address addressData
	}

	var allOrders []models.OrderItem
	var orders orderDetails
	var userOrders []orderDetails
	var count int64

	db := initializers.DB
	r := db.Find(&allOrders)
	if r.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"Message": "Order lIst is empty",
		})
		return
	}
	if r.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}
	count = r.RowsAffected

	//Fetch Payment details for each order and populate the
	for _, order := range allOrders {
		var payment models.Payment
		//	var temp string//
		r := db.Find(&payment, "payment_id = ?", order.Paymentid).Select("payment_method").Scan(&orders.Payment_method)
		if r.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": r.Error.Error(),
			})
			return
		}
		var address models.Address
		r = db.Find(&address, "address_id = ?", order.AddId).Scan(&orders.Shipping_Address)
		if r.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": r.Error.Error(),
			})
			return
		}
		orders.User_id = order.UserIdNo
		orders.Order_id = order.ID
		orders.Status = order.OrderStatus
		orders.Total_amount = order.TotalAmount
		orders.CreatedAt = order.CreatedAt
		orders.UpdatedAt = order.UpdatedAt

		userOrders = append(userOrders, orders)
	}

	c.JSON(http.StatusOK, gin.H{
		"Order details": userOrders,
		"Total orders":  count,
	})
}

// @Summary Cancel Order
// @Description Cancel an order by order ID
// @Tags Orders, Users,
// @Accept json
// @Security ApiKeyAuth
// @Produce json
// @Param order_id query int true "Order ID to cancel"
// @Success 200 {object} map[string]interface{} "Order cancellation success message"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Order not found"
// @Router /user/cancelorder [patch]
func CancelOrder(c *gin.Context) {

	orderItemId, err := strconv.Atoi(c.Query("order_id")) // param
	if err != nil {
		err := map[string]interface{}{
			"Error": "Error in string conversion",
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	var orderItem models.OrderItem
	db := initializers.DB

	err = db.First(&orderItem, orderItemId).Error
	if err != nil {
		msg := map[string]interface{}{
			"Error": "Order id doesn't exist",
		}
		c.JSON(http.StatusBadRequest, msg)
		return
	}

	if orderItem.OrderStatus == "Cancelled" {
		msg := map[string]interface{}{
			"Message": "Order already Cancelled",
		}
		c.JSON(http.StatusBadRequest, msg)
		return
	}

	r := db.Model(&orderItem).Where("id = ?", orderItemId).Update("order_status", "Cancelled")
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	//	adding the balance to wallet
	typE := "credit"
	err = addToWallet(orderItem.UserIdNo, float64(orderItem.TotalAmount), typE)
	if err != nil {
		err := map[string]interface{}{
			"Error": err.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	oK := map[string]interface{}{
		"Message": "Order Cancelled",
	}
	c.JSON(http.StatusOK, oK)
}

func addToWallet(id uint, amount float64, Type string) error {
	var wallet models.Wallet
	db := initializers.DB
	r := db.Where("userid = ?", id).First(&wallet)
	if r.Error != nil {
		walletData := models.Wallet{
			Userid: uint(id),
			Amount: float64(amount),
		}
		r = db.Create(&walletData)
		if r.Error != nil {
			return r.Error
		}
	} else {
		totalAmount := wallet.Amount + float64(amount)
		fmt.Println("Total amount: ", totalAmount)

		r = db.Model(&wallet).Where("userid = ?", id).Update("amount", totalAmount)
		if r.Error != nil {
			return r.Error
		}
	}
	// history
	whistory := models.WalletHistory{
		User_id:         id,
		Amount:          float64(amount), //Amount:          float64(orderItem.TotalAmount),
		TransactionType: Type,
		Data:            time.Now(),
	}
	r = db.Create(&whistory)
	if r.Error != nil {
		return r.Error
	}
	return nil
}


type coupon_Data struct {
	couponid   int
	quantity uint
}
// @Summary Return Order
// @Description Request to return an order by providing order ID and product ID
// @Tags Orders, Users
// @Accept json
// @Security ApiKeyAuth
// @Produce json
// @Param product_id query int true "Product ID of the ordered product"
// @Param order_id query int true "Order ID to be returned"
// @Success 200 {object} map[string]interface{} "Order return success message"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Order not found"
// @Router /user/returnorder [put]
func ReturnOrder(c *gin.Context) {
	product_id, err := strconv.Atoi(c.Query("product_id"))
	if err != nil {
		err := map[string]interface{}{
			"Error": "String convertion failed",
		}
		c.JSON(http.StatusInternalServerError, err)
	}
	orderItemId, errr := strconv.Atoi(c.Query("order_id"))
	if errr != nil {
		err := map[string]interface{}{
			"Error": "String convertion failed",
		}
		c.JSON(http.StatusInternalServerError, err)
	}

	var orderItem models.OrderItem
	db := initializers.DB

	err = db.First(&orderItem, orderItemId).Error
	if err != nil {
		err := map[string]interface{}{
			"Error": "Order id doesn't exist",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	if orderItem.OrderStatus == "Returned" {
		err := map[string]interface{}{
			"Error": "Order id doesn't exist",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	var price uint
	var couponAmount uint
	var returnData coupon_Data
	var amount uint

	_ = db.Table("products").Select("special_price").Where("product_id = ?", product_id).Scan(&price)
	_ = db.Table("order_details").Select("couponid", "quantity").Where("productid = ?", product_id).Scan(&returnData)
	if returnData.couponid != 0 {
		_ = db.Table("coupons").Where("id = ?",returnData.couponid).Scan(&couponAmount)
		amount = (price - couponAmount) * returnData.quantity
		r := db.Table("payments").Where("id = ?",orderItemId).Update("total_amount", amount)
		if r.Error != nil {
			err := map[string]interface{}{
				"Error": r.Error.Error(),
			}
			c.JSON(http.StatusInternalServerError, err)
		}
	} else {
		amount = price * returnData.quantity
		r := db.Table("payments").Where("id = ?",orderItemId).Update("total_amount", amount)
		if r.Error != nil {
			err := map[string]interface{}{
				"Error": r.Error.Error(),
			}
			c.JSON(http.StatusInternalServerError, err)
		}
	}

	r := db.Model(&orderItem).Where("id = ?", orderItemId).Update("order_status", "Returned")
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	//	adding the balance to wallet
	typE := "credit"
	err = addToWallet(orderItem.UserIdNo, float64(amount), typE)
	if err != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	ok := map[string]interface{}{
		"Message": "Order return request success",
	}
	c.JSON(http.StatusOK, ok)
}
