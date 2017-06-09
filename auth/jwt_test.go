package auth

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/nidhik/poker-client/models"
)

var tests = []struct {
	expiry    time.Time
	user      *models.User
	errCreate error
	errVerify error
}{
	{time.Now().AddDate(1, 0, 0), models.NewUser("foobar"), nil, nil},
	{time.Now().Add(time.Hour * -1), models.NewUser("foobar"), nil, ERR_INVALID_TOKEN},
	{time.Now().AddDate(1, 0, 0), models.NewEmptyUser(), ERR_MISSING_ID, ERR_INVALID_TOKEN},
}

func TestJWT(t *testing.T) {

	for _, test := range tests {

		tok, err := CreateToken(test.user, test.expiry)

		if err != test.errCreate {
			t.Fatal("Expected create error to be", test.errCreate, "Actual: ", err)
		}

		u, err := VerifyToken(tok)

		if err != test.errVerify {
			t.Fatal("Expected verify error to be", test.errVerify, "Actual: ", err)
		}

		if u == nil && test.errVerify == nil {
			t.Fatal("Expected user to be returned by verify.")
		}

		if u != nil && (u.ObjectId() != test.user.ObjectId()) {
			t.Fatal("Expected user id to be", test.user.ObjectId(), "Actual:", u.ObjectId())
		}

	}

}

var validTokenHMAC256 = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	"userId": "erin",
	"exp":    time.Now().AddDate(1, 0, 0).Unix(),
})

var tokenHMAC384 = jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.MapClaims{
	"userId": "erin",
	"exp":    time.Now().AddDate(1, 0, 0).Unix(),
})

var expiredToken = jwt.NewWithClaims(jwt.SigningMethodHS384, jwt.MapClaims{
	"userId": "erin",
	"exp":    time.Now().Add(time.Minute * -1).Unix(),
})

var tokenNoUser = jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
	"exp": time.Now().AddDate(1, 0, 0).Unix(),
})

var tokenNoClaims = jwt.New(jwt.SigningMethodHS256)

var realKey = []byte(os.Getenv("JWT_SECRET"))
var fakeKey = []byte("nope")

var verifyTests = []struct {
	desc       string
	token      *jwt.Token
	signingKey []byte
	user       *models.User
	err        error
}{
	{"well formed token signed with real key", validTokenHMAC256, realKey, models.NewUser("erin"), nil},
	{"well formed token signed with fake key", validTokenHMAC256, fakeKey, nil, ERR_INVALID_TOKEN},

	{"HMAC384 token signed with real key", tokenHMAC384, realKey, nil, ERR_INVALID_TOKEN},
	{"Missing user token signed with real key", tokenNoUser, realKey, nil, ERR_INVALID_TOKEN},
	{"No claims token signed with real key", tokenNoClaims, realKey, nil, ERR_INVALID_TOKEN},
	{"Expired token signed with real key", expiredToken, realKey, nil, ERR_INVALID_TOKEN},

	{"HMAC384 token signed with fake key", tokenHMAC384, fakeKey, nil, ERR_INVALID_TOKEN},
	{"Missing user token signed with fake key", tokenNoUser, fakeKey, nil, ERR_INVALID_TOKEN},
	{"No claims token signed with fake key", tokenNoClaims, fakeKey, nil, ERR_INVALID_TOKEN},
	{"Expired token signed with fake key", expiredToken, fakeKey, nil, ERR_INVALID_TOKEN},
}

func TestTokenVerification(t *testing.T) {

	for _, test := range verifyTests {
		fmt.Printf("Testing: %s: \n", test.desc)

		tokenString, err := test.token.SignedString(test.signingKey)

		if err != nil {
			t.Fatal("Unexpcted error when signing token:", err)
		}

		u, err := VerifyToken(tokenString)

		if err != test.err {
			t.Fatal("Expected verify error to be", test.err, "Actual: ", err)
		}

		if test.user == nil && u != nil {
			t.Fatal("Expected user to be nil", "Actual:", u.ObjectId())
		}

		if u == nil && test.user != nil {
			t.Fatal("Expected user id to be", test.user.ObjectId(), "Actual is nil")
		}

		if (test.user != nil) && u.ObjectId() != test.user.ObjectId() {
			t.Fatal("Expected user id to be", test.user.ObjectId(), "Actual:", u.ObjectId())
		}
	}

}
