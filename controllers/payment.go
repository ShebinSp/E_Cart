package controllers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/ShebinSp/e-cart/initializers"
	"github.com/ShebinSp/e-cart/models"
	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
)

type updates struct {
	Productid uint
	Quantity  int
}

// @Summary Cash On Delivery
// @Description Place an order using Cash On Delivery payment method
// @Tags Orders, Users
// @Accept json
// @Security BearerToken
// @Produce json
// @Success 200 {object} map[string]interface{} "Order details and status"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Cart is empty"
// @Router /user/payment/cashOnDelivery [get]
func CashOnDelivery(c *gin.Context) {
	email := c.GetString("user")
	db := initializers.DB
	id := getId(email, db)

	var productUpdate []updates

	result := db.Table("carts").Select("productid", "quantity").Where("userid = ?", id).Scan(&productUpdate)
	if result.RowsAffected == 0 {
		msg := map[string]interface{}{
			"Message": "Cart is empty",
		}
		c.JSON(http.StatusNotFound, msg)
		return
	}
	// for _,details := range productUpdate {
	// 	fmt.Println("Product id : ",details.Productid)
	// 	fmt.Println("Quantity : ",details.Quantity)
	// }

	var totalAmount float64
	r := db.Table("carts").Where("userid = ?", id).Select("SUM(total_price)").Scan(&totalAmount)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": "Can not fetch total amount",
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	date := time.Now()
	paymentData := models.Payment{
		PaymentMethod: "COD",
		TotalAmount:   uint(totalAmount),
		Date:          date,
		Status:        "pending",
		Userid:        id,
	}
	r = db.Create(&paymentData)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	var addressData models.Address
	r = db.First(&addressData, "userid = ?", id)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// var count int64
	// _ = db.Find(&models.Cart{}, "userid = ?",id).Where("total_price"<"price").Count(&count)
	// var couponApplied bool
	// if count == 0{

	// }
	orderData := models.OrderItem{
		UserIdNo:    id,
		TotalAmount: uint(totalAmount),
		Paymentid:   paymentData.PaymentId,
		AddId:       addressData.AddressId,
		OrderStatus: "Processing",
	}
	r = db.Create(&orderData)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error":    r.Error.Error(),
			"Message ": "Order table creation",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	err := OrderDetails(email)
	if err != nil {
		err := map[string]interface{}{
			"Error": err.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	// Updating stock
	err = updateProductQty(productUpdate)
	if err != nil {
		msg := map[string]interface{}{
			"Error": "Product updation failed",
		}
		c.JSON(http.StatusBadRequest, msg)
		return
	}

	result = db.Exec("delete from carts where userid = ?", id)
	count := result.RowsAffected
	if count == 0 {
		msg := map[string]interface{}{
			"Message": "Cart does not exist",
		}
		c.JSON(http.StatusBadRequest, msg)
		return
	}
	if result.Error != nil {
		err := map[string]interface{}{
			"Error": result.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	var orders models.OrderDetails

	r = db.Last(&orders).Where("userid = ?", id)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": "Can not fetch order details",
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	oK := map[string]interface{}{
		"Payment Method":  "COD",
		"Message":         "Item removed from cart",
		"Order Status":    orders.Status,
		"Order id":        orders.Orderid,
		"Order placed at": orders.CreatedAt,
	}
	c.JSON(http.StatusOK, oK)
}

func RazorPay(c *gin.Context) {

	email := c.GetString("user")
	db := initializers.DB
	id := getId(email, db)
	//	id := 1

	var userdata models.User
	r := db.Find(&userdata, "id = ?", id)
	if r.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}

	// Fetching the total amount from cart
	var amount uint
	row := db.Table("carts").Where("userid = ?", id).Select("SUM(total_price)").Row()
	err := row.Scan(&amount)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": err.Error(),
		})
		return
	}

	//Sending payment details to razorpay
	client := razorpay.NewClient(os.Getenv("RAZOR_KEY"), os.Getenv("RAZOR_SECRET"))
	data := map[string]interface{}{
		"amount":   amount * 100,
		"currency": "INR",
		"receipt":  "some_receipt_id",
	}

	// Creating the payment details to client order
	body, errr := client.Order.Create(data, nil) // err is not a new var
	if errr != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": errr,
		})
		return
	}

	// To rendering the HTML page with user and payment details
	value := body["id"]
	//	fmt.Println("value-", value)

	c.HTML(http.StatusOK, "app.html", gin.H{
		"userid":     userdata.ID,
		"totalprice": amount,
		"paymentid":  value,
	})
}

