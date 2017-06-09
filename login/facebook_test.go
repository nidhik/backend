package login

import (
	"os"
	"testing"
)

func TestMe(t *testing.T) {
	var token = os.Getenv("MY_FB_TOKEN")

	user, err := me(token)
	if err != nil {
		t.Fatal("Could not make request:", err)
	}

	if user.Email != "nidhikulkarni82@gmail.com" {
		t.Fatal("Wrong email", user.Email)
	}

	if user.FirstName != "Nidhi" {
		t.Fatal("Wrong first name", user.FirstName)
	}

	if user.Gender != "female" {
		t.Fatal("Wrong gender", user.Gender)
	}
}
