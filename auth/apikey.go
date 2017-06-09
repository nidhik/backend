package auth

import (
	"fmt"
	"os"
)

var clientKey = os.Getenv("CLIENT_KEY")
var ERR_INVALID_CLIENT_KEY = fmt.Errorf("Invalid client key.\n")

func IsApprovedAPIConsumer(key string) error {
	if key == clientKey {
		return nil
	}

	return ERR_INVALID_CLIENT_KEY
}
