package models

import (
	"github.com/nidhik/backend/db"
	"gopkg.in/mgo.v2/bson"
)

const (
	CollectionEmailRecord = "EmailRecord"
)

type Substitution struct {
	Tag   string `json:"tag" bson:"tag"`
	Value string `json:"val" bson:"val"`
}

func NewSubstitution(tag string, value string) *Substitution {
	return &Substitution{Tag: tag, Value: value}
}

type EmailRecord struct {
	To            string          `json:"to" bson:"to"`
	From          string          `json:"from" bson:"from"`
	TemplateId    string          `json:"templateId" bson:"templateId"`
	Subject       string          `json:"subject" bson:"subject"`
	Category      []string        `json:"category" bson:"category"`
	Substitutions []*Substitution `json:"substitutions" bson:"substitutions"`
	IsActive      bool            `json:"isActive" bson:"isActive"`
	db.BaseModel  `bson:",inline"`
}

func NewEmptyEmailRecord() *EmailRecord {
	return &EmailRecord{BaseModel: db.BaseModel{CollectionName: CollectionEmailRecord}}
}

func NewEmailRecord(id string) *EmailRecord {
	return &EmailRecord{BaseModel: db.BaseModel{
		Id: id, CollectionName: CollectionEmailRecord},
	}
}

func NewEmailRecordForUser(user *User, emailType string, templateId string, subject string, substitutions []*Substitution) *EmailRecord {

	record := NewEmptyEmailRecord()

	record.Set("To", user.Email)
	record.Set("From", "founders@spitfireathlete.com")
	record.Set("Subject", subject)
	record.Set("Category", []string{"transactional", emailType})
	record.Set("TemplateId", templateId)
	record.Set("Substitutions", substitutions)

	return record
}

func (model *EmailRecord) Fetch(ds db.DataStore) error {
	return model.BaseModel.Fetch(model, ds)
}

func (model *EmailRecord) Save(ds db.DataStore) error {
	return model.BaseModel.Save(model, ds)
}

func (model *EmailRecord) Delete(ds db.DataStore) error {
	return model.BaseModel.Delete(model, ds)
}

func (model *EmailRecord) Set(fieldName string, value interface{}) {
	model.BaseModel.Set(model, fieldName, value)
}

func (model *EmailRecord) Unset(fieldName string) {
	model.BaseModel.Unset(model, fieldName)
}

func (model *EmailRecord) Get(fieldName string) interface{} {
	return model.BaseModel.Get(model, fieldName)
}

func (model *EmailRecord) Increment(fieldName string, amount int) {
	model.BaseModel.Increment(model, fieldName, amount)
}

func (model *EmailRecord) CustomUnmarshall() {
	model.CollectionName = CollectionEmailRecord
}

func FindEmailRecordsForUser(ds db.DataStore, user *User) ([]*EmailRecord, error) {
	var emails []*EmailRecord

	err := ds.FindEach(CollectionEmailRecord, bson.M{"to": user.Email}, func(model db.Model) {

		var e = model.(*EmailRecord)

		ptr := NewEmptyEmailRecord()
		*ptr = *e
		emails = append(emails, ptr)

	}, &EmailRecord{})

	return emails, err
}
