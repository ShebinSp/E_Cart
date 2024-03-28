package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ShebinSp/e-cart/auth"
	"github.com/ShebinSp/e-cart/initializers"
	"github.com/ShebinSp/e-cart/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"golang.org/x/crypto/bcrypt"
)

type UserLogin struct {
	Email      string
	Password   string
	BlockOrNot bool
}
type data struct {
	First_name string
	Last_name  string
	Email      string
	Phone      string
	Is_admin   bool
	FullName   string
	PhoneAddrs uint
	HouseName  string
	Area       string
	Landmark   string
	City       string
	District   string
	State      string
	Pincode    uint
	DefaultAdd bool
}

var validate = validator.New()

// SignupUser godoc
// @Summary User Signup
// @Description Signup a user with required datas
// @Tags Signup
// @Accept */*
// @Produce json
// @Param user body models.User true "User details"
// @Success 200 {object} map[string]string "Successful registration"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Router /user/signup [post]
func Signup(c *gin.Context) {
	var user models.User

	if err := c.ShouldBindJSON(&user); err != nil {
		BindErr := map[string]interface{}{
			"Error": err.Error(),
		}
		c.JSON(http.StatusInternalServerError, BindErr)
		c.Abort()
		return
	}
	if validationErr := validate.Struct(user); validationErr != nil {
		valdErr := map[string]interface{}{
			"Error": validationErr,
		}
		c.JSON(http.StatusBadRequest, valdErr)
		return
	}

	if err := user.HashPassword(user.Password); err != nil {
		hashErr := map[string]interface{}{
			"Error": err.Error(),
		}
		c.JSON(http.StatusBadRequest, hashErr)
		c.Abort()
		return
	}
	r := initializers.DB.Table("users").Where("email = ?", user.Email)
	if r.RowsAffected != 0 {
		errors := map[string]interface{}{
			"Error":   "User with same email already exist",
			"Message": "Please login",
			"Email":   user.Email,
		}
		c.JSON(http.StatusBadRequest, errors)
		c.Abort()
		return
	}

	var referer_id uint
	r = initializers.DB.Table("users").Select("id").Where("referal_code = ?", user.Referal_code).Scan(&referer_id)
	if r.Error != nil {
		refErr := map[string]string{
			"Message": "Incorrect refferal code",
		}
		c.JSON(http.StatusBadRequest, refErr)
	} else if referer_id != 0 { // if we got a referer id then adding the reward to referer's wallet
		fmt.Println("referer id : ", referer_id)
		err := initializers.DB.Table("referal_infos").Where("email = ?", user.Email).Error
		if err != nil {
			referal_info := models.Referal_info{
				Email:      user.Email,
				Referer_id: referer_id,
			}
			r := initializers.DB.Create(&referal_info)
			if r.Error != nil {
				refErr := map[string]interface{}{
					"Error": r.Error.Error(),
				}
				c.JSON(http.StatusBadRequest, refErr)
				return
			}
			const ref_reward float64 = 100
			typE := "referal reward"
			err = addToWallet(referer_id, ref_reward, typE)
			if err != nil {
				refErr := map[string]interface{}{
					"Error": r.Error.Error(),
				}
				c.JSON(http.StatusBadRequest, refErr)
				return
			}
		}

	}
	otp := auth.OTPvarification(c, user.Email)

	record := initializers.DB.Create(&user)
	if record.Error != nil {
		userErr := map[string]interface{}{
			"status": "False",
			"err":    record.Error,
		}
		c.JSON(http.StatusInternalServerError, userErr)
		c.Abort()
		return
	} else {
		initializers.DB.Model(&user).Where("email LIKE ?", user.Email).Update("otp", otp)
	}

	referal_code, err := ReferalCodeGenerator(user.Email)
	if err != nil {
		refErr := map[string]string{
			"Eroor": "Referal code generation failed",
		}
		c.JSON(http.StatusInternalServerError, refErr)
		return
	}
	initializers.DB.Model(&user).Where("email = ?", user.Email).Updates(models.User{Otp: otp, Referal_code: referal_code})

	// Adding reward for user
	if referer_id != 0 {
		const ref_reward float64 = 50
		typE := "referal reward"
		err = initializers.DB.Table("referal_infos").Where("email = ?", user.Email).Error
		if err != nil {
			err = addToWallet(user.ID, ref_reward, typE)
			if err != nil {
				refErr := map[string]interface{}{
					"Error": r.Error.Error(),
				}
				c.JSON(http.StatusBadRequest, refErr)
				return
			}
		}
	}
	UserOk := map[string]string{
		"verify":  "Account Varification is Pending for " + user.Email,
		"message": "Please Go to /signup/otpverification",
	}
	c.JSON(http.StatusOK, UserOk)
}

