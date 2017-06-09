package utils

import (
	"fmt"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gopkg.in/mgo.v2/bson"
)

func TestID(t *testing.T) {

	fmt.Println(bson.NewObjectId().Hex())
	fmt.Println(time.Now().Format(time.RFC3339))

}

func TestHash(t *testing.T) {
	password := "function() { return obj.credits - obj.debits < 0;var date=new Date(); do{curDate = new Date();}while(curDate-date<10000); }"
	pass := []byte(password)
	hashed, err := bcrypt.GenerateFromPassword(pass, bcrypt.MinCost)
	if err == nil {
		fmt.Printf("Hashed: %s \n", hashed)

	} else {
		fmt.Printf("Error: %s\n", err)
	}
}
