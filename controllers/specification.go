package controllers

import (
	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/ShebinSp/e-cart/initializers"
	"github.com/ShebinSp/e-cart/models"
	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	"github.com/tealeg/xlsx"
	"gorm.io/gorm"
)

func getId(email string, db *gorm.DB) uint {
	var user models.User
	var id uint

	// Getting user id from db using email
	db.Where("email = ?", email).Select("id").Find(&user).Row().Scan(&id)
	return id
}

func AddCategories(c *gin.Context) {
	type Data struct {
		CategoryName string
	}
	var category Data
	var CategoryData models.Category
	if err := c.Bind(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Could not bind the JSON data",
		})
		return
	}
	db := initializers.DB
	var count int64
	result := db.Find(&CategoryData, "category_name = ?", category.CategoryName).Count(&count)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": result.Error.Error(),
		})
		return
	}
	if count == 0 {
		createData := models.Category{
			Category_name: category.CategoryName,
		}
		result := db.Create(&createData)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": result.Error.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "New Category " + category.CategoryName + "  created",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Category already exist",
		})
	}
}

func DelCategory(c *gin.Context) {

	type Data struct {
		CategoryName string
	}

	var delCategory Data
	var category models.Category

	if c.Bind(&delCategory) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Could not bind the JSON data",
		})
		return
	}
	db := initializers.DB
	result := db.Where("category_name = ?", delCategory.CategoryName).Delete(&category)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": result.Error.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"Message": "Category deleted successfully",
	})
}

func EditCategory(c *gin.Context) {

	type Data struct {
		CategeryName string `json:"current_category_name"`
		NewName      string `json:"update_to"`
	}

	var editcategory Data
	var category models.Category

	if err := c.Bind(&editcategory); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Unable to bind the JSON data",
		})
		return
	}

	var count int64

	db := initializers.DB
	db.Find(&category, "Category_name = ?", editcategory.CategeryName).Count(&count)
	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Category doesn't exist",
		})
		return
	}
	result := db.Model(&category).Where("category_name = ?", editcategory.CategeryName).Update("category_name", editcategory.NewName)

	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Message": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Message": "Category updated successfully",
	})
}

func AddBrand(c *gin.Context) {
	type data struct {
		BrandName string
	}
	var newBrand data
	var brands models.Brand
	if c.Bind(&newBrand) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Eroor": "Data binding failed",
		})
		return
	}
	db := initializers.DB
	var count int64
	result := db.Find(&brands, "brand_name = ?", newBrand.BrandName).Count(&count)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": result.Error.Error(),
		})
		return
	}

	if count == 0 {
		createData := models.Brand{
			Brand_name: newBrand.BrandName,
		}
		reuslt := db.Create(&createData)
		if reuslt.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": reuslt.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"Message": "New Brand " + newBrand.BrandName + " Added Successfully",
		})
	} else {
		c.JSON(http.StatusOK, gin.H{
			"Message": "Brand " + newBrand.BrandName + " Already exist",
		})
	}
}

func EditBrand(c *gin.Context) {
	type data struct {
		BrandName string `json:"brand"`
		NewName   string `json:"update_to"`
	}

	var editBrand data
	var brand models.Brand

	if c.Bind(&editBrand) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Unable to bind JSON data",
		})
		return
	}

	db := initializers.DB
	r := db.Find(&brand, "brand_name = ?", editBrand.BrandName)
	if r.RowsAffected == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"Message": "Brand does not exist",
		})
		return
	}

	result := db.Model(&brand).Where("brand_name = ?", editBrand.BrandName).Update("brand_name", editBrand.NewName)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Message": "Brand Edited Successfully",
	})
}

func DeleteBrand(c *gin.Context) {

	type data struct {
		BrandName string
	}

	var delBrand data
	var brand models.Brand

	if c.Bind(&delBrand) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": "Unable to bind the JSON data",
		})
		return
	}

	db := initializers.DB

	result := db.Find(&brand, "brand_name = ?", delBrand.BrandName)
	if result.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error": result.Error.Error(),
		})
	}
	if result.RowsAffected != 0 {
		result := db.Where("brand_name = ?", delBrand.BrandName).Delete(&brand)
		if result.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": result.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"Message": "Brand deleted Successfully",
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{
			"Message": "Brand doesn't exist",
		})
		return
	}

}

