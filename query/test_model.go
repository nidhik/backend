package query

import (
	"github.com/nidhik/backend/db"
)

const (
	join_related_TestModel = "_Join:related:_TestModel"
	TestCollection         = "TestCollection"
)

type TestModel struct {
	TestField    string      `json:"testField,omitempty" bson:"testField"`
	Related      db.Relation `json:"-" bson:"-"`
	db.BaseModel `bson:",inline"`
}

func NewTestModelWithWrite(id string, perms ...string) *TestModel {
	m := NewTestModel(id)
	acl := db.NewACL()
	for _, p := range perms {
		acl.AddWrite(p)
	}

	m.SetAccessControlList(acl)
	return m
}

func NewTestModel(id string) *TestModel {
	model := &TestModel{
		BaseModel: db.BaseModel{
			Id:             id,
			CollectionName: TestCollection},
	}
	model.Related = newTestRelation(model)
	return model
}

func EmptyTestModel() *TestModel {
	return &TestModel{BaseModel: db.BaseModel{CollectionName: "TestCollection"}}
}

func (model *TestModel) Fetch(ds db.DataStore) error {
	return model.BaseModel.Fetch(model, ds)
}

func (model *TestModel) Save(ds db.DataStore) error {
	return model.BaseModel.Save(model, ds)
}

func (model *TestModel) Delete(ds db.DataStore) error {
	return model.BaseModel.Delete(model, ds)
}

func (model *TestModel) Set(fieldName string, value interface{}) {
	model.BaseModel.Set(model, fieldName, value)
}

func (model *TestModel) Unset(fieldName string) {
	model.BaseModel.Unset(model, fieldName)
}

func (model *TestModel) Get(fieldName string) interface{} {
	return model.BaseModel.Get(model, fieldName)
}

func (model *TestModel) Increment(fieldName string, amount int) {
	model.BaseModel.Increment(model, fieldName, amount)
}

func (model *TestModel) CustomUnmarshall() {
	model.CollectionName = TestCollection
	model.Related = newTestRelation(model)
}

// Relation

type testRelation struct {
	db.BaseRelation
}

func newTestRelation(owner db.Model) *testRelation {
	return &testRelation{BaseRelation: db.NewBaseRelation(owner, join_related_TestModel, TestCollection)}
}

func (r *testRelation) Find(rds db.RelationalDataStore) ([]db.Model, error) {
	var user = EmptyTestModel()

	var models []db.Model
	err := rds.FindRelatedObjects(r, func(model db.Model) {

		u := model.(*TestModel)
		ptr := EmptyTestModel()
		*ptr = *u

		models = append(models, ptr)

	}, user)

	return models, err
}
