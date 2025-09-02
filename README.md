# URL Shortener with Web Interface

A complete, production-ready URL shortener service built with Go backend and modern HTML/CSS/JavaScript frontend.

## 🌟 Features

- ✅ **Beautiful Web Interface** - Modern, responsive design
- ✅ **Instant URL Shortening** - Create short URLs from long URLs
- ✅ **Click Tracking** - Monitor access statistics
- ✅ **Recent URLs History** - View and manage your shortened URLs
- ✅ **One-Click Copying** - Easy clipboard integration
- ✅ **Mobile Friendly** - Works perfectly on all devices
- ✅ **RESTful API** - Full API access for developers
- ✅ **Real-time Updates** - Live statistics and feedback

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher

### Installation & Usage

1. **Install Dependencies:**
   ```bash
   go mod tidy
   ```

2. **Start the Server:**
   ```bash
   go run main.go
   ```

3. **Open Your Browser:**
   Visit `http://localhost:8080` to access the web interface

4. **Start Shortening URLs:**
   - Enter any URL (e.g., `https://www.google.com` or just `google.com`)
   - Click "Shorten URL"
   - Copy and share your short link!

## 🌐 Web Interface Features

### Main Interface
- **Clean, modern design** with gradient background
- **Smart URL validation** with real-time feedback
- **Loading animations** and smooth transitions
- **Error handling** with helpful messages

### URL Management
- **Instant shortening** with visual feedback
- **One-click copying** to clipboard
- **Visit short URLs** directly from the interface
- **View detailed statistics** for each URL

### Recent URLs
- **Local storage** keeps track of your URLs
- **Quick access** to previously shortened URLs
- **Batch operations** - copy and view stats
- **Automatic cleanup** (keeps last 10 URLs)

### Statistics Modal
- **Real-time click tracking**
- **Creation timestamps**
- **Access counts**
- **Detailed URL information**

## 📱 Mobile Experience

The interface is fully responsive and optimized for:
- **Smartphones** - Touch-friendly buttons and inputs
- **Tablets** - Optimized layout and spacing
- **Desktop** - Full-featured experience

## 🔧 Easy Run Scripts

- **Windows Users:**
  - `start-server.bat` - Start the server
  - `run-tests.bat` - Run all tests
  - `run-demo.bat` - Run demo client

- **Cross-platform:**
  ```bash
  make run    # Start server
  make test   # Run tests
  make build  # Build binary
  ```

## API Endpoints

### 1. Create Short URL
```bash
POST /api/shorten
Content-Type: application/json

{
  "url": "https://www.example.com"
}
```

**Response:**
```json
{
  "short_code": "abc123",
  "original_url": "https://www.example.com",
  "short_url": "http://localhost:8080/abc123"
}
```

### 2. Redirect to Original URL
```bash
GET /{shortCode}
```

This will redirect (301) to the original URL and increment the access counter.

### 3. Get URL Statistics
```bash
GET /api/stats/{shortCode}
```

**Response:**
```json
{
  "short_code": "abc123",
  "original_url": "https://www.example.com",
  "created_at": "2025-09-02T10:30:00Z",
  "access_count": 5
}
```

### 4. Get All URLs (Admin)
```bash
GET /api/urls
```

Returns an array of all stored URL mappings.

### 5. Health Check
```bash
GET /api/health
```

**Response:**
```json
{
  "status": "healthy",
  "service": "URL Shortener",
  "time": "2025-09-02T10:30:00Z"
}
```

## Usage Examples

### Using curl

1. **Create a short URL:**
   ```bash
   curl -X POST http://localhost:8080/api/shorten \
        -H "Content-Type: application/json" \
        -d '{"url":"https://www.google.com"}'
   ```

2. **Get statistics:**
   ```bash
   curl http://localhost:8080/api/stats/abc123
   ```

3. **Access short URL:**
   ```bash
   curl -L http://localhost:8080/abc123
   ```

### Using PowerShell (Windows)

1. **Create a short URL:**
   ```powershell
   $body = @{url="https://www.google.com"} | ConvertTo-Json
   Invoke-RestMethod -Uri "http://localhost:8080/api/shorten" -Method Post -Body $body -ContentType "application/json"
   ```

2. **Get statistics:**
   ```powershell
   Invoke-RestMethod -Uri "http://localhost:8080/api/stats/abc123"
   ```

## Features Explained

### URL Validation
- Validates URL format before creating short URLs
- Automatically adds `http://` if no scheme is provided
- Prevents creation of invalid URLs

### Duplicate Detection
- Checks for existing URLs before creating new short codes
- Returns existing short code if URL already exists

### Thread Safety
- Uses mutex locks for concurrent access
- Safe for high-traffic scenarios

### Short Code Generation
- Uses cryptographically secure random generation
- Base62 encoding (a-z, A-Z, 0-9)
- 6-character codes provide 56+ billion combinations

### Access Tracking
- Tracks how many times each short URL is accessed
- Provides detailed statistics via API

## Project Structure

```
url-shortener/
├── main.go      # Main application file
├── go.mod       # Go module definition
└── README.md    # This file
```

## Configuration

The application uses the following default configuration:
- **Port:** 8080
- **Short code length:** 6 characters
- **Storage:** In-memory (can be extended to use databases)

To modify these settings, edit the constants in `main.go`.

## Extending the Application

### Adding Database Support

To add persistent storage, replace the in-memory map with a database:

1. Add database dependencies to `go.mod`
2. Create database connection in `main()`
3. Modify `URLShortener` struct to use database client
4. Update CRUD operations to use database queries

### Adding Custom Short Codes

To allow users to specify custom short codes:

1. Add `custom_code` field to `CreateURLRequest`
2. Modify `CreateShortURL` to check for custom code availability
3. Add validation for custom code format

### Adding Expiration

To add URL expiration:

1. Add `expires_at` field to `URLMapping`
2. Add expiration parameter to create request
3. Check expiration in redirect handler

## Production Considerations

1. **Database:** Replace in-memory storage with Redis/PostgreSQL
2. **Caching:** Add Redis for frequently accessed URLs
3. **Rate Limiting:** Implement rate limiting for API endpoints
4. **Authentication:** Add API key authentication for admin endpoints
5. **Monitoring:** Add logging and metrics collection
6. **HTTPS:** Use TLS certificates in production
7. **Load Balancing:** Deploy multiple instances behind a load balancer

## License

This project is open source and available under the MIT License.