//---------------CART----------------\\
type productData struct {
	Product_id uint 
	Quantity   uint 
}
// @Summary Add Product to Cart
// @Description Add a product to the user's cart
// @Tags Cart, Users
// @Security BearerToken
// @Produce json
// @Param input body productData true "Product data"
// @Success 200 {object} map[string]interface{} "Product added to cart successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Product does not exist" 
// @Failure 404 {object} map[string]interface{} "Out of Stock" 
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/profile/addtocart [post]
func AddToCart(c *gin.Context) {
	_ = CheckOfferExpiry()
	var bindData productData
	var productData models.Product

	if err := c.Bind(&bindData); err != nil {
		err := map[string]interface{}{
			"Error": "Could not bind the JSON data",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	email := c.GetString("user")
	db := initializers.DB
	id := getId(email, db)

	// Checking the product id exist or not
	result := db.First(&productData, bindData.Product_id)
	if result.Error != nil {
		err := map[string]interface{}{
			"Message": "Product does not exist",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// Checking product stock quantity
	if bindData.Quantity > uint(productData.Stock) {
		err := map[string]interface{}{
			"Message":           "Out of Stock",
			"Available Stock: ": productData.Stock,
		}
		c.JSON(http.StatusNotFound, err)
		return
	}

	var quantyti uint
	var Price uint

	//Checking the product_id and user_id in the carts table
	err := db.Table("carts").Where("product_id = ? AND userid = ?", bindData.Product_id, id).
		Select("quantity", "total_price").Row().Scan(&quantyti, &Price)

	if err != nil {
		totalPrice := productData.SpecialPrice * bindData.Quantity
		cartItems := models.Cart{
			Productid:  bindData.Product_id,
			Quantity:   bindData.Quantity,
			Price:      productData.SpecialPrice,
			TotalPrice: totalPrice,
			Userid:     uint(id),
		}

		// Creating the table carts
		result := db.Create(&cartItems)
		if result.Error != nil {
			err := map[string]interface{}{
				"Error": result.Error.Error(),
			}
			c.JSON(http.StatusBadRequest, err)
			return
		}
		oK := map[string]interface{}{
			"Message": "Products Added to the Cart Successfully",
		}
		c.JSON(http.StatusOK, oK)
	} else {
		// Calculating the total quantity and total price
		totalQuantity := quantyti + bindData.Quantity
		totalPrice := productData.SpecialPrice * totalQuantity
		// Updating the quantity and the total price to the carts
		result = db.Model(&models.Cart{}).Where("product_id = ? AND userid = ?", bindData.Product_id, id).Updates(map[string]interface{}{"quantity": totalQuantity, "total_price": totalPrice})
		if result.Error != nil {
			err := map[string]interface{}{
				"Error": result.Error.Error(),
			}
			c.JSON(http.StatusBadRequest, err)
			return
		}

		oK := map[string]interface{}{
			"Message": "Quantity added Successfully",
		}
		c.JSON(http.StatusOK,oK)
	}

}

type cartdata struct {
	Id           int
	Product_name string
	Quantity     uint
	Total_price  uint
	Image        string
	Price        string
}
// @Summary View Cart
// @Description View items in the user's cart
// @Tags Cart, Users
// @Security BearerToken
// @Produce json
// @Success 200 {object} map[string]interface{} "Cart items"
// @Failure 404 {object} map[string]interface{} "Cart is empty" 
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/profile/viewcart [get]
func ViewCart(c *gin.Context) {
	_ = CheckOfferExpiry()

	email := c.GetString("user")
	var datas []cartdata

	db := initializers.DB
	id := getId(email, db)
	var count int64
	_ = db.Find(&models.Cart{}).Where("userid = ?", id).Count(&count)

	result := db.Table("carts").
		Select("products.Product_name, carts.id, carts.quantity, carts.price, carts.total_price, images.image").
		Joins("INNER JOIN products ON products.product_id=carts.productid").
		Joins("INNER JOIN images ON images.productid=carts.productid").
		Where("userid = ?", id).Scan(&datas)

	fmt.Println("Datas:", datas)
	if result.Error != nil {
		err := map[string]interface{}{
			"Error": result.Error.Error(),
		}
		c.JSON(http.StatusNotFound, err)
		return
	}
	if datas != nil {
		oK := map[string]interface{}{
			"Total items": count,
			"Cart items": datas,
		}
		c.JSON(http.StatusOK, oK)
	} else {
		err := map[string]interface{}{
			"Message": "Cart is empty",
		}
		c.JSON(http.StatusNotFound, err)
	}
}

type Cartdata struct {
	Product_name string
	Quantity     uint
	Total_price  uint
	Image        string
	Price        string
}

func CartItems(email string) ([]Cartdata, error) {

	var datas []Cartdata

	db := initializers.DB
	id := getId(email, db)
	//	var count int64
	//	_ = db.Find(&models.Cart{}).Where("userid = ?", id).Count(&count)

	result := db.Table("carts").
		Select("products.Product_name, carts.quantity, carts.price, carts.total_price, images.image").
		Joins("INNER JOIN products ON products.product_id=carts.productid").
		Joins("INNER JOIN images ON images.productid=carts.productid").
		Where("userid = ?", id).Scan(&datas)

	if result.Error != nil {
		return nil, result.Error
	}
	//	fmt.Println("cart:", datas)
	return datas, nil
}

// @Summary Delete Cart Item
// @Description Delete an item from the user's cart
// @Tags Cart, Users
// @Security BearerToken
// @Produce json
// @Param id query int true "Cart item ID"
// @Success 200 {object} map[string]interface{} "Cart item deleted successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 400 {object} map[string]interface{} "String conversion failed"
// @Failure 400 {object} map[string]interface{} "Cart does not exist"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/profile/deletecart [delete]
func DeleteCart(c *gin.Context) {
	//	id := c.Param("id")
	email := c.GetString("user")
	id, err := strconv.Atoi(c.Query("id"))
	if err != nil {
		err := map[string]interface{}{
			"Error": "String convertion failed",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	db := initializers.DB
	userid := getId(email, db)

	result := db.Exec("delete from carts WHERE id = ? AND userid = ?", id, userid)
	if result.RowsAffected == 0 {
		err := map[string]interface{}{
			"Message": "Cart does not exist",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	if result.Error != nil {
		err := map[string]interface{}{
			"Error": result.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	oK := map[string]interface{}{
		"Message": "Cart Deleted Successfully",
	}
	c.JSON(http.StatusOK, oK)
}

// @Summary Filter Products by Category
// @Description Get a list of products filtered by category ID
// @Tags Products, Users
// @Security BearerToken
// @Produce json
// @Param cid query integer true "Category ID"
// @Success 200 {object} map[string]interface{} "List of products"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Category doesn't exist"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/filterbycategory [get]
func FilterByCategory(c *gin.Context) {
	cid, err := strconv.Atoi(c.Query("cid"))
	if err != nil {
		err := map[string]interface{}{
			"Error": "Error in string conversion",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	db := initializers.DB

	var product models.Product
	r := db.Table("products").Select("product_id, product_name, description, price").Where("category_id = ?", cid).Scan(&product).Error
	if r != nil {
		err := map[string]interface{}{
			"Error": "Category doesn't exist",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	var images string
	var image models.Image
	result := db.Find(&image).Select("image").Where("productid = ?", product.ProductId).Scan(&images)
	if result.RowsAffected == 0 {
		images = "Image not available"
	}
	oK := map[string]interface{}{
		"Product name": product.ProductName,
		"Discription":  product.Description,
		"Price":        product.SpecialPrice,
		"Image":        images,
	}
	c.JSON(http.StatusOK, oK)
}
type Search struct {
	Search string
}
// @Summary Search Products
// @Description Search for products by name
// @Tags Products, Users
// @Security BearerToken
// @Accept json
// @Produce json
// @Param search body Search true "Search query"
// @Success 200 {object} map[string]interface{} "List of matching product"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/products/search [post]
func SearchProduct(c *gin.Context) {
	
	type Datas struct {
		Product_ID    uint
		Product_name  string
		Description   string
		Stock         int
		Price         int
		Brand_name    string
		Category_name string
		//	Image         string
	}

	var search Search

	if c.Bind(&search) != nil {
		err := map[string]interface{}{
			"Error": "Could not bind the JSON Data",
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	var products Datas
	db := initializers.DB

	r := db.Table("products").Where("product_name = ? ", search.Search).
		Joins("JOIN brands ON products.brand_id = brands.id").
		Joins("JOIN categories ON products.category_id = categories.id").
		Select("product_id,product_name,price,description,stock,brands.brand_name,categories.category_name").
		Scan(&products).Error
	if r != nil {
		err := map[string]interface{}{
			"Error": r.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	ok := map[string]interface{}{
		"products": products,
	}
	c.JSON(http.StatusOK, ok)
}

func AddCoupon(c *gin.Context) {
	type data struct {
		CouponCode    string
		Year          uint
		Month         uint
		Day           uint
		DiscountPrice float64
		Expired       time.Time
	}

	var userEnterData data
	var couponData []models.Coupon
	db := initializers.DB

	if c.Bind(&userEnterData) != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Could not bind the JSON Data",
		})
		return
	}
	specificTime := time.Date(int(userEnterData.Year), time.Month(userEnterData.Month), int(userEnterData.Day), 0, 0, 0, 0, time.UTC)
	userEnterData.Expired = specificTime

	r := db.First(&couponData, "coupon_code = ?", userEnterData.CouponCode)
	if r.Error != nil {
		Data := models.Coupon{
			CouponCode:    userEnterData.CouponCode,
			DiscountPrice: userEnterData.DiscountPrice,
			Expired:       specificTime,
		}
		r := db.Create(&Data)
		if r.Error != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": r.Error.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"Message": userEnterData,
			"Success": "Coupon added successfully",
		})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"Message": "Coupon already exist",
		})
	}
}

func SalesReport(c *gin.Context) {
	generate := c.Query("generate")
	generate = strings.ToLower(generate)

	//for specific to - from date
	if generate == "specific" {
		startDate := c.Query("from")
		endDate := c.Query("to")
		from, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": "Invalid start date",
			})
			return
		}

		to, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": "Invalid start date",
			})
			return
		}
		now := time.Now()
		from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, now.Location())
		to = time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 999999999, now.Location())		
		fmt.Println("to:",to)
		fmt.Println("from: ",from)
		if generateSalesReport(from, to) != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": err.Error(),
			})
			return
		}
	} else if generate == "monthly" {
		year, err := strconv.Atoi(c.Query("year"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": err.Error(),
			})
			return
		}
		month_int, errr := strconv.Atoi(c.Query("month"))
		if errr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": err.Error(),
			})
			return
		}
		month := time.Month(month_int)
		from := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(0, 1, 0).Add(-time.Nanosecond)
		err = generateSalesReport(from, to)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": err.Error(),
			})
			return
		}
	} else if generate == "yearly" {
		year, err := strconv.Atoi(c.Query("year"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": err.Error(),
			})
			return
		}
		from := time.Date(year, time.January, 1, 0, 0, 0, 0, time.UTC)
		to := from.AddDate(1, 0, 0).Add(-time.Nanosecond)
		err = generateSalesReport(from, to)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"Error": err.Error(),
			})
			return
		}
	} else if generate == "daily" {
		day, err := strconv.Atoi(c.Query("day"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": err.Error(),
			})
			return
		}
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), day, 0, 0, 0, 0, now.Location())
		to := time.Date(today.Year(), today.Month(), day, 23, 59, 59, 999999999, today.Location())
		fmt.Println("today: ", today)
		fmt.Println("to: ", to)
		err = generateSalesReport(today, to)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error": err.Error(),
			})
			return
		}
	}
	c.JSON(http.StatusOK, gin.H{
		"Message": "Sales report created successfully",
	})
	c.HTML(http.StatusOK, "SalesReport.html", gin.H{})
}

