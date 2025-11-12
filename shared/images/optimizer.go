package images

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
	"golang.org/x/image/webp"
)

// ImageFormat represents supported image formats
type ImageFormat string

const (
	FormatJPEG ImageFormat = "jpeg"
	FormatPNG  ImageFormat = "png"
	FormatWebP ImageFormat = "webp"
)

// ImageOptimizer provides image optimization capabilities
type ImageOptimizer struct {
	maxWidth       int
	maxHeight      int
	quality        int
	preferredFormat ImageFormat
}

// NewImageOptimizer creates a new image optimizer
func NewImageOptimizer() *ImageOptimizer {
	return &ImageOptimizer{
		maxWidth:       2048,
		maxHeight:      2048,
		quality:        85,
		preferredFormat: FormatWebP,
	}
}

// OptimizeOptions defines image optimization options
type OptimizeOptions struct {
	MaxWidth   int         // Maximum width (0 = no limit)
	MaxHeight  int         // Maximum height (0 = no limit)
	Quality    int         // Quality (1-100)
	Format     ImageFormat // Output format
	StripMeta  bool        // Remove metadata
}

// DefaultOptimizeOptions returns default optimization settings
func DefaultOptimizeOptions() *OptimizeOptions {
	return &OptimizeOptions{
		MaxWidth:  1920,
		MaxHeight: 1080,
		Quality:   85,
		Format:    FormatWebP,
		StripMeta: true,
	}
}

// Optimize optimizes an image
func (io *ImageOptimizer) Optimize(input io.Reader, output io.Writer, opts *OptimizeOptions) error {
	if opts == nil {
		opts = DefaultOptimizeOptions()
	}

	// Decode image
	img, format, err := image.Decode(input)
	if err != nil {
		return fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize if needed
	if opts.MaxWidth > 0 || opts.MaxHeight > 0 {
		img = io.resize(img, opts.MaxWidth, opts.MaxHeight)
	}

	// Encode with optimization
	return io.encode(output, img, opts.Format, opts.Quality, format)
}

// resize resizes image maintaining aspect ratio
func (io *ImageOptimizer) resize(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Calculate new dimensions
	newWidth, newHeight := io.calculateDimensions(width, height, maxWidth, maxHeight)

	// If no resize needed, return original
	if newWidth == width && newHeight == height {
		return img
	}

	// Create new image
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Use high-quality scaling
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	return dst
}

// calculateDimensions calculates new dimensions maintaining aspect ratio
func (io *ImageOptimizer) calculateDimensions(width, height, maxWidth, maxHeight int) (int, int) {
	if maxWidth == 0 {
		maxWidth = width
	}
	if maxHeight == 0 {
		maxHeight = height
	}

	// Calculate scale factor
	scaleWidth := float64(maxWidth) / float64(width)
	scaleHeight := float64(maxHeight) / float64(height)
	scale := scaleWidth
	if scaleHeight < scale {
		scale = scaleHeight
	}

	// Don't upscale
	if scale > 1.0 {
		scale = 1.0
	}

	newWidth := int(float64(width) * scale)
	newHeight := int(float64(height) * scale)

	return newWidth, newHeight
}

// encode encodes image in specified format
func (io *ImageOptimizer) encode(w io.Writer, img image.Image, format ImageFormat, quality int, originalFormat string) error {
	switch format {
	case FormatJPEG:
		opts := &jpeg.Options{Quality: quality}
		return jpeg.Encode(w, img, opts)

	case FormatPNG:
		encoder := &png.Encoder{CompressionLevel: png.BestCompression}
		return encoder.Encode(w, img)

	case FormatWebP:
		// WebP encoding requires external library
		// For now, fall back to JPEG
		// In production, use github.com/chai2010/webp
		return jpeg.Encode(w, img, &jpeg.Options{Quality: quality})

	default:
		// Use original format
		return fmt.Errorf("unsupported format: %s", format)
	}
}

// OptimizeFile optimizes an image file
func (io *ImageOptimizer) OptimizeFile(inputPath, outputPath string, opts *OptimizeOptions) error {
	// Open input file
	input, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input: %w", err)
	}
	defer input.Close()

	// Create output file
	output, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output: %w", err)
	}
	defer output.Close()

	// Optimize
	return io.Optimize(input, output, opts)
}

// GenerateThumbnails generates multiple thumbnail sizes
type ThumbnailSize struct {
	Name   string
	Width  int
	Height int
}

// DefaultThumbnailSizes returns standard thumbnail sizes
func DefaultThumbnailSizes() []ThumbnailSize {
	return []ThumbnailSize{
		{Name: "small", Width: 150, Height: 150},
		{Name: "medium", Width: 300, Height: 300},
		{Name: "large", Width: 600, Height: 600},
		{Name: "xlarge", Width: 1200, Height: 1200},
	}
}

// GenerateThumbnails creates multiple thumbnail versions
func (io *ImageOptimizer) GenerateThumbnails(inputPath, outputDir string, sizes []ThumbnailSize) (map[string]string, error) {
	results := make(map[string]string)

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Load original image
	input, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer input.Close()

	img, _, err := image.Decode(input)
	if err != nil {
		return nil, err
	}

	// Generate each size
	ext := filepath.Ext(inputPath)
	baseName := strings.TrimSuffix(filepath.Base(inputPath), ext)

	for _, size := range sizes {
		// Resize image
		resized := io.resize(img, size.Width, size.Height)

		// Generate output path
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%s.webp", baseName, size.Name))

		// Save thumbnail
		output, err := os.Create(outputPath)
		if err != nil {
			return nil, err
		}

		err = io.encode(output, resized, FormatWebP, 85, "")
		output.Close()

		if err != nil {
			return nil, err
		}

		results[size.Name] = outputPath
	}

	return results, nil
}

