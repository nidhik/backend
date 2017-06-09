package query

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/nidhik/backend/db"

	"gopkg.in/mgo.v2/bson"
)

const (
	QUERY_BY_IDS              = iota
	QUERY_BY_RELATED          = iota
	QUERY_BY_OWNER            = iota
	RELATION_UPDATE_DOCUMENTS = iota
)

type InsertTest struct {
	qb         db.DataStoreQueryBuilder
	model      *TestModel
	modifyTime time.Time
	insertId   string
	err        error
}

var insertId = "lasdfbur80w8jfr"

var insertTests = []InsertTest{
	{unrestrictedQB, EmptyTestModel(), modTime, insertId, nil},
	{restrictedQB, EmptyTestModel(), modTime, insertId, nil},
}

type RelationalQueryTest struct {
	qb          db.DataStoreQueryBuilder
	queryType   int
	collection  string
	owner       *TestModel
	related     *TestModel
	ids         []string
	outputQuery bson.M
}

var relationalTests = []RelationalQueryTest{
	{unrestrictedQB, QUERY_BY_RELATED, join_related_TestModel, nil, NewTestModel("0987654321"), nil, bson.M{"relatedId": bson.M{"$in": []string{"0987654321"}}}},
	{unrestrictedQB, QUERY_BY_OWNER, join_related_TestModel, NewTestModel("1234567890"), nil, nil, bson.M{"owningId": bson.M{"$in": []string{"1234567890"}}}},
	{unrestrictedQB, QUERY_BY_IDS, TestCollection, nil, nil, []string{"foo", "bar", "baz"}, bson.M{"_id": bson.M{"$in": []string{"foo", "bar", "baz"}}}},

	{restrictedQB, QUERY_BY_RELATED, join_related_TestModel, nil, NewTestModel("0987654321"), nil, bson.M{"relatedId": bson.M{"$in": []string{"0987654321"}}}},
	{restrictedQB, QUERY_BY_OWNER, join_related_TestModel, NewTestModel("1234567890"), nil, nil, bson.M{"owningId": bson.M{"$in": []string{"1234567890"}}}},
	{restrictedQB, QUERY_BY_IDS, TestCollection, nil, nil, []string{"foo", "bar", "baz"}, bson.M{"_id": bson.M{"$in": []string{"foo", "bar", "baz"}}, "$or": []bson.M{
		bson.M{"_rperm": bson.M{"$exists": false}},
		bson.M{"_rperm": bson.M{"$in": []interface{}{user.ObjectId(), "*", "role:activeProUser", "role:admin"}}}}}},
}

type RelationUpdateTest struct {
	qb           db.DataStoreQueryBuilder
	owner        *TestModel
	toAdd        []*TestModel
	toRemove     []*TestModel
	outputAdd    []interface{}
	outputDelete bson.M
	err          error
}

var relationUpdateTests = []RelationUpdateTest{
	{unrestrictedQB,
		NewTestModel("Jane Austen"),
		[]*TestModel{NewTestModel("Pride And Prejudice"), NewTestModel("Sense And Sensibility")},
		nil,
		[]interface{}{
			&db.JoinEntry{Related: "Pride And Prejudice", Owning: "Jane Austen"},
			&db.JoinEntry{Related: "Sense And Sensibility", Owning: "Jane Austen"},
		},
		nil,
		nil},

	{unrestrictedQB,
		NewTestModel("Jane Austen"),
		[]*TestModel{NewTestModel("Pride And Prejudice"), NewTestModel("Sense And Sensibility")},
		[]*TestModel{NewTestModel("Of Mice and Men")},
		[]interface{}{
			&db.JoinEntry{Related: "Pride And Prejudice", Owning: "Jane Austen"},
			&db.JoinEntry{Related: "Sense And Sensibility", Owning: "Jane Austen"},
		},
		bson.M{"$or": []bson.M{bson.M{"relatedId": "Of Mice and Men", "owningId": "Jane Austen"}}},
		nil},

	{unrestrictedQB,
		NewTestModel("Jane Austen"),
		nil,
		[]*TestModel{NewTestModel("Of Mice and Men"), NewTestModel("Dune")},
		nil,
		bson.M{"$or": []bson.M{
			bson.M{"relatedId": "Of Mice and Men", "owningId": "Jane Austen"},
			bson.M{"relatedId": "Dune", "owningId": "Jane Austen"},
		}},
		nil},

	{unrestrictedQB,
		NewTestModel("Jane Austen"),
		[]*TestModel{EmptyTestModel(), NewTestModel("Sense And Sensibility")},
		nil,
		nil,
		nil,
		ERR_INSERT_UNSAVED_INTO_RELATION},

	{unrestrictedQB,
		NewTestModel("Jane Austen"),
		nil,
		nil,
		nil,
		nil,
		nil},

	{restrictedQB,
		NewTestModelWithWrite("Jane Austen", user.ObjectId()),
		[]*TestModel{NewTestModel("Pride And Prejudice"), NewTestModel("Sense And Sensibility")},
		nil,
		[]interface{}{
			&db.JoinEntry{Related: "Pride And Prejudice", Owning: "Jane Austen"},
			&db.JoinEntry{Related: "Sense And Sensibility", Owning: "Jane Austen"},
		},
		nil,
		nil},

	{restrictedQB,
		NewTestModel("Jane Austen"),
		[]*TestModel{NewTestModel("Pride And Prejudice"), NewTestModel("Sense And Sensibility")},
		nil,
		nil,
		nil,
		ERR_ACCESS_DENIED},

	{restrictedQB,
		NewTestModelWithWrite("Jane Austen", user.ObjectId()),
		[]*TestModel{NewTestModel("Pride And Prejudice"), NewTestModel("Sense And Sensibility")},
		[]*TestModel{NewTestModel("Of Mice and Men")},
		[]interface{}{
			&db.JoinEntry{Related: "Pride And Prejudice", Owning: "Jane Austen"},
			&db.JoinEntry{Related: "Sense And Sensibility", Owning: "Jane Austen"},
		},
		bson.M{"$or": []bson.M{
			bson.M{"relatedId": "Of Mice and Men", "owningId": "Jane Austen"},
		}},
		nil},

	{restrictedQB,
		NewTestModelWithWrite("Jane Austen", user.ObjectId()),
		nil,
		[]*TestModel{NewTestModel("Of Mice and Men"), NewTestModel("Dune")},
		nil,
		bson.M{"$or": []bson.M{
			bson.M{"relatedId": "Of Mice and Men", "owningId": "Jane Austen"},
			bson.M{"relatedId": "Dune", "owningId": "Jane Austen"},
		}},
		nil},

	{restrictedQB,
		NewTestModelWithWrite("Jane Austen", user.ObjectId()),
		[]*TestModel{EmptyTestModel(), NewTestModel("Sense And Sensibility")},
		nil,
		nil,
		nil,
		ERR_INSERT_UNSAVED_INTO_RELATION},

	{restrictedQB,
		NewTestModelWithWrite("Jane Austen", user.ObjectId()),
		nil,
		nil,
		nil,
		nil,
		nil},

	{restrictedQB,
		NewTestModel("Jane Austen"),
		nil,
		nil,
		nil,
		nil,
		ERR_ACCESS_DENIED},

	{restrictedQB,
		NewTestModel(""),
		nil,
		nil,
		nil,
		nil,
		ERR_UNSAVED_OWNER},
}

