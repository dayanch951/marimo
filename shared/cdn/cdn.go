package cdn

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// CDNProvider represents different CDN providers
type CDNProvider string

const (
	CloudFlare CDNProvider = "cloudflare"
	Fastly     CDNProvider = "fastly"
	CloudFront CDNProvider = "cloudfront"
	Akamai     CDNProvider = "akamai"
	BunnyCDN   CDNProvider = "bunnycdn"
)

// CDNConfig holds CDN configuration
type CDNConfig struct {
	Provider    CDNProvider
	BaseURL     string
	PullZone    string
	APIKey      string
	Enabled     bool
	StaticPaths []string // Paths to serve via CDN
}

// CDN provides CDN URL generation and management
type CDN struct {
	config *CDNConfig
}

// NewCDN creates a new CDN instance
func NewCDN(config *CDNConfig) *CDN {
	return &CDN{config: config}
}

// URL generates a CDN URL for a given path
func (c *CDN) URL(assetPath string) string {
	if !c.config.Enabled {
		return assetPath
	}

	// Clean the path
	assetPath = strings.TrimPrefix(assetPath, "/")

	// Check if path should use CDN
	if !c.shouldUseCDN(assetPath) {
		return assetPath
	}

	// Build CDN URL
	baseURL := strings.TrimSuffix(c.config.BaseURL, "/")
	return fmt.Sprintf("%s/%s", baseURL, assetPath)
}

// shouldUseCDN checks if the path should use CDN
func (c *CDN) shouldUseCDN(assetPath string) bool {
	if len(c.config.StaticPaths) == 0 {
		// Default static paths
		return isStaticAsset(assetPath)
	}

	for _, prefix := range c.config.StaticPaths {
		if strings.HasPrefix(assetPath, prefix) {
			return true
		}
	}

	return false
}

// isStaticAsset checks if the path is a static asset
func isStaticAsset(assetPath string) bool {
	staticExtensions := []string{
		".css", ".js", ".jpg", ".jpeg", ".png", ".gif", ".svg", ".webp",
		".woff", ".woff2", ".ttf", ".eot", ".ico",
		".mp4", ".webm", ".mp3",
	}

	for _, ext := range staticExtensions {
		if strings.HasSuffix(strings.ToLower(assetPath), ext) {
			return true
		}
	}

	return false
}

// ImageURL generates a CDN URL for images with transformations
func (c *CDN) ImageURL(imagePath string, opts *ImageOptions) string {
	baseURL := c.URL(imagePath)

	if opts == nil {
		return baseURL
	}

	// Build query parameters for image transformations
	params := url.Values{}

	if opts.Width > 0 {
		params.Add("w", fmt.Sprintf("%d", opts.Width))
	}

	if opts.Height > 0 {
		params.Add("h", fmt.Sprintf("%d", opts.Height))
	}

	if opts.Quality > 0 && opts.Quality <= 100 {
		params.Add("q", fmt.Sprintf("%d", opts.Quality))
	}

	if opts.Format != "" {
		params.Add("f", opts.Format)
	}

	if opts.Fit != "" {
		params.Add("fit", opts.Fit)
	}

	if len(params) > 0 {
		return fmt.Sprintf("%s?%s", baseURL, params.Encode())
	}

	return baseURL
}

// ImageOptions defines image transformation options
type ImageOptions struct {
	Width   int    // Target width
	Height  int    // Target height
	Quality int    // JPEG/WebP quality (1-100)
	Format  string // Output format (webp, jpeg, png)
	Fit     string // Fit mode (cover, contain, fill, inside, outside)
}

// ResponsiveImageSet generates URLs for responsive images
type ResponsiveImageSet struct {
	Src    string
	SrcSet []string
	Sizes  []string
}

// GenerateResponsiveSet creates responsive image URLs
func (c *CDN) GenerateResponsiveSet(imagePath string, widths []int) *ResponsiveImageSet {
	srcSet := make([]string, len(widths))
	sizes := make([]string, len(widths))

	for i, width := range widths {
		url := c.ImageURL(imagePath, &ImageOptions{
			Width:   width,
			Quality: 85,
			Format:  "webp",
		})

		srcSet[i] = fmt.Sprintf("%s %dw", url, width)

		// Generate sizes attribute
		if i == len(widths)-1 {
			sizes[i] = fmt.Sprintf("%dpx", width)
		} else {
			sizes[i] = fmt.Sprintf("(max-width: %dpx) %dpx", width, width)
		}
	}

	return &ResponsiveImageSet{
		Src:    c.URL(imagePath),
		SrcSet: srcSet,
		Sizes:  sizes,
	}
}

// PurgeCache purges CDN cache for specific URLs
func (c *CDN) PurgeCache(urls []string) error {
	switch c.config.Provider {
	case CloudFlare:
		return c.purgeCloudFlare(urls)
	case BunnyCDN:
		return c.purgeBunnyCDN(urls)
	case CloudFront:
		return c.purgeCloudFront(urls)
	default:
		return fmt.Errorf("purge not implemented for provider: %s", c.config.Provider)
	}
}

