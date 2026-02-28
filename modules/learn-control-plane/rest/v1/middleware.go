package v1

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type bodyWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
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

func IdempotencyMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only apply to state-changing methods
		// EXCEPTION: Skip for /v1/desired as it's a control-plane update that shouldn't require client-side keys for learning
		if c.Request.Method == http.MethodGet || c.Request.URL.Path == "/v1/desired" {
			c.Next()
			return
		}

		key := c.Request.Header.Get("X-Idempotency-Key")
		if key == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "X-Idempotency-Key header is required"})
			return
		}

		// 1. Check if we have a cached result
		var execution IdempotencyExecution
		err := db.Where("key = ?", key).First(&execution).Error
		if err == nil {
			log.Printf("[IDEMPOTENCY] Returning cached result for key: %s", key)
			c.Data(execution.StatusCode, "application/json", execution.ResponseBody)
			c.Abort()
			return
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("[IDEMPOTENCY] Failed to check idempotency for key %s: %v", key, err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "idempotency check failed"})
			return
		}

		// 2. Wrap the response writer to capture the result
		bw := &bodyWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = bw

		c.Next()

		// 3. Store the result if the request was successful
		if c.Writer.Status() < 400 {
			log.Printf("[IDEMPOTENCY] Caching result for key: %s", key)
			capture := IdempotencyExecution{
				Key:          key,
				StatusCode:   c.Writer.Status(),
				ResponseBody: bw.body.Bytes(),
				CreatedAt:    time.Now(),
			}
			if err := db.Create(&capture).Error; err != nil {
				log.Printf("[IDEMPOTENCY] Failed to cache result for key %s: %v", key, err)
			}
		}
	}
}

func MaxConcurrentMiddleware(admissionTokens int) gin.HandlerFunc {
	concurrencyLimiter := make(chan struct{}, admissionTokens)
	return func(c *gin.Context) {
		select {
		case concurrencyLimiter <- struct{}{}:
			defer func() { <-concurrencyLimiter }()
			c.Next()
		default:
			c.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
}

func RateLimiterMiddleware(requestPerSecond int, rate time.Duration) gin.HandlerFunc {
	rateLimiter := make(chan struct{}, requestPerSecond)

	// Fill the bucket with initial tokens
	for range requestPerSecond {
		rateLimiter <- struct{}{}
	}

	go func() {
		t := time.NewTicker(rate)
		defer t.Stop()
		for range t.C {
			for range requestPerSecond {
				select {
				case rateLimiter <- struct{}{}:
				default:
					goto nextTick
				}
			}
		nextTick:
		}
	}()

	return func(c *gin.Context) {
		select {
		case <-rateLimiter:
			c.Next()
		default:
			c.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
}
