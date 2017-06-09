package db

import (
	"strings"
	"time"

	"github.com/fatih/structs"
	"gopkg.in/mgo.v2/bson"
)

type BaseModel struct {
	Id             string     `json:"id" bson:"_id"`
	CollectionName string     `json:"collectionName" bson:"-"`
	CreatedAt      *time.Time `json:"createdAt,omitempty" bson:"_created_at"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty" bson:"_updated_at"`
	changes        bson.M     `json:"-" bson:"-"`
	New            bool       `json:"isNew" bson:"-"`
	ACL            `bson:",inline"`
}

// Model Interface methods

func (model *BaseModel) ObjectId() string {
	return model.Id
}

func (model *BaseModel) CreatedDate() time.Time {
	return *model.CreatedAt
}

func (model *BaseModel) UpdatedDate() time.Time {
	return *model.UpdatedAt
}

func (model *BaseModel) Collection() string {
	return model.CollectionName
}

func (model *BaseModel) IsNew() bool {
	return model.New
}

func (model *BaseModel) AccessControlList() *ACL {
	return &model.ACL
}

func (model *BaseModel) SetObjectId(id string) {
	model.Id = id
}

func (model *BaseModel) SetCollection(name string) {
	model.CollectionName = name
}

func (model *BaseModel) SetIsNew(isNew bool) {
	model.New = isNew
}

func (model *BaseModel) SetCreatedDate(date time.Time) {
	model.CreatedAt = &date
}

func (model *BaseModel) SetUpdatedDate(date time.Time) {
	model.UpdatedAt = &date
}

func (model *BaseModel) SetAccessControlList(acl *ACL) {
	model.ACL = *acl
	model.setOnUpdate("_acl", acl.ACL)
	model.setOnUpdate("_rperm", acl.ReadAccess)
	model.setOnUpdate("_wperm", acl.WriteAccess)
}

func (model *BaseModel) CustomUnmarshall() {
	// do nothing by default
}

func (model *BaseModel) Update(t time.Time) map[string]interface{} {

	model.setOnUpdate("_updated_at", t.UTC())
	return model.changes
}

func (model *BaseModel) Upsert(t time.Time, id string) map[string]interface{} {
	model.setOnInsert("_updated_at", t.UTC())
	model.setOnInsert("_created_at", t.UTC())
	model.setOnInsert("_id", id)
	return model.changes
}

// Convenience methods for use from specific Models

func (model *BaseModel) Fetch(object Model, ds DataStore) error {
	return ds.Fetch(object)
}

func (model *BaseModel) Save(object Model, ds DataStore) error {
	if len(object.ObjectId()) == 0 {
		return ds.InsertObject(object)
	}
	return ds.UpdateObject(object)
}

func (model *BaseModel) Delete(object Model, ds DataStore) error {
	return ds.RemoveObject(object)
}

func (model *BaseModel) Set(object Model, fieldName string, value interface{}) {
	tag, field := getBSONTagAndField(object, fieldName)
	field.Set(value)
	model.setOnUpdate(tag, value)
}

func (model *BaseModel) Unset(object Model, fieldName string) {
	tag, field := getBSONTagAndField(object, fieldName)
	field.Zero()
	model.unset(tag)
}

func (model *BaseModel) Get(object Model, fieldName string) interface{} {
	_, field := getBSONTagAndField(object, fieldName)
	return field.Value()
}

func (model *BaseModel) Increment(object Model, fieldName string, amount int) {
	tag, _ := getBSONTagAndField(object, fieldName)
	model.incOnUpdate(tag, amount)

}

// Private methods, useful only from within BaseModel

func (model *BaseModel) change(key string, tagValue string, value interface{}) {
	model.initChanges(key)
	model.changes[key].(bson.M)[tagValue] = value
}

func (model *BaseModel) deleteChange(key string, tagValue string) {
	model.initChanges(key)
	updates := model.changes[key].(bson.M)
	delete(updates, tagValue)
	if len(updates) == 0 {
		delete(model.changes, key)
	}
}

func (model *BaseModel) setOnInsert(tagValue string, value interface{}) {
	model.change("$setOnInsert", tagValue, value)
}

func (model *BaseModel) setOnUpdate(tagValue string, value interface{}) {
	model.change("$set", tagValue, value)
	model.deleteChange("$unset", tagValue)
}

func (model *BaseModel) unset(tagValue string) {
	model.change("$unset", tagValue, "")
	model.deleteChange("$set", tagValue)
}

func (model *BaseModel) incOnUpdate(tagValue string, amount int) {
	model.change("$inc", tagValue, amount)
}

func (model *BaseModel) initChanges(changesKey string) {
	if model.changes == nil {
		model.changes = bson.M{}
	}
	if model.changes[changesKey] == nil {
		model.changes[changesKey] = bson.M{}
	}

}

func getBSONTagAndField(object Model, fieldName string) (string, *structs.Field) {
	s := structs.New(object)
	field := s.Field(fieldName)
	tagValue := field.Tag("bson")
	res := strings.TrimSuffix(tagValue, ",omitempty")
	return res, field
}
