package metron

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
)

// newMockServer creates a mock HTTP server that simulates the ComicVine API for testing purposes.
func newMockServer() *httptest.Server {
	mux := http.NewServeMux()

	// Mock for Issue Search (for SearchByUPC())
	mux.HandleFunc("/api/issue/", func(w http.ResponseWriter, r *http.Request) {
		// Basic Auth prüfen
		username, password, ok := r.BasicAuth()
		if !ok || username != "testuser" || password != "testpass" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"detail": "Authentication credentials were not provided.",
			})
			return
		}

		// UPC query parameter
		upc := r.URL.Query().Get("upc")
		if upc != "" {
			handleUPCSearch(w, upc)
			return
		}

		// Issue-Detail after ID
		if strings.HasSuffix(r.URL.Path, "/") && r.URL.Path != "/api/issue/" {
			handleIssueDetail(w, r.URL.Path)
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	})

	// Mock for Publisher (for Validate())
	mux.HandleFunc("/api/publisher/", func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "testuser" || password != "testpass" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"count":   1,
			"results": []map[string]interface{}{},
		})
	})

	return httptest.NewServer(mux)
}

func handleUPCSearch(w http.ResponseWriter, upc string) {
	w.Header().Set("Content-Type", "application/json")

	switch upc {
	case "75960609558200111":
		// Successful search with one result
		json.NewEncoder(w).Encode(IssueSearchResponse{
			Count: 1,
			Results: []Issue{
				{
					ID:     162239,
					Number: "1",
					Series: Series{
						ID:     1234,
						Name:   "Amazing Spider-Man",
						Volume: 1,
					},
					CoverDate: "2024-03-01",
					StoreDate: "2024-01-10",
				},
			},
		})
	case "00000000000000000":
		// No results found
		json.NewEncoder(w).Encode(IssueSearchResponse{
			Count:   0,
			Results: []Issue{},
		})
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func handleIssueDetail(w http.ResponseWriter, path string) {
	w.Header().Set("Content-Type", "application/json")

	// Extract ID from /api/issue/162239/
	if strings.HasSuffix(path, "162239/") {
		json.NewEncoder(w).Encode(Issue{
			ID:          162239,
			Number:      "1",
			Description: "The amazing Spider-Man returns!",
			Series: Series{
				ID:     1234,
				Name:   "Amazing Spider-Man",
				Volume: 1,
			},
			CoverDate: "2024-03-01",
			StoreDate: "2024-01-10",
			PageCount: 32,
			Price:     "3.99",
			UPC:       "75960609558200111",
			Credits: []Credit{
				{
					Creator: Creator{ID: 1, Name: "Zeb Wells"},
					Role:    []Role{{ID: 1, Name: "Writer"}},
				},
				{
					Creator: Creator{ID: 2, Name: "John Romita Jr."},
					Role:    []Role{{ID: 2, Name: "Penciller"}},
				},
			},
			Characters: []Character{
				{ID: 1, Name: "Spider-Man (Peter Parker)"},
				{ID: 2, Name: "Mary Jane Watson"},
			},
			Teams: []Team{
				{ID: 1, Name: "Avengers"},
			},
			Arcs: []Arc{
				{ID: 1, Name: "Gang War"},
			},
		})
		return
	}

	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]string{
		"detail": "Not found.",
	})
}
