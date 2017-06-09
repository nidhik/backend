package query

import (
	"testing"

	"fmt"

	"reflect"

	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"

	"gopkg.in/mgo.v2/bson"
)

const (
	COUNT_QUERY      = iota
	FIND_QUERY       = iota
	FIND_ID_QUERY    = iota
	REMOVE_QUERY     = iota
	REMOVE_ALL_QUERY = iota
)

type QueryTest struct {
	qb          db.DataStoreQueryBuilder
	model       db.Model
	queryType   int
	outputQuery bson.M
	inputQuery  bson.M
	err         error
}

var queryTests = []QueryTest{

	{unrestrictedQB, NewTestModel("1234"), FIND_ID_QUERY, bson.M{"_id": "1234"}, nil, nil},
	{unrestrictedQB, NewTestModel("6789"), FIND_ID_QUERY, bson.M{"_id": "6789"}, nil, nil},
	{unrestrictedQB, nil, FIND_QUERY, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}}, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}}, nil},
	{unrestrictedQB, nil, COUNT_QUERY, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}}, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}}, nil},
	{unrestrictedQB, nil, COUNT_QUERY, bson.M{"$or": []bson.M{bson.M{"username": "nidhi"}, bson.M{"email": "nidhi@foo.com"}}}, bson.M{"$or": []bson.M{bson.M{"username": "nidhi"}, bson.M{"email": "nidhi@foo.com"}}}, nil},
	{unrestrictedQB, NewTestModel("1234"), REMOVE_QUERY, bson.M{"_id": "1234"}, nil, nil},
	{unrestrictedQB, EmptyTestModel(), REMOVE_QUERY, nil, nil, db.ERR_MISSING_ID},
	{unrestrictedQB, nil, REMOVE_ALL_QUERY, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}}, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}}, nil},

	{restrictedQB, NewTestModel("1234"), FIND_ID_QUERY, bson.M{"$or": []bson.M{
		bson.M{"_rperm": bson.M{"$exists": false}},
		bson.M{"_rperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}}}, "_id": "1234"}, nil, nil},

	{restrictedQB, NewTestModel("6789"), FIND_ID_QUERY, bson.M{"$or": []bson.M{
		bson.M{"_rperm": bson.M{"$exists": false}},
		bson.M{"_rperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}}}, "_id": "6789"}, nil, nil},

	{restrictedQB, nil, FIND_QUERY, bson.M{"$or": []bson.M{
		bson.M{"_rperm": bson.M{"$exists": false}},
		bson.M{"_rperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}}}, "_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}}, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}}, nil},

	{restrictedQB, nil, COUNT_QUERY, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}, "$or": []bson.M{
		bson.M{"_rperm": bson.M{"$exists": false}},
		bson.M{"_rperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}}}}, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}}, nil},

	//	// FIXME: this won't work anymore
	//	{restrictedQB, nil, COUNT_QUERY, bson.M{"_rperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}, "$or": []bson.M{bson.M{"username": "nidhi"}, bson.M{"email": "nidhi@foo.com"}}}, bson.M{"$or": []bson.M{bson.M{"username": "nidhi"}, bson.M{"email": "nidhi@foo.com"}}}, nil},

	{restrictedQB, NewTestModel("1234"), REMOVE_QUERY, nil, nil, ERR_ACCESS_DENIED},
	{restrictedQB, EmptyTestModel(), REMOVE_QUERY, nil, nil, db.ERR_MISSING_ID},
	{restrictedQB, publicWriteModel, REMOVE_QUERY, bson.M{"_id": publicWriteModel.ObjectId(), "_wperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}}, nil, nil},
	{restrictedQB, userWriteModel, REMOVE_QUERY, bson.M{"_id": userWriteModel.ObjectId(), "_wperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}}, nil, nil},
	{restrictedQB, mixedPermModel, REMOVE_QUERY, bson.M{"_id": mixedPermModel.ObjectId(), "_wperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}}, nil, nil},

	{restrictedQB, user, FIND_ID_QUERY, bson.M{"_id": user.ObjectId()}, nil, nil},
	{restrictedQB, user, REMOVE_QUERY, bson.M{"_id": user.ObjectId()}, nil, nil},

	{restrictedQB, nil, REMOVE_ALL_QUERY, nil, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}}, ERR_ACCESS_DENIED},
	{restrictedQB, nil, REMOVE_ALL_QUERY, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}, "_p_user": models.PointerString(user), "_wperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}}, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}, "_p_user": models.PointerString(user)}, nil},
	{restrictedQB, nil, REMOVE_ALL_QUERY, nil, bson.M{"_created_at": bson.M{"$gt": "2016-02-08T06:33:04.074Z"}, "_p_user": "_User$foobar"}, ERR_ACCESS_DENIED},
}

func TestQuery(t *testing.T) {
	for _, test := range queryTests {
		testQuery(t, test)
	}
}

func testQuery(t *testing.T, test QueryTest) {

	var q bson.M
	var err error
	switch test.queryType {
	case COUNT_QUERY:
		q = test.qb.MakeFindQuery("TestCollection", test.inputQuery)
		break
	case FIND_QUERY:
		q = test.qb.MakeFindQuery("TestCollection", test.inputQuery)
		break
	case FIND_ID_QUERY:
		q, err = test.qb.MakeFindByIdQuery(test.model)
		break
	case REMOVE_QUERY:
		q, err = test.qb.MakeRemoveQuery(test.model)

	case REMOVE_ALL_QUERY:
		q, err = test.qb.MakeRemoveAllQuery("TestCollection", test.inputQuery)
		break
	}

	fmt.Printf("Output Query: %s \n", q)

	if !reflect.DeepEqual(q, test.outputQuery) {
		t.Fatal("Expected:", test.outputQuery, "Actual:", q)
	}

	if err != test.err {
		t.Fatal("Expected:", test.err, "Actual:", err)
	}
}
