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

type SignUpTests struct {
	username     string
	payload      []byte
	responseCode int
	expectedName string
}

var signupTests = []SignUpTests{
	{"invalid1", []byte(`{"foo" : "bar"}`), 400, ""},

	{"invalid2", []byte(`{"username" : " ", "password" : "abc", "email":"erin@foo.com"}`), 401, ""},
	{"invalid3", []byte(`{"username" : "", "password" : "abc", "email":"erin@foo.com"}`), 400, ""},

	{"invalid4", []byte(`{"username" : "nidhi", "password": " ", "email": "nidhi@foo.com"}`), 401, ""},
	{"invalid5", []byte(`{"username" : "nidhi", "password": "", "email": "nidhi@foo.com"}`), 400, ""},
	{"invalid6", []byte(`{"username" : "nidhi", "password": "   spaces around it", "email": "nidhi@foo.com"}`), 401, ""},

	{"invalid7", []byte(`{"username" : "amy", "password": "abc", "email": "invalid email"}`), 400, ""},
	{"invalid8", []byte(`{"username" : "amy", "password": "abcy", "email": "   "}`), 400, ""},
	{"invalid9", []byte(`{"username" : "amy", "password": "abc", "email": ""}`), 400, ""},

	{"xssattempt", []byte(`{"username" : "Hello <STYLE>.XSS{background-image:url(\"javascript:alert('XSS')\");}</STYLE><A CLASS=XSS></A>World", "password": "abc", "email": "xss@foo.com"}`), 401, ""},

	{"evil1", []byte(`{"username" : ";var date=new Date(); do{curDate = new Date();}while(curDate-date<10000)", "password": "abc", "email": "evil@foo.com"}`), 401, ""},
	{"evil2", []byte(`{"username" : "evil2", "password": "abc", "email": "function() { return obj.credits - obj.debits < 0;var date=new Date(); do{curDate = new Date();}while(curDate-date<10000); }"}`), 400, ""},

	{"realderp1", []byte(`{"username" : ">)bummy(<", "password" : "abc", "email":"bummy@foo.com"}`), 401, ""},
	{"realderp2", []byte(`{"username" : "kamy<3", "password" : "abc", "email":"kamy@foo.com"}`), 401, ""},

	{"mo", []byte(`{"username" : "mo", "password": "123456", "email": "mo@foo.com"}`), 401, ""},
	{"mo", []byte(`{"username" : "mo     ", "password": "123456", "email": "mo@foo.com"}`), 401, ""},

	{"connie2", []byte(`{"username" : "connie2", "password": "123456", "email": "connie@bar.com"}`), 401, ""},
	{"connie2", []byte(`{"username" : "connie2", "password": "123456", "email": "connie@bar.com    "}`), 400, ""},

	{"badFirstName", []byte(`{"username" : "karen", "password" :"127fj7%$", "email":"karen@foo.com", "firstName": "   "}`), 401, ""},
	{"xssattempt2", []byte(`{"username" : "xss", "password": "abc", "email": "xss@foo.com", "firstName": "Hello <STYLE>.XSS{background-image:url(\"javascript:alert('XSS')\");}</STYLE><A CLASS=XSS></A>World"}`), 401, ""},
	{"evil3", []byte(`{"username" : "evil3", "password": "abc", "email": "avil3@foo.com", "firstName": ";var date=new Date(); do{curDate = new Date();}while(curDate-date<10000)"}`), 401, ""},

	{"goodName", []byte(`{"username" : "goodname", "password" :"127fj7%$>>>>", "email":"goodname@foo.com", "firstName": "Cindy Mukau"}`), 200, "Cindy Mukau"},
	{"derpgoodName", []byte(`{"username" : "derpgoodname", "password" :"127fj7%$>>>>", "email":"derpgoodname@foo.com", "firstName": "   Layla   "}`), 200, "Layla"},

	{"karen", []byte(`{"username" : "karen", "password" :"127fj7%$", "email":"karen@foo.com"}`), 200, ""},

	{"nidhi", []byte(`{"username" : "nidhi", "password" :"abcd1234", "email":"nidhi@foo.com", "firstName": "nidhi"}`), 200, "Nidhi"},
}

