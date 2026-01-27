package media

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/Secure-Website-Builder/Backend/internal/storage"
)

type Service struct {
	storage storage.ObjectStorage
}

func New(storage storage.ObjectStorage) *Service {
	return &Service{
		storage: storage,
	}
}

const MaxImageSize = 5 * 1024 * 1024

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// UploadImage validates the image, appends the correct extension,
// uploads it, and returns the final URL and MIME type.
func (s *Service) UploadImage(
	ctx context.Context,
	key string,
	r io.Reader,
) (string, string, error) {

	// Validate the image first
	validated, mime, err := ValidateImage(r)
	if err != nil {
		return "", "", fmt.Errorf("invalid image: %w", err)
	}

	ext := allowedImageTypes[mime]
	key = key + ext

	// Upload to storage
	url, err := s.storage.Upload(ctx, key, validated, -1, mime)
	if err != nil {
		return "", "", err
	}

	return url, mime, nil
}

// ValidateImage consumes r and returns a new reader that:
//   - enforces MaxImageSize
//   - guarantees a valid image MIME
//   - must be used downstream (single-pass)
func ValidateImage(r io.Reader) (io.Reader, string, error) {

	header := make([]byte, 512)
	n, err := io.ReadFull(r,header)
	if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
		return nil, "", fmt.Errorf("reading image header: %w", err)
	}

	mime := http.DetectContentType(header[:n])
	if _, ok := allowedImageTypes[mime]; !ok {
		return nil, "", fmt.Errorf("unsupported image type: %s", mime)
	}

	validatedReader := io.MultiReader(bytes.NewReader(header[:n]), r)

	// Wrap with size-checking reader
	// Enforce size with hard error (prevents silent truncation)
	sizeChecker := &sizeLimitedReader{
		R: validatedReader,
		N: MaxImageSize,
	}

	return sizeChecker, mime, nil
}

type sizeLimitedReader struct {
	R io.Reader
	N int64 
}

func (s *sizeLimitedReader) Read(p []byte) (int, error) {
	if s.N <= 0 {
		return 0, fmt.Errorf("file too large")
	}
	if int64(len(p)) > s.N {
		p = p[:s.N]
	}
	n, err := s.R.Read(p)
	s.N -= int64(n)
	if s.N <= 0 && err == nil {
		// If we reached the limit, signal error on next read
		err = fmt.Errorf("file too large")
	}
	return n, err
}