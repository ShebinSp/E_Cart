package routes

import (
	"github.com/ShebinSp/e-cart/auth"
	"github.com/ShebinSp/e-cart/controllers"
	"github.com/ShebinSp/e-cart/middleware"
	"github.com/gin-gonic/gin"
)

func UserRoutes(c *gin.Engine) {
	user := c.Group("/user")

	{
		// add swagger

		user.POST("/signup", controllers.Signup)
		user.POST("/login", controllers.LoginUser)
		user.POST("/signup/otpverification", auth.OtpValidation)
		user.GET("/signout", middleware.UserAuth(), controllers.UserSignout)
		user.PATCH("/changepassword", middleware.UserAuth(), controllers.ChangePassword)
		user.PATCH("/forgotpassword", controllers.ForgotPassword)

		user.GET("/viewprofile", middleware.UserAuth(), controllers.ShowUserDetails)
		user.PATCH("/editprofile", middleware.UserAuth(), controllers.EditProfile)

		user.POST("/addaddress", middleware.UserAuth(), controllers.AddAddress)

		user.POST("products/search", middleware.UserAuth(), controllers.SearchProduct) //
		user.GET("products/view", middleware.UserAuth(), controllers.ViewProducts)    //ProductView
		user.GET("/products/details", middleware.UserAuth(), controllers.ProductDetails)

		user.GET("/filterbycategory", middleware.UserAuth(), controllers.FilterByCategory) //
		user.POST("/profile/addtocart", middleware.UserAuth(), controllers.AddToCart)
		user.GET("/profile/viewcart", middleware.UserAuth(), controllers.ViewCart)         // added a func to view cart seperatly
		user.DELETE("/profile/deletecart", middleware.UserAuth(), controllers.DeleteCart)
		user.GET("/cart/checkout", middleware.UserAuth(), controllers.CheckOut)

		user.POST("/applycoupon", middleware.UserAuth(), controllers.ApplayCoupon) //
		user.POST("/checkcoupon", middleware.UserAuth(), controllers.CheckCoupon)  //

		user.GET("/showorders", middleware.UserAuth(), controllers.ShowOrder)
		user.PATCH("/cancelorder", middleware.UserAuth(), controllers.CancelOrder)
		user.PUT("/returnorder", middleware.UserAuth(), controllers.ReturnOrder)

		user.GET("/payment/cashOnDelivery", middleware.UserAuth(), controllers.CashOnDelivery)
		user.GET("/payment/razorpay", middleware.UserAuth(), controllers.RazorPay)       //
		user.GET("/payment/success", middleware.UserAuth(), controllers.RazorpaySuccess) //
		user.GET("/success", middleware.UserAuth(), controllers.Success)                 //

		user.GET("/payment/showwallet", middleware.UserAuth(), controllers.ShowWallet)
		user.GET("/invoice", middleware.UserAuth(), controllers.PurchaseInvoice)
		user.GET("/invoice/download", middleware.UserAuth(), controllers.DownloadInvoice)
	}

}