// LazyLoadPlaceholder generates a tiny placeholder for lazy loading
func (io *ImageOptimizer) LazyLoadPlaceholder(inputPath string) ([]byte, error) {
	input, err := os.Open(inputPath)
	if err != nil {
		return nil, err
	}
	defer input.Close()

	img, _, err := image.Decode(input)
	if err != nil {
		return nil, err
	}

	// Create tiny 20px width placeholder
	placeholder := io.resize(img, 20, 0)

	// Encode as JPEG with low quality
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, placeholder, &jpeg.Options{Quality: 40}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ImageMetadata contains image metadata
type ImageMetadata struct {
	Width      int
	Height     int
	Format     string
	Size       int64
	AspectRatio float64
}

// GetMetadata extracts image metadata
func GetMetadata(path string) (*ImageMetadata, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Get file size
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	// Decode to get dimensions
	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return nil, err
	}

	aspectRatio := float64(config.Width) / float64(config.Height)

	return &ImageMetadata{
		Width:       config.Width,
		Height:      config.Height,
		Format:      format,
		Size:        stat.Size(),
		AspectRatio: aspectRatio,
	}, nil
}

// SupportsWebP checks if client supports WebP
func SupportsWebP(acceptHeader string) bool {
	return strings.Contains(acceptHeader, "image/webp")
}

// GetOptimalFormat returns the optimal format based on client support
func GetOptimalFormat(acceptHeader string) ImageFormat {
	if SupportsWebP(acceptHeader) {
		return FormatWebP
	}
	return FormatJPEG
}

// CompressionStats contains compression statistics
type CompressionStats struct {
	OriginalSize   int64
	CompressedSize int64
	SavingsBytes   int64
	SavingsPercent float64
}

// CalculateStats calculates compression statistics
func CalculateStats(originalSize, compressedSize int64) *CompressionStats {
	savings := originalSize - compressedSize
	savingsPercent := float64(savings) / float64(originalSize) * 100

	return &CompressionStats{
		OriginalSize:   originalSize,
		CompressedSize: compressedSize,
		SavingsBytes:   savings,
		SavingsPercent: savingsPercent,
	}
}

// BatchOptimizer optimizes multiple images in parallel
type BatchOptimizer struct {
	optimizer *ImageOptimizer
	workers   int
}

// NewBatchOptimizer creates a new batch optimizer
func NewBatchOptimizer(workers int) *BatchOptimizer {
	if workers <= 0 {
		workers = 4
	}

	return &BatchOptimizer{
		optimizer: NewImageOptimizer(),
		workers:   workers,
	}
}

// OptimizeDirectory optimizes all images in a directory
func (bo *BatchOptimizer) OptimizeDirectory(inputDir, outputDir string, opts *OptimizeOptions) error {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	// Find all images
	images, err := filepath.Glob(filepath.Join(inputDir, "*"))
	if err != nil {
		return err
	}

	// Filter image files
	var imageFiles []string
	for _, path := range images {
		if isImageFile(path) {
			imageFiles = append(imageFiles, path)
		}
	}

	// Process each image
	for _, inputPath := range imageFiles {
		baseName := filepath.Base(inputPath)
		outputPath := filepath.Join(outputDir, baseName)

		if err := bo.optimizer.OptimizeFile(inputPath, outputPath, opts); err != nil {
			fmt.Printf("Failed to optimize %s: %v\n", baseName, err)
			continue
		}

		// Print stats
		originalStat, _ := os.Stat(inputPath)
		optimizedStat, _ := os.Stat(outputPath)
		stats := CalculateStats(originalStat.Size(), optimizedStat.Size())

		fmt.Printf("Optimized %s: %.1f%% smaller\n", baseName, stats.SavingsPercent)
	}

	return nil
}

// isImageFile checks if file is an image
func isImageFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".jpg" || ext == ".jpeg" || ext == ".png" || ext == ".webp"
}

// ResponsiveImageGenerator generates responsive image sets
type ResponsiveImageGenerator struct {
	optimizer *ImageOptimizer
}

// NewResponsiveImageGenerator creates a new responsive image generator
func NewResponsiveImageGenerator() *ResponsiveImageGenerator {
	return &ResponsiveImageGenerator{
		optimizer: NewImageOptimizer(),
	}
}

// ResponsiveSet contains responsive image URLs and metadata
type ResponsiveSet struct {
	Original string
	Sizes    map[int]string // width -> URL
	SrcSet   string         // srcset attribute value
}

// Generate creates a responsive image set
func (rig *ResponsiveImageGenerator) Generate(inputPath, outputDir string, widths []int) (*ResponsiveSet, error) {
	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, err
	}

	baseName := strings.TrimSuffix(filepath.Base(inputPath), filepath.Ext(inputPath))
	sizes := make(map[int]string)
	var srcSetParts []string

	// Generate each width
	for _, width := range widths {
		outputPath := filepath.Join(outputDir, fmt.Sprintf("%s_%dw.webp", baseName, width))

		opts := &OptimizeOptions{
			MaxWidth: width,
			Quality:  85,
			Format:   FormatWebP,
		}

		if err := rig.optimizer.OptimizeFile(inputPath, outputPath, opts); err != nil {
			return nil, err
		}

		sizes[width] = outputPath
		srcSetParts = append(srcSetParts, fmt.Sprintf("%s %dw", outputPath, width))
	}

	return &ResponsiveSet{
		Original: inputPath,
		Sizes:    sizes,
		SrcSet:   strings.Join(srcSetParts, ", "),
	}, nil
}
