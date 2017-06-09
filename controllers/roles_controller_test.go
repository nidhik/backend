package controllers

import (
	"bytes"
	"encoding/json"

	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
	"github.com/nidhik/backend/routes"
)

// Test Cases

type GetRoleTest struct {
	desc         string
	id_param     string
	role_id      string
	name         string
	acl          *db.ACL
	users        []*models.User
	responseCode int
}
type PostRoleTest struct {
	desc         string
	payload      []byte
	name         string
	acl          *db.ACL
	responseCode int
}

func (t *GetRoleTest) description() string {
	return t.desc
}

func (t *PostRoleTest) description() string {
	return t.desc
}

var roleA = models.NewEmptyRole()
var roleB = models.NewEmptyRole()
var roleWithUsers = models.NewEmptyRole()

var userFoo = models.NewEmptyUser()
var userBar = models.NewEmptyUser()

var testACLA = &db.ACL{ACL: map[string]db.Permission{"ab9pnlkj4c": db.Permission{Read: true, Write: true}}}
var testACLB = &db.ACL{ACL: map[string]db.Permission{"abdedefec": db.Permission{Read: true, Write: true}}}
var testACLC = &db.ACL{ACL: map[string]db.Permission{"ldsfniruneo": db.Permission{Read: true, Write: true}}}

var getRoleTests []TestCase
var postRoleTests []TestCase

// Test Info
type RolesControllerTest struct{}

func (c *RolesControllerTest) routeAndHandler(method int) (string, gin.HandlerFunc) {

	switch method {
	case GET:
		return routes.GET_ROLE, GetRole
	case POST:
		return routes.ROLES, CreateRole
	default:
		return "", nil
	}

}

func (c *RolesControllerTest) testCases(method int) []TestCase {

	switch method {
	case GET:
		return getRoleTests
	case POST:
		return postRoleTests
	default:
		return nil
	}
}

func (c *RolesControllerTest) setupDataStore(t *testing.T, ds db.DataStore) {

	roleA.Set("Name", "roleA")
	roleA.SetAccessControlList(testACLA)
	if err := roleA.Save(ds); err != nil {
		t.Fatal("Could not setup test database:", err)
	}

	roleB.Set("Name", "roleB")
	roleA.SetAccessControlList(testACLB)
	if err := roleB.Save(ds); err != nil {
		t.Fatal("Could not setup test database:", err)
	}

	roleWithUsers.Set("Name", "roleWithUsers")
	roleWithUsers.SetAccessControlList(testACLC)
	if err := roleWithUsers.Save(ds); err != nil {
		t.Fatal("Could not setup test database:", err)
	}

	if err := userFoo.Save(ds); err != nil {
		t.Fatal("Could not setup test database:", err)
	}

	if err := userBar.Save(ds); err != nil {
		t.Fatal("Could not setup test database:", err)
	}

	roleWithUsers.Users.Add(userFoo)
	roleWithUsers.Users.Add(userBar)

	if err := ds.SaveRelatedObjects(roleWithUsers.Users); err != nil {
		t.Fatal("Could not setup test database:", err)
	}

	getRoleTests = []TestCase{
		&GetRoleTest{"that existing role is found", roleA.ObjectId(), roleA.ObjectId(), roleA.Name, testACLA, nil, 200},
		&GetRoleTest{"that invalid role is not found", "some_invalid_id", roleB.ObjectId(), roleB.Name, testACLB, nil, 404},
		&GetRoleTest{"that existing role with users is found and users are returned", roleWithUsers.ObjectId(), roleWithUsers.ObjectId(), roleWithUsers.Name, testACLC, []*models.User{userFoo, userBar}, 200}}

	postRoleTests = []TestCase{
		&PostRoleTest{"that a role is created from a valid request", []byte(`{"name":"activeProUser_lkjsdnlksjan"}`), "activeProUser_lkjsdnlksjan", db.NewACL(), 200},
		&PostRoleTest{"that a bad request returns a 400", []byte(`{"foo":"bar"}`), "", nil, 400},
	}
}

