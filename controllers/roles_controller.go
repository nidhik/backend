package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
)

type RoleInfo struct {
	Name string `json:"name" binding:"required"`
}

type RoleWithUsers struct {
	*models.Role `json:",inline"`
	Users        []db.Model `json:"users"`
}

type RoleUpdateInfo struct {
	Name   string   `json:"name"`
	Add    []string `json:"addUsers"`
	Remove []string `json:"removeUsers"`
	Read   []string `json:"readAccess"`
	Write  []string `json:"writeAccess"`
}

func GetRole(c *gin.Context) {

	ds := c.MustGet("ds").(db.DataStore)
	id := c.Param("id")

	role := models.NewRole(id)

	var roleAndUsers RoleWithUsers
	if err := role.Fetch(ds); err != nil {
		c.AbortWithError(http.StatusNotFound, err)
	} else {
		models, err2 := role.Users.Find(ds)
		if err2 != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}

		roleAndUsers.Role = role
		roleAndUsers.Users = models

		c.JSON(http.StatusOK, roleAndUsers)
	}
}

func CreateRole(c *gin.Context) {

	var json RoleInfo
	ds := c.MustGet("ds").(db.DataStore)

	if c.BindJSON(&json) == nil {

		role := models.NewEmptyRole()
		role.Set("Name", json.Name)
		role.SetAccessControlList(db.NewACL())

		if err := role.Save(ds); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
		} else {
			c.JSON(http.StatusOK, role)
		}

		return

	}

	c.JSON(http.StatusBadRequest, "Bad request.")

}

func UpdateRole(c *gin.Context) {
	ds := c.MustGet("ds").(db.DataStore)
	id := c.Param("id")

	role := models.NewRole(id)
	// FIXME: if we don't fetch first we have no ACL
	if ferr := role.Fetch(ds); ferr != nil {
		c.AbortWithError(http.StatusInternalServerError, ferr)
	}

	var json RoleUpdateInfo
	if c.BindJSON(&json) == nil {
		if len(json.Name) > 0 {
			role.Set("Name", json.Name)
		}

		for _, userId := range json.Add {
			role.Users.Add(models.NewUser(userId))
		}

		for _, userId := range json.Remove {
			role.Users.Remove(models.NewUser(userId))
		}

		for _, reader := range json.Read {
			role.ACL.AddRead(reader)
		}

		for _, writer := range json.Write {
			role.ACL.AddWrite(writer)
		}

		var err error

		if err = ds.SaveRelatedObjects(role.Users); err == nil {
			if err = role.Save(ds); err == nil {
				c.JSON(http.StatusOK, role)
				return
			}
		}

		c.AbortWithError(http.StatusInternalServerError, err)
		return

	}

	c.JSON(http.StatusBadRequest, "Bad request.")
}
