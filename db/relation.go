package db

import (
	"gopkg.in/mgo.v2/bson"
)

// Joins

type Join interface {
	ObjectId() string
	RelatedId() string
	OwningId() string
}

type JoinEntry struct {
	Id      bson.ObjectId `json:"id" bson:"_id"`
	Related string        `json:"relatedId" bson:"relatedId"`
	Owning  string        `json:"owningId" bson:"owningId"`
}

func (j *JoinEntry) ObjectId() string {
	return j.Id.String()
}
func (j *JoinEntry) RelatedId() string {
	return j.Related
}
func (j *JoinEntry) OwningId() string {
	return j.Owning
}

//Relation

type BaseRelation struct {
	JoinTableName    string
	RelatedTableName string
	OwningModel      Model
	operations       map[Model]bool
}

func NewBaseRelation(owner Model, joinTableName string, relatedTableName string) BaseRelation {
	return BaseRelation{
		JoinTableName:    joinTableName,
		RelatedTableName: relatedTableName,
		OwningModel:      owner,
		operations:       make(map[Model]bool)}
}

func (r *BaseRelation) Add(model Model) {
	r.operations[model] = true
}

func (r *BaseRelation) Remove(model Model) {
	r.operations[model] = false
}

func (r *BaseRelation) JoinCollection() string {
	return r.JoinTableName
}

func (r *BaseRelation) RelatedCollection() string {
	return r.RelatedTableName
}

func (r *BaseRelation) Owner() Model {
	return r.OwningModel
}

func (r *BaseRelation) Inserting() []Model {

	var keys []Model
	for k := range r.operations {
		if r.operations[k] {
			keys = append(keys, k)
		}

	}

	return keys
}

func (r *BaseRelation) Removing() []Model {
	var keys []Model
	for k := range r.operations {
		if !r.operations[k] {
			keys = append(keys, k)
		}

	}

	return keys
}