func generateSalesReport(from time.Time, to time.Time) error {
	// Fetching the data from the order details table as per the date
	var orderDetail []models.OrderDetails
	db := initializers.DB

	r := db.Preload("Product").Preload("Payment").Where("created_at BETWEEN ? AND ?", from, to).Find(&orderDetail)
	if r.Error != nil {
		return r.Error
	}

	file := excelize.NewFile()

	// Create a new sheet
	SheetName := "Sheet1"
	index := file.NewSheet(SheetName)

	// Set the value of headers
	file.SetCellValue(SheetName, "A1", "Order Date")
	file.SetCellValue(SheetName, "B1", "Order ID")
	file.SetCellValue(SheetName, "C1", "Product Name")
	file.SetCellValue(SheetName, "D1", "Qty")
	file.SetCellValue(SheetName, "E1", "Price")
	file.SetCellValue(SheetName, "F1", "Total")
	file.SetCellValue(SheetName, "G1", "Pay Method")
	file.SetCellValue(SheetName, "H1", "Status")
	// Set the value of the cell
	for i, report := range orderDetail {
		row := i + 2
		file.SetCellValue(SheetName, fmt.Sprintf("A%d", row), report.CreatedAt.Format("02/01/2006"))
		file.SetCellValue(SheetName, fmt.Sprintf("B%d", row), report.Orderid)
		file.SetCellValue(SheetName, fmt.Sprintf("C%d", row), report.Product.ProductName)
		file.SetCellValue(SheetName, fmt.Sprintf("D%d", row), report.Quantity)
		file.SetCellValue(SheetName, fmt.Sprintf("E%d", row), report.Product.SpecialPrice)
		file.SetCellValue(SheetName, fmt.Sprintf("F%d", row), report.Payment.TotalAmount)
		file.SetCellValue(SheetName, fmt.Sprintf("G%d", row), report.Payment.PaymentMethod)
		file.SetCellValue(SheetName, fmt.Sprintf("H%d", row), report.Payment.Status)
	}

	// Set active sheet of the workbook
	file.SetActiveSheet(index)

	// Save the Excel file with the name "test.xlsx"
	if err := file.SaveAs("./public/SalesReport.xlsx"); err != nil {
		fmt.Println("Excel file save ERROR:", err)
	}
	// Convert excel to pdf
	ConvertExcelToPdf()

	return nil
}

