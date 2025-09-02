// Global variables
let currentShortCode = '';
let recentUrls = JSON.parse(localStorage.getItem('recentUrls') || '[]');

// DOM elements
const urlForm = document.getElementById('urlForm');
const urlInput = document.getElementById('urlInput');
const shortenBtn = document.getElementById('shortenBtn');
const loading = document.getElementById('loading');
const result = document.getElementById('result');
const error = document.getElementById('error');
const originalUrlSpan = document.getElementById('originalUrl');
const shortUrlSpan = document.getElementById('shortUrl');
const copyBtn = document.getElementById('copyBtn');
const visitBtn = document.getElementById('visitBtn');
const statsBtn = document.getElementById('statsBtn');
const newUrlBtn = document.getElementById('newUrlBtn');
const retryBtn = document.getElementById('retryBtn');
const toast = document.getElementById('toast');
const toastMessage = document.getElementById('toastMessage');

// Modal elements
const statsModal = document.getElementById('statsModal');
const closeModal = document.getElementById('closeModal');
const statShortCode = document.getElementById('statShortCode');
const statOriginalUrl = document.getElementById('statOriginalUrl');
const statCreatedAt = document.getElementById('statCreatedAt');
const statAccessCount = document.getElementById('statAccessCount');

// Recent URLs elements
const recentUrlsContainer = document.getElementById('recentUrls');

// API Configuration
const API_BASE = window.location.origin;

// Initialize the app
document.addEventListener('DOMContentLoaded', function() {
    renderRecentUrls();
    setupEventListeners();
    
    // Auto-focus the input
    urlInput.focus();
});

// Setup all event listeners
function setupEventListeners() {
    // Form submission
    urlForm.addEventListener('submit', handleFormSubmit);
    
    // Button clicks
    copyBtn.addEventListener('click', copyToClipboard);
    visitBtn.addEventListener('click', visitShortUrl);
    statsBtn.addEventListener('click', showStats);
    newUrlBtn.addEventListener('click', resetForm);
    retryBtn.addEventListener('click', resetForm);
    
    // Modal events
    closeModal.addEventListener('click', hideStatsModal);
    statsModal.addEventListener('click', function(e) {
        if (e.target === statsModal) {
            hideStatsModal();
        }
    });
    
    // Keyboard shortcuts
    document.addEventListener('keydown', function(e) {
        if (e.key === 'Escape') {
            hideStatsModal();
        }
        if (e.ctrlKey && e.key === 'Enter') {
            if (!result.classList.contains('hidden')) {
                copyToClipboard();
            }
        }
    });
    
    // Input validation
    urlInput.addEventListener('input', validateInput);
    urlInput.addEventListener('paste', function(e) {
        setTimeout(validateInput, 100);
    });
}

// Handle form submission
async function handleFormSubmit(e) {
    e.preventDefault();
    
    const url = urlInput.value.trim();
    if (!url) {
        showError('Please enter a URL');
        return;
    }
    
    if (!isValidUrl(url)) {
        showError('Please enter a valid URL (e.g., https://www.example.com)');
        return;
    }
    
    await shortenUrl(url);
}

// Validate URL input
function isValidUrl(string) {
    try {
        // Add protocol if missing
        let testUrl = string;
        if (!testUrl.match(/^https?:\/\//)) {
            testUrl = 'http://' + testUrl;
        }
        
        new URL(testUrl);
        
        // Additional validation for domain
        const domain = testUrl.replace(/^https?:\/\//, '').split('/')[0];
        return domain.includes('.') || domain === 'localhost';
    } catch (_) {
        return false;
    }
}

// Validate input and update UI
function validateInput() {
    const url = urlInput.value.trim();
    const isValid = url === '' || isValidUrl(url);
    
    if (url && !isValid) {
        urlInput.style.borderColor = '#ef4444';
        urlInput.style.boxShadow = '0 0 0 3px rgba(239, 68, 68, 0.1)';
    } else {
        urlInput.style.borderColor = '#e1e5e9';
        urlInput.style.boxShadow = 'none';
    }
}

// Shorten URL via API
async function shortenUrl(url) {
    showLoading();
    
    try {
        const response = await fetch(`${API_BASE}/api/shorten`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ url: url })
        });
        
        if (!response.ok) {
            const errorData = await response.text();
            throw new Error(errorData || `HTTP ${response.status}`);
        }
        
        const data = await response.json();
        currentShortCode = data.short_code;
        
        showResult(data);
        addToRecentUrls(data);
        
    } catch (err) {
        console.error('Error shortening URL:', err);
        showError(`Failed to shorten URL: ${err.message}`);
    }
}

