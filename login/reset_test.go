package login

import (
	"testing"

	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/middleware"
	"github.com/nidhik/backend/models"
	"github.com/nidhik/backend/query"
	"github.com/nidhik/backend/routes"
)

func TestGetResetPassword(t *testing.T) {
	query.RunTest(t, func(t *testing.T, ds db.DataStore) {

		router := setupResetTests(t, ds)
		user, token := setupUsers(t, ds, "bae", "bae@foo.com")

		tests := []struct {
			token    string
			user     *models.User
			respCode int
			body     string
		}{
			{token, user, 200, user.Username + "/" + token + "/" + ""},
			{"foobar", user, 404, ""},
		}

		for _, test := range tests {
			resp := recordGet(router, "/reset?token="+test.token, nil)
			if test.respCode != resp.Code {
				t.Fatal("Expected code:", test.respCode, "Actual: ", resp.Code, "Response:", resp)
			} else {

				if resp.Body.String() != test.body {
					t.Fatal("Expected response:\n", test.body, "Actual:\n", resp.Body.String())
				}
			}
		}

	})
}

func TestFinishResetPassword(t *testing.T) {
	query.RunTest(t, func(t *testing.T, ds db.DataStore) {

		router := setupResetTests(t, ds)
		user, token := setupUsers(t, ds, "bae", "bae@foo.com")
		tests := []struct {
			token      string
			formValues url.Values
			respCode   int
			body       string
		}{

			{token, url.Values{"password": {"abcd1234"}, "password2": {"abcd1234"}, "username": {user.Username}}, 200, "Reset password successfully."},
			{token, url.Values{"password": {"foo"}, "password2": {"abcd1234"}, "username": {user.Username}}, 400, user.Username + "/" + token + "/" + "Passwords do not match"},
			{token, url.Values{"password": {"    "}, "password2": {"    "}, "username": {user.Username}}, 400, user.Username + "/" + token + "/" + "Error changing password. Please try again."},
			{token, url.Values{"password": {"abcd1234"}, "password2": {"abcd1234"}}, 400, ""},
			{token, url.Values{"password": {";var date=new Date(); do{curDate = new Date();}while(curDate-date<10000)"}, "password2": {";var date=new Date(); do{curDate = new Date();}while(curDate-date<10000)"}, "username": {user.Username}}, 400, user.Username + "/" + token + "/" + "Invalid password. Please provide a different one."},
			{token, url.Values{"password": {"abcd1234"}, "password2": {"abcd1234"}, "username": {";var date=new Date(); do{curDate = new Date();}while(curDate-date<10000)"}}, 403, ""},
			{token, url.Values{"password": {"abcd1234"}, "password2": {"abcd1234"}, "username": {"function() { return obj.credits - obj.debits < 0;var date=new Date(); do{curDate = new Date();}while(curDate-date<10000); }"}}, 403, ""},
		}

		for _, test := range tests {
			resp := recordPostForm(router, "/reset/finish?token="+test.token, test.formValues)
			if test.respCode != resp.Code {
				t.Fatal("Expected code:", test.respCode, "Actual: ", resp.Code, "Response:", resp)
			} else {

				if resp.Body.String() != test.body {
					t.Fatal("Expected response:\n", test.body, "Actual:\n", resp.Body.String())
				}
			}
		}

	})
}

// Setup

func setupResetTests(t *testing.T, ds db.DataStore) *gin.Engine {

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Custom Middleware
	router.Use(middleware.Connect())

	// Templates
	router.LoadHTMLGlob("templates/*")

	resetGroup := router.Group(routes.RESET)
	resetGroup.Use(middleware.AuthorizedLink())
	{
		resetGroup.GET("", ResetPassword)
		resetGroup.POST(routes.FINISH, FinishResetPassword)
	}

	return router

}
