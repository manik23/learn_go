package v1

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
//
// 💡 CHALLENGE: Implement Consistent Hashing.
//
// RULES:
//   - Use the FNV-1a hash function (already imported: hash/fnv) to hash the resourceID.
//   - A resource belongs to a node if: hash(resourceID) % TotalNodes == NodeIndex
//
// EXAMPLE:
//
//	cfg := ShardConfig{NodeIndex: 0, TotalNodes: 3}
//	cfg.OwnsShard("resource-abc")  // true if fnv("resource-abc") % 3 == 0
//
// WHY FNV?
//
//	It's fast and distributes strings uniformly.
//	Every node given the same input will always get the same output.
//	No coordination needed — it's pure math.
//
// TODO: Implement this function.
func (cfg ShardConfig) OwnsShard(resourceID string) bool {
	panic("TODO: implement OwnsShard using FNV-1a hash")
}

// ParseShardConfig reads NODE_INDEX and TOTAL_NODES from the environment.
//
// 💡 CHALLENGE: Implement this function.
//
// RULES:
//   - Read NODE_INDEX (int) and TOTAL_NODES (int) from environment variables.
//   - If NODE_INDEX is missing or invalid, default to 0.
//   - If TOTAL_NODES is missing, invalid, or <= 0, default to 1 (single-node mode = owns everything).
//
// HINT: Use strconv.Atoi to parse integers from strings.
//
// TODO: Implement this function.
func ParseShardConfig() ShardConfig {
	panic("TODO: implement ParseShardConfig")
}
