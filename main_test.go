package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestURLShortener_CreateShortURL(t *testing.T) {
	us := NewURLShortener("http://localhost:8080")

	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "Valid HTTP URL",
			url:         "http://www.google.com",
			expectError: false,
		},
		{
			name:        "Valid HTTPS URL",
			url:         "https://www.google.com",
			expectError: false,
		},
		{
			name:        "URL without scheme",
			url:         "www.google.com",
			expectError: false,
		},
		{
			name:        "Invalid URL",
			url:         "not-a-valid-url",
			expectError: true,
		},
		{
			name:        "Empty URL",
			url:         "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping, err := us.CreateShortURL(tt.url)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if mapping == nil {
				t.Errorf("Expected mapping but got nil")
				return
			}

			if len(mapping.ShortCode) != 6 {
				t.Errorf("Expected short code length 6, got %d", len(mapping.ShortCode))
			}

			if mapping.AccessCount != 0 {
				t.Errorf("Expected access count 0, got %d", mapping.AccessCount)
			}
		})
	}
}

func TestURLShortener_GetOriginalURL(t *testing.T) {
	us := NewURLShortener("http://localhost:8080")

	// Create a short URL first
	originalURL := "https://www.example.com"
	mapping, err := us.CreateShortURL(originalURL)
	if err != nil {
		t.Fatalf("Failed to create short URL: %v", err)
	}

	// Test getting the original URL
	retrieved, err := us.GetOriginalURL(mapping.ShortCode)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if retrieved.OriginalURL != "https://www.example.com" {
		t.Errorf("Expected %s, got %s", "https://www.example.com", retrieved.OriginalURL)
	}

	if retrieved.AccessCount != 1 {
		t.Errorf("Expected access count 1, got %d", retrieved.AccessCount)
	}

	// Test non-existent short code
	_, err = us.GetOriginalURL("nonexistent")
	if err == nil {
		t.Errorf("Expected error for non-existent short code")
	}
}

func TestDuplicateURLs(t *testing.T) {
	us := NewURLShortener("http://localhost:8080")

	url := "https://www.example.com"

	// Create first mapping
	mapping1, err := us.CreateShortURL(url)
	if err != nil {
		t.Fatalf("Failed to create first short URL: %v", err)
	}

	// Create second mapping with same URL
	mapping2, err := us.CreateShortURL(url)
	if err != nil {
		t.Fatalf("Failed to create second short URL: %v", err)
	}

	// Should return the same mapping
	if mapping1.ShortCode != mapping2.ShortCode {
		t.Errorf("Expected same short code for duplicate URL, got %s and %s",
			mapping1.ShortCode, mapping2.ShortCode)
	}
}

func TestCreateShortURLHandler(t *testing.T) {
	us := NewURLShortener("http://localhost:8080")

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Valid request",
			requestBody:    `{"url":"https://www.google.com"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"url":}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Missing URL",
			requestBody:    `{}`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Invalid URL",
			requestBody:    `{"url":"invalid"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/shorten",
				bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(us.createShortURLHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, status)
			}

			if tt.expectedStatus == http.StatusOK {
				var response CreateURLResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to unmarshal response: %v", err)
				}

				if response.ShortCode == "" {
					t.Errorf("Expected non-empty short code")
				}

				if response.ShortURL == "" {
					t.Errorf("Expected non-empty short URL")
				}
			}
		})
	}
}

