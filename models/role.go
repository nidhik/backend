package models

import (
	"fmt"

	"github.com/nidhik/backend/db"
	"gopkg.in/mgo.v2/bson"
)

const (
	CollectionRole = "_Role"
)

type Role struct {
	Name         string      `json:"name" bson:"name"`
	Users        db.Relation `json:"-" bson:"-"`
	db.BaseModel `bson:",inline"`
}

func NewRoleWithName(id string, name string) *Role {
	role := NewRole(id)
	role.Name = name
	return role
}

func NewRole(id string) *Role {
	role := &Role{
		BaseModel: db.BaseModel{
			Id:             id,
			CollectionName: CollectionRole},
	}
	role.Users = newRelationRoleUsers(role)
	return role
}

func NewEmptyRole() *Role {
	return &Role{
		BaseModel: db.BaseModel{
			CollectionName: CollectionRole},
	}
}

func (role *Role) Fetch(ds db.DataStore) error {
	return role.BaseModel.Fetch(role, ds)
}

func (role *Role) Save(ds db.DataStore) error {
	if err := role.BaseModel.Save(role, ds); err != nil {
		return err
	}

	return nil

}

func (role *Role) Delete(ds db.DataStore) error {
	return role.BaseModel.Delete(role, ds)
}

func (role *Role) Set(fieldName string, value interface{}) {
	role.BaseModel.Set(role, fieldName, value)
}

func (role *Role) Unset(fieldName string) {
	role.BaseModel.Unset(role, fieldName)
}

func (role *Role) Get(fieldName string) interface{} {
	return role.BaseModel.Get(role, fieldName)
}

func (role *Role) Increment(fieldName string, amount int) {
	role.BaseModel.Increment(role, fieldName, amount)
}

func (role *Role) CustomUnmarshall() {
	role.CollectionName = CollectionRole
	role.Users = newRelationRoleUsers(role)
}

// Queries
func UpsertRoleByName(ds db.DataStore, name string, acl *db.ACL) (*Role, error) {
	role := NewEmptyRole()
	role.Set("Name", name)
	role.SetAccessControlList(acl)

	if err := ds.UpsertObject(role, bson.M{"name": name}); err != nil {
		return nil, err
	}
	return role, nil
}

func FindRoleByName(ds db.DataStore, name string) (*Role, error) {

	var m Role
	if err := ds.FindObject(CollectionRole, bson.M{"name": name}, &m); err != nil {
		return nil, err
	}
	return &m, nil
}

// Users Relation

const (
	join_users_Role = "_Join:users:_Role"
)

type relationRoleUsers struct {
	db.BaseRelation
}

func newRelationRoleUsers(owner db.Model) *relationRoleUsers {
	return &relationRoleUsers{BaseRelation: db.NewBaseRelation(owner, join_users_Role, CollectionUser)}
}

func (r *relationRoleUsers) Find(rds db.RelationalDataStore) ([]db.Model, error) {
	var user = NewEmptyUser()

	var models []db.Model
	err := rds.FindRelatedObjects(r, func(model db.Model) {

		u := model.(*User)
		ptr := NewEmptyUser()
		*ptr = *u

		models = append(models, ptr)

	}, user)

	return models, err
}

// General

func FindRolesForUser(user *User, rds db.RelationalDataStore) ([]*Role, error) {
	var roles []*Role
	var role = NewEmptyRole()

	err := rds.FindOwningObjects(join_users_Role, user, func(model db.Model) {
		r := model.(*Role)

		var ptr = NewEmptyRole()
		*ptr = *r
		roles = append(roles, ptr)

	}, role)

	fmt.Printf("Found %d roles for user %s\n", len(roles), user.ObjectId())
	return roles, err

}