func TestSignup(t *testing.T) {

	query.RunTest(t, func(t *testing.T, ds db.DataStore) {
		router, tests := setupSignupTests(t, ds)

		for _, test := range tests {
			fmt.Println("Testing Signup for: " + test.username)

			req, _ := http.NewRequest("POST", routes.SIGNUP, bytes.NewBuffer(test.payload))
			req.Header.Set("Content-Type", "application/json")
			resp := httptest.NewRecorder()
			router.ServeHTTP(resp, req)

			if resp.Code != test.responseCode {
				t.Fatal("Expected:", test.responseCode, "got:", resp.Code, "for test username: ", test.username)
			}

			if resp.Code == 200 {

				var sessionInfo UserSession
				json.Unmarshal(resp.Body.Bytes(), &sessionInfo)

				u, _ := auth.VerifyToken(sessionInfo.Token)

				if len(u.ObjectId()) == 0 {
					t.Fatal("Did not set user id.")
				}

				if u.ObjectId() != sessionInfo.User.ObjectId() {
					t.Fatal("User Ids do not match. Returned", sessionInfo.User.ObjectId(), "Token user id is", u.ObjectId())
				}

				if u.IsFacebook() != sessionInfo.User.IsFacebook {
					t.Fatal("isFacebook does not match. Expected:", u.IsFacebook(), "Actual:", sessionInfo.User.IsFacebook)
				}

				if !sessionInfo.User.IsNew() {
					t.Fatal("Expected isNew to be true.")
				}

				if sessionInfo.User.FirstName != test.expectedName {
					t.Fatal("Expected name to be", test.expectedName, "Actual:", sessionInfo.User.FirstName)
				}

				if u.Collection() != models.CollectionUser {
					t.Fatal("Expected:n", models.CollectionUser, "got:", u.Collection())
				}

				if err := u.Fetch(ds); err != nil {
					t.Fatal("Could not lookup created user in database:", err)
				}

				if u.AccessControlList() == nil {
					t.Fatal("No ACL set for user.")
				}

				if !u.AccessControlList().CanRead(u.ObjectId()) {
					t.Fatal("User does not have read perm: ", u.ObjectId(), "ACL map", u.AccessControlList().ACL, "ACL read list:", u.AccessControlList().ReadAccess, "ACL write list:", u.AccessControlList().WriteAccess)
				}

				if !u.AccessControlList().CanWrite(u.ObjectId()) {
					t.Fatal("User does not have write perm", "ACL map", u.AccessControlList().ACL, "ACL read list:", u.AccessControlList().ReadAccess, "ACL write list:", u.AccessControlList().WriteAccess)
				}

				if !u.AccessControlList().CanRead("anyone else") {
					t.Fatal("User is not public read.", "ACL map", u.AccessControlList().ACL, "ACL read list:", u.AccessControlList().ReadAccess, "ACL write list:", u.AccessControlList().WriteAccess)
				}

				if u.AccessControlList().CanWrite("anyone else") {
					t.Fatal("User is public write!", "ACL map", u.AccessControlList().ACL, "ACL read list:", u.AccessControlList().ReadAccess, "ACL write list:", u.AccessControlList().WriteAccess)
				}

			}

		}
	})

}

// Setup

func setupSignupTests(t *testing.T, ds db.DataStore) (*gin.Engine, []SignUpTests) {

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Connect())
	router.POST(routes.SIGNUP, Signup)

	var mo = &models.User{
		Username:  "mo",
		Email:     "mo@bar.com",
		BaseModel: db.BaseModel{CollectionName: models.CollectionUser},
	}

	var connie = &models.User{
		Username:  "connie1",
		Email:     "connie@bar.com",
		BaseModel: db.BaseModel{CollectionName: models.CollectionUser},
	}

	if err := mo.Save(ds); err != nil {
		t.Fatal("Could not setup test datastore:", err)
	}

	if err := connie.Save(ds); err != nil {
		t.Fatal("Could not setup test datastore:", err)
	}

	return router, signupTests

}
