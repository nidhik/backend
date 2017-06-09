package login

import (
	"bytes"
	"encoding/json"

	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nidhik/backend/auth"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/middleware"
	"github.com/nidhik/backend/models"
	"github.com/nidhik/backend/query"
	"github.com/nidhik/backend/routes"
)

// Tests

type LoginTest struct {
	id           string
	email        string
	username     string
	password     string
	payload      []byte
	responseCode int
}

var loginTests = []LoginTest{

	LoginTest{"", "nidhi@foo.com", "nidhi", "ab9pnlkj4c", []byte(`{"foo" : "bar"}`), http.StatusBadRequest},
	LoginTest{"", "erin@foo.com", "erin", "qwerlugkuytty", []byte(`{"username" : "erin", "password" : "abc"}`), http.StatusUnauthorized},
	LoginTest{"", "amy@foo.com", "amy", "po6hkuygiuy", []byte(`{"username" : "not amy", "password": "po6hkuygiuy"}`), http.StatusUnauthorized},
	LoginTest{"", "karen@foo.com", "karen", "127fj7%$", []byte(`{"username" : "    karen   ", "password" :"127fj7%$"}`), http.StatusOK},
	LoginTest{"", "karen@foo.com", "karen2", "127fj7%$", []byte(`{"username" : "karen2", "password" :"127fj7%$"}`), http.StatusOK},
	LoginTest{"", "derp@foo.com", "derp", "127fj7%$", []byte(`{"username" : "<derp", "password" :"127fj7%$"}`), http.StatusForbidden},
	LoginTest{"", "derp2@foo.com", "derp2", "127fj7%$", []byte(`{"username" : "Hello <STYLE>.XSS{background-image:url(\"javascript:alert('XSS')\");}</STYLE><A CLASS=XSS></A>World", "password" :"127fj7%$"}`), http.StatusForbidden},
	LoginTest{"", "evil@foo.com", "derp", "127fj7%$", []byte(`{"username" : ";var date=new Date(); do{curDate = new Date();}while(curDate-date<10000)", "password" :"127fj7%$"}`), http.StatusForbidden},
	LoginTest{"", "real1@foo.com", "bummy", "abc", []byte(`{"username" : ">)bummy(<", "password" :"abc"}`), http.StatusForbidden},
	LoginTest{"", "real2@foo.com", "kamy", "abc", []byte(`{"username" : "kamy<3", "password" :"abc"}`), http.StatusForbidden},
}

func TestLogin(t *testing.T) {

	query.RunTest(t, func(t *testing.T, ds db.DataStore) {

		router, tests := setupLoginTests(t, ds)

		for _, test := range tests {
			req, _ := http.NewRequest("POST", routes.LOGIN, bytes.NewBuffer(test.payload))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			if resp.Code != test.responseCode {
				t.Fatal("Expected:", test.responseCode, "got:", resp.Code)
			}

			if resp.Code == 200 {

				var sessionInfo UserSession
				json.Unmarshal(resp.Body.Bytes(), &sessionInfo)

				u, _ := auth.VerifyToken(sessionInfo.Token)

				if u.ObjectId() != test.id {
					t.Fatal("Expected:", test.id, "got:", u.ObjectId())
				}

				if u.IsFacebook() != sessionInfo.User.IsFacebook {
					t.Fatal("isFacebook does not match. Expected:", u.IsFacebook(), "Actual:", sessionInfo.User.IsFacebook)
				}

				if sessionInfo.User.IsNew() {
					t.Fatal("Expected isNew to be false.")
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

func setupLoginTests(t *testing.T, ds db.DataStore) (*gin.Engine, []LoginTest) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Connect())
	router.POST(routes.LOGIN, Login)

	var tests []LoginTest

	for _, test := range loginTests {
		user, err := models.NewUserFromEmail(test.email, test.username, test.password, "")

		if err != nil {
			t.Fatal("Could not setup test datastore:", err)
		}

		if err := user.Save(ds); err != nil {
			t.Fatal("Could not setup test datastore:", err)
		}

		fmt.Printf("Created test user: %s: %s %s \n", user.ObjectId(), user.Username, user.AuthData["id"])
		tests = append(tests, LoginTest{user.ObjectId(), test.email, test.username, test.password, test.payload, test.responseCode})
	}

	return router, tests
}
