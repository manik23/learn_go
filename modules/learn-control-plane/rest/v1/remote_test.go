package v1

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getBaseURL() string {
	url := os.Getenv("SERVER_URL")
	if url == "" {
		url = "http://localhost:8080"
	}
	return url
}

func TestRemoteProvisioning(t *testing.T) {
	baseURL := getBaseURL()
	client := &http.Client{Timeout: 15 * time.Second}

	t.Run("Unauthorized Access", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/v1/provision", nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Skip("Server not running at " + baseURL)
		}
		defer resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})

	t.Run("Missing Idempotency Key", func(t *testing.T) {
		req, _ := http.NewRequest("POST", baseURL+"/v1/provision", nil)
		req.Header.Set("X-Auth-Token", "secret")
		resp, err := client.Do(req)
		assert.NoError(t, err)
		if err == nil {
			defer resp.Body.Close()
			assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("Full Provisioning Cycle", func(t *testing.T) {
		resourceID := fmt.Sprintf("remote-res-%d", time.Now().UnixNano())
		body, _ := json.Marshal(ResourceRequest{ID: resourceID})

		// 1. Initial Request
		req, _ := http.NewRequest("POST", baseURL+"/v1/provision", bytes.NewBuffer(body))
		req.Header.Set("X-Auth-Token", "secret")
		req.Header.Set("X-Idempotency-Key", "key-"+resourceID)

		resp, err := client.Do(req)
		assert.NoError(t, err)
		if err == nil {
			defer resp.Body.Close()
			assert.Contains(t, []int{http.StatusCreated, http.StatusAccepted}, resp.StatusCode)

			// If it's Accepted, wait and poll or check final state
			if resp.StatusCode == http.StatusCreated {
				b, _ := io.ReadAll(resp.Body)
				assert.Contains(t, string(b), "successfully provisioned")
			}
		}

		// 2. Exact Duplicate Call (Same Idempotency Key) -> Returns CACHED result
		req2, _ := http.NewRequest("POST", baseURL+"/v1/provision", bytes.NewBuffer(body))
		req2.Header.Set("X-Auth-Token", "secret")
		req2.Header.Set("X-Idempotency-Key", "key-"+resourceID)

		resp2, err := client.Do(req2)
		assert.NoError(t, err)
		if err == nil {
			defer resp2.Body.Close()
			// Should be identical to the first response due to caching
			assert.Equal(t, resp.StatusCode, resp2.StatusCode)
		}

		// 3. Different Idempotency Key (Same Resource ID) -> Hits the LEDGER logic
		req3, _ := http.NewRequest("POST", baseURL+"/v1/provision", bytes.NewBuffer(body))
		req3.Header.Set("X-Auth-Token", "secret")
		req3.Header.Set("X-Idempotency-Key", "other-key-"+resourceID)

		resp3, err := client.Do(req3)
		assert.NoError(t, err)
		if err == nil {
			defer resp3.Body.Close()
			// Should return 200 (Already Provisioned) or 202 (Already in Progress)
			assert.Contains(t, []int{http.StatusOK, http.StatusAccepted}, resp3.StatusCode)
		}
	})
}
