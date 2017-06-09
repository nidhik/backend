package controllers

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/middleware"
	"github.com/nidhik/backend/query"
)

// Test Harness
func TestControllers(t *testing.T) {

	fmt.Println()
	fmt.Println("Testing: Role Controller:")
	testCRUD(t, &RolesControllerTest{})

}

const (
	GET    = iota
	POST   = iota
	PUT    = iota
	DELETE = iota
)

type TestCase interface {
	description() string
}

type ControllerTest interface {
	routeAndHandler(method int) (string, gin.HandlerFunc)
	setupDataStore(t *testing.T, ds db.DataStore)
	testCases(method int) []TestCase
	runGetTest(t *testing.T, test interface{}, router *gin.Engine)
	runPostTest(t *testing.T, test interface{}, router *gin.Engine)
	runDeleteTest(t *testing.T, test interface{}, router *gin.Engine)
	runUpdateTest(t *testing.T, test interface{}, router *gin.Engine)
}

func setup(c ControllerTest) *gin.Engine {

	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware.Connect())

	r, e := c.routeAndHandler(GET)
	router.GET(r, e)

	r, e = c.routeAndHandler(POST)
	router.POST(r, e)

	r, e = c.routeAndHandler(PUT)
	router.PUT(r, e)

	r, e = c.routeAndHandler(DELETE)
	router.DELETE(r, e)

	return router
}

func testCRUD(t *testing.T, c ControllerTest) {

	query.RunTest(t, func(t *testing.T, ds db.DataStore) {
		c.setupDataStore(t, ds)
		router := setup(c)

		for _, test := range c.testCases(GET) {
			fmt.Printf("Test %s \n", test.description())
			c.runGetTest(t, test, router)
		}

		for _, test := range c.testCases(POST) {
			fmt.Printf("Test %s \n", test.description())
			c.runPostTest(t, test, router)
		}

		for _, test := range c.testCases(PUT) {
			fmt.Printf("Test %s \n", test.description())
			c.runUpdateTest(t, test, router)
		}

		for _, test := range c.testCases(DELETE) {
			fmt.Printf("Test %s \n", test.description())
			c.runDeleteTest(t, test, router)
		}
	})

}

// Helpers

func recordGet(router *gin.Engine, url string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("GET", url, body)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func recordPost(router *gin.Engine, url string, body io.Reader) *httptest.ResponseRecorder {
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)
	return resp
}

func getDateFromTime(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}
