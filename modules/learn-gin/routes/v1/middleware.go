package v1

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Logic before the handler (e.g., log start time, IP)
		log.Printf("Request received: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next() // Pass control to the next middleware or handler
		// Logic after the handler (e.g., log response time, status)
		log.Printf("Request finished: %s %s", c.Request.Method, c.Request.URL.Path)
	}
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: Implement the logic:
		//  1) Grab the "X-Auth-Token" header
		//  2) Compare against validToken
		//  3) If mismatch or missing, respond with 401
		//  4) Otherwise pass to next handler

		if c.Request.Header == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		if token := c.Request.Header.Get("X-Auth-Token"); token != "secret" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

func MaxConcurrentMiddleware(addmissionTokens int) gin.HandlerFunc {
	rateLimiter := make(chan struct{}, addmissionTokens)
	return func(c *gin.Context) {

		select {
		case rateLimiter <- struct{}{}:
			defer func() { <-rateLimiter }()
			c.Next()
		default:
			c.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
}
