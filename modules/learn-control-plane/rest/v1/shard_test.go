package v1

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOwnsShard(t *testing.T) {
	t.Run("Single node owns everything", func(t *testing.T) {
		cfg := ShardConfig{NodeIndex: 0, TotalNodes: 1}
		assert.True(t, cfg.OwnsShard("resource-abc"))
		assert.True(t, cfg.OwnsShard("resource-xyz"))
		assert.True(t, cfg.OwnsShard("resource-123"))
	})

	t.Run("Three nodes split ownership", func(t *testing.T) {
		cfg0 := ShardConfig{NodeIndex: 0, TotalNodes: 3}
		cfg1 := ShardConfig{NodeIndex: 1, TotalNodes: 3}
		cfg2 := ShardConfig{NodeIndex: 2, TotalNodes: 3}

		// Every resource must be owned by exactly ONE node
		resources := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
		for _, r := range resources {
			own0 := cfg0.OwnsShard(r)
			own1 := cfg1.OwnsShard(r)
			own2 := cfg2.OwnsShard(r)

			// XOR check: exactly one must be true
			ownedByCount := 0
			if own0 {
				ownedByCount++
			}
			if own1 {
				ownedByCount++
			}
			if own2 {
				ownedByCount++
			}
			assert.Equal(t, 1, ownedByCount, "Resource %q must be owned by exactly 1 node", r)
		}
	})

	t.Run("Ownership is deterministic (same input = same output)", func(t *testing.T) {
		cfg := ShardConfig{NodeIndex: 0, TotalNodes: 3}
		first := cfg.OwnsShard("resource-deterministic")
		second := cfg.OwnsShard("resource-deterministic")
		assert.Equal(t, first, second, "OwnsShard must be deterministic")
	})

	t.Run("Distribution is roughly even across nodes", func(t *testing.T) {
		total := 300
		counts := [3]int{}
		for i := 0; i < total; i++ {
			id := ResourceLedger{ID: "resource-" + string(rune('a'+i%26)) + string(rune('a'+i/26))}.ID
			for n := 0; n < 3; n++ {
				cfg := ShardConfig{NodeIndex: n, TotalNodes: 3}
				if cfg.OwnsShard(id) {
					counts[n]++
				}
			}
		}
		// Each node should own roughly 33% (±10%)
		for n, count := range counts {
			assert.InDelta(t, total/3, count, float64(total)*0.1,
				"Node %d owns %d/%d resources — distribution is skewed", n, count, total)
		}
	})
}

func TestParseShardConfig(t *testing.T) {
	t.Run("Defaults to single-node mode", func(t *testing.T) {
		// No env vars set
		cfg := ParseShardConfig()
		assert.Equal(t, 0, cfg.NodeIndex)
		assert.Equal(t, 1, cfg.TotalNodes)
		// in single-node mode, owns all shards
		assert.True(t, cfg.OwnsShard("any-resource"))
	})

	t.Run("Reads NODE_INDEX and TOTAL_NODES from env", func(t *testing.T) {
		t.Setenv("NODE_INDEX", "2")
		t.Setenv("TOTAL_NODES", "5")
		cfg := ParseShardConfig()
		assert.Equal(t, 2, cfg.NodeIndex)
		assert.Equal(t, 5, cfg.TotalNodes)
	})
}