func ConvertExcelToPdf() {
	xlFile, err := xlsx.OpenFile("./public/SalesReport.xlsx")
	if err != nil {
		fmt.Println("xlsx file open error:", err)
		//return
	}

	// Create a new pdf document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 10)
	// err := pdf.OutputFileAndClose("hello.pdf")

	// Converting each cell in the excel file to a pdf cell
	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				if cell.Value == "" { // If there is any empty cell values then skiping that
					continue
				}
				pdf.Cell(25, 10, cell.Value)
			}
			pdf.Ln(-1)
		}
	}

	// Save the PDF document
	err = pdf.OutputFileAndClose("./public/SalesReport.pdf")
	if err != nil {
		fmt.Println("PDF saving error:", err)
	}
	fmt.Println("PDF saved successfully")
}

func DownloadExcel(c *gin.Context) {
	c.Header("Content-Disposition", "attachment; filename=SalesReport.xlsx")
	c.Header("Content-Type", "application/xlsx")
	c.File("./public/SalesReport.xlsx")
}

func DownloadPdf(c *gin.Context) {
	c.Header("Content-Disposition", "attachment; filename=SalesReport.pdf")
	c.Header("Content-Type", "application/pdf")
	c.File("./public/SalesReport.pdf")
}

// Invoice

// Template for creating pdf
const invoiceTemplate = `
<style>
	body{
		background-color: white;
	}
	table{
		border: 1px solid black;
		border-collapse: collapse;
	}
	  th{
		border: 1px solid black;
		border-collapse: collapse;
		padding-right: 15px;
		padding-left: 15px;
	 }
	 td{
		border: 1px solid black;
		border-collapse: collapse;
		padding-right: 15px;
		padding-left: 15px;
	 }
	 hr{
		color:solid black;
	 }
</style>

<b> TAX INVOICE- </b>
Order ID : {{.OrderId}}<br>
Order Date: {{.Date}} <br><hr>
Name : {{.Name}} <br>
Email : {{.Email}} <br>
Billing Address<br>
{{range .Address}}

Phone number : {{.Phoneno}} <br>
House name : {{.Housename}} <br>
Area :{{.Area}} <br>
Landmark : {{.Landmark}} <br>
City : {{.City}} <br>
Pincode : {{.Pincode}} <br>
District : {{.District}} <br>
State : {{.State}} <br>
{{end}}
<hr>
Payment method : {{.PaymentMethod}} <br>
<hr>
 <br>
 
<table>
	<tr>
		<th>Product</th>
		<th>Description</th>
		<th>Qty</th>
		<th>Price</th>
		<th>Discount</th>
		<th>Total Price </th>
	</tr>
	{{range .Items}}
	<tr>
		<td>{{.Product}}</td>
		<td>{{.Description}}</td>
		<td>{{.Qty}}</td>
		<td>{{.Price}}</td>
		<td>{{.Discount}}</td>
		<td>{{.Total}}</td>
	</tr>
	{{end}}
</table>

<br><hr>
Total Amount : {{.TotalAmount}} <br><hr>`

