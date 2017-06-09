package query

import (
	"testing"

	"fmt"

	"reflect"
	"time"

	"github.com/nidhik/backend/db"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	UPDATE_DOCUMENT = iota
	UPSERT_DOCUMENT = iota
)

type ModifyTest struct {
	qb             db.DataStoreQueryBuilder
	model          db.Model
	modifyType     int
	modifyTime     time.Time
	upsertId       string
	upsertQuery    bson.M
	outputDocument mgo.Change
	outputQuery    bson.M
	err            error
}

var modifyTests = []ModifyTest{
	{unrestrictedQB, NewTestModel("qwerty"), UPDATE_DOCUMENT, modTime, "", nil,
		mgo.Change{
			Update: toMap(bson.M{
				"$set": bson.M{"_updated_at": utcModTime}}),
			ReturnNew: true},
		bson.M{"_id": "qwerty"},
		nil},

	{unrestrictedQB, EmptyTestModel(), UPDATE_DOCUMENT, modTime, "", nil,
		mgo.Change{},
		nil,
		db.ERR_MISSING_ID},

	{unrestrictedQB, EmptyTestModel(), UPSERT_DOCUMENT, modTime, upsertId, bson.M{"_auth_data_facebook.id": "1244567"},
		mgo.Change{
			Update: toMap(bson.M{
				"$setOnInsert": bson.M{"_updated_at": utcModTime, "_created_at": utcModTime, "_id": upsertId}}),
			Upsert:    true,
			ReturnNew: true},
		bson.M{"_auth_data_facebook.id": "1244567"},
		nil},

	{restrictedQB, NewTestModel("qwerty"), UPDATE_DOCUMENT, modTime, "", nil,
		mgo.Change{},
		nil,
		ERR_ACCESS_DENIED},

	{restrictedQB, publicWriteModel, UPDATE_DOCUMENT, modTime, "", nil,
		mgo.Change{
			Update: toMap(bson.M{
				"$set": bson.M{
					"_updated_at": utcModTime,
					"_acl":        map[string]db.Permission{"*": db.Permission{Read: false, Write: true}},
					"_rperm":      []string{},
					"_wperm":      []string{"*"}}}),
			ReturnNew: true},
		bson.M{"_id": publicWriteModel.ObjectId(), "_wperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}},
		nil},

	{restrictedQB, userWriteModel, UPDATE_DOCUMENT, modTime, "", nil,
		mgo.Change{
			Update: toMap(bson.M{
				"$set": bson.M{
					"_updated_at": utcModTime,
					"_acl":        map[string]db.Permission{user.ObjectId(): db.Permission{Read: false, Write: true}},
					"_rperm":      []string{},
					"_wperm":      []string{user.ObjectId()}}}),
			ReturnNew: true},
		bson.M{"_id": userWriteModel.ObjectId(), "_wperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}},
		nil},

	{restrictedQB, mixedPermModel, UPDATE_DOCUMENT, modTime, "", nil,
		mgo.Change{
			Update: toMap(bson.M{
				"$set": bson.M{
					"_updated_at": utcModTime,
					"_acl":        map[string]db.Permission{user.ObjectId(): db.Permission{Read: false, Write: true}, "*": db.Permission{Read: false, Write: true}},
					"_rperm":      []string{},
					"_wperm":      []string{user.ObjectId(), "*"}}}),
			ReturnNew: true},
		bson.M{"_id": mixedPermModel.ObjectId(), "_wperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}},
		nil},

	{restrictedQB, EmptyTestModel(), UPSERT_DOCUMENT, modTime, upsertId, bson.M{"_auth_data_facebook.id": "1244567"},
		mgo.Change{
			Update: toMap(bson.M{
				"$setOnInsert": bson.M{"_updated_at": utcModTime, "_created_at": utcModTime, "_id": upsertId}}),
			Upsert:    true,
			ReturnNew: true},
		bson.M{"_auth_data_facebook.id": "1244567", "_wperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}},
		nil},

	{restrictedQB, NewTestModel("qwerty"), UPSERT_DOCUMENT, modTime, upsertId, bson.M{"_auth_data_facebook.id": "1244567"},
		mgo.Change{},
		nil,
		ERR_ACCESS_DENIED},

	{restrictedQB, user, UPDATE_DOCUMENT, modTime, "", nil,
		mgo.Change{
			Update: toMap(bson.M{
				"$set": bson.M{"_updated_at": utcModTime}}),
			ReturnNew: true},
		bson.M{"_id": user.ObjectId()},
		nil},
}

func TestModify(t *testing.T) {
	for _, test := range modifyTests {
		testModify(t, test)
	}
}

func testModify(t *testing.T, test ModifyTest) {
	var doc mgo.Change
	var q bson.M
	var err error

	switch test.modifyType {
	case UPDATE_DOCUMENT:
		q, doc, err = test.qb.MakeChangeDocument(test.model, test.modifyTime)
		break
	case UPSERT_DOCUMENT:
		q, doc, err = test.qb.MakeUpsertDocument(test.model, test.upsertQuery, test.modifyTime, upsertId)
		break
	}

	fmt.Printf("Apply changes: %s \n", doc)

	if !assertUpdateEqual(t, test.outputDocument.Update, doc.Update) {
		t.Fatal("Expected Update to be:", test.outputDocument.Update, "Actual:", doc.Update)
	}

	if doc.ReturnNew != test.outputDocument.ReturnNew {
		t.Fatal("Expected ReturnNew to be:", test.outputDocument.ReturnNew, "Actual:", doc.ReturnNew)
	}

	if doc.Upsert != test.outputDocument.Upsert {
		t.Fatal("Expected Upsert to be:", test.outputDocument.Upsert, "Actual:", doc.Upsert)
	}

	if !reflect.DeepEqual(test.outputQuery, q) {
		t.Fatal("Expected output query to be:", test.outputQuery, "Actual:", q)
	}

	if err != test.err {
		t.Fatal("Expected error to be:", test.err, "Actual:", err)
	}
}

func getModifyValues(doc map[string]interface{}) (interface{}, interface{}) {
	return doc["$set"], doc["$setOnInsert"]
}

func assertEqualSlice(perm []string, perm2 []string) bool {

	if len(perm) != len(perm2) {
		return false
	}

	for i, e := range perm {
		if perm2[i] != e {
			return false
		}
	}

	return true

}
func assertACLFieldsEqual(exp bson.M, act bson.M) bool {
	er := exp["_rperm"]
	ew := exp["_wperm"]
	eacl := exp["_acl"]

	ar := act["_rperm"]
	aw := act["_wperm"]
	aacl := act["_acl"]

	if er == nil {
		er = []string{}
	}

	if ar == nil {
		ar = []string{}
	}

	if ew == nil {
		ew = []string{}
	}

	if aw == nil {
		aw = []string{}
	}

	if len(er.([]string)) != len(ar.([]string)) {
		fmt.Println("Read perms do not match nil case.")
		fmt.Println("Expected:")
		fmt.Println(er)
		fmt.Println("Actual:")
		fmt.Println(ar)
		return false
	}

	if len(ew.([]string)) != len(aw.([]string)) {
		fmt.Println("Write perms do not match nil case.")
		fmt.Println("Expected:")
		fmt.Println(ew)
		fmt.Println("Actual:")
		fmt.Println(aw)
		return false
	}

	if er != nil && ar != nil && aw != nil && ew != nil {
		if !(assertEqualSlice(er.([]string), ar.([]string)) && assertEqualSlice(ew.([]string), aw.([]string)) && reflect.DeepEqual(eacl, aacl)) {
			fmt.Println("ACLs do not match.")
			return false
		}
	}

	return true

}

func assertSetFieldsEqual(exp bson.M, act bson.M) bool {
	return exp["_updated_at"] == act["_updated_at"] && assertACLFieldsEqual(exp, act)
}

func assertSetOnInsertFieldsEqual(exp bson.M, act bson.M) bool {
	return exp["_updated_at"] == act["_updated_at"] &&
		exp["_created_at"] == act["_created_at"] &&
		exp["_id"] == act["_id"] &&
		assertACLFieldsEqual(exp, act)
}
func assertUpdateEqual(t *testing.T, expected interface{}, actual interface{}) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}

	if expected, ok := expected.(map[string]interface{}); ok {

		if actual, ok := actual.(map[string]interface{}); ok {

			eset, esetOnInsert := getModifyValues(expected)
			aset, asetOnInsert := getModifyValues(actual)

			if (eset == nil || aset == nil) && eset != aset {
				fmt.Println("Set fields not equal nil case")
				return false
			}

			if (esetOnInsert == nil || asetOnInsert == nil) && esetOnInsert != asetOnInsert {
				fmt.Println("Set on insert fields not equal nil case")
				return false
			}

			if esetOnInsert != nil && asetOnInsert != nil {
				if !assertSetOnInsertFieldsEqual(esetOnInsert.(bson.M), asetOnInsert.(bson.M)) {
					fmt.Println("Set on insert fields not equal")
					return false
				}
			}
			if eset != nil && aset != nil {
				if !assertSetFieldsEqual(eset.(bson.M), aset.(bson.M)) {
					fmt.Println("Set fields not equal")
					return false
				}
			}

			return true
		}

		t.Fatal("Actual Update doc is not a map.")
		return false
	}

	t.Fatal("Expected Update doc is not a map.", reflect.TypeOf(expected))
	return false
}
