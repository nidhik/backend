package query

import (
	"time"

	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
)

var user = models.NewUser("testrestrictedquerybuilderuser")
var roles = []*models.Role{models.NewRoleWithName("1234", "activeProUser"), models.NewRoleWithName("4567", "admin")}

var unrestrictedQB = NewMongoQueryBuilder()
var restrictedQB = NewRestrictedQueryBuilder(user, roles)

var publicWriteModel = NewTestModelWithWrite(",jfbjdbjdlfkn", db.PUBLIC_KEY)
var userWriteModel = NewTestModelWithWrite("kdajfhiudj.clmzd;.", user.ObjectId())
var mixedPermModel = NewTestModelWithWrite("lskdjfni;", user.ObjectId(), db.PUBLIC_KEY)

var modTime = time.Now()
var utcModTime = modTime.UTC()
var upsertId = "i87hkdzsucg^"
