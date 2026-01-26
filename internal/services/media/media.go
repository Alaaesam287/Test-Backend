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

func (s *Service) UploadImage(
	ctx context.Context,
	key string,
	r io.Reader,
	maxSize int64,
) (string, error) {

	validated, mime, err := ValidateImage(r, maxSize)
	if err != nil {
		return "", err
	}

	url, err := s.storage.Upload(ctx, key, validated, -1, mime)
	if err != nil {
		return "", err
	}

	return url, nil
}

// Image Validation

var allowedMIMEs = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

func ValidateImage(
	r io.Reader,
	maxSize int64,
) (io.Reader, string, error) {

	// Enforce size at I/O level
	limited := io.LimitReader(r, maxSize+1)

	// Read header for MIME sniffing
	header := make([]byte, 512)
	n, err := limited.Read(header)
	if err != nil && err != io.EOF {
		return nil, "", fmt.Errorf("reading image header: %w", err)
	}

	if int64(n) > maxSize {
		return nil, "", fmt.Errorf("file too large")
	}

	mime := http.DetectContentType(header[:n])
	if !allowedMIMEs[mime] {
		return nil, "", fmt.Errorf("unsupported image type: %s", mime)
	}

	// Rebuild reader so downstream sees full content
	validatedReader := io.MultiReader(
		bytes.NewReader(header[:n]),
		limited,
	)

	return validatedReader, mime, nil
}