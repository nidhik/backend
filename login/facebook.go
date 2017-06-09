package login

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"time"
)

var ERR_FB_APP_ACCESS = errors.New("Error with FB app access.")
var ERR_FB_TOKEN = errors.New("Error with FB User token")
var debugTokenURL = "https://graph.facebook.com/debug_token"
var meURL = "https://graph.facebook.com/me"

type FBVerifiedToken struct {
	AccessToken    string    `json:"accessToken"`
	ExpirationDate time.Time `json:"expirationDate"`
	Id             string    `json:"id"`
}

type FBUser struct {
	ProfileId string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	Gender    string `json:"gender"`
}

func decodeFBResponseIntoMap(resp *http.Response) (map[string]interface{}, error) {

	d := json.NewDecoder(resp.Body)
	d.UseNumber()

	var x map[string]interface{}
	if decodeErr := d.Decode(&x); decodeErr != nil {
		return nil, decodeErr
	}

	data := x["data"].(map[string]interface{})

	if apiErr := data["error"]; apiErr != nil {
		message := apiErr.(map[string]interface{})["message"].(string)
		return nil, errors.New(message)
	}

	return data, nil
}

func makeRequest(token string) (*http.Response, error) {

	appAccessToken := os.Getenv("FB_APP_ACCESS_TOKEN")

	if len(appAccessToken) == 0 {
		return nil, ERR_FB_APP_ACCESS
	}

	inputTokenParam := "input_token=" + token
	accessTokenParam := "access_token=" + appAccessToken
	return http.Get(debugTokenURL + "?" + inputTokenParam + "&" + accessTokenParam)
}

func validateFbUserAccessToken(token string) (*FBVerifiedToken, error) {

	// Make request
	resp, nErr := makeRequest(token)

	if nErr != nil {
		return nil, nErr
	}

	// Decode response
	defer resp.Body.Close()
	data, dErr := decodeFBResponseIntoMap(resp)
	if dErr != nil {
		return nil, dErr
	}

	// Construct and return verified token
	expiresAt := data["expires_at"].(json.Number)

	seconds, pErr := strconv.ParseInt(expiresAt.String(), 10, 64)

	if pErr != nil {
		return nil, pErr
	}
	expiry := time.Unix(seconds, 0).UTC()
	return &FBVerifiedToken{AccessToken: token, ExpirationDate: expiry, Id: data["user_id"].(string)}, nil

}

func makeMeRequest(token string) (*http.Response, error) {
	if len(token) == 0 {
		return nil, ERR_FB_TOKEN
	}

	tokenParam := "access_token=" + token
	return http.Get(meURL + "?" + tokenParam)
}

func me(token string) (*FBUser, error) {
	// Make request
	resp, nErr := makeMeRequest(token)

	if nErr != nil {
		return nil, nErr
	}

	// Decode response
	defer resp.Body.Close()

	d := json.NewDecoder(resp.Body)
	d.UseNumber()

	var x FBUser
	if decodeErr := d.Decode(&x); decodeErr != nil {
		return nil, decodeErr
	}

	return &x, nil

}
