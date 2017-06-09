package query

import (
	"fmt"
	"strconv"
	"testing"

	"reflect"

	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
	"gopkg.in/mgo.v2/bson"
)

var runAtomicIncTest bool = true

func TestAtomicIncrement(t *testing.T) {

	if runAtomicIncTest {

		for i := 0; i < 10; i++ {
			RunTest(t, func(t *testing.T, ds db.DataStore) {

				task := models.NewEmptyTask()
				if error := task.Save(ds); error != nil {
					t.Fatal("Could not create task:", error)
				}

				messages := make(chan string)

				for i := 0; i < 20; i++ {
					taskCopy := models.NewTask(task.ObjectId())
					go claimTask(t, i, taskCopy, ds, messages)
				}

				for i := 0; i < 20; i++ {
					msg := <-messages
					fmt.Printf("goroutine: %s \n", msg)
				}

				// refresh the task
				if error := task.Fetch(ds); error != nil {
					t.Fatal("Could not refresh task:", error)
				}

				if task.Claimed != 20 {
					t.Fatal("Expected task to be claimed 20 separate times. Expected task.Claimed = 20. Actual was:", task.Claimed)
				}

				fmt.Printf("Task claimed value is: %d \n", task.Claimed)
				fmt.Printf("Task message: %s \n", task.Message)

			})

		}
	}

}

func claimTask(t *testing.T, i int, task *models.Task, ds db.DataStore, done chan string) {

	id := strconv.Itoa(i)
	task.Increment("Claimed", 1)
	err := task.Save(ds)

	if err != nil {
		t.Fatal("Error saving task %s", err)
	}

	if err == nil {
		msg := id
		pos := strconv.Itoa(task.Claimed)

		if task.Claimed == 1 {
			msg = msg + " claimed task first!"
			task.Set("Message", msg)
			if err := task.Save(ds); err != nil {
				t.Fatal("Error saving task %s", err)
			} else {
				done <- msg
			}

		} else {
			msg = msg + " was #" + pos + " to claim task "
			done <- msg
		}
	}
}

func TestTaskParams(t *testing.T) {

	RunTest(t, func(t *testing.T, ds db.DataStore) {

		// set up related data
		user := models.NewEmptyUser()
		user.Set("Name", "FooBarBaz")
		if err := user.Save(ds); err != nil {
			t.Fatal("Could not set up datastore for test.", err)
		}

		task0 := models.NewTaskForUser(user, "TEST_TASK0", "TEST_ACTION", "PLAN_COMPLETED", models.AsPointer(user))
		if error := task0.Save(ds); error != nil {
			t.Fatal("Could not create task:", error)
		}

		task1 := models.NewTaskForUser(user, "TEST_TASK1", "TEST_ACTION", bson.M{"foo": "bar"})
		if error := task1.Save(ds); error != nil {
			t.Fatal("Could not create task:", error)
		}

		task2 := models.NewTaskForUser(user, "TEST_TASK2", "TEST_ACTION", []string{"foo", "bar", "baz"})
		if error := task2.Save(ds); error != nil {
			t.Fatal("Could not create task:", error)
		}

		task3 := models.NewTaskForUser(models.NewUser("abc"), "TEST_TASK3", "TEST_ACTION", []string{"foo", "bar", "baz"})
		if error := task3.Save(ds); error != nil {
			t.Fatal("Could not create task:", error)
		}

		tests := []struct {
			user         *models.User
			task         *models.Task
			hasErr       bool
			loadedValues []interface{}
		}{
			{user, task0, false, []interface{}{"PLAN_COMPLETED", user}},
			{user, task1, false, []interface{}{bson.M{"foo": "bar"}}},
			{user, task2, false, []interface{}{[]string{"foo", "bar", "baz"}}},
			{task3.User, task3, true, nil},
		}

		// Run Tests

		for _, test := range tests {
			fetched, user, err := test.task.FetchParameters(ds)

			if test.hasErr == (err == nil) {
				t.Fatal("Expected to have error:", test.hasErr, "Got error:", err)
			}

			if test.user != nil && test.user.Name != "" {
				if user == nil || user.Name != test.user.Name {
					t.Fatal("Expected to load user with name", test.user.Name, ".Actual user was:", user)
				}
			}

			if !reflect.DeepEqual(test.loadedValues, fetched) {
				t.Fatal("Expected to load:", test.loadedValues, "Actual:", fetched)
			}

		}

	})

}
