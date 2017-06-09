package db

import (
	"time"
)

type Model interface {
	ObjectId() string
	Collection() string
	CreatedDate() time.Time
	UpdatedDate() time.Time
	AccessControlList() *ACL
	IsNew() bool
	// https://godoc.org/gopkg.in/mgo.v2#Change
	Update(t time.Time) map[string]interface{}
	Upsert(t time.Time, id string) map[string]interface{}

	SetObjectId(id string)
	SetIsNew(isNew bool)
	SetCollection(name string)
	SetCreatedDate(date time.Time)
	SetUpdatedDate(date time.Time)
	SetAccessControlList(acl *ACL)

	Unset(fieldName string)
	Set(fieldName string, value interface{})
	Get(fieldName string) interface{}
	Increment(fieldName string, amount int)

	Fetch(ds DataStore) error
	Save(ds DataStore) error
	Delete(ds DataStore) error
	CustomUnmarshall()
}

type Relation interface {
	Owner() Model
	Add(model Model)
	Remove(model Model)
	JoinCollection() string
	RelatedCollection() string
	Find(rds RelationalDataStore) ([]Model, error)

	Inserting() []Model
	Removing() []Model
}

type RelationalDataStore interface {
	FindRelatedObjects(relation Relation, f func(Model), result Model, sortFields ...string) error
	FindOwningObjects(joinCollection string, relatedModel Model, f func(Model), result Model) error

	SaveRelatedObjects(relation Relation) error
}

type DataStore interface {
	RelationalDataStore

	Close()
	SetQueryBuilder(qb DataStoreQueryBuilder)

	InsertObject(model Model) error
	RemoveObject(model Model) error
	UpdateObject(model Model) error

	InsertAll(collectionName string, models []Model) error
	RemoveAll(collectionName string, query map[string]interface{}) error
	Count(collectionName string, query map[string]interface{}) (int, error)
	Fetch(result Model) error
	FindObject(collectionName string, query map[string]interface{}, result Model) error
	FindEach(collectionName string, query map[string]interface{}, f func(Model), result Model, sortFields ...string) error
	UpsertObject(model Model, query map[string]interface{}) error
}
