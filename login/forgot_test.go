package login

import (
	"bytes"
	"encoding/json"

	"testing"

	"github.com/gin-gonic/gin"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/middleware"

	"github.com/nidhik/backend/models"
	"github.com/nidhik/backend/query"
	"github.com/nidhik/backend/routes"
	"gopkg.in/mgo.v2/bson"
)

func TestForgotPassword(t *testing.T) {
	query.RunTest(t, func(t *testing.T, ds db.DataStore) {

		router := setupForgotTests(t, ds)
		foo, _ := setupUsers(t, ds, "foo", "foo@bar.com")
		fb, _ := setupFBUser(t, ds, "facebook@bar.com")

		tests := []struct {
			token        string
			user         *models.User
			payload      []byte
			respCode     int
			response     string
			checkForTask bool
		}{
			{"", foo, []byte(`{"email": "foo@bar.com"}`), 200, "Please check your email for password reset instructions.", true},
			{"", nil, []byte(`{"email": "invalid@bar.com"}`), 404, "", false},
			{"", fb, []byte(`{"email": "facebook@bar.com"}`), 200, "You created an account with Facebook. Please press the Facebook button to log in.", false},
			{"", nil, []byte(`{"email": "function() { return obj.credits - obj.debits < 0;var date=new Date(); do{curDate = new Date();}while(curDate-date<10000); }" }`), 400, "", false},
		}

		for _, test := range tests {
			resp := recordPost(router, routes.FORGOT, bytes.NewBuffer(test.payload))
			if test.respCode != resp.Code {
				t.Fatal("Expected code:", test.respCode, "Actual: ", resp.Code, "Response:", resp)
			} else if test.response != "" {

				var message string
				err := json.Unmarshal(resp.Body.Bytes(), &message)

				if err != nil {
					t.Fatal("COuld not unmarshall response:", resp)
				}

				if message != test.response {
					t.Fatal("Expected response:\n", test.response, "Actual:\n", message)
				}
			}

			if test.checkForTask {
				tasks := FindNewTasks(t, ds)
				if len(tasks) == 0 {
					t.Fatal("No tasks created.")
				}

				var created *models.Task
				for _, task := range tasks {
					if task.User.ObjectId() != test.user.ObjectId() {
						continue
					}
					created = task
					break
				}

				if created == nil {
					t.Fatal("Did not create password reset task.")
				}

				if created.Type != models.TaskTypeEmail {
					t.Fatal("Wrong task type:", created.Type)
				}

				if created.Action != models.HandleEmail {
					t.Fatal("Wrong action type:", created.Action)
				}

				if category, ok := created.Parameters[0].(string); !ok || category != "PASSWORD_RESET_GO" {
					t.Fatal("Expected PASSWORD_RESET category as first parameter. Actual parameter:", created.Parameters[0])
				}

				if pointer, ok := created.Parameters[1].(bson.M); !ok || pointer["objectId"] != test.user.ObjectId() {
					t.Fatal("Expected user as second parameter with id", test.user.ObjectId(), "Actual parameter:", created.Parameters[1])
				}

				if data, ok := created.Parameters[2].(bson.M); !ok || data["token"] == nil {
					t.Fatal("Expected third param to have token. Actual param:", created.Parameters[2])
				}

			}

		}

	})
}

// Setup

func setupForgotTests(t *testing.T, ds db.DataStore) *gin.Engine {

	gin.SetMode(gin.TestMode)
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Custom Middleware
	router.Use(middleware.Connect())
	router.POST(routes.FORGOT, ForgotPassword)

	return router

}
