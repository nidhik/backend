package middleware

import (
	"fmt"
	"net/http"

	"os"

	"github.com/gin-gonic/gin"
	"github.com/nidhik/backend/auth"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
	"github.com/nidhik/backend/query"
)

// Database Connection

var AllowedOrigin = os.Getenv("ALLOWED_ORIGIN")

func Preflight() gin.HandlerFunc {

	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", AllowedOrigin)
		c.Header("Access-Control-Allow-Headers", fmt.Sprintf("Content-Type,Access-Control-Allow-Origin,Access-Control-Allow-Headers,%s,%s", CLIENT_KEY_HEADER, SESSION_HEADER))
		c.Header("Access-Control-Allow-Methods", "POST,PUT,DELETE,GET")
		c.JSON(http.StatusOK, struct{}{})
	}
}

// Database

func Connect() gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method == http.MethodOptions {
			fmt.Println("Preflight request, allowing through.")
			c.Next()
			return
		}

		datastore := db.GetDataStore(query.NewMongoQueryBuilder())

		defer datastore.Close()
		c.Set("ds", datastore)

		c.Next()
	}
}

// Approved API Consumers

var CLIENT_KEY_HEADER = "X-Client-Key"

func API() gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method == http.MethodOptions {
			fmt.Println("Preflight request, allowing through.")
			c.Next()
			return
		}

		key := c.Request.Header.Get(CLIENT_KEY_HEADER)
		if err := auth.IsApprovedAPIConsumer(key); err == nil {
			c.Header("Access-Control-Allow-Origin", AllowedOrigin)
			c.Next()

		} else {
			fmt.Printf("Error: %s", err)
			c.AbortWithStatus(http.StatusNotFound)
		}
	}
}

// Authorization

var SESSION_HEADER = "X-Session-Token"

func authenticateToken(ds db.DataStore, token string) (*models.User, []*models.Role, error) {
	user, err := auth.VerifyToken(token)
	if err != nil {
		return nil, nil, err
	}

	if err := user.Fetch(ds); err != nil {
		return nil, nil, err
	}

	roles, err := models.FindRolesForUser(user, ds)
	if err != nil {
		return nil, nil, err
	}

	return user, roles, nil
}

func AuthorizedLink() gin.HandlerFunc {
	return func(c *gin.Context) {
		ds := c.MustGet("ds").(db.DataStore)
		token := c.Query("token")

		if user, roles, err := authenticateToken(ds, token); err == nil {
			c.Set("user", user)
			c.Set("roles", roles)

			ds.SetQueryBuilder(query.NewRestrictedQueryBuilder(user, roles))
		} else {
			fmt.Printf("Error: %s", err)
			c.AbortWithStatus(http.StatusNotFound)
		}
	}
}

func AdminFunctionUserAuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		ds := c.MustGet("ds").(db.DataStore)
		token := c.Request.Header.Get(SESSION_HEADER)

		if user, roles, err := authenticateToken(ds, token); err == nil {
			c.Set("user", user)
			c.Set("roles", roles)
		} else {
			fmt.Printf("Error: %s", err)
			c.AbortWithStatus(http.StatusForbidden)
		}
	}
}

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method == http.MethodOptions {
			fmt.Println("Preflight request, allowing through.")
			c.Next()
			return
		}

		ds := c.MustGet("ds").(db.DataStore)
		token := c.Request.Header.Get(SESSION_HEADER)
		if user, roles, err := authenticateToken(ds, token); err == nil {
			c.Set("user", user)
			c.Set("roles", roles)

			ds.SetQueryBuilder(query.NewRestrictedQueryBuilder(user, roles))
		} else {
			fmt.Printf("Error: %s", err)
			c.AbortWithStatus(http.StatusForbidden)
		}
	}
}
