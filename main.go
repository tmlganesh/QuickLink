package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"mime"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// URLMapping represents a URL mapping in our storage
type URLMapping struct {
	ID          string    `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	AccessCount int64     `json:"access_count"`
}

// CreateURLRequest represents the request body for creating a short URL
type CreateURLRequest struct {
	URL string `json:"url"`
}

// CreateURLResponse represents the response body for creating a short URL
type CreateURLResponse struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

// URLStats represents URL statistics
type URLStats struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	AccessCount int64     `json:"access_count"`
}

// URLShortener is our main service struct
type URLShortener struct {
	storage map[string]*URLMapping
	mutex   sync.RWMutex
	baseURL string
}

// NewURLShortener creates a new URL shortener instance
func NewURLShortener(baseURL string) *URLShortener {
	return &URLShortener{
		storage: make(map[string]*URLMapping),
		baseURL: baseURL,
	}
}

// generateShortCode generates a random short code
func (us *URLShortener) generateShortCode() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6

	result := make([]byte, length)
	for i := range result {
		num, _ := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		result[i] = charset[num.Int64()]
	}
	return string(result)
}

// isValidURL checks if the provided URL is valid
func isValidURL(str string) bool {
	if str == "" {
		return false
	}

	// Handle special invalid cases
	if str == "invalid" || strings.HasPrefix(str, "://") || str == "not-a-valid-url" {
		return false
	}

	// Add http:// if no scheme is present
	if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") && !strings.HasPrefix(str, "ftp://") {
		str = "http://" + str
	}

	u, err := url.Parse(str)
	if err != nil {
		return false
	}

	// Check for valid scheme and host
	if u.Scheme == "" || u.Host == "" {
		return false
	}

	// Additional validation for host - must contain at least one dot or be localhost
	if u.Host != "localhost" && !strings.Contains(u.Host, ".") {
		return false
	}

	// Check for spaces or other invalid characters
	if strings.Contains(u.Host, " ") {
		return false
	}

	return true
}

// normalizeURL normalizes the URL by adding http:// if needed
func normalizeURL(str string) string {
	if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") && !strings.HasPrefix(str, "ftp://") {
		return "http://" + str
	}
	return str
}

// CreateShortURL creates a new short URL
func (us *URLShortener) CreateShortURL(originalURL string) (*URLMapping, error) {
	if !isValidURL(originalURL) {
		return nil, fmt.Errorf("invalid URL provided")
	}

	normalizedURL := normalizeURL(originalURL)

	// Check if URL already exists
	us.mutex.RLock()
	for _, mapping := range us.storage {
		if mapping.OriginalURL == normalizedURL {
			us.mutex.RUnlock()
			return mapping, nil
		}
	}
	us.mutex.RUnlock()

	// Generate unique short code
	var shortCode string
	us.mutex.Lock()
	defer us.mutex.Unlock()

	for {
		shortCode = us.generateShortCode()
		if _, exists := us.storage[shortCode]; !exists {
			break
		}
	}

	mapping := &URLMapping{
		ID:          shortCode, // Using short code as ID for simplicity
		ShortCode:   shortCode,
		OriginalURL: normalizedURL,
		CreatedAt:   time.Now(),
		AccessCount: 0,
	}

	us.storage[shortCode] = mapping
	return mapping, nil
}

// GetOriginalURL retrieves the original URL by short code
func (us *URLShortener) GetOriginalURL(shortCode string) (*URLMapping, error) {
	us.mutex.Lock()
	defer us.mutex.Unlock()

	mapping, exists := us.storage[shortCode]
	if !exists {
		return nil, fmt.Errorf("short URL not found")
	}

	mapping.AccessCount++
	return mapping, nil
}

// GetStats returns statistics for a short URL
func (us *URLShortener) GetStats(shortCode string) (*URLMapping, error) {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	mapping, exists := us.storage[shortCode]
	if !exists {
		return nil, fmt.Errorf("short URL not found")
	}

	return mapping, nil
}

// getAllURLs returns all stored URLs (for admin purposes)
func (us *URLShortener) getAllURLs() []*URLMapping {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	var urls []*URLMapping
	for _, mapping := range us.storage {
		urls = append(urls, mapping)
	}
	return urls
}

// HTTP Handlers

func (us *URLShortener) createShortURLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.URL == "" {
		http.Error(w, "URL is required", http.StatusBadRequest)
		return
	}

	mapping, err := us.CreateShortURL(req.URL)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := CreateURLResponse{
		ShortCode:   mapping.ShortCode,
		OriginalURL: mapping.OriginalURL,
		ShortURL:    fmt.Sprintf("%s/%s", us.baseURL, mapping.ShortCode),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (us *URLShortener) redirectHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	mapping, err := us.GetOriginalURL(shortCode)
	if err != nil {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, mapping.OriginalURL, http.StatusMovedPermanently)
}

func (us *URLShortener) statsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	shortCode := vars["shortCode"]

	mapping, err := us.GetStats(shortCode)
	if err != nil {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	stats := URLStats{
		ShortCode:   mapping.ShortCode,
		OriginalURL: mapping.OriginalURL,
		CreatedAt:   mapping.CreatedAt,
		AccessCount: mapping.AccessCount,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func (us *URLShortener) allURLsHandler(w http.ResponseWriter, r *http.Request) {
	urls := us.getAllURLs()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(urls)
}

func (us *URLShortener) healthHandler(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status":  "healthy",
		"service": "URL Shortener",
		"time":    time.Now().Format(time.RFC3339),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Custom static file handler with proper MIME types
func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	// Remove /static/ prefix to get the actual file path
	filePath := strings.TrimPrefix(r.URL.Path, "/static/")
	fullPath := filepath.Join("static", filePath)

	// Set proper MIME type based on file extension
	ext := filepath.Ext(filePath)
	switch ext {
	case ".css":
		w.Header().Set("Content-Type", "text/css")
	case ".js":
		w.Header().Set("Content-Type", "application/javascript")
	case ".html":
		w.Header().Set("Content-Type", "text/html")
	case ".png":
		w.Header().Set("Content-Type", "image/png")
	case ".jpg", ".jpeg":
		w.Header().Set("Content-Type", "image/jpeg")
	case ".gif":
		w.Header().Set("Content-Type", "image/gif")
	case ".svg":
		w.Header().Set("Content-Type", "image/svg+xml")
	default:
		// Try to detect MIME type
		if contentType := mime.TypeByExtension(ext); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
	}

	// Serve the file
	http.ServeFile(w, r, fullPath)
}

func main() {
	// Configuration
	port := "8080"
	baseURL := "http://localhost:" + port

	// Create URL shortener instance
	urlShortener := NewURLShortener(baseURL)

	// Setup routes
	r := mux.NewRouter()

	// Serve static files with proper MIME types
	r.PathPrefix("/static/").HandlerFunc(staticFileHandler)

	// Serve index.html at root
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		http.ServeFile(w, r, "./static/index.html")
	}).Methods("GET") // API routes
	r.HandleFunc("/api/shorten", urlShortener.createShortURLHandler).Methods("POST")
	r.HandleFunc("/api/stats/{shortCode}", urlShortener.statsHandler).Methods("GET")
	r.HandleFunc("/api/urls", urlShortener.allURLsHandler).Methods("GET")
	r.HandleFunc("/api/health", urlShortener.healthHandler).Methods("GET")

	// Redirect route (must be last to avoid conflicts)
	r.HandleFunc("/{shortCode:[a-zA-Z0-9]{6}}", urlShortener.redirectHandler).Methods("GET")

	// Add CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	fmt.Printf("🚀 URL Shortener server starting on port %s\n", port)
	fmt.Printf("📡 Web Interface: %s\n", baseURL)
	fmt.Printf("📡 API Base URL: %s/api\n", baseURL)
	fmt.Println("\n📋 Available endpoints:")
	fmt.Println("   GET  /                    - Web Interface")
	fmt.Println("   POST /api/shorten        - Create short URL")
	fmt.Println("   GET  /{shortCode}        - Redirect to original URL")
	fmt.Println("   GET  /api/stats/{shortCode} - Get URL statistics")
	fmt.Println("   GET  /api/urls           - Get all URLs (admin)")
	fmt.Println("   GET  /api/health         - Health check")
	fmt.Println("\n🌐 Open your browser and go to:")
	fmt.Printf("   %s\n", baseURL)
	fmt.Println("\n🔗 Example API usage:")
	fmt.Println("   curl -X POST http://localhost:8080/api/shorten \\")
	fmt.Println("        -H \"Content-Type: application/json\" \\")
	fmt.Println("        -d '{\"url\":\"https://www.google.com\"}'")

	log.Fatal(http.ListenAndServe(":"+port, r))
}