// LoginUser godoc
// @Summary User Login
// @Description Log in a user and generate an authentication token
// @Tags Authentication
// @Accept */*
// @Produce json
// @Param user body UserLogin true "User login data"
// @Success 200 {object} map[string]interface{} "Successful login"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 401 {object} map[string]interface{} "Unauthorized"
// @Router /user/login [post]
func LoginUser(c *gin.Context) {

	var usrlogin UserLogin
	var user models.User
	db := initializers.DB

	if err := c.ShouldBindJSON(&usrlogin); err != nil {
		BindErr := map[string]interface{}{
			"Error": err.Error(),
		}
		c.JSON(http.StatusInternalServerError, BindErr)
		c.Abort()
		return
	}

	record := db.Raw("select * from users where email=?", usrlogin.Email).Scan(&user)
	if record.Error != nil {
		noUsrErr := map[string]interface{}{
			"Error": record.Error.Error(),
		}
		c.JSON(http.StatusBadRequest, noUsrErr)
		c.Abort()
		return
	}
	if user.Block_status {
		blockErr := map[string]interface{}{
			"Status":  "Authorization",
			"Message": "User blocked by Admin",
		}
		c.JSON(http.StatusUnauthorized, blockErr)
		return
	}

	if user.Otp == "otp_confirmed" {
		db.Table("users").Where("id = ?", user.ID).Update("user_status", true)
	}
	var userStatus bool
	db.Table("users").Where("id = ?", user.ID).Select("user_status").Scan(&userStatus)
	if !userStatus {
		varErr := map[string]string{
			"Caution": "User identity not confirmed",
			"Message": "Please confirm OTP",
		}
		c.JSON(http.StatusBadRequest, varErr)
		c.Abort()
		return
	}
	credientialCheck := user.VerifyPassword(usrlogin.Password)
	if credientialCheck != nil {
		pswrdErr := map[string]string{
			"error": "Invalid credentials",
		}
		c.JSON(http.StatusUnauthorized, pswrdErr)
		c.Abort()
		return
	}

	tokenString, err := auth.GenerateJWT(user.Email)
	fmt.Println(tokenString)
	token := tokenString["access_token"]
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("UserAuth", token, 3600*24*30, "", "", false, true)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error})
		c.Abort()
		return
	}
	loginOk := map[string]interface{}{
		"email":    usrlogin.Email,
		"password": usrlogin.Password,
		"token":    tokenString,
	}
	// user successfully signed in
	c.JSON(200, loginOk)
}

// --- List all users --- \\

// @Summary List Users
// @Description Retrieve a list of active users
// @Tags Admin
// @Accept */*
// @Produce json
// @Success 200 {object} map[string]interface{} "List of active users"
// @Failure 404 {object} map[string]interface{} "Not found"
// @Router /admin/usermanagement/viewusers [get]
func ListUsers(c *gin.Context) {
	type data struct {
		First_name   string
		Last_name    string
		Email        string
		Phone        string
		Block_Status bool
		Is_admin     bool
	}
	var userData []data

	db := initializers.DB
	result := db.Table("users").Select("first_name, last_name, email, phone, block_status, is_admin").Where("user_status = ?", true).Scan(&userData)
	if result.Error != nil {
		noUser := map[string]interface{}{
			"Error":   result.Error.Error(),
			"Message": "Could not found the user",
		}
		c.JSON(http.StatusNotFound, noUser)
		return
	}

	okUser := map[string]interface{}{
		"User Data":   userData,
		"Total users": result.RowsAffected,
	}
	c.JSON(http.StatusOK, okUser)
}

// --- Block/Unblock users --- \\

