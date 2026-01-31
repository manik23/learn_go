package main

import (
	"log"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var localUserCache sync.Map

type User struct {
	Name string `json:"name,omitempty"`
}

func loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Logic before the handler (e.g., log start time, IP)
		log.Printf("Request received: %s %s", c.Request.Method, c.Request.URL.Path)
		c.Next() // Pass control to the next middleware or handler
		// Logic after the handler (e.g., log response time, status)
		log.Printf("Request finished: %s %s", c.Request.Method, c.Request.URL.Path)
	}
}

func setupV1Routes(router *gin.Engine) {

	v1 := router.Group("v1")
	v1.Use(loggerMiddleware())
	{
		v1.GET("/user/:id", getUserByID)
		v1.POST("/user", createUser)
	}

}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement the logic:
		//  1) Grab the "X-Auth-Token" header
		//  2) Compare against validToken
		//  3) If mismatch or missing, respond with 401
		//  4) Otherwise pass to next handler

		if r.Header == nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if token := r.Header.Get("X-Auth-Token"); token != "" {
			log.Println("ok")
		}
		next.ServeHTTP(w, r)
	})
}

func getUserByID(c *gin.Context) {
	id := c.Param("id")

	if val, ok := localUserCache.Load(id); !ok {
		c.AbortWithStatus(http.StatusBadRequest)
	} else {
		c.JSON(http.StatusOK, gin.H{
			"created_at": val,
		})
	}

}

func createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	localUserCache.Store(user.Name, time.Now())

	c.JSON(http.StatusOK, nil)

}

func CountWordFrequency(text string) map[string]int {

	wordFreq := make(map[string]int)

	// Create a regex to match only letters and digits
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)

	// Split using regex
	words := re.Split("Hello,world!123", -1) // ["Hello", "world", "123"]

	wordFreq[word]++
	return wordFreq
}

func checkAlphanumeric(b byte) string {
	if ((b-0) >= 'a' && (b-0) <= 'z') || ((b-0) >= 'A' && (b-0) <= 'Z') || ((b-0) >= '0' && (b-0) <= '9') {
		return string(b)
	}
	return ""
}

// BinarySearchRecursive performs binary search using recursion.
// Returns the index of the target if found, or -1 if not found.
func BinarySearchRecursive(arr []int, target int, left int, right int) int {

	if left > right {
		return -1
	}

	mid := int((left + right) / 2)

	if arr[mid] == target {
		return mid
	} else if arr[mid] > target {
		return BinarySearchRecursive(arr, target, left, mid-1)
	} else {
		return BinarySearchRecursive(arr, target, mid+1, right)
	}
}