func TestQueryBuilder(t *testing.T) {

	for _, test := range insertTests {
		testInsert(t, test)
	}

	for _, test := range relationalTests {
		testRelationalQuery(t, test)
	}

	for _, test := range relationUpdateTests {
		testRelationUpdate(t, test)
	}
}

func testRelationalQuery(t *testing.T, test RelationalQueryTest) {

	var q bson.M
	switch test.queryType {
	case QUERY_BY_IDS:
		q = test.qb.QueryByIds(test.collection, test.ids)
		break
	case QUERY_BY_OWNER:
		q = test.qb.QueryByOwningModels(test.collection, []db.Model{test.owner})
		break
	case QUERY_BY_RELATED:
		q = test.qb.QueryByRelatedModels(test.collection, []db.Model{test.related})
		break
	}

	fmt.Printf("Output Rel. Query: %s \n", q)

	if !reflect.DeepEqual(q, test.outputQuery) {
		t.Fatal("Expected:", test.outputQuery, "Actual:", q)
	}
}

func testInsert(t *testing.T, test InsertTest) {
	model, err := test.qb.MakeInsertDocument(test.model, test.modifyTime, test.insertId)

	fmt.Printf("Insert: %s \n", model)

	if model.ObjectId() != test.insertId {
		t.Fatal("Expected object id to be:", test.insertId, "Actual:", model.ObjectId())
	}

	if model.CreatedDate() != test.modifyTime {
		t.Fatal("Expected created date to be:", test.modifyTime, "Actual:", model.CreatedDate())
	}

	if model.UpdatedDate() != test.modifyTime {
		t.Fatal("Expected updated date to be:", test.modifyTime, "Actual:", model.UpdatedDate())
	}

	if err != test.err {
		t.Fatal("Expected error to be:", test.err, "Actual:", err)
	}
}

func testRelationUpdate(t *testing.T, test RelationUpdateTest) {

	// Prepare the relation
	relation := test.owner.Related
	for _, model := range test.toAdd {
		relation.Add(model)
	}

	for _, model := range test.toRemove {
		relation.Remove(model)
	}

	toAdd, toDelete, err := test.qb.MakeRelationUpdateDocuments(relation)

	fmt.Printf("Add all in: %s\n", toAdd)
	fmt.Printf("Remove all in: %s\n", toDelete)

	if err != test.err {
		t.Fatal("Expected error to be:", test.err, "Actual:", err)
	}

	if len(toAdd) != len(test.outputAdd) {
		t.Fatal("Expected number of insert documents to be:", len(test.outputAdd), "Actual:", len(toAdd))
	}

	for _, testOutput := range test.outputAdd {
		assertResultsContainInsertDocument(t, testOutput.(*db.JoinEntry), toAdd)
	}

	if !reflect.DeepEqual(test.outputDelete, toDelete) {
		t.Fatal("Expected delete query to be: ", test.outputDelete, "Actual:", toDelete)
	}

}

func assertResultsContainInsertDocument(t *testing.T, join *db.JoinEntry, results []interface{}) {

	for _, res := range results {

		if j, ok := res.(*db.JoinEntry); ok {
			if j.Related == join.Related && j.Owning == join.Owning {
				return
			}
		} else {
			t.Fatal("Unexpected insert document type found in results %s \n", res)
		}
	}

	t.Fatal("Could not find join in results %s \n", join)
}
