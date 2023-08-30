package routes

import (
	"github.com/ShebinSp/e-cart/controllers"
	"github.com/ShebinSp/e-cart/middleware"
	"github.com/gin-gonic/gin"
)

func AdminRoutes(c *gin.Engine) {
	admin := c.Group("/admin")
	{
		admin.POST("/signup", controllers.AdminSignup)
		admin.POST("/login", controllers.AdminLogin)
		admin.GET("/signout",middleware.AdminAuth(), controllers.AdminSignout)

		admin.PATCH("usermanagement/manageblock",middleware.AdminAuth(),controllers.BlockUser)
		admin.GET("usermanagement/viewusers",middleware.AdminAuth(),controllers.ListUsers)

		admin.POST("/addcategory",middleware.AdminAuth(),controllers.AddCategories)
		admin.DELETE("/deletecategory",middleware.AdminAuth(),controllers.DelCategory)
		admin.PUT("/editcategory",middleware.AdminAuth(),controllers.EditCategory)

		admin.POST("/addbrand",middleware.AdminAuth(),controllers.AddBrand)
		admin.PATCH("/editbrand",middleware.AdminAuth(),controllers.EditBrand)
		admin.DELETE("/deletebrand",middleware.AdminAuth(),controllers.DeleteBrand)

		admin.POST("/addproduct",middleware.AdminAuth(),controllers.AddProduct)
		admin.GET("products/view", middleware.UserAuth(), controllers.ViewProducts)
		admin.POST("/addimage",middleware.AdminAuth(),controllers.AddImage)
		admin.PATCH("/updateproduct",middleware.AdminAuth(),controllers.EditProduct)
		admin.DELETE("products/delete",middleware.AdminAuth(),controllers.DelProduct)

		admin.POST("/addcoupon",middleware.AdminAuth(),controllers.AddCoupon)//

		admin.GET("/show-orders", middleware.AdminAuth(), controllers.ShowOrders)
		admin.PUT("/change-orderstatus", middleware.AdminAuth(), controllers.ChangeOrderStatus)
		admin.PATCH("/cancel-order", middleware.AdminAuth(),controllers.CancelOrder)

		admin.GET("/dashboard", middleware.AdminAuth(),controllers.AdminDashBoard)
		admin.POST("/applyoffers", middleware.AdminAuth(), controllers.ApplyOffers)///
		admin.GET("/showoffers", middleware.AdminAuth(), controllers.ShowOffers)///
		admin.PATCH("/canceloffers", middleware.AdminAuth(), controllers.CancelOffer)///

		admin.POST("/order/salesreport",middleware.AdminAuth(), controllers.SalesReport)
		admin.GET("/order/Salesreport/download/excel",middleware.AdminAuth(),controllers.DownloadExcel)
		admin.GET("/order/salesreport/download/pdf",middleware.AdminAuth(), controllers.DownloadPdf)

	}
	
}