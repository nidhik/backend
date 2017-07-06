package models

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/utils"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

var ERR_INVALID_EMAIL = errors.New("Invalid email")
var ERR_INVALID_PASSWORD = errors.New("Invalid password.")
var ERR_INVALID_USERNAME = errors.New("Invalid username.")
var ERR_INVALID_FIRST_NAME = errors.New("Invalid First Name.")

var strict = bluemonday.StrictPolicy()

const (
	// CollectionUser holds the name of the users collection
	CollectionUser = "_User"
)

type User struct {
	Email          string                 `json:"email,omitempty" bson:"email"`
	Username       string                 `json:"username,omitempty" binding:"required" bson:"username"`
	HashedPassword []byte                 `json:"-" binding:"required" bson:"_hashed_password"`
	Name           string                 `json:"name,omitempty" bson:"name"`
	FirstName      string                 `json:"firstName,omitempty" bson:"firstName"`
	Gender         string                 `json:"gender,omitempty" bson:"gender"`
	AuthData       map[string]interface{} `json:"-" bson:"_auth_data_facebook,omitempty"`
	CustomFields   map[string]interface{} `json:"customFields" bson:"customFields,omitempty"`
	db.BaseModel   `bson:",inline"`
}

func NewUser(id string) *User {
	model := &User{BaseModel: db.BaseModel{
		Id: id, CollectionName: CollectionUser},
	}
	return model
}

func NewEmptyUser() *User {

	return &User{
		BaseModel: db.BaseModel{CollectionName: CollectionUser},
	}
}

func NewUserFromFacebookAuth(token string, profileId string, expiresAt time.Time) (*User, error) {

	data := make(map[string]interface{})
	data["access_token"] = token
	data["expiration_date"] = expiresAt.String()
	data["id"] = profileId

	user := NewEmptyUser()
	user.Set("AuthData", data)

	return user, nil
}

func NewUserFromEmail(email string, username string, password string, firstName string) (*User, error) {

	// Required Fields
	e := strings.TrimSpace(email)
	u := strings.TrimSpace(username)
	p := strings.TrimSpace(password)

	if len(e) == 0 || !strings.Contains(email, "@") {
		return nil, ERR_INVALID_EMAIL
	}

	if len(u) == 0 {
		return nil, ERR_INVALID_USERNAME
	}

	sanitizedUsername := strict.Sanitize(username)
	if len(sanitizedUsername) != len(username) {
		fmt.Println("Username sanitize check failed:" + email)
		return nil, ERR_INVALID_USERNAME
	}

	if len(p) == 0 || len(p) != len(password) {
		return nil, ERR_INVALID_PASSWORD
	}

	pass := []byte(password)
	hashed, err := bcrypt.GenerateFromPassword(pass, bcrypt.MinCost)
	if err != nil {
		return nil, err
	}

	user := &User{
		BaseModel: db.BaseModel{CollectionName: CollectionUser},
	}

	user.Set("Email", e)
	user.Set("Username", u)
	user.Set("HashedPassword", hashed)

	// Optional Fields
	if len(firstName) > 0 {
		fn := strings.TrimSpace(firstName)

		if len(fn) == 0 {
			return nil, ERR_INVALID_FIRST_NAME
		}

		sanitizedFn := strict.Sanitize(fn)
		if len(sanitizedFn) != len(fn) {
			fmt.Println("First name sanitize check failed:" + email)
			return nil, ERR_INVALID_FIRST_NAME
		}

		formattedName := utils.Capitalize(fn)
		user.Set("FirstName", formattedName)
	}

	return user, nil
}

func (user *User) ChangePassword(password string, ds db.DataStore) error {
	p := strings.TrimSpace(password)

	if len(p) == 0 || len(p) != len(password) {
		return ERR_INVALID_PASSWORD
	}
	pass := []byte(password)
	hashed, err := bcrypt.GenerateFromPassword(pass, bcrypt.MinCost)
	if err != nil {
		return err
	}

	user.Set("HashedPassword", hashed)
	return user.Save(ds)
}

func (user *User) Fetch(ds db.DataStore) error {
	return user.BaseModel.Fetch(user, ds)
}

func (user *User) Save(ds db.DataStore) error {
	return user.BaseModel.Save(user, ds)
}

func (user *User) Delete(ds db.DataStore) error {
	return user.BaseModel.Delete(user, ds)
}

func (user *User) Set(fieldName string, value interface{}) {
	user.BaseModel.Set(user, fieldName, value)
}

func (user *User) Unset(fieldName string) {
	user.BaseModel.Unset(user, fieldName)
}

func (user *User) Get(fieldName string) interface{} {
	return user.BaseModel.Get(user, fieldName)
}

func (user *User) Increment(fieldName string, amount int) {
	user.BaseModel.Increment(user, fieldName, amount)
}

func (user *User) CustomUnmarshall() {
	user.CollectionName = CollectionUser
}

func (user *User) IsFacebook() bool {
	return user.AuthData["id"] != nil
}

func (user *User) SetCustomField(key string, val interface{}) {
	dst := make(map[string]interface{})
	for k, v := range user.CustomFields {
		dst[k] = v
	}

	dst[key] = val
	user.Set("CustomFields", dst)
}

func CountUsersWithUsernameOrEmail(ds db.DataStore, username string, email string) (int, error) {

	e := strings.TrimSpace(email)
	u := strings.TrimSpace(username)

	return ds.Count(CollectionUser, bson.M{"$or": []bson.M{bson.M{"username": u}, bson.M{"email": e}}})
}

func UpsertUserByFacebookProfile(ds db.DataStore, fbProfileId string, token string, expiresAt time.Time) (*User, error) {
	if user, err := NewUserFromFacebookAuth(token, fbProfileId, expiresAt); err != nil {
		return nil, err
	} else {
		if err2 := ds.UpsertObject(user, bson.M{"_auth_data_facebook.id": fbProfileId}); err2 != nil {
			return nil, err2
		}
		return user, nil
	}
}

func FindUserByUsername(ds db.DataStore, username string) (*User, error) {
	var user User
	err := ds.FindObject(CollectionUser, bson.M{"username": username}, &user)
	return &user, err
}

func FindUserByEmail(ds db.DataStore, email string) (*User, error) {
	var user User
	err := ds.FindObject(CollectionUser, bson.M{"email": email}, &user)
	return &user, err
}

func (user *User) CheckPassword(password string) error {

	pass := []byte(password)
	perr := bcrypt.CompareHashAndPassword(user.HashedPassword, pass)
	return perr

}
