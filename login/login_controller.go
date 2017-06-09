package login

import (
	"fmt"
	"net/http"
	"time"

	valid "github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
)

// Exports
var strict = bluemonday.StrictPolicy()

type ForgotPasswordInfo struct {
	Email string `json:"email" valid:"email" binding:"required"`
}

type ResetPasswordInfo struct {
	Password  string `form:"password" binding:"required"`
	Password2 string `form:"password2" binding:"required"`
	Username  string `form:"username" binding:"required"`
}

type LoginCredentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type SignupInfo struct {
	FirstName string `json:"firstName"`
	Username  string `json:"username" binding:"required"`
	Password  string `json:"password" binding:"required"`
	Email     string `json:"email" valid:"email" binding:"required"`
}

type FacebookAuthData struct {
	AccessToken    string     `json:"accessToken" binding:"required"`
	ExpirationDate *time.Time `json:"expirationDate" binding:"required"`
	Id             string     `json:"id" binding:"required"`
}

type UserSession struct {
	User  UserWithAuthInfo `json:"user"`
	Token string           `json:"sessionToken"`
}

type UserWithAuthInfo struct {
	*models.User `json:",inline"`
	IsFacebook   bool `json:"isFacebook"`
}

func FacebookLogin(c *gin.Context) {

	var json FacebookAuthData
	ds := c.MustGet("ds").(db.DataStore)

	if c.BindJSON(&json) == nil {

		if user, token, err := loginWithFacebook(json, ds); err != nil {
			fmt.Printf("Facebook Login Error: %s \n", err)
			c.JSON(http.StatusUnauthorized, err.Error())
		} else {
			c.JSON(http.StatusOK, UserSession{UserWithAuthInfo{user, user.IsFacebook()}, token})
		}

		return
	}

	c.JSON(http.StatusBadRequest, "Bad request.")

}

func Login(c *gin.Context) {

	var json LoginCredentials
	ds := c.MustGet("ds").(db.DataStore)

	if c.BindJSON(&json) == nil {

		sanitized := strict.Sanitize(json.Username)
		if len(sanitized) != len(json.Username) {
			c.JSON(http.StatusForbidden, models.ERR_INVALID_USERNAME.Error())
			return
		}

		if user, token, err := login(json, ds); err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			c.JSON(http.StatusOK, UserSession{UserWithAuthInfo{user, user.IsFacebook()}, token})
		}

		return
	}

	c.JSON(http.StatusBadRequest, "Bad request.")

}

func Logout(c *gin.Context) {
	ds := c.MustGet("ds").(db.DataStore)
	user := c.MustGet("user").(*models.User)

	if err := logout(user, ds); err != nil {
		fmt.Printf("Error on logout: %s", err)
		c.AbortWithError(http.StatusInternalServerError, err)
	} else {
		c.JSON(http.StatusOK, user)
	}
}

func Signup(c *gin.Context) {
	var json SignupInfo
	ds := c.MustGet("ds").(db.DataStore)

	if c.BindJSON(&json) == nil {
		if isValid, _ := valid.ValidateStruct(json); isValid {

			if user, token, err := signup(json, ds); err != nil {
				fmt.Printf("Sign up error: %s \n", err)
				c.JSON(http.StatusUnauthorized, err.Error())
			} else {
				c.JSON(http.StatusOK, UserSession{UserWithAuthInfo{user, user.IsFacebook()}, token})
			}
			return
		}
	}

	c.JSON(http.StatusBadRequest, "Bad request.")
}

// Forgot Password

func FinishResetPassword(c *gin.Context) {
	token := c.Query("token")
	ds := c.MustGet("ds").(db.DataStore)
	user := c.MustGet("user").(*models.User)

	var form ResetPasswordInfo

	if c.Bind(&form) == nil {

		sanitized := strict.Sanitize(form.Username)
		if len(form.Username) != len(sanitized) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		if form.Password != form.Password2 {
			c.HTML(http.StatusBadRequest, "password_reset.tmpl", gin.H{
				"status":   "Passwords do not match",
				"username": form.Username,
				"token":    token,
			})
			return
		}

		sanitized = strict.Sanitize(form.Password)
		if len(form.Password) != len(sanitized) {
			c.HTML(http.StatusBadRequest, "password_reset.tmpl", gin.H{
				"status":   "Invalid password. Please provide a different one.",
				"username": form.Username,
				"token":    token,
			})
			return
		}

		if user.ChangePassword(form.Password, ds) == nil {
			fmt.Println("Changed password for user: " + form.Username)
			c.HTML(http.StatusOK, "password_reset_success.tmpl", gin.H{
				"status": "Reset password successfully.",
			})
		} else {
			c.HTML(http.StatusBadRequest, "password_reset.tmpl", gin.H{
				"status":   "Error changing password. Please try again.",
				"username": form.Username,
				"token":    token,
			})
		}

		return

	}
	fmt.Println("Bad request for password reset.")
	c.AbortWithStatus(http.StatusBadRequest)

}

func ResetPassword(c *gin.Context) {
	ds := c.MustGet("ds").(db.DataStore)
	user := c.MustGet("user").(*models.User)
	token := c.Query("token")

	if err := user.Fetch(ds); err != nil {
		fmt.Printf("Could not find user %s error: %s \n", user.ObjectId(), err)
		c.AbortWithStatus(http.StatusNotFound)
		return
	}

	c.HTML(http.StatusOK, "password_reset.tmpl", gin.H{
		"username": user.Username,
		"token":    token,
		"status":   "",
	})

}

func ForgotPassword(c *gin.Context) {

	ds := c.MustGet("ds").(db.DataStore)
	var json ForgotPasswordInfo

	if c.BindJSON(&json) == nil {
		if isValid, _ := valid.ValidateStruct(json); isValid {

			user, err := models.FindUserByEmail(ds, json.Email)
			if err != nil {
				fmt.Printf("Could not find user for email <%s>: %s \n", json.Email, err)
				c.AbortWithStatus(http.StatusNotFound)
				return
			}

			if user.IsFacebook() {
				c.JSON(http.StatusOK, "You created an account with Facebook. Please press the Facebook button to log in.")
				return
			}

			err = initiateReset(user, ds)
			if err != nil {
				fmt.Printf("Initiate password reset error: %s \n", err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}

			c.JSON(http.StatusOK, "Please check your email for password reset instructions.")
			return
		}

	}

	c.JSON(http.StatusBadRequest, "Bad request.")

}
