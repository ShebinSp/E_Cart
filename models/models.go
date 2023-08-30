package models

import "golang.org/x/crypto/bcrypt"

type User struct {
	ID           uint   `json:"id" gorm:"primaryKey;unique"`
	First_Name   string `json:"first_name" gorm:"not null" validate:"required,min=2,max=50"`
	Last_Name    string `json:"last_name" gorm:"not null" validate:"required,min=1,max=50"`
	Email        string `json:"email" gorm:"not null;unique" validate:"email,required"`
	Password     string `json:"password" gorm:"not null"  validate:"required"`
	Phone        string `json:"phone" gorm:"not null;unique" validate:"required"`
	Otp          string `json:"otp"`
	Block_status bool   `json:"blockOrNot"`
	User_status  bool   `gorm:"default:false"`
	Is_admin     bool   `json:"is_admin"`
	Referal_code string `json:"referal_code"`
}

// type Admin struct {
// 	Email    string
// 	Password string
// }

// func (admin *Admin) HashPassword(password string) error {
// 	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
// 	if err != nil {
// 		return err
// 	}
// 	admin.Password = string(bytes)
// 	return nil
// }
// func (Admin *Admin) VerifyPassword(givenPassword string) error {
// 	err := bcrypt.CompareHashAndPassword([]byte(Admin.Password), []byte(givenPassword))
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

type Address struct {
	AddressId uint `gorm:"primaryKey;unique"`
	User      User `gorm:"ForeignKey:Userid"`
	Userid    uint

	FullName   string `json:"fullname" gorm:"not null"`
	Phone      uint   `json:"phone" gorm:"not null"`
	HouseName  string `json:"housename" gorm:"not null"`
	Area       string `json:"area" gorm:"not null"`
	Landmark   string `json:"landmark" gorm:"not null"`
	City       string `json:"city" gorm:"not null"`
	District   string `json:"district" gorm:"not null"`
	State      string `json:"state" gorm:"not null"`
	Pincode    uint   `json:"pincode" gorm:"not null"`
	DefaultAdd bool   `json:"defaultadd" gorm:"default:false"`
}

func (user *User) HashPassword(password string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	user.Password = string(bytes)
	return nil
}

func (user *User) VerifyPassword(givenPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(givenPassword))
	if err != nil {
		return err
	}
	return nil
}

type Referal_info struct {
	Email      string
	Referer_id uint
}
