package login

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"time"

	"github.com/nidhik/backend/auth"

	"github.com/gin-gonic/gin"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
	"github.com/nidhik/backend/query"
)

func recordPost(router *gin.Engine, url string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func recordPostForm(router *gin.Engine, url string, form url.Values) *httptest.ResponseRecorder {

	req, _ := http.NewRequest("POST", url, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
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

func setupFBUser(t *testing.T, ds db.DataStore, email string) (*models.User, string) {
	user, err := models.NewUserFromFacebookAuth("abcd", "123456", time.Now().AddDate(1, 0, 0))
	query.AssertNoError(t, "Could not set up test user:", err)

	user.Set("Email", email)
	err = user.Save(ds)
	query.AssertNoError(t, "Could not set up test user:", err)

	validToken, err := auth.CreateToken(user, time.Now().AddDate(1, 0, 0))
	query.AssertNoError(t, "Could not set up test token:", err)

	return user, validToken
}
func setupUsers(t *testing.T, ds db.DataStore, username string, email string) (*models.User, string) {
	user := models.NewEmptyUser()
	user.Set("Username", username)
	user.Set("Email", email)
	err := user.Save(ds)
	query.AssertNoError(t, "Could not set up test user:", err)

	validToken, err := auth.CreateToken(user, time.Now().AddDate(1, 0, 0))
	query.AssertNoError(t, "Could not set up test token:", err)

	return user, validToken
}

func FindNewTasks(t *testing.T, ds db.DataStore) []*models.Task {
	var actual []*models.Task

	err := models.FindEachNewTask(ds, func(t *models.Task) {
		actual = append(actual, t)
	})

	query.AssertNoError(t, "Could not find new tasks:", err)
	return actual
}