// Show loading state
function showLoading() {
    hideAllStates();
    loading.classList.remove('hidden');
    shortenBtn.disabled = true;
    shortenBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i><span>Creating...</span>';
}

// Show result state
function showResult(data) {
    hideAllStates();
    
    originalUrlSpan.textContent = data.original_url;
    shortUrlSpan.textContent = data.short_url;
    
    // Make short URL clickable
    shortUrlSpan.onclick = () => window.open(data.short_url, '_blank');
    shortUrlSpan.style.cursor = 'pointer';
    
    result.classList.remove('hidden');
    result.classList.add('fade-in');
    
    // Focus on copy button for accessibility
    setTimeout(() => copyBtn.focus(), 100);
}

// Show error state
function showError(message) {
    hideAllStates();
    
    const errorMessage = document.getElementById('errorMessage');
    errorMessage.textContent = message;
    
    error.classList.remove('hidden');
    error.classList.add('fade-in');
    
    // Focus on retry button
    setTimeout(() => retryBtn.focus(), 100);
}

// Hide all states
function hideAllStates() {
    loading.classList.add('hidden');
    result.classList.add('hidden');
    error.classList.add('hidden');
    
    shortenBtn.disabled = false;
    shortenBtn.innerHTML = '<i class="fas fa-magic"></i><span>Shorten URL</span>';
}

// Reset form to initial state
function resetForm() {
    hideAllStates();
    urlInput.value = '';
    urlInput.focus();
    currentShortCode = '';
    
    // Reset input styling
    urlInput.style.borderColor = '#e1e5e9';
    urlInput.style.boxShadow = 'none';
}

// Copy short URL to clipboard
async function copyToClipboard() {
    const shortUrl = shortUrlSpan.textContent;
    
    try {
        await navigator.clipboard.writeText(shortUrl);
        showToast('URL copied to clipboard!');
        
        // Visual feedback
        copyBtn.innerHTML = '<i class="fas fa-check"></i>';
        copyBtn.style.background = '#10b981';
        
        setTimeout(() => {
            copyBtn.innerHTML = '<i class="fas fa-copy"></i>';
            copyBtn.style.background = '#667eea';
        }, 2000);
        
    } catch (err) {
        console.error('Failed to copy:', err);
        
        // Fallback for older browsers
        const textArea = document.createElement('textarea');
        textArea.value = shortUrl;
        document.body.appendChild(textArea);
        textArea.focus();
        textArea.select();
        
        try {
            document.execCommand('copy');
            showToast('URL copied to clipboard!');
        } catch (fallbackErr) {
            showToast('Failed to copy URL', 'error');
        }
        
        document.body.removeChild(textArea);
    }
}

// Visit short URL in new tab
function visitShortUrl() {
    const shortUrl = shortUrlSpan.textContent;
    window.open(shortUrl, '_blank');
}

