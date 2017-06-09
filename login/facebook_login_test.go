package login

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nidhik/backend/auth"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/middleware"
	"github.com/nidhik/backend/models"
	"github.com/nidhik/backend/query"
	"github.com/nidhik/backend/routes"
)

// Tests

var myFBToken = os.Getenv("MY_FB_TOKEN")
var myFBId = os.Getenv("MY_FB_ID")
var validPaylod = `{"accessToken":"` + myFBToken + `", "expirationDate":"2015-01-20T02:46:18.684Z", "id": "639083781"}`
var onlyToken = `{"accessToken":"` + myFBToken + `}`

type FBLoginTest struct {
	id           string
	token        string
	profileId    string
	expiry       time.Time
	payload      []byte
	responseCode int
}

var fbLoginTests = []FBLoginTest{
	{"", "", "", time.Now(), []byte(`{"foo" : "bar"}`), 400},
	{"", "", "", time.Now(), []byte(`{"accessToken": "blah", "expirationDate":"2015-01-20T02:46:18.684Z", "id": "1234"}`), 401},
	{"", myFBToken, myFBId, time.Now(), []byte(validPaylod), 200},
	{"", myFBToken, myFBId, time.Now(), []byte(onlyToken), 400},
}

func TestFBLoginWithInsert(t *testing.T) {

	query.RunTest(t, func(t *testing.T, ds db.DataStore) {

		gin.SetMode(gin.TestMode)
		router := gin.New()
		router.Use(middleware.Connect())
		router.POST(routes.FACEBOOK_LOGIN, FacebookLogin)

		for _, test := range fbLoginTests {

			req, _ := http.NewRequest("POST", routes.FACEBOOK_LOGIN, bytes.NewBuffer(test.payload))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			if resp.Code != test.responseCode {
				t.Fatal("Expected:", test.responseCode, "got:", resp.Code)
			}

			if resp.Code == 200 {

				var sessionInfo UserSession
				json.Unmarshal(resp.Body.Bytes(), &sessionInfo)

				u, err := auth.VerifyToken(sessionInfo.Token)
				if err != nil {
					t.Fatal("Unexpected error when verifying JWT token", err)
				}

				if !sessionInfo.User.IsFacebook {
					t.Fatal("Expected isFacebook to be set to true.")
				}

				if !sessionInfo.User.IsNew() {
					t.Fatal("Expected isNew to be true.")
				}

				if u.Collection() != models.CollectionUser {
					t.Fatal("Expected:n", models.CollectionUser, "got:", u.Collection())
				}

				if sessionInfo.User.Email != "nidhikulkarni82@gmail.com" {
					t.Fatal("Wrong email", sessionInfo.User.Email)
				}

				if sessionInfo.User.FirstName != "Nidhi" {
					t.Fatal("Wrong first name", sessionInfo.User.FirstName)
				}

				if sessionInfo.User.Gender != "female" {
					t.Fatal("Wrong gender", sessionInfo.User.Gender)
				}
			}

		}

	})

}

func TestFBLoginWithUpdate(t *testing.T) {

	query.RunTest(t, func(t *testing.T, ds db.DataStore) {

		router, tests := setupFBLoginTests(t, ds)
		for _, test := range tests {

			fmt.Println("Test id: " + test.id)
			req, _ := http.NewRequest("POST", routes.FACEBOOK_LOGIN, bytes.NewBuffer(test.payload))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			if resp.Code != test.responseCode {
				t.Fatal("Expected:", test.responseCode, "got:", resp.Code)
			}

			if resp.Code == 200 {

				var sessionInfo UserSession
				json.Unmarshal(resp.Body.Bytes(), &sessionInfo)

				u, err := auth.VerifyToken(sessionInfo.Token)
				if err != nil {
					t.Fatal("Unexpected error when verifying JWT token", err)
				}

				if !sessionInfo.User.IsFacebook {
					t.Fatal("Expected isFacebook to be set to true.")
				}

				if sessionInfo.User.IsNew() {
					t.Fatal("Expected isNew to be false.")
				}

				if u.ObjectId() != test.id {
					t.Fatal("Expected:", test.id, "got:", u.ObjectId())
				}

				if sessionInfo.User.ObjectId() != test.id {
					t.Fatal("Expected:", test.id, "got:", u.ObjectId())
				}

				if u.Collection() != models.CollectionUser {
					t.Fatal("Expected:n", models.CollectionUser, "got:", u.Collection())
				}

			}

		}

	})
}

// Setup

func setupFBLoginTests(t *testing.T, ds db.DataStore) (*gin.Engine, []*FBLoginTest) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Connect())
	router.POST(routes.FACEBOOK_LOGIN, FacebookLogin)

	var tests []*FBLoginTest

	for _, test := range fbLoginTests {
		user, err := models.NewUserFromFacebookAuth(test.token, test.profileId, test.expiry)

		if err != nil {
			t.Fatal("Could not setup test datastore:", err)
		}

		if err := user.Save(ds); err != nil {
			t.Fatal("Could not setup test datastore:", err)
		}
		fmt.Printf("Created test user: %s: %s \n", user.ObjectId(), user.AuthData["id"])
		tests = append(tests, &FBLoginTest{user.ObjectId(), test.token, test.profileId, test.expiry, test.payload, test.responseCode})
	}

	return router, tests

}
