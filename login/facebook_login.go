package login

import (
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
)

// Facebook Login
func loginWithFacebook(authData FacebookAuthData, ds db.DataStore) (*models.User, string, error) {
	verifiedToken, err := validateFbUserAccessToken(authData.AccessToken)
	if err != nil {
		return nil, "", err
	}

	user, err := models.UpsertUserByFacebookProfile(ds, verifiedToken.Id, verifiedToken.AccessToken, verifiedToken.ExpirationDate)
	if err != nil {
		return nil, "", err
	}
	isNew := user.IsNew()

	if user.AccessControlList().IsZero() {

		acl := db.NewACL()
		acl.SetPublicRead()
		acl.AddRead(user.ObjectId())
		acl.AddWrite(user.ObjectId())
		user.SetAccessControlList(acl)

		if saveErr := user.Save(ds); saveErr != nil {
			return nil, "", saveErr
		}

	}

	if isNew {
		// Note: we do no care if there is an error since this is just extra info, return the user anyway
		if fbuser, err := me(verifiedToken.AccessToken); err == nil {
			if fbuser.Email != "" {
				user.Set("Email", fbuser.Email)
			}

			if fbuser.FirstName != "" {
				user.Set("FirstName", fbuser.FirstName)
			}

			if fbuser.Gender != "" {
				user.Set("Gender", fbuser.Gender)
			}

			if saveErr := user.Save(ds); saveErr != nil {
				return nil, "", saveErr
			}
		}
	}

	token, err := createToken(user)
	user.SetIsNew(isNew)
	return user, token, err

}
