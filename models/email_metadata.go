package models

import (
	"github.com/nidhik/backend/db"
	"gopkg.in/mgo.v2/bson"
)

const (
	CollectionEmailMetadata = "EmailMetadata"
)

type EmailMetadataParameter struct {
	AttributeName  string `json:"attributeName" bson:"attributeName"`
	Tag            string `json:"tag" bson:"tag"`
	Transformation string `json:"transformation" bson:"transformation"`
}

func NewTagParameter(attribute string, tag string) *EmailMetadataParameter {
	return &EmailMetadataParameter{AttributeName: attribute, Tag: tag}
}

func NewTransformParameter(attribute string, tag string, transformation string) *EmailMetadataParameter {
	return &EmailMetadataParameter{AttributeName: attribute, Tag: tag, Transformation: transformation}
}

type EmailMetadata struct {
	EmailType    string                    `json:"emailType" bson:"emailType"`
	TemplateId   string                    `json:"templateId" bson:"templateId"`
	Subject      string                    `json:"subject" bson:"subject"`
	Parameters   []*EmailMetadataParameter `json:"parameters" bson:"parameters"`
	IsActive     bool                      `json:"isActive" bson:"isActive"`
	db.BaseModel `bson:",inline"`
}

func NewEmptyEmailMetadata() *EmailMetadata {
	return &EmailMetadata{BaseModel: db.BaseModel{CollectionName: CollectionEmailMetadata}}
}

func NewEmailMetadata(id string) *EmailMetadata {
	return &EmailMetadata{BaseModel: db.BaseModel{
		Id: id, CollectionName: CollectionEmailMetadata},
	}
}

func NewEmailMetadataFromTemplate(template string, emailType string, subject string, params []*EmailMetadataParameter, active bool) *EmailMetadata {
	m := NewEmptyEmailMetadata()

	m.Set("TemplateId", template)
	m.Set("EmailType", emailType)
	m.Set("Subject", subject)
	m.Set("Parameters", params)
	m.Set("IsActive", active)

	return m
}

func (model *EmailMetadata) Fetch(ds db.DataStore) error {
	return model.BaseModel.Fetch(model, ds)
}

func (model *EmailMetadata) Save(ds db.DataStore) error {
	return model.BaseModel.Save(model, ds)
}

func (model *EmailMetadata) Delete(ds db.DataStore) error {
	return model.BaseModel.Delete(model, ds)
}

func (model *EmailMetadata) Set(fieldName string, value interface{}) {
	model.BaseModel.Set(model, fieldName, value)
}

func (model *EmailMetadata) Unset(fieldName string) {
	model.BaseModel.Unset(model, fieldName)
}

func (model *EmailMetadata) Get(fieldName string) interface{} {
	return model.BaseModel.Get(model, fieldName)
}

func (model *EmailMetadata) Increment(fieldName string, amount int) {
	model.BaseModel.Increment(model, fieldName, amount)
}

func (model *EmailMetadata) CustomUnmarshall() {
	model.CollectionName = CollectionEmailMetadata
}

func FindMetadataForEmailType(ds db.DataStore, emailType string) (*EmailMetadata, error) {
	var m EmailMetadata
	err := ds.FindObject(CollectionEmailMetadata, bson.M{"emailType": emailType}, &m)
	return &m, err
}
