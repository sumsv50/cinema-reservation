package reservation_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"
)

func TestLoopingSeatRequestPattern(t *testing.T) {
	const (
		totalRequests = 10000
		rows          = 10
		columns       = 15
		targetURL     = "http://localhost:8080/api/v1/reservations"
		cinemaSlug    = "grand-cinema-downtown"
	)

	type SeatRequest struct {
		Row    int `json:"row"`
		Column int `json:"column"`
	}

	type ReservationRequest struct {
		CinemaSlug string        `json:"cinema_slug"`
		Seats      []SeatRequest `json:"seats"`
	}

	client := &http.Client{}
	var wg sync.WaitGroup
	var mu sync.Mutex

	counts := make(map[int]int)
	expected := map[int]int{
		200: 0,
		409: 0,
		400: 0,
		-1:  0,
	}

	// Track reservation counts for each seat (row-column)
	seatMap := make(map[string]bool)

	sent := 0
	fire := make(chan struct{})
	for sent < totalRequests {
		for i := 0; i < rows && sent < totalRequests; i++ {
			for j := 0; j <= columns && sent < totalRequests; j++ {
				row, col := i, j
				key := fmt.Sprintf("%d-%d", row, col)

				// --- Update expected counts
				if col >= columns {
					expected[400]++
				} else {
					if !seatMap[key] {
						expected[200]++
						seatMap[key] = true
					} else {
						expected[409]++
					}
				}

				// --- Send request
				req := ReservationRequest{
					CinemaSlug: cinemaSlug,
					Seats:      []SeatRequest{{Row: row, Column: col}},
				}
				wg.Add(1)
				go func(r ReservationRequest) {
					defer wg.Done()
					body, _ := json.Marshal(r)
					<-fire
					resp, err := client.Post(targetURL, "application/json", bytes.NewBuffer(body))
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
				}(req)

				sent++
			}
		}
	}

	close(fire)
	wg.Wait()

	// --- Report actual vs. expected
	t.Logf("HTTP Response Summary:")
	for code, count := range counts {
		if code == -1 {
			t.Logf("  Error: %d responses", count)
		} else {
			t.Logf("  HTTP Code %d: %d responses", code, count)
		}
	}
	for code, exp := range expected {
		if counts[code] != exp {
			t.Errorf("❌ Mismatch: got %d, expected %d for HTTP %d", counts[code], exp, code)
		}
	}
}
