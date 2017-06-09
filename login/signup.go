package login

import (
	"errors"

	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
)

// Email Signup
var ERR_USER_EXISTS = errors.New("This email or username is already taken.")

func createUser(info SignupInfo, ds db.DataStore) (*models.User, error) {
	if user, createErr := models.NewUserFromEmail(info.Email, info.Username, info.Password, info.FirstName); createErr != nil {
		return nil, createErr
	} else {

		if saveErr := user.Save(ds); saveErr != nil {
			return nil, saveErr
		}

		acl := db.NewACL()
		acl.SetPublicRead()
		acl.AddRead(user.ObjectId())
		acl.AddWrite(user.ObjectId())
		user.SetAccessControlList(acl)

		if saveErr := user.Save(ds); saveErr != nil {
			return nil, saveErr
		}

		user.SetIsNew(true)
		return user, nil
	}
}

func validateUser(info SignupInfo, ds db.DataStore) (*models.User, error) {
	n, err := models.CountUsersWithUsernameOrEmail(ds, info.Username, info.Email)

	if err != nil {
		return nil, err
	}

	if n > 0 {
		return nil, ERR_USER_EXISTS
	}

	return createUser(info, ds)
}

func signup(info SignupInfo, ds db.DataStore) (*models.User, string, error) {
	// basic email, username & password validation
	// make sure username & email is not already in database
	// if available, create user with this info
	user, err := validateUser(info, ds)

	if err != nil {
		return nil, "", err
	}

	task0 := models.NewTaskForUser(user, models.TaskTypeEmail, models.HandleEmail, "SIGN_UP_V2_GO", models.AsPointer(user))
	task1 := models.NewTaskForUser(user, models.TaskTypeEmail, models.HandleEmail, "PRO_AWARENESS_V2_GO", models.AsPointer(user))
	task0.Save(ds)
	task1.Save(ds)

	token, err := createToken(user)
	return user, token, err
}
