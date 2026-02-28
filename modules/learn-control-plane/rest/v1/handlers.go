package v1

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProvisioningState int

const (
	PROVISIONING ProvisioningState = iota
	PROVISIONED
	FAILED
)

type Provisioner struct {
	Desired  int64 `json:"desired"`
	Observed int64 `json:"observed"`
	DB       *gorm.DB
	mu       sync.RWMutex
}

func (p *Provisioner) incDesired() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Desired++
}

func (p *Provisioner) decDesired() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Desired--
}

func (p *Provisioner) incObserved() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Observed++
}

func (p *Provisioner) decObserved() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Observed--
}

func (p *Provisioner) getDesired() int64 {
	p.mu.RLock()
	desired := p.Desired
	p.mu.RUnlock()
	return desired
}

func (p *Provisioner) getObserved() int64 {
	p.mu.RLock()
	observed := p.Observed
	p.mu.RUnlock()
	return observed
}

type ResourceRequest struct {
	ID string `json:"id"`
}

type DesiredRequest struct {
	Count int64 `json:"count"`
}

type ResourceResponse struct {
	ID           string            `json:"id"`
	State        ProvisioningState `json:"state"`
	LastUpdateAt time.Time         `json:"last_update_at"`
}

type ResourceLedger struct {
	ID        string            `json:"id"`
	State     ProvisioningState `json:"state"`
	CreatedAt time.Time         `json:"created_at"`
	UpdatedAt time.Time         `json:"updated_at"`
}

type IdempotencyExecution struct {
	Key          string    `gorm:"primaryKey"`
	StatusCode   int       `json:"status_code"`
	ResponseBody []byte    `json:"response_body"`
	CreatedAt    time.Time `json:"created_at"`
}

func SetupV1(serverCtx context.Context, r *gin.Engine, db *gorm.DB) {
	p := &Provisioner{
		Desired:  0,
		Observed: 0,
		DB:       db,
		mu:       sync.RWMutex{},
	}

	p.DB.AutoMigrate(&ResourceLedger{}, &IdempotencyExecution{}, &ControlPlaneLease{})

	// Sync state from Database (Source of Truth)
	p.DB.Model(&ResourceLedger{}).Count(&p.Desired)
	p.DB.Model(&ResourceLedger{}).Where("state = ?", PROVISIONED).Count(&p.Observed)

	go startReconciler(serverCtx, p)

	v1 := r.Group("v1")
	v1.Use(AuthMiddleware())
	// Phase 5.1 Idempotency Key Implementation with Caching
	v1.Use(IdempotencyMiddleware(db))

	v1.GET("/state", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"desired":  p.Desired,
			"observed": p.Observed,
			"status":   "reconciling",
		})
	})

	v1.POST("/provision", p.resourceProvisioningHandler)
	v1.POST("/desired", p.setDesiredHandler)
}

func (p *Provisioner) resourceProvisioningHandler(c *gin.Context) {
	// Phase 1: Increment the desired state
	p.incDesired()

	var req ResourceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var resourceLedger ResourceLedger
	err := p.DB.Where("id = ?", req.ID).First(&resourceLedger).Error

	if err == nil {
		// ALREADY EXISTS: Check the state
		log.Printf("[IDEMPOTENCY] Resource found for Id %s, current state: %v", req.ID, resourceLedger.State)

		if resourceLedger.State == PROVISIONED {
			c.JSON(http.StatusOK, ResourceResponse{
				ID:           resourceLedger.ID,
				State:        resourceLedger.State,
				LastUpdateAt: resourceLedger.UpdatedAt,
			})
			return
		}

		if resourceLedger.State == PROVISIONING {
			c.JSON(http.StatusAccepted, gin.H{"message": "Resource provisioning already in progress"})
			return
		}
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		resourceLedger = ResourceLedger{
			ID:    req.ID,
			State: PROVISIONING,
		}
		if err := p.DB.Create(&resourceLedger).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create resource ledger"})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	select {
	case <-c.Done():
		log.Printf("Client Disconnected for Id %s", req.ID)
		return
	case <-time.After(time.Duration(rand.Intn(5)) * time.Second):
		log.Printf("Resource provisioning completed for Id %s", req.ID)
		p.DB.Model(&resourceLedger).Update("state", PROVISIONED)
		p.incObserved()
		c.JSON(http.StatusCreated, gin.H{"message": "successfully provisioned"})
	}
}

func (p *Provisioner) setDesiredHandler(c *gin.Context) {
	var req DesiredRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	p.mu.Lock()
	p.Desired = req.Count
	p.mu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"message": "Desired state updated",
		"desired": p.Desired,
	})
}
