package models

import (
	"errors"

	"strings"

	"github.com/nidhik/backend/db"
	"gopkg.in/mgo.v2/bson"
)

var ERR_UNKNOWN_POINTER_TYPE = errors.New("Unknown object pointer type")
var ERR_UNKNOWN_COLL_NAME = errors.New("Unknown object className")

type Pointer struct {
	TypeName  string `json:"__type" bson:"__type"`
	ClassName string `json:"className" bson:"className"`
	ObjectId  string `json:"objectId" bson:"objectId"`
}

func NewPointer(className string, objectId string) *Pointer {
	return &Pointer{TypeName: "Pointer", ClassName: className, ObjectId: objectId}
}

func AsPointer(model db.Model) bson.M {

	return bson.M{
		"__type":    "Pointer",
		"className": model.Collection(),
		"objectId":  model.ObjectId(),
	}

}

func PointerString(model db.Model) string {
	return model.Collection() + "$" + model.ObjectId()
}

func (p *Pointer) model() (db.Model, error) {
	if p.TypeName == "Pointer" {
		switch p.ClassName {
		case CollectionTask:
			return NewTask(p.ObjectId), nil
		case CollectionRole:
			return NewRole(p.ObjectId), nil
		case CollectionUser:
			return NewUser(p.ObjectId), nil
		default:
			return nil, ERR_UNKNOWN_COLL_NAME
		}
	}

	return nil, ERR_UNKNOWN_POINTER_TYPE
}

func (p *Pointer) Fetch(ds db.DataStore) (db.Model, error) {
	if m, err := p.model(); err != nil {
		return nil, err
	} else {
		return m, m.Fetch(ds)
	}
}

func UnmarshallPointer(ptr string) *Pointer {
	if len(ptr) == 0 {
		return nil
	}

	slices := strings.SplitN(ptr, "$", 2)
	collection := slices[0]
	id := slices[1]
	res := NewPointer(collection, id)

	return res

}
