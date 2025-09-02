package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// Example client demonstrating how to use the URL shortener API
func runClient() {
	// Wait for server to start
	fmt.Println("🔍 Testing URL Shortener API...")
	time.Sleep(2 * time.Second)

	baseURL := "http://localhost:8080"

	// Test 1: Health check
	fmt.Println("\n1. Testing health endpoint...")
	resp, err := http.Get(baseURL + "/api/health")
	if err != nil {
		log.Printf("❌ Health check failed: %v", err)
	} else {
		fmt.Printf("✅ Health check passed: %s\n", resp.Status)
		resp.Body.Close()
	}

	// Test 2: Create short URL
	fmt.Println("\n2. Creating short URL...")
	// Note: In a real client, you would send a proper POST request with JSON body
	// This is just a demonstration of what the client would do
	fmt.Println("   POST /api/shorten")
	fmt.Println("   Body: {\"url\":\"https://www.google.com\"}")
	fmt.Println("   Expected: Returns short URL")

	// Test 3: Access statistics
	fmt.Println("\n3. Getting URL statistics...")
	fmt.Println("   GET /api/stats/{shortCode}")
	fmt.Println("   Expected: Returns access count and metadata")

	// Test 4: List all URLs
	fmt.Println("\n4. Getting all URLs...")
	resp, err = http.Get(baseURL + "/api/urls")
	if err != nil {
		log.Printf("❌ Failed to get all URLs: %v", err)
	} else {
		fmt.Printf("✅ All URLs endpoint accessible: %s\n", resp.Status)
		resp.Body.Close()
	}

	fmt.Println("\n🎯 To fully test the API, use curl commands from the README.md")
	fmt.Println("📖 Example:")
	fmt.Println("   curl -X POST http://localhost:8080/api/shorten \\")
	fmt.Println("        -H \"Content-Type: application/json\" \\")
	fmt.Println("        -d '{\"url\":\"https://www.google.com\"}'")
}
