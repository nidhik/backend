package query

import (
	"testing"

	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
)

func TestChangePassword(t *testing.T) {

	RunTest(t, func(t *testing.T, ds db.DataStore) {
		user, err := models.NewUserFromEmail("foo@bar.com", "foofoo", "abc", "")
		AssertNoError(t, "Could not set up test user:", err)

		err = user.Save(ds)
		AssertNoError(t, "Could not set up test user:", err)

		tests := []struct {
			user        *models.User
			newPassword string
			err         error
		}{
			{user, "ABCDEFG", nil},
			{user, "abc", nil},
			{user, "          ", models.ERR_INVALID_PASSWORD},
			{user, "    whitespace      ", models.ERR_INVALID_PASSWORD},
			// Note, there is no error here bc hashing the password is harmless. Hashed: $2a$04$YhQlD5R29WaASG12he23H.gVtSwpUf7YdjLEhQIPy3ZD4O.f/Sl.W
			{user, "function() { return obj.credits - obj.debits < 0;var date=new Date(); do{curDate = new Date();}while(curDate-date<10000); }", nil},
		}

		for _, test := range tests {
			err := user.ChangePassword(test.newPassword, ds)

			if test.err == nil {
				AssertNoError(t, "Unexpected error when changing password", err)

				err = user.CheckPassword(test.newPassword)
				AssertNoError(t, "New password does not authenticate.", err)

			} else if err != test.err {
				t.Fatal("Expected error on changing password:", test.err, "Actual:", err)
			}

		}
	})

}