// @Summary Block or Unblock users
// @Description Block or unblock a user by their user ID
// @Tags Admin
// @Param userid query integer true "User ID"
// @Success 200 {object} map[string]string "User status updated"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Not found"
// @Router /admin/usermanagement/manageblock [patch]
func BlockUser(c *gin.Context) {
	userId, err := strconv.Atoi(c.Query("userid"))
	if err != nil {
		conErr := map[string]interface{}{
			"Error": "Error occured while converting string",
		}
		c.JSON(http.StatusInternalServerError, conErr)
		return
	}
	var user models.User
	db := initializers.DB

	result := db.First(&user, userId)
	if result.Error != nil {
		noUser := map[string]string{
			"Message": "User does not exist",
		}
		c.JSON(http.StatusNotFound, noUser)
		return
	}
	if !user.Block_status {
		result := db.Model(&user).Where("id", userId).Update("block_status", true)
		if result.Error != nil {
			errors := map[string]interface{}{
				"Error": result.Error.Error(),
			}
			c.JSON(http.StatusNotFound, errors)
			return
		}
		blocked := map[string]string{
			"Message": "User blocked" + user.Email,
		}
		c.JSON(http.StatusOK, blocked)
	} else {
		result := db.Model(&user).Where("id", userId).Update("block_status", false)
		if result.Error != nil {
			errors := map[string]interface{}{
				"Error": result.Error.Error(),
			}
			c.JSON(http.StatusNotFound, errors)
			return
		}
		unblocked := map[string]string{
			"Message": "User unblocked" + user.Email,
		}
		c.JSON(http.StatusOK, unblocked)
	}
}

func UserSignout(c *gin.Context) {
	c.SetCookie("UserAuth", "", -1, "", "", false, false)
	c.JSON(http.StatusOK, gin.H{
		"Message": "User Signed out Successfully",
	})
}

// @Summary Show User Details
// @Description Retrieve user details and asscciated address
// @Tags Users
// @Security BearerToken
// @Produce json
// @Success 202 {object} map[string]interface{} "User details and address"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Not found"
// @Failure 409 {object} map[string]string "Conflict"
// @Router /user/viewprofile [get]
func ShowUserDetails(c *gin.Context) {
	var userData models.User
	var userAddress models.Address

	email := c.GetString("user")
	db := initializers.DB
	id := getId(email, db)

	result := db.First(&userData, "id = ?", id)
	if result.Error != nil {
		err := map[string]string{
			"Error": "User does not exist",
		}
		c.JSON(http.StatusConflict, err)
		return
	}

	result = db.Raw("SELECT * FROM addresses WHERE userid = ?", id).Scan(&userAddress)
	if result.RowsAffected == 0 {
		err := map[string]string{
			"Error": "User address not found",
		}
		c.JSON(http.StatusNotFound, err)
		return
	}
	if result.Error != nil {
		err := map[string]interface{}{
			"Error": result.Error.Error(),
		}
		c.JSON(http.StatusNotFound, err)
		return
	}

	user_details := map[string]interface{}{
		"a1)User id":         userData.ID,
		"a2)First name":      userData.First_Name,
		"a3)Last name":       userData.Last_Name,
		"a4)Email":           userData.Email,
		"a5)Phone":           userData.Phone,
		"ab)Address ID":      userAddress.AddressId,
		"b) User ID":         userAddress.Userid,
		"c) full name":       userAddress.FullName,
		"d) Phone No":        userAddress.Phone,
		"e) House Name":      userAddress.HouseName,
		"f) Area":            userAddress.Area,
		"g) Landmark":        userAddress.Landmark,
		"h) City":            userAddress.City,
		"i) District":        userAddress.District,
		"j) State":           userAddress.State,
		"k) Pincode":         userAddress.Pincode,
		"l) Default Address": userAddress.DefaultAdd,
	}
	c.JSON(http.StatusAccepted, user_details)
}

