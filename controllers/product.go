package controllers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ShebinSp/e-cart/initializers"
	"github.com/ShebinSp/e-cart/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)
const Dis uint = 10

func AddProduct(c *gin.Context) {
	var product models.Product

	if err := c.Bind(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{ // 400
			"error":   err,
			"err.msg": "Failed to bind the data",
		})
		fmt.Println("Data binding err: ", err)
		return
	}

	db := initializers.DB
	var stock int
	var currentStock = int(product.Stock)

	if currentStock < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error":   "Please check Stock value",
			"Message": "Can not add negative values as quantity",
		})
		c.Abort()
		return
	}

	result := db.Find(&product, "product_name = ?", product.ProductName).Select("stock").Scan(&stock)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{ // 404
			"error": result.Error.Error(),
		})
		return
	}
	if result.RowsAffected == 0 {

		if product.SpecialPrice <= 0 {
			product.SpecialPrice = product.ActualPrice - Dis
		}
	
		product.Offer_details = "No offer available"
	
		fmt.Println("sprice:",product.SpecialPrice)///
		if product.ActualPrice <= product.SpecialPrice {
			c.JSON(http.StatusBadRequest, gin.H{
				"Caution": "Special price should be less than actual price",
			})
			return
		}
		fmt.Println("Rows Affected == 0")
		result := db.Create(&product)
		if result.Error != nil {
			c.JSON(http.StatusNotFound, gin.H{
				"Error": result.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"Message":    "Successfully Added the Product",
			"Product id": product.ProductId,
		})
	} else {

		totalStock := int(currentStock) + stock

		result := db.Model(&models.Product{}).Where("product_name = ?", product.ProductName).Updates(models.Product{
			Stock: totalStock,
		})
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": result.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"Message": "Product already exist, Updating Stock",
		})
	}
}

// --- Add Image ---\\
func AddImage(c *gin.Context) {
	imagepath, _ := c.FormFile("image")
	//Image of the product should be added once a product added. Or add product id manually to add pic
	//var ProductId uint
	pr_id := c.PostForm("product_id")
	productId, _ := strconv.Atoi(pr_id)
	var lastProduct models.Product

	db := initializers.DB
	db.First(&lastProduct).Select("product_id").Where("product_id = ?", pr_id).Scan(&productId)

	var product models.Product
	result := db.First(&product, productId)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": result.Error.Error(),
		})
		return
	}

	extension := filepath.Ext(imagepath.Filename)
	image := uuid.New().String() + extension
	c.SaveUploadedFile(imagepath, "./public/images"+image)

	imagedata := models.Image{
		Image:     image,
		Productid: uint(productId),
	}
	result = db.Create(&imagedata)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": result.Error.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Message": "Image Added Successfully",
	})
}

// --- Edit Product --- \\
func EditProduct(c *gin.Context) {
	err := CheckOfferExpiry()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": err.Error(),
		})
		return
	}
	type editProduct struct {
		ProductId      uint   `json:"product_id"`
		NewProductName string `json:"new_name"`
		ActualPrice    uint   `json:"new_actual_price"`
		SpecialPrice   uint   `json:"new_special_price"`
		Stock          int    `json:"new_stock"`
		Description    string `json:"new_description"`
	}
	var editproduct editProduct
	var product models.Product

	if c.Bind(&editproduct) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Data binding error",
		})
		return
	}
	// if editproduct.ActualPrice != 0 {
	// 	editproduct.SpecialPrice =
	// }

	var count int64
	db := initializers.DB
	result := db.Find(&product).Where("product_id = ?", editproduct.ProductId).Count(&count)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": result.Error.Error(),
		})
		return
	}

	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"Message":      "Product doesn't exist",
			"Please go to": "/addproduct",
		})
		return
	}
	if editproduct.ActualPrice <= editproduct.SpecialPrice {
		c.JSON(http.StatusBadRequest, gin.H{
			"Caution": "Special price should be less than actual price",
		})
		return
	}

	result = db.Model(&models.Product{}).Where("product_id = ?", editproduct.ProductId).Updates(models.Product{
		ProductId:    editproduct.ProductId,
		ProductName:  editproduct.NewProductName,
		ActualPrice:  editproduct.ActualPrice,
		SpecialPrice: editproduct.SpecialPrice,
		Stock:        editproduct.Stock,
		Description:  editproduct.Description,
	})
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Message": "Successfully updated the Product",
	})
}

// --- Delete Product ---\\
func DelProduct(c *gin.Context) {
	type delProduct struct {
		DelProduct string //`json:"delete_product"`
	}
	var delproduct delProduct
	var product models.Product

	if c.Bind(&delproduct) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "JSON Data binding failed",
		})
		return
	}
	db := initializers.DB
	result := db.Where("product_name = ?", delproduct.DelProduct).Delete(&product)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": result.Error.Error(),
		})
		return
	}
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"Message": "Product doesn't exist",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Success": "Product Deleted Successfully",
	})
}

