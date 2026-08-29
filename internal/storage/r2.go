package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nfnt/resize"
)

type R2Client struct {
	client     *s3.Client
	bucketName string
	publicURL  string // Custom public URL for serving images (e.g., R2.dev domain with custom subdomain)
}

// ImageDimensions defines target sizes for image processing
type ImageDimensions struct {
	MaxWidth  uint
	MaxHeight uint
	Quality   uint
}

// Standard dimensions for gallery images
var (
	// MainImage: 4K-friendly, longest edge 2560px
	MainImageDimensions = ImageDimensions{MaxWidth: 2560, MaxHeight: 2560, Quality: 85}
	// ThumbnailImage: square 400x400 for nice grid layout
	ThumbnailImageDimensions = ImageDimensions{MaxWidth: 400, MaxHeight: 400, Quality: 80}
)

func NewR2Client(accountID, accessKeyID, secretAccessKey, bucketName, publicURL string) (*R2Client, error) {
	customResolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               fmt.Sprintf("https://%s.eu.r2.cloudflarestorage.com", accountID),
			HostnameImmutable: true,
			Source:            aws.EndpointSourceCustom,
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("auto"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithEndpointResolverWithOptions(customResolver),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	return &R2Client{
		client:     client,
		bucketName: bucketName,
		publicURL:  publicURL,
	}, nil
}

func (r *R2Client) EnsureBucketStructure(ctx context.Context) error {
	folders := []string{"images", "thumbnails"}
	for _, folder := range folders {
		// Create folder marker (empty object with key ending in /)
		_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket: aws.String(r.bucketName),
			Key:    aws.String(folder + "/"),
		})
		if err != nil {
			// Ignore if already exists
		}
	}
	return nil
}

func (r *R2Client) UploadFile(ctx context.Context, key string, body io.Reader, contentType string) (string, error) {
	_, err := r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucketName),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	// Return public URL (publicURL already includes bucket path handling)
	url := fmt.Sprintf("%s/%s", r.publicURL, key)
	return url, nil
}

func (r *R2Client) UploadFileFromPath(ctx context.Context, key, filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	return r.UploadFile(ctx, key, file, "application/octet-stream")
}

func (r *R2Client) GetPublicURL(key string) string {
	return fmt.Sprintf("%s/%s", r.publicURL, key)
}

func (r *R2Client) GetImageURL(filename string) string {
	return r.GetPublicURL("images/" + filename)
}

func (r *R2Client) GetThumbnailURL(filename string) string {
	return r.GetPublicURL("thumbnails/" + filename)
}

// ProcessAndUploadImage reads an image from a reader, resizes it to the specified dimensions,
// and uploads to R2. Returns the R2 URL of the uploaded image.
func (r *R2Client) ProcessAndUploadImage(ctx context.Context, key string, reader io.Reader, dims ImageDimensions, contentType string) (string, error) {
	// Decode image
	img, format, err := image.Decode(reader)
	if err != nil {
		return "", fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize image - use 0 for height to maintain aspect ratio based on width
	// or 0 for width to maintain aspect ratio based on height
	resized := resize.Resize(0, dims.MaxHeight, img, resize.NearestNeighbor)

	// Encode to buffer
	var buf bytes.Buffer
	switch format {
	case "png":
		err = png.Encode(&buf, resized)
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: int(dims.Quality)})
	default:
		// Default to JPEG for photos
		err = jpeg.Encode(&buf, resized, &jpeg.Options{Quality: int(dims.Quality)})
	}
	if err != nil {
		return "", fmt.Errorf("failed to encode resized image: %w", err)
	}

	// Determine content type
	uploadContentType := contentType
	if uploadContentType == "" || uploadContentType == "application/octet-stream" {
		if format == "png" {
			uploadContentType = "image/png"
		} else {
			uploadContentType = "image/jpeg"
		}
	}

	// Upload to R2
	_, err = r.UploadFile(ctx, key, &buf, uploadContentType)
	if err != nil {
		return "", err
	}
	// Return just the key (e.g., "images/abc123.jpg"), not the full URL
	return key, nil
}

// UploadImageAndThumbnail takes an image file, creates both a main image and thumbnail,
// and uploads both to R2. Returns the keys for both (not full URLs).
func (r *R2Client) UploadImageAndThumbnail(ctx context.Context, fileReader io.Reader, filename string, contentType string) (imageKey string, thumbnailKey string, err error) {
	// Read all bytes from the reader
	data, err := io.ReadAll(fileReader)
	if err != nil {
		return "", "", fmt.Errorf("failed to read image data: %w", err)
	}

	// Determine file extension for keys
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}

	// Generate unique base name (without extension - we'll add it per image)
	baseName := fmt.Sprintf("%d", time.Now().UnixNano())

	imageKey = "images/" + baseName + ext
	thumbnailKey = "thumbnails/" + baseName + "_thumb" + ext

	// Process and upload main image
	_, err = r.ProcessAndUploadImage(ctx, imageKey, bytes.NewReader(data), MainImageDimensions, contentType)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload main image: %w", err)
	}

	// Process and upload thumbnail
	_, err = r.ProcessAndUploadImage(ctx, thumbnailKey, bytes.NewReader(data), ThumbnailImageDimensions, contentType)
	if err != nil {
		return "", "", fmt.Errorf("failed to upload thumbnail: %w", err)
	}

	// Return only the keys (paths), not full URLs
	return imageKey, thumbnailKey, nil
}
