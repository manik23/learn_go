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
	nodeID := os.Getenv("NODE_ID")
	if nodeID == "" {
		nodeID = "local"
	}

	// --- Architecture Decision ---
	//
	// Phase 5.3 (Leader Election) and Phase 5.4 (Sharding) solve DIFFERENT problems:
	//
	//   Leader Election  → ONE node controls GLOBAL state (e.g., setting Desired count).
	//                      Prevents two nodes from simultaneously deciding to scale.
	//
	//   Sharding         → ALL nodes reconcile in PARALLEL, each owning a unique shard.
	//                      The deployment layer (K8s: one pod per NODE_INDEX) prevents
	//                      two nodes with the same index from running simultaneously.
	//                      This is "natural partitioning" — no lock needed per shard.
	//
	// Production pattern (Kubernetes):
	//   Each shard gets its OWN lease: "reconciler-lock-shard-0", "reconciler-lock-shard-1".
	//   Here we use the simpler model: global lease for global ops, sharding for local work.

	shard := ParseShardConfig()

	// STEP 1: Global gate — only the leader adjusts Desired state cluster-wide.
	isLeader := tryAcquireLease(nodeID, p.DB)

	if isLeader {
		p.reconcileGlobalState(nodeID)
	} else {
		log.Printf("[NODE %s][RECONCILER] Follower — skipping global state management", nodeID)
	}

	// STEP 2: Per-shard work — ALL nodes do this, regardless of leader status.
	// Safety: two nodes share a shard ONLY if they have the same NODE_INDEX.
	// The deployment layer must prevent this (K8s StatefulSet, Docker --name uniqueness).
	p.reconcileShard(nodeID, shard)
}

// reconcileGlobalState is only run by the current leader.
// It is responsible for cluster-wide decisions: scaling up/down total resource count.
// It is also the SOLE authority on updating p.Observed (the cluster-wide reality).
func (p *Provisioner) reconcileGlobalState(nodeID string) {
	var totalCount int64
	var observedCount int64
	p.DB.Model(&ResourceLedger{}).Count(&totalCount)
	p.DB.Model(&ResourceLedger{}).Where("state = ?", PROVISIONED).Count(&observedCount)

	// Keep in-memory p.Observed in sync with cluster-wide DB reality.
	// Only the leader does this to avoid race conditions across nodes.
	p.mu.Lock()
	p.Observed = observedCount
	p.mu.Unlock()

	desired := p.getDesired()

	log.Printf("[NODE %s][LEADER] Global state: Desired=%d TotalInDB=%d Observed=%d",
		nodeID, desired, totalCount, observedCount)

	if desired > totalCount {
		diff := desired - totalCount
		log.Printf("[NODE %s][LEADER] ScaleUp: Creating %d new resource stubs", nodeID, diff)
		for i := 0; i < int(diff); i++ {
			id := fmt.Sprintf("global-auto-%d-%d", time.Now().UnixNano(), i)
			p.DB.Create(&ResourceLedger{ID: id, State: PROVISIONING})
		}
	} else if desired < totalCount {
		diff := totalCount - desired
		log.Printf("[NODE %s][LEADER] ScaleDown: Marking %d resources for deletion", nodeID, diff)
		var surplus []ResourceLedger
		p.DB.Where("state = ?", PROVISIONED).Limit(int(diff)).Find(&surplus)
		for _, r := range surplus {
			p.DB.Delete(&r)
		}
	}
}

// reconcileShard is run by ALL nodes. Each node only processes resources
// whose ID hashes to its NODE_INDEX. No lock needed — natural partitioning.
//
// IMPORTANT: This function does NOT write to p.Observed.
// p.Observed is a cluster-wide counter owned exclusively by reconcileGlobalState().
// Using a local variable here prevents stomping the global count with a per-shard slice.
func (p *Provisioner) reconcileShard(nodeID string, shard ShardConfig) {
	// Load all resources, filter to this node's shard in-memory.
	// (DB doesn't understand our hash function — this is the standard pattern.)
	var allResources []ResourceLedger
	p.DB.Find(&allResources)

	// Local count — never written to p.Observed
	myObserved := int64(0)
	for _, r := range allResources {
		if !shard.OwnsShard(r.ID) {
			continue
		}

		switch r.State {
		case PROVISIONED:
			myObserved++
		case PROVISIONING:
			// Complete in-flight work for this shard's resources
			log.Printf("[NODE %s][SHARD %d/%d] Completing resource: %s",
				nodeID, shard.NodeIndex, shard.TotalNodes, r.ID)
			p.DB.Model(&r).Update("state", PROVISIONED)
			myObserved++
		}
	}

	// Log only — this is a shard-local metric, not the cluster-wide p.Observed
	log.Printf("[NODE %s][SHARD Index %d/%d] My shard observed = %d (cluster p.Observed = %d)",
		nodeID, shard.NodeIndex, shard.TotalNodes, myObserved, p.getObserved())
}
