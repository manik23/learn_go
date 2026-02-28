package v1

import (
	"errors"
	"log"
	"time"

	"gorm.io/gorm"
)

type ControlPlaneLease struct {
	ID        string    `gorm:"primaryKey"` // Always "reconciler-lock"
	NodeID    string    `json:"node_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

func tryAcquireLease(nodeID string, db *gorm.DB) bool {
	var lease ControlPlaneLease
	now := time.Now()
	leaseDuration := 15 * time.Second

	// 1. Try to Refresh or Takeover using a single Atomic UPDATE
	// We only succeed if:
	//   a) We are the current leader (Heartbeat)
	//   b) The current lease has expired (Takeover)
	result := db.Model(&ControlPlaneLease{}).
		Where("id = ?", "reconciler-lock").
		Where("(node_id = ? OR expires_at < ?)", nodeID, now).
		Updates(map[string]interface{}{
			"node_id":    nodeID,
			"expires_at": now.Add(leaseDuration),
		})

	if result.Error != nil {
		log.Printf("[NODE %s][LEASE] DB error during lease attempt: %v", nodeID, result.Error)
		return false
	}

	if result.RowsAffected > 0 {
		// Log only when we take over, not on every heartbeat to keep logs clean
		// But for learning purposes, let's log the current leader
		log.Printf("[NODE %s][LEASE] Node %s acquired lease", nodeID, nodeID)
		return true
	}

	// 2. If no rows were affected, the lease might not exist at all OR it's held by another active node
	// Check if it exists
	err := db.Where("id = ?", "reconciler-lock").First(&lease).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Create the initial lease
		err = db.Create(&ControlPlaneLease{
			ID:        "reconciler-lock",
			NodeID:    nodeID,
			ExpiresAt: now.Add(leaseDuration),
		}).Error
		return err == nil
	}

	log.Printf("[NODE %s][LEASE] Held by node: %s (Active for %v more)", nodeID, lease.NodeID, time.Until(lease.ExpiresAt).Round(time.Second))
	return false
}
