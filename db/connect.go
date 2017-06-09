package db

import (
	"fmt"

	"gopkg.in/mgo.v2"
)

var (
	// Session stores mongo session
	MasterSession *mgo.Session

	// Mongo stores the mongodb connection string information
	Mongo *mgo.DialInfo
)

func Connect(uri string) {
	if len(uri) == 0 {
		panic("No Database connection URL defined.")
		return
	}

	mongo, err := mgo.ParseURL(uri)
	session, err := mgo.Dial(uri)

	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		panic(err.Error())
	}

	fmt.Println("Connected.")

	session.SetSafe(&mgo.Safe{})
	MasterSession = session
	Mongo = mongo

}

// This is an exported var so we can replace this implementation during testing

var GetDataStore = func(qb DataStoreQueryBuilder) DataStore {
	fmt.Println("Creating a new session copy.")
	ds := &MongoDataStore{Session: MasterSession.Copy()}
	ds.SetQueryBuilder(qb)
	return ds
}
