package auth

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/ShebinSp/e-cart/initializers"
	"github.com/ShebinSp/e-cart/models"
	"github.com/gin-gonic/gin"
)

var otp string
// var eMail string

func generateOTP() string {
	seed := time.Now().UnixNano()
	//rand.Seed(time.Now().UnixNano())
	//v := rand.Intn(9999) + 1000
	randomGenerator := rand.New(rand.NewSource(seed))

	return strconv.Itoa(randomGenerator.Intn(8999) + 1000)
}

func OTPvarification(c *gin.Context, email string) string {
	otp = generateOTP()

	sendMail(c, email)

	c.JSON(200, gin.H{
		"success": "OTP send sucessfully, Please check your email",
	})
	return otp
}
func sendMail(c *gin.Context, email string) {
//	eMail = email
	user := os.Getenv("email")
	password := os.Getenv("ePassword")
	host := os.Getenv("eHost")
	to := email
	sub := os.Getenv("eSub")

	auth := smtp.PlainAuth("", user, password, host)

	body := fmt.Sprintf("OTP for varification is: %s", otp)
	fmt.Println("OTP:", otp)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + sub + "\r\n" +
		"\r\n" + body + "\r\n")

	err := smtp.SendMail(host+":587", auth, user, []string{to}, msg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"err": err,
		})
		c.Abort()
		return
	}
}

func OtpValidation(c *gin.Context) {
	type UserOTP struct {
		Otp   string
		Email string
	}

	var userOTP UserOTP
	var userData models.User
//	userOTP.Email = eMail

	if err := c.Bind(&userOTP); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"Eroor": "Could not bind the JSON Data",
		})
		return
	}
	//----------------------------------------------------
	// otp := fmt.Sprintf("User entered otp: %s",userOTP.Otp)
	// em := fmt.Sprintf("User email: %s",userOTP.Email)
	// fmt.Println(otp)
	// fmt.Println(em)
	//----------------------------------------------------
	db := initializers.DB
	result := db.First(&userData, "otp LIKE ? AND email LIKE ?", userOTP.Otp, userOTP.Email)

	if result.Error != nil {
		c.JSON(404, gin.H{
			"Error": result.Error.Error(),
		})

		db.Where("email = ?", userOTP.Email).Delete(&models.User{})
//		eMail = "nil"
		c.JSON(http.StatusNotFound, gin.H{
			"Error": "Incorrect OTP, Please register again",
			"msg":   "Goto /signup",
		})
		return
	}

	db.First(&userData)
	userData.Otp = "otp_confirmed"
	db.Save(&userData)

	c.JSON(http.StatusOK, gin.H{
		"success": "Your Account " + userOTP.Email + " Registered successfully",
		"msg":     "go to /login to continue",
	})
//	eMail = "nil"

}
