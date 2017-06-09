package query

import (
	"errors"
	"time"

	"github.com/nidhik/backend/db"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var ERR_INSERT_UNSAVED_INTO_RELATION = errors.New("You cannot add an unsaved object to a relation.")
var ERR_UNSAVED_OWNER = errors.New("You cannot add related objects to an unsaved object.")

type MongoQueryBuilder struct {
}

func NewMongoQueryBuilder() *MongoQueryBuilder {
	return &MongoQueryBuilder{}
}

// Query & Update Builders

func (m *MongoQueryBuilder) MakeCountQuery(collectionName string, query map[string]interface{}) bson.M {
	return query
}

func (m *MongoQueryBuilder) MakeFindQuery(collectionName string, query map[string]interface{}) bson.M {
	return query
}

func (m *MongoQueryBuilder) MakeFindByIdQuery(model db.Model) (bson.M, error) {
	if len(model.ObjectId()) == 0 {
		return nil, db.ERR_MISSING_ID
	}

	return bson.M{"_id": model.ObjectId()}, nil
}

func (m *MongoQueryBuilder) MakeRemoveQuery(model db.Model) (bson.M, error) {
	if len(model.ObjectId()) == 0 {
		return nil, db.ERR_MISSING_ID
	}

	return bson.M{"_id": model.ObjectId()}, nil
}

func (m *MongoQueryBuilder) MakeRemoveAllQuery(collectionName string, query map[string]interface{}) (bson.M, error) {
	return query, nil
}

func (m *MongoQueryBuilder) MakeInsertDocument(model db.Model, t time.Time, id string) (db.Model, error) {

	model.SetObjectId(id)
	model.SetCreatedDate(t)
	model.SetUpdatedDate(t)

	return model, nil

}

// findAndModify() https://docs.mongodb.com/manual/reference/method/db.collection.findAndModify/
func (m *MongoQueryBuilder) MakeChangeDocument(model db.Model, t time.Time) (bson.M, mgo.Change, error) {
	if len(model.ObjectId()) == 0 {
		return nil, mgo.Change{}, db.ERR_MISSING_ID
	}

	return bson.M{"_id": model.ObjectId()},
		mgo.Change{
			Update:    model.Update(t),
			ReturnNew: true,
		}, nil
}

func (m *MongoQueryBuilder) MakeUpsertDocument(model db.Model, query map[string]interface{}, t time.Time, id string) (bson.M, mgo.Change, error) {

	return query,
		mgo.Change{
			Update:    model.Upsert(t, id),
			Upsert:    true,
			ReturnNew: true,
		}, nil
}

func (m *MongoQueryBuilder) QueryByRelatedModels(joinCollectionName string, related []db.Model) bson.M {
	var relatedIds []string
	for _, r := range related {
		relatedIds = append(relatedIds, r.ObjectId())
	}

	return inArray("relatedId", relatedIds)
}

func (m *MongoQueryBuilder) QueryByOwningModels(joinCollectionName string, owning []db.Model) bson.M {
	var owningIds []string
	for _, o := range owning {
		owningIds = append(owningIds, o.ObjectId())
	}

	return inArray("owningId", owningIds)
}

func (m *MongoQueryBuilder) QueryByIds(collectionName string, ids []string) bson.M {
	return inArray("_id", ids)
}

func inArray(field string, vals []string) bson.M {
	return bson.M{
		field: bson.M{
			"$in": vals,
		},
	}
}

func (m *MongoQueryBuilder) MakeRelationUpdateDocuments(relation db.Relation) ([]interface{}, bson.M, error) {
	if relation == nil || len(relation.Owner().ObjectId()) == 0 {
		return nil, nil, ERR_UNSAVED_OWNER
	}

	inserting := relation.Inserting()
	removing := relation.Removing()
	owner := relation.Owner()

	var toAdd []interface{}
	var toDelete []bson.M

	for _, model := range inserting {
		if len(model.ObjectId()) == 0 {
			return nil, nil, ERR_INSERT_UNSAVED_INTO_RELATION
		}
		toAdd = append(toAdd, insertRelated(model, owner))
	}

	var deleteQuery bson.M
	if len(removing) > 0 {
		for _, model := range removing {
			toDelete = append(toDelete, deleteRelated(model, owner))
		}
		deleteQuery = bson.M{"$or": toDelete}
	}

	return toAdd, deleteQuery, nil
}

func insertRelated(model db.Model, owner db.Model) *db.JoinEntry {
	join := &db.JoinEntry{Id: bson.NewObjectId(), Related: model.ObjectId(), Owning: owner.ObjectId()}
	return join
}

func deleteRelated(model db.Model, owner db.Model) bson.M {
	deleteQ := bson.M{"relatedId": model.ObjectId(), "owningId": owner.ObjectId()}
	return deleteQ
}
