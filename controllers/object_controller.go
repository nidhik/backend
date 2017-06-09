package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/nidhik/backend/db"
	"github.com/nidhik/backend/models"
)

func QueryCollection(c *gin.Context) {
	collection := c.Param("collection")
	switch collection {
	default:
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
}

func GetAll(c *gin.Context) {
	collection := c.Param("collection")
	switch collection {
	default:
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
}

func UpdateModel(c *gin.Context) {
	collection := c.Param("collection")
	switch collection {
	default:
		c.AbortWithStatus(http.StatusNotFound)
		return
	}
}

func GetModel(c *gin.Context) {

	collection := c.Param("collection")

	switch collection {
	case models.CollectionRole:
		GetRole(c)
		return
	case models.CollectionUser:
		GetUser(c)
		return
	default:
		getModelFromPointer(c)
		return
	}
}

func getModelFromPointer(c *gin.Context) {
	ds := c.MustGet("ds").(db.DataStore)
	id := c.Param("id")
	collection := c.Param("collection")

	pointer := models.NewPointer(collection, id)
	model, err := pointer.Fetch(ds)

	if err != nil {
		c.AbortWithError(http.StatusNotFound, err)
	} else {
		c.JSON(http.StatusOK, model)
	}

}
