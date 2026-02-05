package v1

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const (
	MAX_CONCURRENT_REQUESTS = 10
)

type User struct {
	Name      string    `json:"name,omitempty" binding:"required"`
	CreatedAt time.Time `json:"created_at,omitzero"`
}

type UserHandler struct {
	db *gorm.DB
}

func newUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

func setupUserHandler(v1 *gin.RouterGroup, db *gorm.DB) error {
	if err := db.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("failed to migrate user table: %w", err)
	}

	userHandler := newUserHandler(db)

	v1.Use(loggerMiddleware())
	v1.Use(MaxConcurrentMiddleware(MAX_CONCURRENT_REQUESTS))
	v1.Use(AuthMiddleware())
	{
		v1.GET("/user/:id", userHandler.getUserByID)
		v1.POST("/user", userHandler.createUser)
	}

	return nil
}

func (h *UserHandler) getUserByID(c *gin.Context) {
	id := c.Param("id")

	var user User
	if err := h.db.Where("name = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) createUser(c *gin.Context) {
	var user User
	if err := c.ShouldBindBodyWithJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user.CreatedAt = time.Now()

	tx := h.db.FirstOrCreate(&user, User{Name: user.Name})
	if tx.Error != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": tx.Error.Error()})
		return
	}

	c.JSON(http.StatusOK, nil)
}

// func CountWordFrequency(text string) map[string]int {
// 	wordFreq := make(map[string]int)

// 	// Create a regex to match only letters and digits
// 	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)

// 	// Split using regex
// 	words := re.Split("Hello,world!123", -1) // ["Hello", "world", "123"]

// 	wordFreq[word]++
// 	return wordFreq
// }

// func checkAlphanumeric(b byte) string {
// 	if ((b-0) >= 'a' && (b-0) <= 'z') || ((b-0) >= 'A' && (b-0) <= 'Z') || ((b-0) >= '0' && (b-0) <= '9') {
// 		return string(b)
// 	}
// 	return ""
// }

// // BinarySearchRecursive performs binary search using recursion.
// // Returns the index of the target if found, or -1 if not found.
// func BinarySearchRecursive(arr []int, target int, left int, right int) int {
// 	if left > right {
// 		return -1
// 	}

// 	mid := int((left + right) / 2)

// 	if arr[mid] == target {
// 		return mid
// 	} else if arr[mid] > target {
// 		return BinarySearchRecursive(arr, target, left, mid-1)
// 	} else {
// 		return BinarySearchRecursive(arr, target, mid+1, right)
// 	}
// }