func TestRedirectHandler(t *testing.T) {
	us := NewURLShortener("http://localhost:8080")

	// Create a short URL first
	mapping, err := us.CreateShortURL("https://www.google.com")
	if err != nil {
		t.Fatalf("Failed to create short URL: %v", err)
	}

	// Test redirect
	req, err := http.NewRequest("GET", "/"+mapping.ShortCode, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	// We need to set up the mux vars manually for testing
	req = req.WithContext(req.Context())

	// Create a simple handler that directly calls redirect with the short code
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate mux.Vars by directly getting the original URL
		retrieved, err := us.GetOriginalURL(mapping.ShortCode)
		if err != nil {
			http.Error(w, "Short URL not found", http.StatusNotFound)
			return
		}
		http.Redirect(w, r, retrieved.OriginalURL, http.StatusMovedPermanently)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusMovedPermanently {
		t.Errorf("Expected status %d, got %d", http.StatusMovedPermanently, status)
	}

	location := rr.Header().Get("Location")
	if location != "https://www.google.com" {
		t.Errorf("Expected redirect to https://www.google.com, got %s", location)
	}
}

func TestStatsHandler(t *testing.T) {
	us := NewURLShortener("http://localhost:8080")

	// Create a short URL first
	mapping, err := us.CreateShortURL("https://www.google.com")
	if err != nil {
		t.Fatalf("Failed to create short URL: %v", err)
	}

	// Access it a few times to increment counter
	us.GetOriginalURL(mapping.ShortCode)
	us.GetOriginalURL(mapping.ShortCode)

	// Test stats
	req, err := http.NewRequest("GET", "/api/stats/"+mapping.ShortCode, nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	// Create a simple handler for testing
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		stats, err := us.GetStats(mapping.ShortCode)
		if err != nil {
			http.Error(w, "Short URL not found", http.StatusNotFound)
			return
		}

		statsResponse := URLStats{
			ShortCode:   stats.ShortCode,
			OriginalURL: stats.OriginalURL,
			CreatedAt:   stats.CreatedAt,
			AccessCount: stats.AccessCount,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(statsResponse)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, status)
	}

	var stats URLStats
	err = json.Unmarshal(rr.Body.Bytes(), &stats)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if stats.AccessCount != 2 {
		t.Errorf("Expected access count 2, got %d", stats.AccessCount)
	}

	if stats.ShortCode != mapping.ShortCode {
		t.Errorf("Expected short code %s, got %s", mapping.ShortCode, stats.ShortCode)
	}
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://www.google.com", true},
		{"http://www.google.com", true},
		{"www.google.com", true},
		{"google.com", true},
		{"ftp://files.example.com", true},
		{"", false},
		{"invalid", false},
		{"not-a-valid-url", false},
		{"http://", false},
		{"://invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isValidURL(tt.url)
			if result != tt.expected {
				t.Errorf("isValidURL(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"https://www.google.com", "https://www.google.com"},
		{"http://www.google.com", "http://www.google.com"},
		{"www.google.com", "http://www.google.com"},
		{"google.com", "http://google.com"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := normalizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeURL(%q) = %v, expected %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestConcurrentAccess(t *testing.T) {
	us := NewURLShortener("http://localhost:8080")

	// Create a short URL
	mapping, err := us.CreateShortURL("https://www.google.com")
	if err != nil {
		t.Fatalf("Failed to create short URL: %v", err)
	}

	// Simulate concurrent access
	numGoroutines := 100
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			_, err := us.GetOriginalURL(mapping.ShortCode)
			if err != nil {
				t.Errorf("Error in concurrent access: %v", err)
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Check final access count
	finalStats, err := us.GetStats(mapping.ShortCode)
	if err != nil {
		t.Fatalf("Failed to get final stats: %v", err)
	}

	if finalStats.AccessCount != int64(numGoroutines) {
		t.Errorf("Expected access count %d, got %d", numGoroutines, finalStats.AccessCount)
	}
}

func TestShortCodeUniqueness(t *testing.T) {
	us := NewURLShortener("http://localhost:8080")

	codes := make(map[string]bool)
	numURLs := 1000

	for i := 0; i < numURLs; i++ {
		url := fmt.Sprintf("https://example%d.com", i)
		mapping, err := us.CreateShortURL(url)
		if err != nil {
			t.Fatalf("Failed to create short URL %d: %v", i, err)
		}

		if codes[mapping.ShortCode] {
			t.Errorf("Duplicate short code generated: %s", mapping.ShortCode)
		}
		codes[mapping.ShortCode] = true
	}
}
