package middleware

import (
	"net/http"

	"github.com/ShebinSp/e-cart/auth"
	"github.com/gin-gonic/gin"
)

func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("AdminAuth")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"Error": "Request does not contain an access token",
			})
			c.Abort()
			return
		}
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"Error": err.Error(),
			})
		}

		err = auth.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		c.Next()
	}

}

func UserAuth() gin.HandlerFunc{
	return func(c *gin.Context){
		tokenString,_ := c.Cookie("UserAuth")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Request does not contain an access token",
			})
			c.Abort()
			return
		}
		err := auth.ValidateToken(tokenString)
		c.Set("user",auth.P)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
