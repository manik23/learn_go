package v1

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestRouter() (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	SetupV1(r, db)
	return r, db
}

func TestAuthMiddleware(t *testing.T) {
	router, _ := setupTestRouter()

	t.Run("Missing Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/state", nil)
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Invalid Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/state", nil)
		req.Header.Set("X-Auth-Token", "wrong")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("Valid Token", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/v1/state", nil)
		req.Header.Set("X-Auth-Token", "secret")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code) // Fails later at Idempotency check
	})
}

func TestIdempotencyMiddleware(t *testing.T) {
	router, _ := setupTestRouter()

	t.Run("Missing Idempotency Key", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/provision", nil)
		req.Header.Set("X-Auth-Token", "secret")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestProvisioningFlow(t *testing.T) {
	router, db := setupTestRouter()

	t.Run("Successful Provisioning", func(t *testing.T) {
		body, _ := json.Marshal(ResourceRequest{ID: "res-1"})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/provision", bytes.NewBuffer(body))
		req.Header.Set("X-Auth-Token", "secret")
		req.Header.Set("X-Idempotency-Key", "key-1")

		router.ServeHTTP(w, req)

		// Note: The handler has a random sleep. In a real test we might mock time
		// or use a shorter sleep for testing.
		assert.Contains(t, []int{http.StatusCreated, http.StatusOK}, w.Code)
	})

	t.Run("Already Provisioned", func(t *testing.T) {
		db.Create(&ResourceLedger{ID: "res-done", State: PROVISIONED})

		body, _ := json.Marshal(ResourceRequest{ID: "res-done"})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/v1/provision", bytes.NewBuffer(body))
		req.Header.Set("X-Auth-Token", "secret")
		req.Header.Set("X-Idempotency-Key", "key-done-any")
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "res-done")
	})
}
