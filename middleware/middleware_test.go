package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nidhik/backend/auth"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
	"github.com/nidhik/backend/query"
)

func setup(route string, finalHandler gin.HandlerFunc, middleware ...gin.HandlerFunc) *gin.Engine {

	gin.SetMode(gin.TestMode)
	router := gin.New()
	for _, h := range middleware {
		router.Use(h)
	}
	router.GET(route, finalHandler)
	return router
}

func TestAPIConsumer(t *testing.T) {
	query.RunTest(t, func(t *testing.T, ds db.DataStore) {

		var validKey = os.Getenv("CLIENT_KEY")

		tests := []struct {
			handler  gin.HandlerFunc
			route    string
			url      string
			headers  map[string]string
			respCode int
		}{
			{Get200(), "/", "/randomdddakjfnjdns", nil, 404},
			{Get200(), "/", "/", nil, 404},
			{Get200(), "/", "/", map[string]string{CLIENT_KEY_HEADER: validKey}, 200},
			{Get200(), "/", "/", map[string]string{CLIENT_KEY_HEADER: "derpy"}, 404},
		}

		for _, test := range tests {
			router := setup(test.route, test.handler, API())
			resp := recordGet(router, test.url, test.headers)
			if resp.Code != test.respCode {
				t.Fatal("Expected response code", test.respCode, "Got:", resp.Code, "Response:", resp)
			}
		}
	})

}

func TestAuthRequired(t *testing.T) {
	query.RunTest(t, func(t *testing.T, ds db.DataStore) {

		user, roles, validToken := setupUsersAndRoles(t, ds)

		tests := []struct {
			handler  gin.HandlerFunc
			route    string
			url      string
			headers  map[string]string
			respCode int
		}{
			{Get200(), "/", "/randomdddakjfnjdns", nil, 403},
			{Get200(), "/", "/", nil, 403},
			{Get200(), "/", "/", map[string]string{SESSION_HEADER: validToken}, 200},
			{Get200(), "/", "/", map[string]string{SESSION_HEADER: "derpy"}, 403},
			{CheckForUserAndRoles(user, nil), "/", "/", map[string]string{SESSION_HEADER: validToken}, 500},
			{CheckForUserAndRoles(models.NewUser("wrong user"), nil), "/", "/", map[string]string{SESSION_HEADER: validToken}, 500},
			{CheckForUserAndRoles(user, roles...), "/", "/", map[string]string{SESSION_HEADER: validToken}, 200},
		}

		for _, test := range tests {
			router := setup(test.route, test.handler, Connect(), AuthRequired())
			resp := recordGet(router, test.url, test.headers)
			if resp.Code != test.respCode {
				t.Fatal("Expected response code", test.respCode, "Got:", resp.Code, "Response:", resp)
			}
		}
	})

}

func TestAuthorizedLink(t *testing.T) {

	query.RunTest(t, func(t *testing.T, ds db.DataStore) {

		user, roles, validToken := setupUsersAndRoles(t, ds)

		tests := []struct {
			handler  gin.HandlerFunc
			route    string
			url      string
			respCode int
		}{
			{Get200(), "/reset", "/reset?token=foobar", 404},
			{Get200(), "/reset", "/reset?token=foobar", 404},
			{Get200(), "/reset", "/reset?token=" + validToken, 200},
			{CheckForUserAndRoles(user, nil), "/reset", "/reset?token=" + validToken, 500},
			{CheckForUserAndRoles(models.NewUser("Wrong user"), nil), "/reset", "/reset?token=" + validToken, 500},
			{CheckForUserAndRoles(user, roles...), "/reset", "/reset?token=" + validToken, 200},
		}

		for _, test := range tests {
			router := setup(test.route, test.handler, Connect(), AuthorizedLink())
			resp := recordGet(router, test.url, nil)
			if resp.Code != test.respCode {
				t.Fatal("Expected response code", test.respCode, "Got:", resp.Code, "Response:", resp)
			}
		}
	})

}

func TestConnect(t *testing.T) {

	query.RunTest(t, func(t *testing.T, ds db.DataStore) {

		tests := []struct {
			handler  gin.HandlerFunc
			route    string
			url      string
			respCode int
		}{
			{Get200(), "/", "/", 200},
			{CheckForDatabase(), "/", "/", 200},
		}

		for _, test := range tests {
			router := setup(test.route, test.handler, Connect())
			resp := recordGet(router, test.url, nil)
			if resp.Code != test.respCode {
				t.Fatal("Expected response code", test.respCode, "Got:", resp.Code)
			}
		}

	})

}

// Test Handlers

func Get200() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, "Ok")
	}
}

func CheckForDatabase() gin.HandlerFunc {
	return func(c *gin.Context) {

		_, exists := c.Get("ds")

		if !exists {
			c.JSON(http.StatusInternalServerError, "No datastore found.")
		} else {
			c.JSON(http.StatusOK, "Ok")
		}

	}
}

func CheckForUserAndRoles(user *models.User, roles ...*models.Role) gin.HandlerFunc {
	return func(c *gin.Context) {

		u, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusInternalServerError, "User not set.")
			return

		}
		r, exists := c.Get("roles")
		if !exists {
			c.JSON(http.StatusInternalServerError, "Roles not set.")
			return
		}

		if u, ok := u.(*models.User); !ok || u.ObjectId() != user.ObjectId() {
			c.JSON(http.StatusInternalServerError, "User is not set correctly.")
			return
		}
		if r, ok2 := r.([]*models.Role); !ok2 || !equalSlice(roles, r) {
			c.JSON(http.StatusInternalServerError, "Roles are not set correctly.")
			return
		}

		c.JSON(http.StatusOK, "Ok")

	}
}

// Helpers

func setupUsersAndRoles(t *testing.T, ds db.DataStore) (*models.User, []*models.Role, string) {
	user := models.NewEmptyUser()
	if err := user.Save(ds); err != nil {
		t.Fatal("Could not set up test user.", err)
	}

	var roles = []*models.Role{models.NewEmptyRole(), models.NewEmptyRole()}
	for _, r := range roles {

		if err := r.Save(ds); err != nil {
			t.Fatal("Could not set up test role.", err)
		}
		r.Users.Add(user)
		if err := ds.SaveRelatedObjects(r.Users); err != nil {
			t.Fatal("Could not set up test role.", err)
		}
	}

	validToken, err := auth.CreateToken(user, time.Now().AddDate(1, 0, 0))
	if err != nil {
		t.Fatal("Could not set up test token.", err)
	}

	return user, roles, validToken
}

func equalSlice(exp []*models.Role, actual []*models.Role) bool {

	if len(exp) != len(actual) {
		return false
	}

	for i, r := range exp {
		if actual[i].ObjectId() != r.ObjectId() {
			return false
		}
	}

	return true
}

func recordGet(router *gin.Engine, url string, headers map[string]string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	for key, val := range headers {
		req.Header.Set(key, val)
	}
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}