func (c *RolesControllerTest) runGetTest(t *testing.T, test interface{}, router *gin.Engine) {
	testCase := test.(*GetRoleTest)
	resp := recordGet(router, routes.ROLES+"/"+testCase.id_param, nil)
	verifyGetRoleResponse(t, testCase, resp)
}

func (c *RolesControllerTest) runPostTest(t *testing.T, test interface{}, router *gin.Engine) {
	testCase := test.(*PostRoleTest)
	resp := recordPost(router, routes.ROLES, bytes.NewBuffer(testCase.payload))
	verifyPostRoleResponse(t, testCase, resp)
}

type GetRoleResult struct {
	models.Role `json:",inline"`
	Users       []*models.User `json:"users"`
}

func verifyGetRoleResponse(t *testing.T, test *GetRoleTest, resp *httptest.ResponseRecorder) {

	if resp.Code != test.responseCode {
		t.Fatal("Expected: ", test.responseCode, "got: ", resp.Code)
	}

	if resp.Code == 200 {

		var r GetRoleResult
		json.Unmarshal(resp.Body.Bytes(), &r)

		if r.Id != test.role_id {
			t.Fatal("Expected Id:", test.role_id, "got:", r.Id)
		}

		if r.Name != test.name {
			t.Fatal("Expected Name:", test.name, "got:", r.Name)
		}

		if r.Collection() != models.CollectionRole {
			t.Fatal("Expected Collection:", models.CollectionRole, "got:", r.Collection())
		}

		if !reflect.DeepEqual(test.acl.ACL, r.AccessControlList().ACL) {
			t.Fatal("Expected ACL:", test.acl.ACL, "got:", r.AccessControlList().ACL)
		}

		if len(r.Users) != len(test.users) {
			t.Fatal("Expected # related users:", len(test.users), "got:", len(r.Users))
		}

		for i, expectedUser := range test.users {
			returned := r.Users[i]

			if returned == nil {
				t.Fatal("Expected user:", expectedUser.ObjectId(), "at index: ", i, "got nil")
			} else if expectedUser.ObjectId() != returned.ObjectId() {
				t.Fatal("Expected user:", expectedUser.ObjectId(), "at index: ", i, "got:", returned.ObjectId())
			}

		}
	}
}

func verifyPostRoleResponse(t *testing.T, test *PostRoleTest, resp *httptest.ResponseRecorder) {
	if resp.Code != test.responseCode {
		t.Fatal("Expected: ", test.responseCode, "got: ", resp.Code)
	}

	if resp.Code == 200 {

		var r models.Role
		json.Unmarshal(resp.Body.Bytes(), &r)

		if len(r.Id) == 0 {
			t.Fatal("Did not return role id.")
		}

		now := getDateFromTime(time.Now())
		createdAt := getDateFromTime(r.CreatedDate())
		updatedAt := getDateFromTime(r.UpdatedDate())

		if !createdAt.Equal(now) {
			t.Fatal("Expected:", now, "got:", createdAt)
		}

		if !updatedAt.Equal(now) {
			t.Fatal("Expected:", now, "got:", updatedAt)
		}

		if r.Name != test.name {
			t.Fatal("Expected Name:", test.name, "got:", r.Name)
		}

		if r.Collection() != models.CollectionRole {
			t.Fatal("Expected Collection:", models.CollectionRole, "got:", r.Collection())
		}

		if !reflect.DeepEqual(test.acl.ACL, r.AccessControlList().ACL) {
			t.Fatal("Expected ACL:", test.acl.ACL, "got:", r.AccessControlList().ACL)
		}

	}
}

// Not used

func (c *RolesControllerTest) runDeleteTest(t *testing.T, test interface{}, router *gin.Engine) {}

func (c *RolesControllerTest) runUpdateTest(t *testing.T, test interface{}, router *gin.Engine) {}
