package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
)

type createdAfter struct {
	Time string `json:"createdAfter" binding:"required"`
}

type userId struct {
	Id string `json:"id" binding:"required"`
}

func Me(c *gin.Context) {
	user := c.MustGet("user").(*models.User)
	c.JSON(http.StatusOK, user)
}

func GetUser(c *gin.Context) {

	id := c.Param("id")
	ds := c.MustGet("ds").(db.DataStore)

	pointer := models.NewPointer(models.CollectionUser, id)
	model, err := pointer.Fetch(ds)

	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
	} else {
		c.JSON(http.StatusOK, model)
	}

}

func GetRolesForUser(c *gin.Context) {
	ds := c.MustGet("ds").(db.DataStore)
	var json userId

	if c.BindJSON(&json) == nil {
		user := models.NewUser(json.Id)
		roles, err := models.FindRolesForUser(user, ds)
		if err != nil {
			c.AbortWithError(http.StatusNotFound, err)
		} else {
			result := roles
			if roles == nil {
				result = make([]*models.Role, 0)
			}
			c.JSON(http.StatusOK, result)

		}

		return
	}

	c.JSON(http.StatusBadRequest, "Bad request.")

}
