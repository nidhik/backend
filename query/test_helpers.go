package query

import (
	"testing"

	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
	"github.com/nidhik/backend/utils"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type DatastoreTest func(t *testing.T, ds db.DataStore)

const errSetup = "Could not set up mongo database:"

func RunTest(t *testing.T, test DatastoreTest) {
	// SetupMongoContainer may skip or fatal the test if docker isn't found or something goes
	// wrong when setting up the container. Thus, no error is returned
	containerID, _ := utils.SetupMongoContainer(t)
	defer containerID.KillRemove(t)

	db.Connect(utils.MONGO_URI)

	// Create the tables we need
	session := db.MasterSession.Copy()
	database := session.DB(db.Mongo.Database)

	createCollection(t, database, models.CollectionTask)

	createCollection(t, database, models.CollectionEmailRecord)
	createCollection(t, database, models.CollectionEmailMetadata)

	createCollection(t, database, models.CollectionRole)
	createCollection(t, database, models.CollectionUser)

	ds := db.GetDataStore(NewMongoQueryBuilder())

	defer db.MasterSession.Close()

	test(t, ds)

}

func createCollection(t *testing.T, database *mgo.Database, name string) {
	collection := database.C(name)
	err := collection.Create(&mgo.CollectionInfo{})
	AssertNoError(t, errSetup, err)
}

func toMap(m bson.M) map[string]interface{} {
	return m
}

func AssertNoError(t *testing.T, message string, err error) {
	if err != nil {
		t.Fatal(message, err)
	}
}