// --- View Products ---\\
type Datas struct {
	Product_ID    uint
	Product_name  string
	Description   string
	Stock         int
	ActualPrice   uint
	SpecialPrice  uint
	Offer_details string
	Brand_name    string
	Category_name string
	Image         string
}
// @Summary View Products
// @Description View a list of products with pagination support
// @Tags Products, Users
// @Security BearerToken
// @Produce json
// @Param limit query integer false "Number of items per page"
// @Param offset query integer false "Page number"
// @Success 200 {object} map[string]interface{} "List of products"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/products/view [get]
func ViewProducts(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	offset, _ := strconv.Atoi(c.Query("offset"))

	var products []Datas
	var count int64

	db := initializers.DB
	query := "SELECT products.product_id, products.product_name, products.description, products.stock, products.actual_price,products.special_price, products.offer_details, brands.brand_name, categories.category_name, images.image FROM products LEFT JOIN brands ON products.brand_id = brands.id LEFT JOIN categories ON products.category_id = categories.id LEFT JOIN images ON products.product_id = images.productid"
	countQuery := "SELECT COUNT(*) FROM products LEFT JOIN brands ON products.brand_id = brands.id LEFT JOIN categories ON products.category_id = categories.id"

	_ = db.Raw(countQuery).Count(&count)
	if limit != 0 && offset != 0 {
		query = fmt.Sprintf("%s LIMIT %d OFFSET %d", query, limit, offset)
	} else if limit != 0 {
		query = fmt.Sprintf("%s LIMIT %d", query, limit)
	} else if offset != 0 {
		query = fmt.Sprintf("%s OFFSET %d", query, offset)
	}

	result := db.Raw(query).Scan(&products)
	if result.Error != nil {
		err := map[string]interface{}{
			"Error": result.Error.Error(),
		}
		c.JSON(http.StatusNotFound, err)
		return
	}

	oK := map[string]interface{}{
		"Total Products Found": count,
		"Products":             products,
	}
	c.JSON(http.StatusOK, oK)
}

// func ViewProducts(c *gin.Context) {

// 	type Datas struct {
// 		Product_ID    uint
// 		Product_name  string
// 		Description   string
// 		Stock         int
// 		ActualPrice   uint
// 		SpecialPrice  uint
// 		Brand_name    string
// 		Category_name string
// 		Offer_details string
// 		Images        []string
// 	}

// 	limit, _ := strconv.Atoi(c.Query("limit"))
// 	offset, _ := strconv.Atoi(c.Query("offset"))

// 	var product []Datas
// 	var count int64

// 	err := CheckOfferExpiry()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"Error": err.Error(),
// 		})
// 		return
// 	}

// 	db := initializers.DB
// 	r := db.Table("products").
// 		Select(`products.product_id, products.product_name, products.description, products.stock, products.actual_price, products.special_price, products.offer_details, brands.brand_name, categories.category_name, ARRAY_AGG(images.image) as images`).
// 		Joins("LEFT JOIN brands ON products.brand_id = brands.id").
// 		Joins("LEFT JOIN categories ON products.category_id = categories.id").
// 		Joins("LEFT JOIN images ON products.product_id = images.productid").
// 		Group("products.product_id, brands.brand_name, categories.category_name").
// 		Offset(offset).
// 		Limit(limit).
// 		Scan(&product)
// 	if r.Error != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{
// 			"Error": r.Error.Error(),
// 		})
// 		return
// 	}
// 	count = r.RowsAffected

// 	// for i := range product {
// 	// 	imageUrls := strings.Split(product[i].Images, ",")
// 	// }

// 	c.JSON(http.StatusOK, gin.H{
// 		"Total Products Found": count,
// 		"Products":             product,
// 	})
// }

// --- Product details --- \\
type Product_Details struct {
	ProductId    uint
	Product_Name string
	ActualPrice  int
	SpecialPrice uint
	Stock        int
	Description  string
	Image        string
}
// @Summary Get Product Details
// @Description Get details of a specific product by its name
// @Tags Products, Users
// @Security BearerToken
// @Produce json
// @Param product_name query string true "Product name"
// @Success 200 {object} map[string]interface{} "Product details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/products/details [get]
func ProductDetails(c *gin.Context) {

	productName := c.Query("product_name")
	ProductName := strings.ToUpper(productName)

	var productDetails Product_Details
	//	var product []ProductDetails
	err := CheckOfferExpiry()
	if err != nil {
		err := map[string]interface{}{
			"Error": err.Error(),
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	db := initializers.DB

	result := db.Table("products").Select("product_id", "product_name", "actual_price", "special_price", "stock", "description").Where("product_name = ?", ProductName).Find(&productDetails)

	if result.Error != nil {
		err := map[string]interface{}{
			"Error": result.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if result.RowsAffected == 0 {
		err := map[string]interface{}{
			"Message": "Product not found",
		}
		c.JSON(http.StatusNotFound, err)
		return
	}
	result = db.Table("images").Select("image").Where("productid = ?", productDetails.ProductId).Scan(&productDetails.Image)
	if result.Error != nil {
		err := map[string]interface{}{
			"Error": result.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	//	product = append(product, productDetails)
	oK := map[string]interface{}{
		"Product details": productDetails,
	}
	c.JSON(http.StatusOK, oK)
}
