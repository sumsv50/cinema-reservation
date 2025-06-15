package reservation_test

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"testing"
)

func TestConcurrentSeatReservation(t *testing.T) {
	const (
		totalRequests = 10000
		targetURL     = "http://localhost:8080/api/v1/reservations"
		cinemaSlug    = "grand-cinema-downtown"
		row           = 0
		column        = 0
	)

	payload := []byte(fmt.Sprintf(`
		{
			"cinema_slug": "%s",
			"seats": [
					{
							"row": %d,
							"column": %d
					}
			]
		}
	`, cinemaSlug, row, column))
	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		counts = make(map[int]int) // key: HTTP status code, value: count
	)

	client := &http.Client{}
	fire := make(chan struct{})
	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-fire
			resp, err := client.Post(targetURL, "application/json", bytes.NewBuffer(payload))
			if err != nil {
				mu.Lock()
				counts[-1]++
				mu.Unlock()
				t.Errorf("Request failed: %v", err)
				return
			}
			defer resp.Body.Close()

			mu.Lock()
			counts[resp.StatusCode]++
			mu.Unlock()
		}()
	}

	close(fire)
	wg.Wait()

	t.Logf("HTTP Response Summary:")
	for code, count := range counts {
		if code == -1 {
			t.Logf("  Error: %d responses", count)
		} else {
			t.Logf("  HTTP Code %d: %d responses", code, count)
		}
	}

	// Validate test expectations
	if counts[http.StatusOK] != 1 {
		t.Errorf("❌ Mismatch: got %d, expected 1 for HTTP 200 (OK)", counts[http.StatusOK])
	}
	if counts[http.StatusConflict] != totalRequests-1 {
		t.Errorf("❌ Mismatch: got %d, expected %d for HTTP 409 (Conflict)", counts[http.StatusConflict], totalRequests-1)
	}
	if len(counts) > 2 {
		t.Errorf("❌ Unexpected status codes received:")
		for code := range counts {
			if code != http.StatusOK && code != http.StatusConflict && code != -1 {
				t.Errorf("  Unexpected status code: %d (%d times)", code, counts[code])
			}
		}
	}
}
