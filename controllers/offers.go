package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/ShebinSp/e-cart/initializers"
	"github.com/ShebinSp/e-cart/models"
	"github.com/gin-gonic/gin"
)

// Apply offers
func ApplyOffers(c *gin.Context) {
	type OfferDetails struct {
		OfferName     string    `json:"offer_name" gorm:"not null"`
		OfferType     string    `json:"offer_applied_on" gorm:"not null"` // product,brand,category should choose one to apply offer
		OfferDiscount uint      `json:"offer_discount"`
		Expired_on    time.Time `json:"expired_on" gorm:"not null"` // "expired_on": "2023-12-31T23:59:59Z",
		ProductId     uint      `json:"product_id"`
		CategoryId    uint      `json:"category_id"`
		BrandId       uint      `json:"brand_id"`
	}

	var offerDetails OfferDetails
	if c.Bind(&offerDetails) != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "Data binding failed",
		})
		return
	}

	db := initializers.DB
	offer_type := strings.ToLower(offerDetails.OfferType)

	currentTime := time.Now()
	expiredData := offerDetails.Expired_on

	if currentTime.After(expiredData) {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Please check offer validity",
		})
		c.Abort()
		return
	}

	if offer_type == "product" {
		var ActualPrice uint
		var offer_id uint
		var offer models.Offers

		r := db.First(&offer).Where("product_id = ?", offerDetails.ProductId)
		if r.Error != nil {
			offers := models.Offers{
				Offer_name:     offerDetails.OfferName,
				Offer_discount: offerDetails.OfferDiscount,
				Product_id:     offerDetails.ProductId,
				Expired_on:     offerDetails.Expired_on,
			}
			r = db.Create(&offers)
			if r.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Error": r.Error.Error(),
				})
				return
			}
		} else {
			r := db.Find(&offer).Where("product_id = ?", offerDetails.ProductId).Updates(models.Offers{
				Offer_name:     offerDetails.OfferName,
				Offer_discount: offerDetails.OfferDiscount,
				Product_id:     offerDetails.ProductId,
				Expired_on:     offerDetails.Expired_on,
			})
			if r.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Error": r.Error.Error(),
				})
				return
			}
		}

		var pCount int64
		_ = db.Last(&models.Offers{}).Select("id").Where("product_id = ?", offerDetails.ProductId).Scan(&offer_id)
		_ = db.Table("products").Select("actual_price").Where("product_id = ?", offerDetails.ProductId).Scan(&ActualPrice).Count(&pCount)
		if pCount == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"Caution": "Product not found",
			})
			return
		}

		r = db.First(&models.Product{}, offerDetails.ProductId).Updates(models.Product{SpecialPrice: ActualPrice - offerDetails.OfferDiscount, Offer_details: offer_type + " offer available", Offer_id: offer_id})
		if r.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"Error":   r.Error.Error(),
				"Message": "Failed to apply Product Offer",
			})
			return
		}

		var product string
		_ = db.First(&models.Product{}).Select("product_name").Where("product_id = ?", offerDetails.ProductId).Scan(&product)
		c.JSON(http.StatusOK, gin.H{
			"Message": "Offer added successfully to product " + product,
		})

	} else if offer_type == "brand" {
		var productid []uint
		var offer_id uint

		r := db.First(models.Offers{}).Where("brand_id = ?", offerDetails.BrandId)
		if r.Error != nil {
			offers := models.Offers{
				Offer_name:     offerDetails.OfferName,
				Offer_discount: offerDetails.OfferDiscount,
				Brand_id:       offerDetails.BrandId,
				Expired_on:     offerDetails.Expired_on,
			}
			r = db.Create(&offers)
			if r.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Error": r.Error.Error(),
				})
				return
			}
		} else {
			r := db.Find(&models.Offers{}).Where("brand_id = ?", offerDetails.BrandId).Updates(models.Offers{
				Offer_name:     offerDetails.OfferName,
				Offer_discount: offerDetails.OfferDiscount,
				Brand_id:       offerDetails.BrandId,
				Expired_on:     offerDetails.Expired_on,
			})
			if r.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Error": r.Error.Error(),
				})
				return
			}
		}
		_ = db.Last(&models.Offers{}).Select("id").Where("brand_id = ?", offerDetails.BrandId).Scan(&offer_id)

		_ = db.Find(&models.Product{}).Select("product_id").Where("brand_id = ?", offerDetails.BrandId).Scan(&productid)

		for _, id := range productid {
			var ActualPrice uint
			_ = db.Find(&models.Product{}).Select("actual_price").Where("product_id = ?", id).Scan(&ActualPrice)
			r := db.Table("products").Where("product_id = ?", id).Updates(models.Product{SpecialPrice: ActualPrice - offerDetails.OfferDiscount, Offer_details: offer_type + " offer available", Offer_id: offer_id})
			if r.RowsAffected == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"Error": "There is no brand corresponding to entered brand id",
				})
				return
			}
			if r.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Error": r.Error.Error(),
				})
				return
			}
		}

		var brand string
		_ = db.First(&models.Brand{}).Select("brand_name").Where("id = ?", offerDetails.BrandId).Scan(&brand)
		c.JSON(http.StatusOK, gin.H{
			"Message": "Offer added successfully on brand " + brand,
		})

	} else if offer_type == "category" {

		var offer_id uint
		r := db.First(models.Offers{}).Where("category_id = ?", offerDetails.CategoryId)
		if r.Error != nil {
			offers := models.Offers{
				Offer_name:     offerDetails.OfferName,
				Offer_discount: offerDetails.OfferDiscount,
				Category_id:    offerDetails.CategoryId,
				Expired_on:     offerDetails.Expired_on,
			}
			r = db.Create(&offers)
			if r.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Error": r.Error.Error(),
				})
				return
			}
		} else {
			r := db.Find(&models.Offers{}).Where("category_id = ?", offerDetails.CategoryId).Updates(models.Offers{
				Offer_name:     offerDetails.OfferName,
				Offer_discount: offerDetails.OfferDiscount,
				Category_id:    offerDetails.CategoryId,
				Expired_on:     offerDetails.Expired_on,
			})
			if r.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Error": r.Error.Error(),
				})
				return
			}
			_ = db.Last(&models.Offers{}).Select("id").Where("category_id = ?", offerDetails.CategoryId).Scan(&offer_id)
		}

		var productid []uint
		_ = db.Find(&models.Product{}).Select("product_id").Where("category_id = ?", offerDetails.CategoryId).Scan(&productid)

		for _, id := range productid {
			var ActualPrice uint
			_ = db.Find(&models.Product{}).Select("actual_price").Where("product_id = ?", id).Scan(&ActualPrice)
			r := db.Table("products").Where("product_id = ?", id).Updates(models.Product{SpecialPrice: ActualPrice - offerDetails.OfferDiscount, Offer_details: offer_type + " offer available", Offer_id: offer_id})
			if r.RowsAffected == 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"Error": "There is no category corresponding to entered category id",
				})
				return
			}
			if r.Error != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"Error": r.Error.Error(),
				})
				return
			}
		}

		var category string
		_ = db.First(&models.Category{}).Select("category_name").Where("id = ?", offerDetails.ProductId).Scan(&category)
		c.JSON(http.StatusOK, gin.H{
			"Message": "Offer added successfully to category " + category,
		})

	} else {
		c.JSON(http.StatusBadRequest, gin.H{
			"Error":   "Invalid 'Offer Applied ON' entry",
			"Message": "Choose where you want to apply the offer. produt, brand or category(choose one)",
		})
		c.Abort()
		return
	}
}

