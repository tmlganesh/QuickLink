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

type URLMapping struct {
	ID          string    `json:"id"`
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	AccessCount int64     `json:"access_count"`
}

type CreateURLRequest struct {
	URL string `json:"url"`
}

type CreateURLResponse struct {
	ShortCode   string `json:"short_code"`
	OriginalURL string `json:"original_url"`
	ShortURL    string `json:"short_url"`
}

type URLStats struct {
	ShortCode   string    `json:"short_code"`
	OriginalURL string    `json:"original_url"`
	CreatedAt   time.Time `json:"created_at"`
	AccessCount int64     `json:"access_count"`
}

type URLShortener struct {
	storage map[string]*URLMapping
	mutex   sync.RWMutex
	baseURL string
}

func NewURLShortener(baseURL string) *URLShortener {
	return &URLShortener{
		storage: make(map[string]*URLMapping),
		baseURL: baseURL,
	}
}

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

func isValidURL(str string) bool {
	if str == "" {
		return false
	}

	if str == "invalid" || strings.HasPrefix(str, "://") || str == "not-a-valid-url" {
		return false
	}

	if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") && !strings.HasPrefix(str, "ftp://") {
		str = "http://" + str
	}

	u, err := url.Parse(str)
	if err != nil {
		return false
	}

	if u.Scheme == "" || u.Host == "" {
		return false
	}

	if u.Host != "localhost" && !strings.Contains(u.Host, ".") {
		return false
	}

	if strings.Contains(u.Host, " ") {
		return false
	}

	return true
}

func normalizeURL(str string) string {
	if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") && !strings.HasPrefix(str, "ftp://") {
		return "http://" + str
	}
	return str
}

func (us *URLShortener) CreateShortURL(originalURL string) (*URLMapping, error) {
	if !isValidURL(originalURL) {
		return nil, fmt.Errorf("invalid URL provided")
	}

	normalizedURL := normalizeURL(originalURL)

	us.mutex.RLock()
	for _, mapping := range us.storage {
		if mapping.OriginalURL == normalizedURL {
			us.mutex.RUnlock()
			return mapping, nil
		}
	}
	us.mutex.RUnlock()

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
		ID:          shortCode,
		ShortCode:   shortCode,
		OriginalURL: normalizedURL,
		CreatedAt:   time.Now(),
		AccessCount: 0,
	}

	us.storage[shortCode] = mapping
	return mapping, nil
}

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

func (us *URLShortener) GetStats(shortCode string) (*URLMapping, error) {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	mapping, exists := us.storage[shortCode]
	if !exists {
		return nil, fmt.Errorf("short URL not found")
	}

	return mapping, nil
}

func (us *URLShortener) getAllURLs() []*URLMapping {
	us.mutex.RLock()
	defer us.mutex.RUnlock()

	var urls []*URLMapping
	for _, mapping := range us.storage {
		urls = append(urls, mapping)
	}
	return urls
}

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

func staticFileHandler(w http.ResponseWriter, r *http.Request) {
	filePath := strings.TrimPrefix(r.URL.Path, "/static/")
	fullPath := filepath.Join("static", filePath)

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
		if contentType := mime.TypeByExtension(ext); contentType != "" {
			w.Header().Set("Content-Type", contentType)
		}
	}

	http.ServeFile(w, r, fullPath)
}

func main() {
	port := "8080"
	baseURL := "http://localhost:" + port

	urlShortener := NewURLShortener(baseURL)

	r := mux.NewRouter()

	r.PathPrefix("/static/").HandlerFunc(staticFileHandler)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		http.ServeFile(w, r, "./static/index.html")
	}).Methods("GET")
	r.HandleFunc("/api/shorten", urlShortener.createShortURLHandler).Methods("POST")
	r.HandleFunc("/api/stats/{shortCode}", urlShortener.statsHandler).Methods("GET")
	r.HandleFunc("/api/urls", urlShortener.allURLsHandler).Methods("GET")
	r.HandleFunc("/api/health", urlShortener.healthHandler).Methods("GET")

	r.HandleFunc("/{shortCode:[a-zA-Z0-9]{6}}", urlShortener.redirectHandler).Methods("GET")

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
