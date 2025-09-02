//go:build demo
// +build demo

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// DemoClient represents a client for testing the URL shortener
type DemoClient struct {
	baseURL string
	client  *http.Client
}

// CreateURLRequest represents the request to create a short URL
type CreateURLRequestDemo struct {
	URL string `json:"url"`
}

// CreateURLResponse represents the response from creating a short URL
type CreateURLResponseDemo struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

// URLStatsDemo represents URL statistics
type URLStatsDemo struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	AccessCount int64     `json:"access_count"`
}

func NewDemoClient(baseURL string) *DemoClient {
	return &DemoClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *DemoClient) CreateShortURL(originalURL string) (*CreateURLResponseDemo, error) {
	reqBody := CreateURLRequestDemo{URL: originalURL}
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Post(c.baseURL+"/api/shorten", "application/json", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var response CreateURLResponseDemo
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (c *DemoClient) GetStats(shortCode string) (*URLStatsDemo, error) {
	resp, err := c.client.Get(c.baseURL + "/api/stats/" + shortCode)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(body))
	}

	var stats URLStatsDemo
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return &stats, nil
}

func (c *DemoClient) AccessShortURL(shortCode string) error {
	// Use a client that doesn't follow redirects to see the redirect response
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(c.baseURL + "/" + shortCode)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMovedPermanently {
		return fmt.Errorf("expected 301 redirect, got %s", resp.Status)
	}

	return nil
}

func (c *DemoClient) HealthCheck() error {
	resp, err := c.client.Get(c.baseURL + "/api/health")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check failed: %s", resp.Status)
	}

	return nil
}

func main() {
	fmt.Println("🧪 URL Shortener Demo Client")
	fmt.Println("=====================================")

	client := NewDemoClient("http://localhost:8080")

	// Wait a moment for the server to start
	fmt.Println("⏳ Waiting for server to start...")
	time.Sleep(3 * time.Second)

	// Test 1: Health Check
	fmt.Println("\n1. 🏥 Health Check")
	if err := client.HealthCheck(); err != nil {
		log.Printf("❌ Health check failed: %v", err)
		log.Println("💡 Make sure the server is running: go run main.go")
		return
	}
	fmt.Println("✅ Server is healthy!")

	// Test 2: Create Short URLs
	fmt.Println("\n2. 🔗 Creating Short URLs")
	testURLs := []string{
		"https://www.google.com",
		"https://www.github.com",
		"https://www.stackoverflow.com",
		"youtube.com", // Test URL normalization
	}

	var createdCodes []string

	for i, url := range testURLs {
		fmt.Printf("   Creating short URL for: %s\n", url)

		response, err := client.CreateShortURL(url)
		if err != nil {
			log.Printf("❌ Failed to create short URL: %v", err)
			continue
		}

		fmt.Printf("   ✅ Created: %s -> %s\n", response.ShortURL, response.OriginalURL)
		createdCodes = append(createdCodes, response.ShortCode)

		if i < len(testURLs)-1 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	if len(createdCodes) == 0 {
		log.Println("❌ No short URLs were created successfully")
		return
	}

	// Test 3: Access Short URLs
	fmt.Println("\n3. 🌐 Testing URL Redirects")
	for i, code := range createdCodes {
		fmt.Printf("   Accessing short URL: /%s\n", code)

		if err := client.AccessShortURL(code); err != nil {
			log.Printf("❌ Failed to access short URL: %v", err)
			continue
		}

		fmt.Printf("   ✅ Redirect successful for /%s\n", code)

		if i < len(createdCodes)-1 {
			time.Sleep(300 * time.Millisecond)
		}
	}

	// Test 4: Get Statistics
	fmt.Println("\n4. 📊 Getting URL Statistics")
	for _, code := range createdCodes {
		stats, err := client.GetStats(code)
		if err != nil {
			log.Printf("❌ Failed to get stats for %s: %v", code, err)
			continue
		}

		fmt.Printf("   📈 %s:\n", code)
		fmt.Printf("      Original URL: %s\n", stats.OriginalURL)
		fmt.Printf("      Created: %s\n", stats.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("      Access Count: %d\n", stats.AccessCount)
		fmt.Println()
	}

	// Test 5: Test Duplicate URL
	fmt.Println("5. 🔄 Testing Duplicate URL Handling")
	fmt.Println("   Creating short URL for duplicate: https://www.google.com")

	duplicateResponse, err := client.CreateShortURL("https://www.google.com")
	if err != nil {
		log.Printf("❌ Failed to test duplicate URL: %v", err)
	} else {
		fmt.Printf("   ✅ Duplicate handling: %s (should be same as first Google URL)\n",
			duplicateResponse.ShortCode)
	}

	fmt.Println("\n🎉 Demo completed successfully!")
	fmt.Println("\n💡 Manual Testing Commands:")
	fmt.Println("   # Create short URL")
	fmt.Println("   curl -X POST http://localhost:8080/api/shorten \\")
	fmt.Println("        -H \"Content-Type: application/json\" \\")
	fmt.Println("        -d '{\"url\":\"https://example.com\"}'")
	fmt.Println()
	fmt.Println("   # Get all URLs")
	fmt.Println("   curl http://localhost:8080/api/urls")
	fmt.Println()
	fmt.Println("   # Access short URL in browser")
	fmt.Printf("   http://localhost:8080/%s\n", createdCodes[0])
}