func CheckOfferExpiry() error {
	db := initializers.DB
	var offers []models.Offers
	currentTime := time.Now()

	_ = db.Find(&offers).Where("expired_on <= ?", currentTime).Scan(&offers)

	for _, offer := range offers {

		expiredOffer := offer.Expired_on

		if offer.Product_id != 0 {
			var special_price uint

			if currentTime.After(expiredOffer) {
				r := db.Table("products").Select("special_price").Where("product_id = ?", offer.Product_id).Scan(&special_price)
				if r.Error != nil {
					db.Delete(&models.Offers{}).Where("id = ?", offer.ID)
					return fmt.Errorf("product already deleted")
				}
				r = db.Model(&models.Product{}).Where("product_id = ?", offer.Product_id).Updates(models.Product{
					SpecialPrice:  special_price + offer.Offer_discount,
					Offer_details: "nil",
					Offer_id:      0,
				})
				if r.Error != nil {
					return r.Error
				}
				r = db.Delete(&models.Offers{}, offer.ID)//.Where("id = ?", offer.ID)
				if r.Error != nil {
					return r.Error
				}
				return nil
			}
		} else if offer.Brand_id != 0 {
			var special_price uint
			var allProducts []models.Product

			if currentTime.After(expiredOffer) {
				r := db.Find(&models.Product{}).Where("brand_id = ?", offer.Brand_id).Scan(&allProducts)
				if r.Error != nil {
					return r.Error
				}
				for _, products := range allProducts {
					r = db.Table("products").Select("special_price").Where("product_id = ?", products.ProductId).Scan(&special_price)
					if r.Error != nil {
						db.Delete(&models.Offers{}).Where("id = ?", offer.ID)
						return fmt.Errorf("product already deleted")
					}
					r = db.Model(&models.Product{}).Where("product_id = ?", products.ProductId).Updates(models.Product{
						SpecialPrice:  special_price + offer.Offer_discount,
						Offer_details: "nil",
						Offer_id:      0,
					})
					if r.Error != nil {
						return r.Error
					}
				}

				r = db.Delete(&models.Offers{}, offer.ID)//.Where("id = ?", offer.ID)
				if r.Error != nil {
					return r.Error
				}
				return nil
			}
		} else if offer.Category_id != 0 {
			var special_price uint
			var allProducts []models.Product

			if currentTime.After(expiredOffer) {
				r := db.Find(&models.Product{}).Where("category_id = ?", offer.Category_id).Scan(&allProducts)
				if r.Error != nil {
					return r.Error
				}
				for _, products := range allProducts {
					r = db.Table("products").Select("special_price").Where("product_id = ?", products.ProductId).Scan(&special_price)
					if r.Error != nil {
						db.Delete(&models.Offers{}).Where("id = ?", offer.ID)
						return fmt.Errorf("product already deleted")
					}
					r = db.Model(&models.Product{}).Where("product_id = ?", products.ProductId).Updates(models.Product{
						SpecialPrice:  (special_price - Dis) + offer.Offer_discount,
						Offer_details: "No offer available",
						Offer_id:      0,
					})
					if r.Error != nil {
						return r.Error
					}
				}
				r = db.Delete(&models.Offers{}, offer.ID)//.Where("id = ?", offer.ID)
				if r.Error != nil {
					return r.Error
				}
				return nil
			}
		}
	}
	return nil

}
func ShowOffers(c *gin.Context){
	var offers []models.Offers

	db := initializers.DB
	r := db.Table("offers").Find(&offers)
	count := r.RowsAffected
	if r.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Number of offers": count,
		"offers": offers,
	})
}

func CancelOffer(c *gin.Context){
	offerid, err := strconv.Atoi(c.Query("offer_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": "String conversion failed",
		})
		return
	}

	db := initializers.DB
	currentTime := time.Now()

	r := db.Table("offers").Where("id = ?",offerid).Update("expired_on", currentTime)
	if r.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": r.Error.Error(),
		})
		return
	}

	err = CheckOfferExpiry()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"Error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"Message": "Offer cancelled",
	})
	
}
