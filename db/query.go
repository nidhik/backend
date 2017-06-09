package db

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type RelationalDataStoreQueryBuilder interface {
	QueryByRelatedModels(joinCollectionName string, related []Model) bson.M
	QueryByOwningModels(joinCollectionName string, owning []Model) bson.M
	QueryByIds(collectionName string, ids []string) bson.M

	MakeRelationUpdateDocuments(relation Relation) ([]interface{}, bson.M, error)
}

type DataStoreQueryBuilder interface {
	RelationalDataStoreQueryBuilder

	// Read
	MakeCountQuery(collectionName string, query map[string]interface{}) bson.M
	MakeFindQuery(collectionName string, query map[string]interface{}) bson.M
	MakeFindByIdQuery(model Model) (bson.M, error)

	// Write
	MakeRemoveQuery(model Model) (bson.M, error)
	MakeInsertDocument(model Model, t time.Time, id string) (Model, error)
	MakeRemoveAllQuery(collectionName string, query map[string]interface{}) (bson.M, error)

	// See findAndModify() https://docs.mongodb.com/manual/reference/method/db.collection.findAndModify/
	MakeChangeDocument(model Model, t time.Time) (bson.M, mgo.Change, error)
	MakeUpsertDocument(model Model, query map[string]interface{}, t time.Time, id string) (bson.M, mgo.Change, error)
}