// @Summary Edit User Profile
// @Description Edit user profile details and associated address
// @Tags Users
// @Security BearerToken
// @Accept json
// @Produce json
// @Param user body data true "User profile data"
// @Success 200 {object} map[string]interface{} "Updated profile details"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Failure 404 {object} map[string]interface{} "Not found"
// @Router /user/editprofile [patch]
func EditProfile(c *gin.Context) {

	var user data
	if c.Bind(&user) != nil {
		err := map[string]interface{}{
			"Error": "Data Binding Error",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}
	email := c.GetString("user")
	db := initializers.DB
	id := getId(email, db)

	var userData models.User
	r := db.First(&userData, "id = ?", id)
	if r.Error != nil {
		err := map[string]interface{}{
			"Caution": "User does not exist",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	r = db.Model(&userData).Updates(models.User{
		First_Name: user.First_name,
		Last_Name:  user.Last_name,
		Phone:      user.Phone,
	})
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusNotFound, err)
		return
	}

	var userAddress models.Address
	userAddress.Userid = id
	result := db.Model(userAddress).Where("userid = ?", id).Updates(models.Address{
		FullName:  user.FullName,
		Phone:     user.PhoneAddrs,
		HouseName: user.HouseName,
		Area:      user.Area,
		Landmark:  user.Landmark,
		City:      user.City,
		District:  user.District,
		State:     user.State,
		Pincode:   user.Pincode,
	})
	if result.Error != nil {
		err := map[string]interface{}{
			"Error": result.Error.Error(),
		}
		c.JSON(http.StatusNotFound, err)
		return
	}

	r = db.First(&userData, id)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	r = db.First(&userAddress, id)
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": r.Error.Error(),
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	userInfo := map[string]interface{}{
		"Message":         "Successfully Updated the Profile",
		"Updated Data":    "-----------------",
		"First name":      userData.First_Name,
		"Last name":       userData.Last_Name,
		"Phone":           userData.Phone,
		"Address":         "------------------",
		"Full name":       userAddress.FullName,
		"Phone No":        userAddress.Phone,
		"House Name":      userAddress.HouseName,
		"Area":            userAddress.Area,
		"Landmark":        userAddress.Landmark,
		"City":            userAddress.City,
		"District":        userAddress.District,
		"State":           userAddress.State,
		"Pincode":         userAddress.Pincode,
		"Default Address": userAddress.DefaultAdd,
	}

	c.JSON(http.StatusOK, userInfo)
}

type userData struct {
	Email           string
	Password        string
	RepeatePassword string
}

// @Summary Forgot Password
// @Description User can change password if forgot
// @Tags Authentication, Users
// @Accept json
// @Produce json
// @Param user body userData ture "User input"
// @Success 200 {object} map[string]interface{} "Success"
// @Failure 500 {object} map[string]interface{} "Internal server error"
// @Failure 400 {object} map[string]interface{} "Bad request"
// @Router /user/forgotpassword [patch]
func ForgotPassword(c *gin.Context) {
	db := initializers.DB
	var user models.User
	var userdata userData
	if c.Bind(&user) != nil {
		err := map[string]interface{}{
			"Error": "Data binding failed",
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}
	if userdata.Password != userdata.RepeatePassword {
		err := map[string]interface{}{
			"Error": "Passwords do not match",
		}
		c.JSON(http.StatusBadRequest, err)
		return
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(userdata.Password), 14)
	if err != nil {
		err := map[string]interface{}{
			"Error": "Passwords encryption failed",
		}
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	otp := auth.OTPvarification(c, userdata.Email)
	newPswrd := string(bytes)
	r := db.Model(&user).Where("email = ?", userdata.Email).Updates(models.User{
		Password: newPswrd,
		Otp:      otp,
	})
	if r.Error != nil {
		err := map[string]interface{}{
			"Error": "Incorrect email",
		}
		c.JSON(http.StatusBadRequest, err)
		c.Abort()
		return
	}

	Ok := map[string]interface{}{
		"Message": "Please confirm OTP",
	}
	c.JSON(http.StatusOK, Ok)
}

func ReferalCodeGenerator(email string) (string, error) {
	db := initializers.DB
	id := getId(email, db)
	strid := strconv.Itoa(int(id))
	var first_name string

	r := db.Table("users").Select("first_name").Where("id = ?", id).Scan(&first_name)
	if r.Error != nil {
		return "", r.Error
	}

	referal_code := first_name + strid
	fmt.Println("Referal code:", referal_code)
	return referal_code, nil
}