// Show URL statistics
async function showStats() {
    if (!currentShortCode) return;
    
    try {
        const response = await fetch(`${API_BASE}/api/stats/${currentShortCode}`);
        
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}`);
        }
        
        const stats = await response.json();
        
        statShortCode.textContent = stats.short_code;
        statOriginalUrl.textContent = stats.original_url;
        statCreatedAt.textContent = formatDate(stats.created_at);
        statAccessCount.textContent = stats.access_count;
        
        showStatsModal();
        
    } catch (err) {
        console.error('Error fetching stats:', err);
        showToast('Failed to load statistics', 'error');
    }
}

// Show stats modal
function showStatsModal() {
    statsModal.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
    
    // Focus on close button for accessibility
    setTimeout(() => closeModal.focus(), 100);
}

// Hide stats modal
function hideStatsModal() {
    statsModal.classList.add('hidden');
    document.body.style.overflow = 'auto';
}

// Add URL to recent URLs
function addToRecentUrls(data) {
    const urlData = {
        shortCode: data.short_code,
        shortUrl: data.short_url,
        originalUrl: data.original_url,
        createdAt: new Date().toISOString()
    };
    
    // Remove duplicate if exists
    recentUrls = recentUrls.filter(url => url.shortCode !== data.short_code);
    
    // Add to beginning
    recentUrls.unshift(urlData);
    
    // Keep only last 10
    recentUrls = recentUrls.slice(0, 10);
    
    // Save to localStorage
    localStorage.setItem('recentUrls', JSON.stringify(recentUrls));
    
    // Re-render
    renderRecentUrls();
}

// Render recent URLs
function renderRecentUrls() {
    if (recentUrls.length === 0) {
        recentUrlsContainer.innerHTML = '<p class="no-urls">No URLs shortened yet. Create your first short URL above!</p>';
        return;
    }
    
    const urlsHtml = recentUrls.map(url => `
        <div class="recent-url-item">
            <div class="recent-url-header">
                <div class="recent-url-info">
                    <a href="${url.shortUrl}" target="_blank" class="recent-short-url">${url.shortUrl}</a>
                    <div class="recent-original-url">${url.originalUrl}</div>
                </div>
                <div class="recent-url-actions">
                    <button onclick="copyUrl('${url.shortUrl}')" title="Copy URL">
                        <i class="fas fa-copy"></i>
                    </button>
                    <button onclick="showUrlStats('${url.shortCode}')" title="View Stats">
                        <i class="fas fa-chart-bar"></i>
                    </button>
                </div>
            </div>
            <div class="recent-url-meta">
                <span>Created: ${formatDate(url.createdAt)}</span>
                <span>Short Code: ${url.shortCode}</span>
            </div>
        </div>
    `).join('');
    
    recentUrlsContainer.innerHTML = urlsHtml;
}

// Copy URL from recent URLs
async function copyUrl(url) {
    try {
        await navigator.clipboard.writeText(url);
        showToast('URL copied to clipboard!');
    } catch (err) {
        showToast('Failed to copy URL', 'error');
    }
}

// Show stats for a specific URL
async function showUrlStats(shortCode) {
    currentShortCode = shortCode;
    await showStats();
}

// Format date for display
function formatDate(dateString) {
    const date = new Date(dateString);
    const now = new Date();
    const diffTime = Math.abs(now - date);
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));
    
    if (diffDays === 1) {
        return 'Today';
    } else if (diffDays === 2) {
        return 'Yesterday';
    } else if (diffDays <= 7) {
        return `${diffDays - 1} days ago`;
    } else {
        return date.toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        });
    }
}

// Show toast notification
function showToast(message, type = 'success') {
    toastMessage.textContent = message;
    
    // Set toast style based on type
    if (type === 'error') {
        toast.style.background = '#ef4444';
        toast.querySelector('i').className = 'fas fa-exclamation-circle';
    } else {
        toast.style.background = '#10b981';
        toast.querySelector('i').className = 'fas fa-check-circle';
    }
    
    toast.classList.remove('hidden');
    
    // Auto-hide after 3 seconds
    setTimeout(() => {
        toast.classList.add('hidden');
    }, 3000);
}

// Features modal
function showFeatures() {
    const features = `
        <div class="modal" id="featuresModal">
            <div class="modal-content">
                <div class="modal-header">
                    <h3><i class="fas fa-star"></i> Features</h3>
                    <button onclick="this.closest('.modal').remove()" class="close-btn">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <div class="modal-body">
                    <ul style="list-style: none; padding: 0;">
                        <li style="margin-bottom: 15px;"><i class="fas fa-check" style="color: #10b981; margin-right: 10px;"></i> Instant URL shortening</li>
                        <li style="margin-bottom: 15px;"><i class="fas fa-check" style="color: #10b981; margin-right: 10px;"></i> Click tracking and analytics</li>
                        <li style="margin-bottom: 15px;"><i class="fas fa-check" style="color: #10b981; margin-right: 10px;"></i> Custom short codes</li>
                        <li style="margin-bottom: 15px;"><i class="fas fa-check" style="color: #10b981; margin-right: 10px;"></i> Recent URLs history</li>
                        <li style="margin-bottom: 15px;"><i class="fas fa-check" style="color: #10b981; margin-right: 10px;"></i> One-click copying</li>
                        <li style="margin-bottom: 15px;"><i class="fas fa-check" style="color: #10b981; margin-right: 10px;"></i> Mobile-friendly design</li>
                    </ul>
                </div>
            </div>
        </div>
    `;
    document.body.insertAdjacentHTML('beforeend', features);
    document.body.style.overflow = 'hidden';
}

// API modal
function showAPI() {
    const apiInfo = `
        <div class="modal" id="apiModal">
            <div class="modal-content">
                <div class="modal-header">
                    <h3><i class="fas fa-code"></i> API Documentation</h3>
                    <button onclick="this.closest('.modal').remove(); document.body.style.overflow = 'auto';" class="close-btn">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <div class="modal-body">
                    <h4>Shorten URL</h4>
                    <code style="background: #f3f4f6; padding: 10px; border-radius: 5px; display: block; margin: 10px 0;">
                        POST /api/shorten<br>
                        Content-Type: application/json<br><br>
                        {"url": "https://example.com"}
                    </code>
                    
                    <h4 style="margin-top: 20px;">Get Statistics</h4>
                    <code style="background: #f3f4f6; padding: 10px; border-radius: 5px; display: block; margin: 10px 0;">
                        GET /api/stats/{shortCode}
                    </code>
                    
                    <h4 style="margin-top: 20px;">Redirect</h4>
                    <code style="background: #f3f4f6; padding: 10px; border-radius: 5px; display: block; margin: 10px 0;">
                        GET /{shortCode}
                    </code>
                </div>
            </div>
        </div>
    `;
    document.body.insertAdjacentHTML('beforeend', apiInfo);
    document.body.style.overflow = 'hidden';
}

// About modal
function showAbout() {
    const aboutInfo = `
        <div class="modal" id="aboutModal">
            <div class="modal-content">
                <div class="modal-header">
                    <h3><i class="fas fa-info-circle"></i> About QuickLink</h3>
                    <button onclick="this.closest('.modal').remove(); document.body.style.overflow = 'auto';" class="close-btn">
                        <i class="fas fa-times"></i>
                    </button>
                </div>
                <div class="modal-body">
                    <p style="margin-bottom: 15px;">QuickLink is a modern, fast, and reliable URL shortener built with Go and modern web technologies.</p>
                    <p style="margin-bottom: 15px;">Features include instant URL shortening, click tracking, and a beautiful responsive interface.</p>
                    <p style="margin-bottom: 15px;">Built with:</p>
                    <ul style="margin-left: 20px;">
                        <li>Go (Golang) backend</li>
                        <li>Gorilla Mux router</li>
                        <li>HTML5, CSS3, JavaScript</li>
                        <li>Font Awesome icons</li>
                        <li>Google Fonts</li>
                    </ul>
                </div>
            </div>
        </div>
    `;
    document.body.insertAdjacentHTML('beforeend', aboutInfo);
    document.body.style.overflow = 'hidden';
}

// Error handling for network issues
window.addEventListener('online', function() {
    showToast('Connection restored!');
});

window.addEventListener('offline', function() {
    showToast('Connection lost. Please check your internet.', 'error');
});

// Keyboard shortcuts info
document.addEventListener('keydown', function(e) {
    if (e.ctrlKey && e.shiftKey && e.key === '?') {
        const shortcuts = `
            <div class="modal" id="shortcutsModal">
                <div class="modal-content">
                    <div class="modal-header">
                        <h3><i class="fas fa-keyboard"></i> Keyboard Shortcuts</h3>
                        <button onclick="this.closest('.modal').remove(); document.body.style.overflow = 'auto';" class="close-btn">
                            <i class="fas fa-times"></i>
                        </button>
                    </div>
                    <div class="modal-body">
                        <div style="display: grid; gap: 10px;">
                            <div><kbd>Ctrl + Enter</kbd> - Copy short URL</div>
                            <div><kbd>Escape</kbd> - Close modal</div>
                            <div><kbd>Ctrl + Shift + ?</kbd> - Show this help</div>
                        </div>
                    </div>
                </div>
            </div>
        `;
        document.body.insertAdjacentHTML('beforeend', shortcuts);
        document.body.style.overflow = 'hidden';
    }
});

// Initialize service worker for offline functionality (if needed)
if ('serviceWorker' in navigator) {
    window.addEventListener('load', function() {
        navigator.serviceWorker.register('/sw.js').catch(function() {
            // Service worker registration failed, but that's okay
        });
    });
}