// When the Razorpay payment is done this function will work
func RazorpaySuccess(c *gin.Context) {
	email := c.GetString("user")
	db := initializers.DB
	id := getId(email, db)

	// Fetching the payment details from razorpay
	orderid := c.Query("order_id")
	paymentid := c.Query("payment_id")
	signature := c.Query("signature")
	totalamount := c.Query("total")

	//Creating table razorpay using the data from razorpay
	Rpay := models.RazorPay{
		User_id:          id,
		RazorPayment_id:  paymentid,
		Signature:        signature,
		RazorPayOrder_id: orderid,
		AmountPaid:       totalamount,
	}
	r := db.Create(&Rpay)
	if r.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}

	today := time.Now()
	totalprice, err := strconv.Atoi(totalamount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "String conversion error",
		})
		return
	}

	//Creating payment table
	paymentdata := models.Payment{
		Userid:        id,
		PaymentMethod: "Razor Pay",
		Status:        "Success",
		Date:          today,
		TotalAmount:   uint(totalprice),
	}
	r = db.Create(&paymentdata)
	if r.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}

	var addressData models.Address
	r = db.First(&addressData, "userid = ?", id)
	if r.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}
	pid := paymentdata.PaymentId

	orderData := models.OrderItem{
		UserIdNo:    id,
		TotalAmount: uint(totalprice),
		Paymentid:   pid,
		AddId:       addressData.AddressId,
		OrderStatus: "Processing",
	}

	r = db.Create(&orderData)
	if r.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}
	err = OrderDetails(email)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": err.Error(),
		})
		return
	}

	// Update product qty
	var productUpdate []updates
	result := db.Table("carts").Select("productid", "quantity").Where("userid = ?", id).Scan(&productUpdate)
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"Message": "Cart is empty",
		})
		return
	}
	err = updateProductQty(productUpdate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Product updation failed",
		})
		return
	}
	//	DeleteCart
	result = db.Exec("delete from carts where userid = ?", id)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": result.Error.Error(),
		})
		return
	}
	var orders models.OrderDetails
	r = db.Last(&orders).Where("userid = ?", id)
	if r.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Can not fetch order details",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Payment Method": "Razor Pay",
		"status":         true,
		"payment_id":     pid,
		"Message":        "Order Placed Successfully",
		"notice":         "Item removed from cart",
		"Order id":       orders.Orderid,
		"Order Place at": orders.CreatedAt,
	})
}

func Success(c *gin.Context) {
	
	pid, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Error in string conversion",
		})
		return
	}
	fmt.Println("id:",pid)
	c.HTML(http.StatusOK, "success.html", gin.H{
		"payment_id": pid,
	})
}

// @Summary Show Wallet Balance
// @Description Retrieve the balance amount in the user's wallet
// @Tags Wallet, Users
// @Accept json
// @Security BearerToken
// @Produce json
// @Success 200 {object} map[string]interface{} "Wallet balance"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Router /user/payment/showwallet [get]
func ShowWallet(c *gin.Context) {
	email := c.GetString("user")
	db := initializers.DB
	id := getId(email, db)

	var wallet models.Wallet

	r := db.First(&wallet).Where("userid = ?", id)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	oK := map[string]interface{}{
		"Balance Amount": wallet.Amount,
	}
	c.JSON(http.StatusOK, oK)
}