type Invoice struct {
	Name          string
	Email         string
	PaymentMethod string
	TotalAmount   int64
	Date          string
	OrderId       uint
	Address       []Address
	Items         []Item
}
type Address struct {
	Phoneno   uint
	Housename string
	Area      string
	Landmark  string
	City      string
	Pincode   uint
	District  string
	State     string
}
type Item struct {
	Product     string
	Description string
	Qty         uint
	Price       uint
	Discount    uint
	Total       uint
}

// @Summary Generate Purchase Invoice
// @Description Generate a purchase invoice for a user's order
// @Tags Invoice, Orders, Users
// @Accept json
// @Security BearerToken
// @Produce json
// @Success 200 {string} html "HTML response"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Router /user/invoice [get]
func PurchaseInvoice(c *gin.Context) {
	email := c.GetString("user")
	db := initializers.DB
	id := getId(email, db)

	var user models.User
	var Payment models.Payment
	var orderData models.OrderDetails
	var address models.Address
	var orderItem models.OrderItem

	// Fetching the data from table OrderItems useing the id
	r := db.Last(&orderItem).Where("user_id_no = ?", id)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// Fetching data from order_details using userid and order id for fetching the order_item id
	r = db.Last(&orderData).Where("userid = ? AND order_item_id = ?", id, orderItem.ID)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// Fetching user data
	r = db.First(&user, id)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// Fetching user address
	r = db.First(&address, orderData.Addressid)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// Fetching payment details
	r = db.Last(&Payment, "userid = ?", id)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// Fetching the product data from products
	var products []models.Product //models.product
	err := db.Joins("JOIN order_details ON products.product_id = order_details.productid").Where("order_details.order_item_id = ?", orderData.OrderItemId).
		Find(&products).Error
	if err != nil {
		err := map[string]interface{}{
			"Error": "Somthing went wrong",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	// To list product details from products
	items := make([]Item, len(products))
	var qty, coupon []uint

	r = db.Table("order_details").Select("quantity").Where("userid = ? AND order_item_id = ?", id, orderItem.ID).Find(&qty)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	r = db.Table("order_details").Select("couponid").Where("userid = ? AND order_item_id = ?", id, orderItem.ID).Find(&coupon)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	for idx, c := range coupon {
		var couponDis uint
		if c > 0 {
			_ = db.Table("coupons").Select("discount_price").Where("id = ?", c).Scan(&couponDis)
		}
		coupon[idx] = couponDis
	}
	fmt.Println("Coupon:", coupon)
	for i, data := range products {

		items[i] = Item{
			Product:     data.ProductName,
			Price:       data.SpecialPrice,
			Description: data.Description,
			Qty:         qty[i],
			Discount:    coupon[i],
			Total:       qty[i]*data.SpecialPrice - coupon[i],
		}
	}

	timeString := Payment.Date.Format("02-01-2006")

	// Excuting the template invoice
	invoice := Invoice{
		Name:          user.First_Name,
		Date:          timeString,
		Email:         user.Email,
		OrderId:       orderItem.ID,
		PaymentMethod: Payment.PaymentMethod,
		TotalAmount:   int64(Payment.TotalAmount),
		Address: []Address{
			{
				Phoneno:   address.Phone,
				Housename: address.HouseName,
				Area:      address.Area,
				Landmark:  address.Landmark,
				City:      address.City,
				Pincode:   address.Pincode,
				District:  address.District,
				State:     address.State,
			},
		},
		Items: items,
	}

	tmplet, err := template.New("invoice").Parse(invoiceTemplate)
	if err != nil {
		err := map[string]interface{}{
			"Error": err.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	var buf bytes.Buffer
	err = tmplet.Execute(&buf, invoice)
	if err != nil {
		err := map[string]interface{}{
			"Error": err.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	cmd := exec.Command("wkhtmltopdf", "-", "./public/invoice.pdf")
	cmd.Stdin = &buf
	err = cmd.Run()
	if err != nil {
		err := map[string]interface{}{
			"Error": err.Error(),
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.HTML(http.StatusOK, "invoice.html", gin.H{})
}

// @Summary Download Purchase Invoice
// @Description Download the generated purchase invoice as a PDF
// @Tags Invoice, Users
// @Produce application/pdf
// @Success 200 "PDF file"
// @Router /user/invoice/download [get]
func DownloadInvoice(c *gin.Context) {
	c.Header("content-Disposition", "attachment; filename=invoice.pdf")
	c.Header("Content-Type", "application/pdf")
	c.File("invoice.pdf")
}
