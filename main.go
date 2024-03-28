package main

import (
	"log"

	"github.com/ShebinSp/e-cart/initializers"
	"github.com/ShebinSp/e-cart/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	_ "github.com/ShebinSp/e-cart/docs"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @Title	E-commerce
// @version 1.0
// @description An e-commerce site in Go using Gin framework

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @host localhost:1111
// @BasePath /
// @schemes http
func main() {

	// Load environment variables from .env file
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error loading env file")
	}

	// Connect to database
	initializers.ConnectToDb()

	// Initialize Gin router
	r := gin.Default()
	r.LoadHTMLGlob("templates/*.html")

	url := ginSwagger.URL("http://localhost:1111/swagger/doc.json") // The url pointing to API definition
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, url))

	// Setup routes for user and admin
	routes.UserRoutes(r)
	routes.AdminRoutes(r)

	// Start the server
	r.Run(":1111")
}
