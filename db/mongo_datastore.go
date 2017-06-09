package db

import (
	"errors"
	"fmt"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var ERR_MISSING_ID = errors.New("No Id provided for object.")
var ERR_OBJECT_EXISTS = errors.New("Cannot insert existing object.")
var ERR_LIMIT_EXCEEDED = errors.New("Provided limit is to high to be used with FindAll")

var DEFAULT_QUERY_LIMIT = 1000

type MongoDataStore struct {
	Session *mgo.Session
	builder DataStoreQueryBuilder
}

func (m *MongoDataStore) Close() {
	fmt.Println("Closing session for request.")
	m.Session.Close()
}

func (m *MongoDataStore) SetQueryBuilder(qb DataStoreQueryBuilder) {
	m.builder = qb
}

// DataStore Interface

func (m *MongoDataStore) Count(collectionName string, query map[string]interface{}) (int, error) {

	q := m.builder.MakeCountQuery(collectionName, query)

	db := m.Session.DB(Mongo.Database)
	n, err := db.C(collectionName).Find(q).Count()

	if err != nil {
		return -1, err
	}

	return n, nil
}

func (m *MongoDataStore) Fetch(result Model) error {
	id := result.ObjectId()
	collectionName := result.Collection()

	if len(id) == 0 {
		return ERR_MISSING_ID
	}

	q, qerr := m.builder.MakeFindByIdQuery(result)

	if qerr != nil {
		return qerr
	}

	if err := m.FindObject(collectionName, q, result); err != nil {
		return err
	}

	result.CustomUnmarshall()
	return nil
}

func (m *MongoDataStore) FindObject(collectionName string, query map[string]interface{}, result Model) error {
	q := m.builder.MakeFindQuery(collectionName, query)

	// Creating this value is a very lightweight operation, and involves no network communication.
	db := m.Session.DB(Mongo.Database)
	if err := db.C(collectionName).Find(q).One(result); err != nil {
		return err
	}

	result.CustomUnmarshall()
	return nil
}

func (m *MongoDataStore) InsertAll(collectionName string, models []Model) error {
	db := m.Session.DB(Mongo.Database)

	var docs []interface{}
	for _, model := range models {
		if len(model.ObjectId()) > 0 {
			return ERR_OBJECT_EXISTS
		}

		d, qerr := m.builder.MakeInsertDocument(model, time.Now(), bson.NewObjectId().Hex())
		if qerr != nil {
			return qerr
		}

		docs = append(docs, d)
	}

	if err := db.C(collectionName).Insert(docs...); err != nil {
		return err
	}

	for _, model := range models {
		model.SetIsNew(true)
		model.CustomUnmarshall()
	}

	return nil

}

func (m *MongoDataStore) InsertObject(model Model) error {

	if len(model.ObjectId()) > 0 {
		return ERR_OBJECT_EXISTS
	}

	db := m.Session.DB(Mongo.Database)

	doc, qerr := m.builder.MakeInsertDocument(model, time.Now(), bson.NewObjectId().Hex())
	if qerr != nil {
		return qerr
	}

	if err := db.C(model.Collection()).Insert(doc); err != nil {
		return err
	}

	doc.SetIsNew(true)
	doc.CustomUnmarshall()
	return nil
}

func (m *MongoDataStore) UpsertObject(model Model, query map[string]interface{}) error {
	db := m.Session.DB(Mongo.Database)
	collectionName := model.Collection()

	q, change, qerr := m.builder.MakeUpsertDocument(model, query, time.Now(), bson.NewObjectId().Hex())

	if qerr != nil {
		return qerr
	}

	if info, err := db.C(collectionName).Find(q).Apply(change, model); err != nil {
		return err
	} else {
		if info.UpsertedId != nil {
			model.SetIsNew(true)
		}
	}

	model.CustomUnmarshall()
	return nil

}

func (m *MongoDataStore) UpdateObject(model Model) error {

	if len(model.ObjectId()) == 0 {
		return ERR_MISSING_ID
	}

	db := m.Session.DB(Mongo.Database)
	collectionName := model.Collection()

	q, change, qerr := m.builder.MakeChangeDocument(model, time.Now())

	if qerr != nil {
		return qerr
	}

	if _, err := db.C(collectionName).Find(q).Apply(change, model); err != nil {
		return err
	}

	model.CustomUnmarshall()
	return nil
}

func (m *MongoDataStore) RemoveObject(model Model) error {
	collectionName := model.Collection()
	id := model.ObjectId()
	q, qerr := m.builder.MakeRemoveQuery(model)

	if qerr != nil {
		return qerr
	}

	db := m.Session.DB(Mongo.Database)

	if len(id) == 0 {
		return ERR_MISSING_ID
	}

	err := db.C(collectionName).Remove(q)
	return err
}

func (m *MongoDataStore) RemoveAll(collectionName string, query map[string]interface{}) error {
	q, err := m.builder.MakeRemoveAllQuery(collectionName, query)
	if err != nil {
		return err
	}

	db := m.Session.DB(Mongo.Database)
	_, err = db.C(collectionName).RemoveAll(q)
	return err
}

func (m *MongoDataStore) FindEach(collectionName string, query map[string]interface{}, f func(Model), result Model, sortFields ...string) error {
	db := m.Session.DB(Mongo.Database)
	q := m.builder.MakeFindQuery(collectionName, query)
	iter := db.C(collectionName).Find(q).Sort(sortFields...).Iter()

	for iter.Next(result) {
		result.CustomUnmarshall()
		f(result)
	}

	if err := iter.Err(); err != nil {
		fmt.Printf("Error %s, collection name: %s", err.Error(), collectionName)
		return err // error on iteration
	}

	if err := iter.Close(); err != nil {
		fmt.Printf("Error %s, collection name: %s", err.Error(), collectionName)
		return err // error on close
	}

	return nil
}

func (m *MongoDataStore) join(joinCollection string, query map[string]interface{}, f func(Join), j Join) error {
	db := m.Session.DB(Mongo.Database)
	iter := db.C(joinCollection).Find(query).Iter()

	for iter.Next(j) {
		f(j)
	}

	if err := iter.Err(); err != nil {
		fmt.Printf("Error %s", err.Error())
		return err // error on iteration
	}

	if err := iter.Close(); err != nil {
		fmt.Printf("Error %s", err.Error())
		return err // error on close
	}

	return nil
}

func (m *MongoDataStore) FindRelatedObjects(relation Relation, f func(Model), result Model, sortFields ...string) error {
	q := m.builder.QueryByOwningModels(relation.JoinCollection(), []Model{relation.Owner()})

	var join JoinEntry
	var ids []string
	err := m.join(relation.JoinCollection(), q, func(j Join) {
		ids = append(ids, join.RelatedId())
	},
		&join)

	if err != nil {
		fmt.Printf("Error getting joins from %s Error: %s", relation.JoinCollection(), err.Error())
		return err
	}
	fmt.Printf("Related ids in %s %s", relation.JoinCollection(), ids)
	qRelated := m.builder.QueryByIds(result.Collection(), ids)
	return m.FindEach(result.Collection(), qRelated, f, result, sortFields...)

}

func (m *MongoDataStore) FindOwningObjects(joinCollection string, relatedModel Model, f func(Model), result Model) error {
	q := m.builder.QueryByRelatedModels(joinCollection, []Model{relatedModel})

	var join JoinEntry
	var ids []string
	err := m.join(joinCollection, q, func(j Join) {
		ids = append(ids, join.OwningId())
	},
		&join)

	if err != nil {
		return err
	}

	qOwning := m.builder.QueryByIds(result.Collection(), ids)
	return m.FindEach(result.Collection(), qOwning, f, result)
}

func (m *MongoDataStore) SaveRelatedObjects(relation Relation) error {

	toAdd, toDelete, err := m.builder.MakeRelationUpdateDocuments(relation)
	if err != nil {
		fmt.Printf("Error while preparing bulk write for %s Error: %s\n", relation.JoinCollection(), err.Error())
		return err
	}

	fmt.Printf("To add to %s: %s\n", relation.JoinCollection(), toAdd)
	fmt.Printf("To delete from %s: %s\n", relation.JoinCollection(), toDelete)

	db := m.Session.DB(Mongo.Database)
	c := db.C(relation.JoinCollection())

	// Add Related
	if toAdd != nil {
		if err = c.Insert(toAdd...); err != nil {
			fmt.Printf("Error on insert for %s Error: %s\n", relation.JoinCollection(), err.Error())
			return err
		}
	}

	// Delete Related
	if toDelete != nil {
		if _, err = c.RemoveAll(toDelete); err != nil {
			fmt.Printf("Error on delete  for %s Error: %s\n", relation.JoinCollection(), err.Error())
			return err
		}
	}

	return nil

}