type couponData struct {
	Coupon string
}
// @Summary Check Coupon Validity
// @Description Check the validity of a coupon code
// @Tags Coupons, Cart, Users
// @Accept json
// @Param data body couponData true "Coupon data"
// @Produce json
// @Success 200 {object} map[string]interface{} "Coupon validity status"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Coupon does not exist"
// @Router /user/checkcoupon [post]
func CheckCoupon(c *gin.Context) {
	var coupon models.Coupon
	var userEnterData couponData

	if c.Bind(&userEnterData) != nil {
		err := map[string]interface{}{
			"Error": "Could not bind the JSON Data",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	db := initializers.DB
	var count int64
	r := db.Find(&coupon, "coupon_code = ?", userEnterData.Coupon).Count(&count)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if count == 0 {
		msg := map[string]interface{}{
			"message": "Coupon does't exist",
		}
		c.JSON(http.StatusNotFound, msg)
		return
	}
	currentTime := time.Now()
	expiredData := coupon.Expired

	if currentTime.Before(expiredData) {
		oK := map[string]interface{}{
			"message":    "Coupon is valid",
			"expirys on": expiredData,
		}
		c.JSON(http.StatusOK, oK)
	} else if currentTime.After(expiredData) {
		msg := map[string]interface{}{
			"message": "Coupon expired",
		}
		c.JSON(http.StatusBadRequest, msg)
	}
}

type uSerData struct {
	Coupon     string
	Product_id int
	FullName   string
	Phone      string
	Area       string
	Landmark   string
	City       string
	Pincode    string
	District   string
	State      string
}
// @Summary Apply Coupon
// @Description Apply a coupon code to a specific product in the cart
// @Tags Coupons, Cart, Users
// @Security BearerToken
// @Accept json
// @Param data body uSerData true "Coupon data"
// @Produce json
// @Success 200 {object} map[string]interface{} "Coupon details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Coupon does not exist"
// @Router /user/applycoupon [post]
func ApplayCoupon(c *gin.Context) {
	email := c.GetString("user")
	db := initializers.DB
	id := getId(email, db)

	var userEnterData uSerData
	var coupon models.Coupon

	var discountPercentage float64
	//	var DiscountPrice float64 // manage this discount price should effect in db when applied
	var couponValidity string
	if c.Bind(&userEnterData) != nil {
		err := map[string]interface{}{
			"Error": "Could not bind the JSON Data",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	var count int64
	r := db.Find(&coupon, "coupon_code = ?", userEnterData.Coupon).Count(&count)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	var price uint
	r = db.Table("carts").Select("price").Where("userid = ? AND productid = ?", id, userEnterData.Product_id).Scan(&price)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	minPurchase := 3333
	str := fmt.Sprint(minPurchase)
	if price < 3333 {
		msg := map[string]interface{}{
			"Message": "Coupon can apply when buying product for minimum rs." + str,
		}
		c.JSON(http.StatusBadRequest, msg)
		return
	}
	if count == 0 {
		err := map[string]interface{}{
			"Message": "Coupon does not exist",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	} else {
		currentTime := time.Now()
		expiredData := coupon.Expired

		if currentTime.Before(expiredData) {
			couponValidity = "Coupon valid"
			discountPercentage = coupon.DiscountPrice
			//fmt.Println("Dis:", discountPercentage)
		} else if currentTime.After(expiredData) {
			msg := map[string]interface{}{
				"message": "Coupon expired",
			}
			c.JSON(http.StatusBadRequest, msg)
			return
		}
	}
	//	ViewCart //-->
	var counter int64
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
		msg := map[string]interface{}{
			"Message": "Cart is empty",
		}
		c.JSON(http.StatusBadRequest, msg)
		return
	}

	var totalPrice float64
	result := db.Table("carts").Where("userid = ?", id).Select("SUM(total_price)").Scan(&totalPrice).Error
	if result != nil {
		err := map[string]interface{}{
			"Error": "Can not fetch total amount",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var product string
	_ = db.First(&models.Product{}, userEnterData.Product_id).Select("product_name").Scan(&product)
	discountPrice := totalPrice - discountPercentage

	oK := map[string]interface{}{
		"Cart Items":                cartdata,
		"Total Items in Cart":       counter,
		"Validity":                  couponValidity,
		"Valid till":                coupon.Expired,
		"Total Price":               totalPrice,
		"Discount":                  discountPercentage, //discountPrice,
		"Total payable amount":      discountPrice,
		"Coupon applied on product": product,
	}
	c.JSON(http.StatusOK, oK)

	var qty int64
	r = db.First(&models.Cart{}).Where("productid = ?", userEnterData.Product_id).Select("quantity").Scan(&qty)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	r = db.First(&models.Cart{}).Where("productid = ?", userEnterData.Product_id).Updates(models.Cart{TotalPrice: uint(price*uint(qty) - uint(discountPercentage)), Coupon: uint(coupon.ID)})
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	now := time.Now()
	r = db.Find(&models.Coupon{}).Where("id = ?", coupon.ID).Update("expired", now)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusInternalServerError, err)
	}

}

func updateProductQty(productUpdate []updates) error {
	var product models.Product
	var stock int
	db := initializers.DB

	for _, details := range productUpdate {
		_ = db.Raw("SELECT stock FROM products WHERE product_id = ?", details.Productid).Select("stock").Scan(&stock)
		//		fmt.Println("Product stock=",stock) //
		currentStock := stock - details.Quantity
		r := db.Model(&product).Where("product_id = ?", details.Productid).Update("stock", currentStock)
		if r.Error != nil {
			return r.Error
		}
		//		fmt.Println("updated stock  =", currentStock) //
	}
	return nil
}