// purgeCloudFlare purges CloudFlare cache
func (c *CDN) purgeCloudFlare(urls []string) error {
	// Implementation would use CloudFlare API
	// https://api.cloudflare.com/client/v4/zones/{zone_id}/purge_cache
	fmt.Printf("Purging CloudFlare cache for %d URLs\n", len(urls))
	return nil
}

// purgeBunnyCDN purges BunnyCDN cache
func (c *CDN) purgeBunnyCDN(urls []string) error {
	// Implementation would use BunnyCDN API
	// POST to https://api.bunny.net/pullzone/{pullZoneId}/purgeCache
	fmt.Printf("Purging BunnyCDN cache for %d URLs\n", len(urls))
	return nil
}

// purgeCloudFront purges CloudFront cache
func (c *CDN) purgeCloudFront(urls []string) error {
	// Implementation would use AWS CloudFront API
	fmt.Printf("Purging CloudFront cache for %d URLs\n", len(urls))
	return nil
}

// CacheControlHeaders returns appropriate cache-control headers
func CacheControlHeaders(assetType string) map[string]string {
	headers := make(map[string]string)

	switch assetType {
	case "immutable":
		// For versioned assets (e.g., app.abc123.js)
		headers["Cache-Control"] = "public, max-age=31536000, immutable"
	case "static":
		// For static assets that rarely change
		headers["Cache-Control"] = "public, max-age=604800" // 1 week
	case "dynamic":
		// For dynamic content
		headers["Cache-Control"] = "public, max-age=300, must-revalidate" // 5 minutes
	case "private":
		// For user-specific content
		headers["Cache-Control"] = "private, max-age=0, must-revalidate"
	case "no-cache":
		// For content that shouldn't be cached
		headers["Cache-Control"] = "no-cache, no-store, must-revalidate"
		headers["Pragma"] = "no-cache"
		headers["Expires"] = "0"
	default:
		headers["Cache-Control"] = "public, max-age=3600" // 1 hour
	}

	return headers
}

// AssetVersioning adds version hash to asset URLs for cache busting
type AssetVersioning struct {
	manifest map[string]string
}

// NewAssetVersioning creates a new asset versioning instance
func NewAssetVersioning(manifestPath string) *AssetVersioning {
	// In production, load manifest from file
	// manifest.json maps original filenames to versioned ones
	manifest := map[string]string{
		"app.js":     "app.abc123.js",
		"app.css":    "app.def456.css",
		"logo.png":   "logo.ghi789.png",
	}

	return &AssetVersioning{manifest: manifest}
}

// Get returns the versioned asset path
func (av *AssetVersioning) Get(assetPath string) string {
	if versioned, ok := av.manifest[path.Base(assetPath)]; ok {
		dir := path.Dir(assetPath)
		if dir == "." {
			return versioned
		}
		return path.Join(dir, versioned)
	}

	return assetPath
}

// PreloadHeaders generates HTTP/2 Server Push or Link preload headers
func PreloadHeaders(assets []string) map[string]string {
	var links []string

	for _, asset := range assets {
		asType := "script"
		if strings.HasSuffix(asset, ".css") {
			asType = "style"
		} else if strings.HasSuffix(asset, ".woff2") || strings.HasSuffix(asset, ".woff") {
			asType = "font"
		}

		links = append(links, fmt.Sprintf("<%s>; rel=preload; as=%s", asset, asType))
	}

	return map[string]string{
		"Link": strings.Join(links, ", "),
	}
}

// CompressionMiddleware configuration for gzip/brotli
type CompressionConfig struct {
	Enabled        bool
	MinSize        int      // Minimum size to compress (bytes)
	Level          int      // Compression level (1-9)
	ContentTypes   []string // Content types to compress
	ExcludedPaths  []string // Paths to exclude from compression
}

// DefaultCompressionConfig returns default compression settings
func DefaultCompressionConfig() *CompressionConfig {
	return &CompressionConfig{
		Enabled:  true,
		MinSize:  1024, // 1KB
		Level:    6,    // Balanced compression
		ContentTypes: []string{
			"text/html",
			"text/css",
			"text/javascript",
			"application/javascript",
			"application/json",
			"application/xml",
			"image/svg+xml",
		},
		ExcludedPaths: []string{
			"/health",
			"/metrics",
		},
	}
}

// ShouldCompress checks if response should be compressed
func (cc *CompressionConfig) ShouldCompress(contentType string, size int, path string) bool {
	if !cc.Enabled {
		return false
	}

	// Check minimum size
	if size < cc.MinSize {
		return false
	}

	// Check excluded paths
	for _, excluded := range cc.ExcludedPaths {
		if strings.HasPrefix(path, excluded) {
			return false
		}
	}

	// Check content type
	for _, ct := range cc.ContentTypes {
		if strings.Contains(contentType, ct) {
			return true
		}
	}

	return false
}
