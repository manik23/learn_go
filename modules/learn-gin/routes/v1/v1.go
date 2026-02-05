package v1

import (
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

func SetupV1Routes(router *gin.Engine, db *gorm.DB) error {
	v1 := router.Group("v1")
	return setupUserHandler(v1, db)
}
