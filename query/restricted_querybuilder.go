package query

import (
	"errors"

	"time"

	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var ERR_ACCESS_DENIED = errors.New("You don't have persmission to carry out this operation.")

type RestrictedMongoQueryBuilder struct {
	User    *models.User
	Roles   []*models.Role
	builder *MongoQueryBuilder
	access  []interface{}
}

func NewRestrictedQueryBuilder(user *models.User, roles []*models.Role) *RestrictedMongoQueryBuilder {
	access := getAccess(user, roles)
	return &RestrictedMongoQueryBuilder{user, roles, &MongoQueryBuilder{}, access}
}

// Query & Update Builders

func getAccess(user *models.User, roles []*models.Role) []interface{} {
	access := []interface{}{user.ObjectId(), db.PUBLIC_KEY}
	for _, r := range roles {
		access = append(access, "role:"+r.Name)
	}
	return access
}

func (m *RestrictedMongoQueryBuilder) addReadCheck(query bson.M) {
	query["$or"] = []bson.M{
		bson.M{db.READ_PERM: bson.M{"$exists": false}},
		bson.M{db.READ_PERM: bson.M{"$in": m.access}},
	}

}

func (m *RestrictedMongoQueryBuilder) addWriteCheck(query bson.M) {
	query[db.WRITE_PERM] = bson.M{
		"$in": m.access,
	}
}

func (m *RestrictedMongoQueryBuilder) checkWrite(model db.Model) error {

	if len(model.ObjectId()) == 0 {
		return nil // this is going to be an insert
	}

	if model.Collection() == models.CollectionUser && m.User.ObjectId() == model.ObjectId() {
		return nil
	}

	if model.AccessControlList() == nil || !model.AccessControlList().CanWrite(m.User.ObjectId()) {
		return ERR_ACCESS_DENIED
	}

	return nil
}

func (m *RestrictedMongoQueryBuilder) MakeCountQuery(collectionName string, query map[string]interface{}) bson.M {
	result := m.builder.MakeCountQuery(collectionName, query)
	m.addReadCheck(result)
	return result
}

func (m *RestrictedMongoQueryBuilder) MakeFindQuery(collectionName string, query map[string]interface{}) bson.M {
	result := m.builder.MakeFindQuery(collectionName, query)

	if collectionName == models.CollectionUser && m.User.ObjectId() == query["_id"] {
		return result
	}

	m.addReadCheck(result)
	return result
}

func (m *RestrictedMongoQueryBuilder) MakeFindByIdQuery(model db.Model) (bson.M, error) {
	result, err := m.builder.MakeFindByIdQuery(model)

	if model.Collection() == models.CollectionUser && m.User.ObjectId() == model.ObjectId() {
		return result, nil
	}

	if err != nil {
		return nil, err
	}
	m.addReadCheck(result)
	return result, nil
}

func (m *RestrictedMongoQueryBuilder) MakeRemoveQuery(model db.Model) (bson.M, error) {

	if perr := m.checkWrite(model); perr != nil {
		return nil, perr
	}

	result, err := m.builder.MakeRemoveQuery(model)

	if model.Collection() == models.CollectionUser && m.User.ObjectId() == model.ObjectId() {
		return result, nil
	}

	if err != nil {
		return nil, err
	}

	m.addWriteCheck(result)
	return result, nil

}

func (m *RestrictedMongoQueryBuilder) MakeRemoveAllQuery(collectionName string, query map[string]interface{}) (bson.M, error) {
	result, err := m.builder.MakeRemoveAllQuery(collectionName, query)
	if err != nil {
		return nil, err
	}

	if collectionName == models.CollectionUser || collectionName == models.CollectionRole {
		return nil, ERR_ACCESS_DENIED
	}

	if query["_p_user"] != models.PointerString(m.User) {
		return nil, ERR_ACCESS_DENIED // cannot delete objects that are not yours
	}

	m.addWriteCheck(result)
	return result, nil

}

func (m *RestrictedMongoQueryBuilder) MakeInsertDocument(model db.Model, t time.Time, id string) (db.Model, error) {

	result, err := m.builder.MakeInsertDocument(model, t, id)

	if err != nil {
		return nil, err
	}

	return result, nil

}

// findAndModify() https://docs.mongodb.com/manual/reference/method/db.collection.findAndModify/
func (m *RestrictedMongoQueryBuilder) MakeChangeDocument(model db.Model, t time.Time) (bson.M, mgo.Change, error) {
	if perr := m.checkWrite(model); perr != nil {
		return nil, mgo.Change{}, perr
	}

	q, c, err := m.builder.MakeChangeDocument(model, t)

	if err != nil {
		return nil, mgo.Change{}, err
	}

	if model.Collection() == models.CollectionUser && m.User.ObjectId() == q["_id"] {
		return q, c, nil
	}

	m.addWriteCheck(q)
	return q, c, nil
}

func (m *RestrictedMongoQueryBuilder) MakeUpsertDocument(model db.Model, query map[string]interface{}, t time.Time, id string) (bson.M, mgo.Change, error) {

	if perr := m.checkWrite(model); perr != nil {
		return nil, mgo.Change{}, perr
	}

	q, c, err := m.builder.MakeUpsertDocument(model, query, t, id)

	if err != nil {
		return nil, mgo.Change{}, err
	}

	m.addWriteCheck(q)
	return q, c, nil
}

func (m *RestrictedMongoQueryBuilder) QueryByRelatedModels(joinCollectionName string, related []db.Model) bson.M {
	return m.builder.QueryByRelatedModels(joinCollectionName, related)
}

func (m *RestrictedMongoQueryBuilder) QueryByOwningModels(joinCollectionName string, owning []db.Model) bson.M {
	return m.builder.QueryByOwningModels(joinCollectionName, owning)
}

func (m *RestrictedMongoQueryBuilder) QueryByIds(collectionName string, ids []string) bson.M {
	result := m.builder.QueryByIds(collectionName, ids)
	m.addReadCheck(result)
	return result
}

func (m *RestrictedMongoQueryBuilder) MakeRelationUpdateDocuments(relation db.Relation) ([]interface{}, bson.M, error) {
	if perr := m.checkWrite(relation.Owner()); perr != nil {
		return nil, nil, perr
	}

	r1, r2, err := m.builder.MakeRelationUpdateDocuments(relation)
	return r1, r2, err
}
