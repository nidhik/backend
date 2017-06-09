package login

import (
	"strings"
	"time"

	"github.com/nidhik/backend/auth"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
	"gopkg.in/mgo.v2/bson"
)

// Forgot Password

func initiateReset(user *models.User, ds db.DataStore) error {
	expiry := time.Now().Add(time.Minute * 20) // 20 min from now
	token, err := auth.CreateToken(user, expiry)

	if err != nil {
		return err
	}

	task := models.NewTaskForUser(user, models.TaskTypeEmail, models.HandleEmail, "PASSWORD_RESET_GO", models.AsPointer(user), bson.M{"token": token})
	return task.Save(ds)

}

// Username & Password Login

func authenticate(creds LoginCredentials, ds db.DataStore) (*models.User, error) {
	username := strings.TrimSpace(creds.Username)
	if user, err := models.FindUserByUsername(ds, username); err != nil {
		return nil, err
	} else {
		authErr := user.CheckPassword(creds.Password)
		return user, authErr
	}
}

func createToken(user *models.User) (string, error) {
	expiry := time.Now().AddDate(1, 0, 0) // 1 year from now
	return auth.CreateToken(user, expiry)
}

func login(creds LoginCredentials, ds db.DataStore) (*models.User, string, error) {
	if user, err := authenticate(creds, ds); err != nil {
		return nil, "", err
	} else {
		token, err := createToken(user)
		return user, token, err
	}
}

// Logout

func logout(user *models.User, ds db.DataStore) error {
	// FIXME: do nothign for now, ideally we would expire the JWT token
	return nil

}
