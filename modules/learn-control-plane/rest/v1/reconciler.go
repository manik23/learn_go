package v1

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"
)

func startReconciler(ctx context.Context, p *Provisioner) {

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		select {
		case <-ctx.Done():
			return
		default:
			p.Reconcile()
		}
	}
}

func (p *Provisioner) Reconcile() {
	nodeId := os.Getenv("NODE_ID")
	if nodeId == "" {
		log.Printf("[RECONCILER] NODE_ID not set, skipping reconciliation")
		return
	}

	if !tryAcquireLease(nodeId, p.DB) {
		log.Printf("[RECONCILER] Failed to acquire lease, skipping reconciliation")
		return
	}

	var observedCount int64
	var totalCount int64
	p.DB.Model(&ResourceLedger{}).Where("state = ?", PROVISIONED).Count(&observedCount)
	p.DB.Model(&ResourceLedger{}).Count(&totalCount)

	// Sync memory state with DB reality
	p.mu.Lock()
	p.Observed = observedCount
	p.mu.Unlock()

	desired := p.getDesired()
	observed := p.getObserved()

	if desired == observed {
		log.Println("[RECONCILER] Status: Stable (Desired == Observed)")
		return
	}

	if desired > totalCount {
		diff := desired - totalCount
		log.Printf("[RECONCILER] ScaleUp: Need %d more resources. Creating records...", diff)
		for i := 0; i < int(diff); i++ {
			id := fmt.Sprintf("auto-%d-%d", time.Now().Unix(), i)
			p.DB.Create(&ResourceLedger{ID: id, State: PROVISIONING})
		}
	} else if desired < observed {
		diff := observed - desired
		log.Printf("[RECONCILER] ScaleDown: Surplus %d resources. Cleaning up...", diff)
		var resources []ResourceLedger
		p.DB.Where("state = ?", PROVISIONED).Limit(int(diff)).Find(&resources)
		for _, r := range resources {
			p.DB.Delete(&r)
			p.decObserved()
		}
	}

	// Any resources stuck in PROVISIONING? Finish them.
	var inFlight []ResourceLedger
	p.DB.Where("state = ?", PROVISIONING).Find(&inFlight)
	for _, r := range inFlight {
		log.Printf("[RECONCILER] Completing work for in-flight resource %s", r.ID)
		p.DB.Model(&r).Update("state", PROVISIONED)
		p.incObserved()
	}
}
