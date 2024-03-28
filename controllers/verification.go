package controllers

import (
	"net/http"

	"github.com/ShebinSp/e-cart/initializers"
	"github.com/ShebinSp/e-cart/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)
type userEnterData struct {
	Email           string
	LastPassword    string
	Password        string
	confirmPassword string
}
// @Summary Change Password
// @Description Change user password
// @Tags Authentication, Users
// @Security BearerToken
// @Accept json
// @Produce json
// @Param password body userEnterData true "Password change data"
// @Success 200 {object} map[string]interface{} "Password changed successfully"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Not found"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/changepassword [patch]
func ChangePassword(c *gin.Context) {
	
	var data userEnterData
	var userData models.User
	db := initializers.DB

	if c.Bind(&data) != nil {
		err := map[string]interface{}{
			"Error": "JSON Data Binding error",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	record := db.Raw("select * from users where email=?", data.Email).Scan(&userData)
	if record.Error != nil {
		err := map[string]interface{}{
			"error": record.Error.Error(),
		}
		c.JSON(http.StatusInternalServerError, err)
		c.Abort()
		return
	}
	if psChk := userData.VerifyPassword(data.LastPassword); psChk != nil {
		err := map[string]interface{}{
			"Message": "Incorrect last password",
		}
		c.JSON(http.StatusNotFound, err)
		return
	}

	if data.Password != data.confirmPassword {
		err := map[string]interface{}{
			"Error": "Passwords not match",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(data.Password), 10)
	if err != nil {
		err := map[string]interface{}{
			"Status": "False",
			"Error":  "Hashing Possword error",
		}
		c.JSON(http.StatusBadRequest,err)
		return
	}
	r := db.Find(&userData, "email = ?", data.Email)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": "User does not exist",
		}
		c.JSON(http.StatusNotFound, err)
		return
	}

	r = db.Model(&userData).Where("email = ?", data.Email).Update("password", hash)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": "User does not exist",
		}
		c.JSON(http.StatusNotFound, err)
		return
	}

	oK := map[string]interface{}{
		"Message": "Password Changed Successfully",
	}
	c.JSON(http.StatusOK, oK)
}
