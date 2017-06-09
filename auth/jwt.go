package auth

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/nidhik/backend/models"
)

var signingKey = []byte(os.Getenv("JWT_SECRET"))
var ERR_MISSING_ID = errors.New("Missing user id.")
var ERR_INVALID_TOKEN = fmt.Errorf("Invalid token.\n")

var signingMethod = jwt.SigningMethodHS256

func CreateToken(user *models.User, expiry time.Time) (string, error) {
	if user.ObjectId() == "" {
		return "", ERR_MISSING_ID
	}
	token := jwt.NewWithClaims(signingMethod, jwt.MapClaims{
		"userId": user.ObjectId(),
		"exp":    expiry.Unix(),
	})

	tokenString, err := token.SignedString(signingKey)

	return tokenString, err
}

func VerifyToken(tokenString string) (*models.User, error) {

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok || token.Method != signingMethod {
			fmt.Printf("Unexpected signing method: %v \n", token.Header["alg"])
			return nil, ERR_INVALID_TOKEN
		}

		return signingKey, nil
	})

	if err != nil || !token.Valid {
		return nil, ERR_INVALID_TOKEN
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		if userId, ok := claims["userId"].(string); ok {
			return models.NewUser(userId), nil
		} else {
			return nil, ERR_INVALID_TOKEN
		}

	}

	return nil, err

}
