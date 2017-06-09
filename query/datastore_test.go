package query

import (
	"testing"

	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
	"gopkg.in/mgo.v2/bson"
)

func TestIsNew(t *testing.T) {
	RunTest(t, func(t *testing.T, ds db.DataStore) {

		u1 := models.NewEmptyUser()
		u1.Save(ds)

		if !u1.IsNew() {
			t.Fatal("Expected isNew to be true.")
		}

		u2 := models.NewUser(u1.ObjectId())
		if u2.IsNew() {
			t.Fatal("Expected isNew to be false.")
		}
		u2.Fetch(ds)
		if u2.IsNew() {
			t.Fatal("Expected isNew to be false.")
		}

		u3 := models.NewEmptyUser()
		u3.Set("Email", "what@baz.com")
		u3.Save(ds)

		if !u3.IsNew() {
			t.Fatal("Expected isNew to be true.")
		}

		// Make sure Upsert works too

		upsertedUser := models.NewEmptyUser()
		upsertedUser.Set("Email", "whatupdates@baz.com")
		ds.UpsertObject(upsertedUser, bson.M{"email": "what@baz.com"})

		if upsertedUser.IsNew() {
			t.Fatal("Expected isNew to be false.")
		}

		if upsertedUser.Email != "whatupdates@baz.com" {
			t.Fatal("Expected email to equal [whatupdates@baz.com]. Actual:", upsertedUser.Email)
		}

		if upsertedUser.ObjectId() != u3.ObjectId() {
			t.Fatal("Expected id of upserted document to be", u3.ObjectId(), "Actual:", upsertedUser.ObjectId())
		}

		upsertedUser2 := models.NewEmptyUser()
		upsertedUser2.Set("Email", "email2@bar.com")
		ds.UpsertObject(upsertedUser2, bson.M{"email": "foo@bar.edu"})

		if !upsertedUser2.IsNew() {
			t.Fatal("Expected isNew to be true.")
		}

		if upsertedUser2.Email != "email2@bar.com" {
			t.Fatal("Expected email to equal [email2@bar.com]. Actual:", upsertedUser2.Email)
		}

		if upsertedUser2.ObjectId() == u3.ObjectId() {
			t.Fatal("Not a new document.", upsertedUser2.ObjectId())
		}

	})
}
