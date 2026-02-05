package v1

import "github.com/gin-gonic/gin"

func SetupV1Routes(router *gin.Engine) {
	userHandler := newUserHandler()
	setupUserHandler(router, userHandler)
}
