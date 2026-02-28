package v1

import (
	"hash/fnv"
	"log"
	"os"
	"strconv"
)

// ShardConfig holds the sharding configuration for this node.
// In a real system this would be discovered via service registry (etcd/consul).
type ShardConfig struct {
	// NodeIndex is this node's position in the cluster (0-based).
	// For node-1: 0, node-2: 1, node-3: 2
	NodeIndex int

	// TotalNodes is the total number of nodes in the cluster.
	TotalNodes int
}

// OwnsShard returns true if this node is responsible for the given resourceID.
// Uses FNV-1a: deterministic, fast, no coordination needed — pure math.
func (cfg ShardConfig) OwnsShard(resourceID string) bool {
	h := fnv.New32a()
	h.Write([]byte(resourceID))
	return h.Sum32()%uint32(cfg.TotalNodes) == uint32(cfg.NodeIndex)
}

// ParseShardConfig reads NODE_INDEX and TOTAL_NODES from the environment.
//   - Missing / invalid NODE_INDEX → defaults to 0
//   - Missing / invalid TOTAL_NODES → defaults to 1 (single-node: owns everything)
func ParseShardConfig() ShardConfig {
	nodeIndex := 0
	if v := os.Getenv("NODE_INDEX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			nodeIndex = n
		}
	}

	totalNodes := 1
	if v := os.Getenv("TOTAL_NODES"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			totalNodes = n
		}
	}

	cfg := ShardConfig{NodeIndex: nodeIndex, TotalNodes: totalNodes}
	log.Printf("[SHARD] Config: node %d of %d (owns ~%.0f%% of resources)",
		nodeIndex, totalNodes, float64(100)/float64(totalNodes))
	return cfg
}
