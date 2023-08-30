package controllers

import (
	"fmt"
	"net/http"

	"github.com/ShebinSp/e-cart/auth"
	"github.com/ShebinSp/e-cart/initializers"
	"github.com/ShebinSp/e-cart/models"
	"github.com/gin-gonic/gin"
)

type AdminLog struct {
	Email    string
	Password string
}

func AdminSignup(c *gin.Context) {
	var admin models.User
	var isExist uint

	if err := c.ShouldBindJSON(&admin); err != nil {
		c.JSON(404, gin.H{
			"err": err.Error(),
		})
		c.Abort()
		return
	}

	initializers.DB.Raw("select isExist(*) from admins where email=?", admin.Email).Scan(&isExist)
	if isExist > 0 {
		c.JSON(400, gin.H{
			"status": "false",
			"msg":    "an admin with same email already exists",
		})
		c.Abort()
		return
	}

	if err := admin.HashPassword(admin.Password); err != nil {
		c.JSON(404, gin.H{
			"error_hash": err.Error(),
		})
	}
	record := initializers.DB.Create(&admin)
	if record.Error != nil {
		c.JSON(400, gin.H{
			"err_i": record.Error.Error(),
			"msg":   "Admin signup failed",
		})
		c.Abort()
		return
	}
	c.JSON(200, gin.H{
		"email":  admin.Email,
		"status": "ok",
		"msg":    "new admin account created successfuly, Please login to continue",
	})
}

func AdminLogin(c *gin.Context) {
	var login AdminLog
	var admin models.User
	//adminEmail := "shebinsp@gmail.com"
	db := initializers.DB
	_ = db.Model(&admin).Where("email = ?", "shebinsp@gmail.com").Update("is_admin", true)

	if err := c.ShouldBindJSON(&login); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"err_bindJSON": err.Error(),
		})
		c.Abort()
		return
	}

	record := db.Raw("select * from users where email = ?", login.Email).Scan(&admin)
	if record.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Invalid email",
		})
		c.Abort()
		return
	}

	if err := admin.VerifyPassword(login.Password); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "Invalid Password",
		})
		c.Abort()
		return
	}

	err := initializers.DB.Model(&admin).Select("is_admin").First(&admin).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"msg": "Failed to retrive user data",
		})
		c.Abort()
		return
	}

	if !admin.Is_admin {
		c.JSON(http.StatusUnauthorized, gin.H{
			"msg": "User is not an admin",
		})
		c.Abort()
		return
	}

	tokenString, err := auth.GenerateJWT(admin.Email)
	fmt.Println(tokenString)

	token := tokenString["access_token"]
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("AdminAuth", token, 3600*24*30, "", "", false, true)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		c.Abort()
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email": admin.Email,
		"msg":   admin.Email + " logged in as admin successfully!",
		"token": tokenString,
	})

}

func AdminSignout(c *gin.Context) {
	c.SetCookie("AdminAuth", "", -1, "", "", false, false)
	c.JSON(http.StatusOK, gin.H{
		"Message": "Admin Successfully Signed Out",
	})
}

func ChangeOrderStatus(c *gin.Context){
	var Order_status struct {
		Order_id int
		Order_Status string
	}
	if err := c.BindJSON(&Order_status); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"Error": err.Error(),
		})
		return
	}
	db := initializers.DB
	r := db.Model(&models.OrderItem{}).Where("id = ?", Order_status.Order_id).Update("order_status", Order_status.Order_Status)
	if r.Error != nil {
		c.JSON(http.StatusOK, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}
	
	var paymentId uint
	var orders models.OrderItem
	_ = db.Find(&orders,"id = ?",Order_status.Order_id).Select("paymentid").Scan(&paymentId)
	fmt.Println("user:",paymentId)
	r = db.Model(&models.Payment{}).Where("payment_id = ?",paymentId).Update("status",Order_status.Order_Status)
	if r.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}
	r = db.Model(&models.OrderDetails{}).Where("paymentid = ?",paymentId).Update("status",Order_status.Order_Status)
	if r.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"order status": Order_status.Order_Status,
		"Message": "Order status changed successfully",
	})
	
}

func AdminDashBoard(c *gin.Context){
	// Total number of products
	var count int64
	db := initializers.DB
	r := db.Find(&models.Product{}).Count(&count)
	if r.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}

	// Total categories and number of categories
	var categories []models.Category
	var cCount int64
	r = db.Find(&categories).Count(&cCount)
	if r.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}

	// Total brands and number of brands
	var brands []models.Brand
	var bCount int64
	r = db.Find(&brands).Count(&bCount)
	if r.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}

	// Total number of users
	var uCount int64
	r = db.Find(&models.User{}).Count(&uCount)
	if r.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}

	// Total orders
	var oCount int64
	r = db.Find(&models.OrderItem{}).Count(&oCount)
	if r.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}

	// Total profit
	var pCount int64
	var totalAmount float64
	r = db.Table("payments").Select("SUM(total_amount)").Scan(&totalAmount).Count(&pCount)
	if r.Error != nil {
		totalAmount = 0
	}
	var rPay_amount float64
	var rpayCount int64
	r = db.Table("payments").Select("SUM(total_amount)").Where("payment_method = ?", "Razor Pay").Scan(&rPay_amount).Count(&rpayCount)
	if r.Error != nil {
		rPay_amount = 0
	}
	var cod_amount float64
	var codCount int64
	if codCount != 0 {
		r = db.Table("payments").Select("SUM(total_amount)").Where("payment_method = ?", "COD").Scan(&cod_amount).Count(&codCount)
	if r.Error != nil {
		cod_amount = 0
	}
	}


	// Total sales
	var pQty int64
	r = db.Table("order_details").Select("SUM(quantity)").Scan(&pQty)
	if r.Error != nil {
		pQty = 0
	}

	c.JSON(http.StatusOK, gin.H{
		"Total number of products": count,
		"All categories": categories,
		"Total number of categories": cCount,
		"All brands": brands,
		"Total number of brands": bCount,
		"Total number of users": uCount,
		"Total orders": oCount,
		"Total products sold": pQty,
		"Total Profit": totalAmount,
		"Total number of payments": pCount,
		"Total amount credited via Razor Pay": rPay_amount,
		"Total Razor pay payments": rpayCount,
		"Total payments as COD": codCount,
		"Total amount credited as COD": cod_amount,

	})
}